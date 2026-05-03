// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"bytes"
	"marktex/internal/ast"
	"strings"
)

// blockParser builds the block-level AST from a stream of lines.
// It maintains an open-container stack and processes lines one at a time.
type blockParser struct {
	scanner *scanner
	ext     Extensions

	// linkRefs is populated as link-reference definitions are encountered.
	linkRefs map[string]linkRef

	// refs holds citation and figure definition maps, pre-populated by preScan.
	refs defRefs

	// stack holds the currently open container nodes (Document at index 0,
	// possibly followed by BlockQuote / BulletList / OrderedList / ListItem).
	stack []openContainer
}

// openContainer tracks an open block container in the parse stack.
type openContainer struct {
	node          ast.Node
	contentIndent int // for list items: minimum indent of continuation lines
}

// linkRef holds a resolved link reference definition.
type linkRef struct {
	dest  string
	title string
}

// fenceState tracks an open fenced code block.
type fenceState struct {
	char     byte // '`' or '~'
	width    int  // number of opening fence characters
	indent   int  // indentation of the opening line
	info     string
	lines    []string
	startPos ast.Pos
}

func newBlockParser(src []byte, ext Extensions, refs defRefs) *blockParser {
	if refs.citations == nil {
		refs.citations = make(map[string]string)
	}
	if refs.figures == nil {
		refs.figures = make(map[string]string)
	}
	if refs.tables == nil {
		refs.tables = make(map[string]string)
	}
	return &blockParser{
		scanner:  newScanner(src),
		ext:      ext,
		linkRefs: make(map[string]linkRef),
		refs:     refs,
		stack:    []openContainer{{node: ast.NewDocument(ast.Pos{Line: 1, Col: 1})}},
	}
}

func (p *blockParser) document() *ast.Document {
	return p.stack[0].node.(*ast.Document)
}

func (p *blockParser) parse() *ast.Document {
	var (
		para      *ast.Paragraph // currently open paragraph
		paraLines []string       // collected raw lines for the paragraph
		paraPos   ast.Pos

		fence *fenceState // currently open fenced code block

		indentedLines []string // accumulating indented code lines
		indentedPos   ast.Pos

		bqLines  []string // accumulated blockquote inner lines
		bqOpen   bool
		bqPos    ast.Pos

		blankCount int // consecutive blank lines
	)

	flushParagraph := func() {
		if para == nil {
			return
		}
		// Join paragraph lines, parse inline content, attach to para
		raw := strings.Join(paraLines, "\n")
		inlines := parseInline(raw, paraPos, p.ext, p.refs.citations, p.refs.figures, p.refs.tables)
		for _, n := range inlines {
			ast.AppendChild(para, n)
		}
		p.appendToTop(para)
		para = nil
		paraLines = nil
	}

	flushIndented := func() {
		if len(indentedLines) == 0 {
			return
		}
		// Strip trailing blank lines
		for len(indentedLines) > 0 && strings.TrimSpace(indentedLines[len(indentedLines)-1]) == "" {
			indentedLines = indentedLines[:len(indentedLines)-1]
		}
		if len(indentedLines) > 0 {
			literal := strings.Join(indentedLines, "\n") + "\n"
			p.appendToTop(ast.NewIndentedCode(indentedPos, literal))
		}
		indentedLines = nil
	}

	flushBlockQuote := func() {
		if !bqOpen {
			return
		}
		bqOpen = false
		// Re-parse the inner content as blocks
		inner := strings.Join(bqLines, "\n")
		innerDoc := parseWithRefs([]byte(inner), p.ext, p.refs)
		bq := ast.NewBlockQuote(bqPos)
		for _, child := range innerDoc.Children() {
			ast.AppendChild(bq, child)
		}
		p.appendToTop(bq)
		bqLines = nil
	}

	for {
		l := p.scanner.next()
		if l == nil {
			break
		}
		nodePos := ast.Pos{Line: l.lineNum, Col: 1, Offset: l.offset}
		content := l.content

		// ── Open fenced code block: accumulate until closing fence ────────────
		if fence != nil {
			if closingFence(content, fence.char, fence.width) {
				literal := strings.Join(fence.lines, "\n")
				if len(fence.lines) > 0 {
					literal += "\n"
				}
				p.appendToTop(ast.NewFencedCode(fence.startPos, fence.info, literal))
				fence = nil
			} else {
				// Strip up to fence.indent leading spaces
				stripped := string(trimLeftSpaces(content, fence.indent))
				fence.lines = append(fence.lines, stripped)
			}
			continue
		}

		// ── Blank line ────────────────────────────────────────────────────────
		if isBlankLine(content) {
			blankCount++
			if bqOpen && blankCount > 0 {
				flushBlockQuote()
			}
			flushParagraph()
			flushIndented()
			continue
		}
		blankCount = 0

		// ── Block quote continuation / opening ────────────────────────────────
		if len(content) > 0 && content[0] == '>' {
			inner, _ := stripBlockQuoteMarker(content)
			if !bqOpen {
				flushParagraph()
				flushIndented()
				bqOpen = true
				bqPos = nodePos
			}
			bqLines = append(bqLines, string(inner))
			continue
		}
		if bqOpen {
			// Lazy continuation: non-blank line continues a blockquote paragraph
			// Only if the last thing in bqLines was a non-blank paragraph line
			if len(bqLines) > 0 && !isBlankLine([]byte(bqLines[len(bqLines)-1])) {
				bqLines = append(bqLines, string(content))
				continue
			}
			flushBlockQuote()
		}

		indent := l.indent()

		// ── Definition blocks (C#/F#/T# entries) — must precede thematic break.
		// A single ---…--- block may contain any mix of entry types.
		if isDefinitionDelimiterLine(content) {
			if next := p.scanner.peek(); next != nil {
				if p.isAnyDefinitionEntry(next.content) {
					if block, ok := p.parseDefinitionBlock(nodePos); ok {
						flushParagraph()
						flushIndented()
						p.appendToTop(block)
						continue
					}
				}
			}
		}

		// ── Thematic break ────────────────────────────────────────────────────
		if isThematicBreak(content) {
			flushParagraph()
			flushIndented()
			p.appendToTop(ast.NewThematicBreak(nodePos))
			continue
		}

		// ── ATX heading ───────────────────────────────────────────────────────
		if level, text := parseATXHeading(content); level > 0 {
			flushParagraph()
			flushIndented()
			h := ast.NewHeading(nodePos, level)
			inlines := parseInline(text, nodePos, p.ext, p.refs.citations, p.refs.figures, p.refs.tables)
			for _, n := range inlines {
				ast.AppendChild(h, n)
			}
			p.appendToTop(h)
			continue
		}

		// ── Setext heading (underline following a paragraph) ──────────────────
		if para != nil {
			if isSetextUnderline(content) {
				level := 1
				if content[0] == '-' {
					level = 2
				}
				raw := strings.Join(paraLines, "\n")
				h := ast.NewHeading(paraPos, level)
				inlines := parseInline(raw, paraPos, p.ext, p.refs.citations, p.refs.figures, p.refs.tables)
				for _, n := range inlines {
					ast.AppendChild(h, n)
				}
				p.appendToTop(h)
				para = nil
				paraLines = nil
				continue
			}
		}

		// ── Fenced code block (opening) ───────────────────────────────────────
		if fs := parseFenceOpening(content); fs != nil {
			flushParagraph()
			flushIndented()
			fs.startPos = nodePos
			fs.indent = indent
			fence = fs
			continue
		}

		// ── Table float with metadata (must precede plain GFM table check) ───
		if p.ext.Has(ExtTableFloat) && isTableMetaRow(content) {
			if block, ok := p.parseTableBlock(l, nodePos); ok {
				flushParagraph()
				flushIndented()
				p.appendToTop(block)
				continue
			}
		}

		// ── GFM Table ─────────────────────────────────────────────────────────
		if p.ext.Has(ExtTables) {
			if isTableSeparator(p.scanner.peek()) {
				flushParagraph()
				flushIndented()
				tableNode := p.parseTable(l, nodePos)
				if tableNode != nil {
					p.appendToTop(tableNode)
					continue
				}
			}
		}

		// ── Ordered list ──────────────────────────────────────────────────────
		if num, delim, width := parseOrderedListMarker(bytes.TrimLeft(content, " ")); width > 0 {
			actualIndent := indent
			{
				flushParagraph()
				flushIndented()
				itemContent := string(content[actualIndent+width:])
				list := ast.NewOrderedList(nodePos, num, delim)
				item := ast.NewListItem(nodePos)
				// Parse item content (may be multi-line; for now single-line)
				inlines := parseInline(itemContent, nodePos, p.ext, p.refs.citations, p.refs.figures, p.refs.tables)
				for _, n := range inlines {
					ast.AppendChild(item, n)
				}
				ast.AppendChild(list, item)
				// Accumulate subsequent list items
				p.parseListItems(list, item, byte(delim), true)
				p.appendToTop(list)
				continue
			}
		}

		// ── Bullet list ───────────────────────────────────────────────────────
		if marker, width := parseBulletListMarker(bytes.TrimLeft(content, " ")); marker != 0 {
			flushParagraph()
			flushIndented()
			actualIndent := indent
			itemContent := string(content[actualIndent+width:])
			list := ast.NewBulletList(nodePos, marker)
			item := ast.NewListItem(nodePos)
			inlines := parseInline(itemContent, nodePos, p.ext, p.refs.citations, p.refs.figures, p.refs.tables)
			for _, n := range inlines {
				ast.AppendChild(item, n)
			}
			ast.AppendChild(list, item)
			p.parseListItems(list, item, byte(marker), false)
			p.appendToTop(list)
			continue
		}

		// ── Indented code block ───────────────────────────────────────────────
		if indent >= 4 && para == nil {
			if len(indentedLines) == 0 {
				indentedPos = nodePos
			}
			// Strip exactly 4 leading spaces
			indentedLines = append(indentedLines, string(content[4:]))
			continue
		}

		// ── Paragraph (or paragraph continuation) ────────────────────────────
		flushIndented()
		if para == nil {
			para = ast.NewParagraph(nodePos)
			paraPos = nodePos
		}
		paraLines = append(paraLines, string(content))
	}

	// Flush any remaining open elements
	flushBlockQuote()
	flushParagraph()
	flushIndented()
	if fence != nil {
		// Unclosed fence: treat as a fenced code block anyway (CommonMark behavior)
		literal := strings.Join(fence.lines, "\n")
		if len(fence.lines) > 0 {
			literal += "\n"
		}
		p.appendToTop(ast.NewFencedCode(fence.startPos, fence.info, literal))
	}

	return p.document()
}

// appendToTop appends node to the innermost open container.
func (p *blockParser) appendToTop(n ast.Node) {
	top := p.stack[len(p.stack)-1].node
	ast.AppendChild(top, n)
}

// isSetextUnderline returns true if content is a setext heading underline
// (`===` or `---` with optional trailing whitespace).
func isSetextUnderline(content []byte) bool {
	s := bytes.TrimRight(content, " \t")
	if len(s) == 0 {
		return false
	}
	ch := s[0]
	if ch != '=' && ch != '-' {
		return false
	}
	for _, b := range s {
		if b != ch {
			return false
		}
	}
	return true
}

// closingFence returns true if content is a valid closing fence.
func closingFence(content []byte, char byte, width int) bool {
	s := bytes.TrimLeft(content, " ")
	if len(s) < width {
		return false
	}
	for i := 0; i < width; i++ {
		if s[i] != char {
			return false
		}
	}
	// Rest must be optional spaces only
	rest := bytes.TrimRight(s[width:], " \t")
	return len(rest) == 0
}

// parseFenceOpening attempts to parse a fenced code block opening line.
// Returns a fenceState on success, nil otherwise.
func parseFenceOpening(content []byte) *fenceState {
	s := bytes.TrimLeft(content, " ")
	if len(s) < 3 {
		return nil
	}
	char := s[0]
	if char != '`' && char != '~' {
		return nil
	}
	width := countLeadingChar(s, char)
	if width < 3 {
		return nil
	}
	info := strings.TrimSpace(string(s[width:]))
	// Backtick info string must not contain a backtick
	if char == '`' && strings.ContainsRune(info, '`') {
		return nil
	}
	return &fenceState{char: char, width: width, info: info}
}

// isTableSeparator returns true if the given line looks like a GFM table
// separator row (`| :--- | :---: | ---: |`).
func isTableSeparator(l *line) bool {
	if l == nil {
		return false
	}
	s := strings.TrimSpace(string(l.content))
	if len(s) == 0 {
		return false
	}
	// Strip surrounding pipes
	if s[0] == '|' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '|' {
		s = s[:len(s)-1]
	}
	cells := strings.Split(s, "|")
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		if !isSeparatorCell(cell) {
			return false
		}
	}
	return true
}

func isSeparatorCell(s string) bool {
	if len(s) == 0 {
		return false
	}
	if s[0] == ':' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == ':' {
		s = s[:len(s)-1]
	}
	if len(s) == 0 {
		return false
	}
	for _, b := range []byte(s) {
		if b != '-' {
			return false
		}
	}
	return true
}

// parseTable consumes the header row (already read as l), the separator row,
// and all subsequent body rows.
func (p *blockParser) parseTable(headerLine *line, pos ast.Pos) ast.Node {
	// Consume separator row
	sepLine := p.scanner.next()
	if sepLine == nil {
		return nil
	}
	aligns := parseSeparatorAlignments(string(sepLine.content))
	ncols := len(aligns)

	table := ast.NewTable(pos, aligns)

	// Header section
	head := ast.NewTableSection(pos, true)
	hrow := p.parseTableRow(string(headerLine.content), aligns, pos)
	if hrow == nil {
		return nil
	}
	ast.AppendChild(head, hrow)
	ast.AppendChild(table, head)

	// Body section
	body := ast.NewTableSection(pos, false)
	for {
		next := p.scanner.peek()
		if next == nil || isBlankLine(next.content) {
			break
		}
		// A line is a body row if it contains at least one '|' or has ncols cells
		s := strings.TrimSpace(string(next.content))
		if !strings.ContainsRune(s, '|') && ncols > 1 {
			break
		}
		l := p.scanner.next()
		row := p.parseTableRow(string(l.content), aligns, ast.Pos{Line: l.lineNum, Col: 1, Offset: l.offset})
		if row != nil {
			ast.AppendChild(body, row)
		}
	}
	if len(body.Children()) > 0 {
		ast.AppendChild(table, body)
	}
	return table
}

func (p *blockParser) parseTableRow(raw string, aligns []ast.Align, pos ast.Pos) *ast.TableRow {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	if len(s) > 0 && s[0] == '|' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '|' {
		s = s[:len(s)-1]
	}
	cells := strings.Split(s, "|")
	row := ast.NewTableRow(pos)
	for i, cellSrc := range cells {
		var align ast.Align
		if i < len(aligns) {
			align = aligns[i]
		}
		cell := ast.NewTableCell(pos, align)
		inlines := parseInline(strings.TrimSpace(cellSrc), pos, p.ext, p.refs.citations, p.refs.figures, p.refs.tables)
		for _, n := range inlines {
			ast.AppendChild(cell, n)
		}
		ast.AppendChild(row, cell)
	}
	return row
}

func parseSeparatorAlignments(sep string) []ast.Align {
	s := strings.TrimSpace(sep)
	if len(s) > 0 && s[0] == '|' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '|' {
		s = s[:len(s)-1]
	}
	cells := strings.Split(s, "|")
	aligns := make([]ast.Align, 0, len(cells))
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		left := len(cell) > 0 && cell[0] == ':'
		right := len(cell) > 0 && cell[len(cell)-1] == ':'
		switch {
		case left && right:
			aligns = append(aligns, ast.AlignCenter)
		case right:
			aligns = append(aligns, ast.AlignRight)
		case left:
			aligns = append(aligns, ast.AlignLeft)
		default:
			aligns = append(aligns, ast.AlignNone)
		}
	}
	return aligns
}

// parseListItems reads subsequent lines that belong to the same list,
// appending new ListItem nodes to list.
func (p *blockParser) parseListItems(list ast.Node, firstItem *ast.ListItem, marker byte, ordered bool) {
	for {
		next := p.scanner.peek()
		if next == nil {
			return
		}
		// Blank line: peek one more ahead
		if isBlankLine(next.content) {
			// Look past blank; if next non-blank line starts another item, continue
			p.scanner.next() // consume blank
			after := p.scanner.peek()
			if after == nil {
				return
			}
			if isNextListItem(after.content, marker, ordered) {
				l := p.scanner.next()
				item := p.parseListItemLine(l, marker, ordered)
				if item != nil {
					// Mark list as loose since there was a blank line
					switch lst := list.(type) {
					case *ast.BulletList:
						lst.Tight = false
					case *ast.OrderedList:
						lst.Tight = false
					}
					ast.AppendChild(list, item)
				}
				continue
			}
			return
		}
		if !isNextListItem(next.content, marker, ordered) {
			return
		}
		l := p.scanner.next()
		item := p.parseListItemLine(l, marker, ordered)
		if item != nil {
			ast.AppendChild(list, item)
		}
	}
}

func isNextListItem(content []byte, marker byte, ordered bool) bool {
	s := bytes.TrimLeft(content, " ")
	if ordered {
		_, _, w := parseOrderedListMarker(s)
		return w > 0
	}
	m, _ := parseBulletListMarker(s)
	return m == rune(marker)
}

func (p *blockParser) parseListItemLine(l *line, marker byte, ordered bool) *ast.ListItem {
	pos := ast.Pos{Line: l.lineNum, Col: 1, Offset: l.offset}
	content := bytes.TrimLeft(l.content, " ")
	var itemContent string
	if ordered {
		_, _, width := parseOrderedListMarker(content)
		if width == 0 {
			return nil
		}
		itemContent = string(content[width:])
	} else {
		_, width := parseBulletListMarker(content)
		if width == 0 {
			return nil
		}
		itemContent = string(content[width:])
	}
	item := ast.NewListItem(pos)
	inlines := parseInline(itemContent, pos, p.ext, p.refs.citations, p.refs.figures, p.refs.tables)
	for _, n := range inlines {
		ast.AppendChild(item, n)
	}
	return item
}

// parseCitationBlock reads lines after the opening `---` delimiter until a
// closing `---` is found. It returns a CitationBlock node and true on success,
// or nil, false if the block is malformed (in which case the caller falls
// through to thematic-break handling).
func (p *blockParser) parseCitationBlock(pos ast.Pos) (*ast.CitationBlock, bool) {
	refs := make(map[string]string)

	for {
		l := p.scanner.next()
		if l == nil {
			// Unclosed citation block — treat opening `---` as a thematic break
			return nil, false
		}

		if isCitationDelimiterLine(l.content) {
			// Closing delimiter: commit the block
			break
		}

		key, bib, ok := parseCitationEntryBytes(l.content)
		if !ok {
			// Non-entry line inside the block — not a valid citation block
			return nil, false
		}
		refs[key] = bib
	}

	if len(refs) == 0 {
		return nil, false
	}
	return ast.NewCitationBlock(pos, refs), true
}

// parseFigureBlock reads F#key:label lines until the closing `---`.
func (p *blockParser) parseFigureBlock(pos ast.Pos) (*ast.FigureBlock, bool) {
	refs := make(map[string]string)

	for {
		l := p.scanner.next()
		if l == nil {
			return nil, false
		}
		if isDefinitionDelimiterLine(l.content) {
			break
		}
		key, label, ok := parseFigureEntryBytes(l.content)
		if !ok {
			return nil, false
		}
		refs[key] = label
	}

	if len(refs) == 0 {
		return nil, false
	}
	return ast.NewFigureBlock(pos, refs), true
}

// parseTableDefBlock reads T#key:label lines until the closing `---` and
// returns a TableDef node (invisible in output, refs pre-populated by preScan).
func (p *blockParser) parseTableDefBlock(pos ast.Pos) (*ast.TableDef, bool) {
	refs := make(map[string]string)
	for {
		l := p.scanner.next()
		if l == nil {
			return nil, false
		}
		if isDefinitionDelimiterLine(l.content) {
			break
		}
		key, label, ok := parseTableEntryBytes(l.content)
		if !ok {
			return nil, false
		}
		refs[key] = label
	}
	if len(refs) == 0 {
		return nil, false
	}
	return ast.NewTableDef(pos, refs), true
}

// ─── Table-float helpers ──────────────────────────────────────────────────────

// isTableMetaRow returns true for lines of the form `|- ... -|`.
func isTableMetaRow(content []byte) bool {
	s := strings.TrimSpace(string(content))
	return len(s) >= 5 && strings.HasPrefix(s, "|-") && strings.HasSuffix(s, "-|")
}

// tableMetaContent strips the `|- ` prefix and ` -|` suffix and returns the
// trimmed inner content of a table meta row.
func tableMetaContent(content []byte) string {
	s := strings.TrimSpace(string(content))
	s = strings.TrimPrefix(s, "|-")
	s = strings.TrimSuffix(s, "-|")
	return strings.TrimSpace(s)
}

// parseTableMeta splits a `T#key:placement` string.
func parseTableMeta(s string) (key, placement string, ok bool) {
	if !strings.HasPrefix(s, "T#") {
		return "", "", false
	}
	idx := strings.IndexByte(s, ':')
	if idx < 3 {
		return "", "", false
	}
	k := strings.TrimSpace(s[:idx])
	pl := strings.TrimSpace(s[idx+1:])
	if k == "" || pl == "" {
		return "", "", false
	}
	return k, pl, true
}

// parseTableBlock reads consecutive `|- ... -|` meta rows (starting from the
// already-consumed firstMeta line), then the table header and body, and wraps
// the result in a TableBlock node.
//
// If the lines following the meta rows do not form a valid GFM table the
// method returns (nil, false) and the meta rows are silently dropped — the
// caller continues in the main parse loop.
func (p *blockParser) parseTableBlock(firstMeta *line, pos ast.Pos) (*ast.TableBlock, bool) {
	metaLines := [][]byte{firstMeta.content}

	// Consume any additional consecutive meta rows.
	for {
		next := p.scanner.peek()
		if next == nil || !isTableMetaRow(next.content) {
			break
		}
		metaLines = append(metaLines, p.scanner.next().content)
	}

	// The next line must be the table header row.
	headerLine := p.scanner.next()
	if headerLine == nil {
		return nil, false
	}

	// The line after that must be the column-separator row.
	if !isTableSeparator(p.scanner.peek()) {
		return nil, false
	}

	// Parse the GFM table (header + separator + body).
	tableHeaderPos := ast.Pos{Line: headerLine.lineNum, Col: 1, Offset: headerLine.offset}
	tableNode := p.parseTable(headerLine, tableHeaderPos)
	if tableNode == nil {
		return nil, false
	}

	// Extract metadata from the |- ... -| rows.
	var key, placement, captionStr string
	for _, raw := range metaLines {
		inner := tableMetaContent(raw)
		if k, pl, ok := parseTableMeta(inner); ok {
			key, placement = k, pl
		} else {
			captionStr = inner
		}
	}

	labelKey := p.refs.tables[key]

	// Parse the caption string as inline so that [C#01], [F#01], etc. resolve.
	captionNodes := parseInline(captionStr, pos, p.ext, p.refs.citations, p.refs.figures, p.refs.tables)

	block := ast.NewTableBlock(pos, key, labelKey, placement, captionNodes)
	ast.AppendChild(block, tableNode)
	return block, true
}

// ─── Unified definition-block helpers ────────────────────────────────────────

// isAnyDefinitionEntry returns true when the line is a recognised C#, F#, or
// T# entry, according to the currently enabled extensions.
func (p *blockParser) isAnyDefinitionEntry(content []byte) bool {
	if p.ext.Has(ExtCitations) {
		if _, _, ok := parseCitationEntryBytes(content); ok {
			return true
		}
	}
	if p.ext.Has(ExtFigures) {
		if _, _, ok := parseFigureEntryBytes(content); ok {
			return true
		}
	}
	if p.ext.Has(ExtTableFloat) {
		if _, _, ok := parseTableEntryBytes(content); ok {
			return true
		}
	}
	return false
}

// parseDefinitionBlock reads any mix of C#, F#, and T# entries until the
// closing `---` and returns a DefinitionBlock node. A line that matches none
// of the enabled entry types causes the block to be rejected (the opening
// `---` will fall through to thematic-break handling).
func (p *blockParser) parseDefinitionBlock(pos ast.Pos) (*ast.DefinitionBlock, bool) {
	citations := make(map[string]string)
	figures := make(map[string]string)
	tables := make(map[string]string)

	for {
		l := p.scanner.next()
		if l == nil {
			return nil, false
		}
		if isDefinitionDelimiterLine(l.content) {
			break
		}
		if p.ext.Has(ExtCitations) {
			if k, v, ok := parseCitationEntryBytes(l.content); ok {
				citations[k] = v
				continue
			}
		}
		if p.ext.Has(ExtFigures) {
			if k, v, ok := parseFigureEntryBytes(l.content); ok {
				figures[k] = v
				continue
			}
		}
		if p.ext.Has(ExtTableFloat) {
			if k, v, ok := parseTableEntryBytes(l.content); ok {
				tables[k] = v
				continue
			}
		}
		// Unrecognised line — not a valid definition block
		return nil, false
	}

	if len(citations)+len(figures)+len(tables) == 0 {
		return nil, false
	}
	return ast.NewDefinitionBlock(pos, citations, figures, tables), true
}
