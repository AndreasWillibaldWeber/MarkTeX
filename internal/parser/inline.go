// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/andreaswillibaldweber/marktex/internal/ast"
)

// parseInline converts a raw text string (a paragraph's collected lines) into a
// slice of inline AST nodes. It handles emphasis, code spans, links, images,
// escapes, and hard/soft breaks.
func parseInline(src string, pos ast.Pos, ext Extensions, citationRefs, figureRefs, tableRefs map[string]string) []ast.Node {
	p := &inlineParser{src: src, basePos: pos, ext: ext, citationRefs: citationRefs, figureRefs: figureRefs, tableRefs: tableRefs}
	return p.parse()
}

// inlineParser holds the state for inline parsing.
type inlineParser struct {
	src          string
	basePos      ast.Pos
	ext          Extensions
	pos          int
	citationRefs map[string]string
	figureRefs   map[string]string
	tableRefs    map[string]string

	// delims is the CommonMark delimiter stack for the emphasis scopes that
	// are currently open (the whole input, and each link text inside it).
	delims []*delimiter
}

func (p *inlineParser) parse() []ast.Node {
	var nodes []ast.Node
	for p.pos < len(p.src) {
		nodes = append(nodes, p.parseOne()...)
	}
	return p.finishInline(nodes, 0)
}

// parseOne parses the next inline element starting at p.pos and returns it.
func (p *inlineParser) parseOne() []ast.Node {
	ch := p.src[p.pos]

	switch {
	case ch == '\\' && p.pos+1 < len(p.src):
		return p.parseEscape()
	case ch == '`':
		return p.parseCodeSpan()
	case ch == '!' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '[':
		// Extended figure image takes priority: ![F#key:w:pos][caption](url)
		if p.ext.Has(ExtFigures) {
			if nodes, ok := p.parseFigureImage(); ok {
				return nodes
			}
		}
		if nodes, ok := p.parseImage(); ok {
			return nodes
		}
	case ch == '[':
		// Citation and figure refs take priority over regular links.
		if p.ext.Has(ExtCitations) {
			if nodes, ok := p.parseCitationRef(); ok {
				return nodes
			}
		}
		if p.ext.Has(ExtFigures) {
			if nodes, ok := p.parseFigureRef(); ok {
				return nodes
			}
		}
		if p.ext.Has(ExtTableFloat) {
			if nodes, ok := p.parseTableRef(); ok {
				return nodes
			}
		}
		if nodes, ok := p.parseLink(); ok {
			return nodes
		}
	case ch == '*' || ch == '_':
		return p.parseDelimiterRun(ch)
	case p.ext.Has(ExtStrikethrough) && ch == '~' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '~':
		if nodes, ok := p.parseStrikethrough(); ok {
			return nodes
		}
	case ch == '&':
		if nodes, ok := p.parseEntity(); ok {
			return nodes
		}
	case p.ext.Has(ExtMath) && ch == '$':
		if nodes, ok := p.parseMath(); ok {
			return nodes
		}
	case ch == '\n':
		return p.parseLineBreak()
	}

	// Fall through: consume one rune as text.
	r, size := utf8.DecodeRuneInString(p.src[p.pos:])
	nodePos := p.posAt(p.pos)
	p.pos += size
	return []ast.Node{ast.NewText(nodePos, string(r))}
}

// parseEscape handles `\X` where X is an ASCII punctuation character.
func (p *inlineParser) parseEscape() []ast.Node {
	nodePos := p.posAt(p.pos)
	next := p.src[p.pos+1]
	p.pos += 2
	if next == '\n' {
		return []ast.Node{ast.NewHardBreak(nodePos)}
	}
	if isASCIIPunctuation(rune(next)) {
		return []ast.Node{ast.NewEscape(nodePos, rune(next))}
	}
	// Not a valid escape; emit both characters as text.
	return []ast.Node{ast.NewText(nodePos, string([]byte{'\\', next}))}
}

// parseCodeSpan handles “ `code` “ and ``` “code“ ```.
func (p *inlineParser) parseCodeSpan() []ast.Node {
	nodePos := p.posAt(p.pos)
	// Count opening backticks
	start := p.pos
	n := 0
	for p.pos < len(p.src) && p.src[p.pos] == '`' {
		n++
		p.pos++
	}
	// Search for a matching run of exactly n backticks
	content := &strings.Builder{}
	for p.pos < len(p.src) {
		if p.src[p.pos] == '`' {
			closeStart := p.pos
			m := 0
			for p.pos < len(p.src) && p.src[p.pos] == '`' {
				m++
				p.pos++
			}
			if m == n {
				// Normalize: collapse interior whitespace/newlines per CommonMark
				raw := content.String()
				raw = strings.ReplaceAll(raw, "\n", " ")
				if len(raw) >= 2 && raw[0] == ' ' && raw[len(raw)-1] == ' ' {
					allSpaces := true
					for _, r := range raw {
						if r != ' ' {
							allSpaces = false
							break
						}
					}
					if !allSpaces {
						raw = raw[1 : len(raw)-1]
					}
				}
				return []ast.Node{ast.NewCodeSpan(nodePos, raw)}
			}
			// Not a match; include the backticks in the content
			content.WriteString(p.src[closeStart:p.pos])
		} else {
			r, size := utf8.DecodeRuneInString(p.src[p.pos:])
			content.WriteRune(r)
			p.pos += size
		}
	}
	// No closing delimiter found; treat the opening backticks as literal text
	p.pos = start + n
	return []ast.Node{ast.NewText(nodePos, strings.Repeat("`", n))}
}

// parseImage handles `![alt](url "title")`.
func (p *inlineParser) parseImage() ([]ast.Node, bool) {
	nodePos := p.posAt(p.pos)
	save := p.pos
	p.pos += 2 // consume `![`

	altNodes, ok := p.parseInlineContent(']')
	if !ok {
		p.pos = save
		return nil, false
	}

	dest, title, ok := p.parseLinkDestinationAndTitle()
	if !ok {
		p.pos = save
		return nil, false
	}

	img := ast.NewImage(nodePos, dest, title)
	for _, n := range altNodes {
		ast.AppendChild(img, n)
	}
	return []ast.Node{img}, true
}

// parseLink handles `[text](url "title")`.
func (p *inlineParser) parseLink() ([]ast.Node, bool) {
	nodePos := p.posAt(p.pos)
	save := p.pos
	p.pos++ // consume `[`

	textNodes, ok := p.parseInlineContent(']')
	if !ok {
		p.pos = save
		return nil, false
	}

	dest, title, ok := p.parseLinkDestinationAndTitle()
	if !ok {
		p.pos = save
		return nil, false
	}

	link := ast.NewLink(nodePos, dest, title)
	for _, n := range textNodes {
		ast.AppendChild(link, n)
	}
	return []ast.Node{link}, true
}

// parseInlineContent parses inline nodes up to (but not including) the
// closing byte 'close'. Consumes the closing byte.
func (p *inlineParser) parseInlineContent(close byte) ([]ast.Node, bool) {
	save := p.pos
	bottom := len(p.delims) // link text is a separate emphasis scope (CommonMark 6.3)
	var nodes []ast.Node
	depth := 0
	for p.pos < len(p.src) {
		if p.src[p.pos] == '[' {
			depth++
			nodes = append(nodes, ast.NewText(p.posAt(p.pos), "["))
			p.pos++
		} else if p.src[p.pos] == ']' {
			if depth > 0 {
				depth--
				nodes = append(nodes, ast.NewText(p.posAt(p.pos), "]"))
				p.pos++
			} else {
				p.pos++ // consume ']'
				return p.finishInline(nodes, bottom), true
			}
		} else {
			more := p.parseOne()
			nodes = append(nodes, more...)
		}
	}
	p.pos = save
	p.delims = p.delims[:bottom]
	return nil, false
}

// parseLinkDestinationAndTitle expects `(url)` or `(url "title")` at the
// current position. Returns the destination, title, and whether parsing
// succeeded.
func (p *inlineParser) parseLinkDestinationAndTitle() (dest, title string, ok bool) {
	if p.pos >= len(p.src) || p.src[p.pos] != '(' {
		return "", "", false
	}
	p.pos++ // consume '('

	// Skip leading whitespace
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}

	// Parse destination: either angle-bracket form or plain
	if p.pos < len(p.src) && p.src[p.pos] == '<' {
		dest, ok = p.parseAngleBracketDest()
	} else {
		dest, ok = p.parsePlainDest()
	}
	if !ok {
		return "", "", false
	}

	// Skip whitespace before optional title
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t' || p.src[p.pos] == '\n') {
		p.pos++
	}

	// Parse optional title
	if p.pos < len(p.src) && (p.src[p.pos] == '"' || p.src[p.pos] == '\'' || p.src[p.pos] == '(') {
		title, ok = p.parseLinkTitle()
		if !ok {
			return "", "", false
		}
	}

	// Skip trailing whitespace
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}

	if p.pos >= len(p.src) || p.src[p.pos] != ')' {
		return "", "", false
	}
	p.pos++ // consume ')'
	return dest, title, true
}

func (p *inlineParser) parseAngleBracketDest() (string, bool) {
	p.pos++ // consume '<'
	start := p.pos
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		if ch == '>' {
			dest := p.src[start:p.pos]
			p.pos++ // consume '>'
			return string(dest), true
		}
		if ch == '\n' || ch == '<' {
			return "", false
		}
		p.pos++
	}
	return "", false
}

func (p *inlineParser) parsePlainDest() (string, bool) {
	start := p.pos
	depth := 0
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		if ch == '(' {
			depth++
		} else if ch == ')' {
			if depth == 0 {
				break
			}
			depth--
		} else if ch == ' ' || ch == '\t' || ch == '\n' {
			break
		}
		if ch == '\\' && p.pos+1 < len(p.src) {
			p.pos += 2
			continue
		}
		p.pos++
	}
	return string(p.src[start:p.pos]), true
}

func (p *inlineParser) parseLinkTitle() (string, bool) {
	open := p.src[p.pos]
	var close byte
	switch open {
	case '"':
		close = '"'
	case '\'':
		close = '\''
	case '(':
		close = ')'
	default:
		return "", false
	}
	p.pos++ // consume opening delimiter
	var b strings.Builder
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		if ch == close {
			p.pos++ // consume closing delimiter
			return b.String(), true
		}
		if ch == '\\' && p.pos+1 < len(p.src) {
			p.pos++
			b.WriteByte(p.src[p.pos])
			p.pos++
			continue
		}
		if open == '(' && ch == '(' {
			return "", false // unescaped '(' inside paren title is not allowed
		}
		b.WriteByte(ch)
		p.pos++
	}
	return "", false
}

// ─── Emphasis: CommonMark delimiter algorithm (spec 6.2 and appendix A) ─────

// delimiter is a run of `*` or `_` that may open or close emphasis. The run
// stays in the node list as plain text until processEmphasis decides.
type delimiter struct {
	node     *ast.Text
	char     byte
	length   int // delimiters not yet consumed
	origLen  int // length of the run in the source, for the rule of three
	canOpen  bool
	canClose bool
}

// parseDelimiterRun consumes a run of `*` or `_`, emits it as text and records
// it on the delimiter stack when the flanking rules allow it to open or close.
func (p *inlineParser) parseDelimiterRun(ch byte) []ast.Node {
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] == ch {
		p.pos++
	}
	node := ast.NewText(p.posAt(start), p.src[start:p.pos])

	before, after := ' ', ' ' // start and end of input count as whitespace
	if start > 0 {
		before, _ = utf8.DecodeLastRuneInString(p.src[:start])
	}
	if p.pos < len(p.src) {
		after, _ = utf8.DecodeRuneInString(p.src[p.pos:])
	}
	canOpen, canClose := delimiterCanOpenClose(ch, before, after)
	if canOpen || canClose {
		n := p.pos - start
		p.delims = append(p.delims, &delimiter{node: node, char: ch, length: n, origLen: n, canOpen: canOpen, canClose: canClose})
	}
	return []ast.Node{node}
}

// delimiterCanOpenClose applies the left-/right-flanking definitions of
// CommonMark 6.2 to a run of ch surrounded by the runes before and after.
// `*` opens when left-flanking and closes when right-flanking; `_` additionally
// refuses to open or close inside a word unless punctuation is adjacent.
func delimiterCanOpenClose(ch byte, before, after rune) (canOpen, canClose bool) {
	leftFlanking := !unicode.IsSpace(after) &&
		(!isUnicodePunctuation(after) || unicode.IsSpace(before) || isUnicodePunctuation(before))
	rightFlanking := !unicode.IsSpace(before) &&
		(!isUnicodePunctuation(before) || unicode.IsSpace(after) || isUnicodePunctuation(after))
	if ch == '_' {
		return leftFlanking && (!rightFlanking || isUnicodePunctuation(before)),
			rightFlanking && (!leftFlanking || isUnicodePunctuation(after))
	}
	return leftFlanking, rightFlanking
}

// finishInline closes an emphasis scope: the delimiters pushed since bottom
// are matched, the scope is popped, adjacent text is merged and line breaks
// are resolved.
func (p *inlineParser) finishInline(nodes []ast.Node, bottom int) []ast.Node {
	nodes = processEmphasis(nodes, p.delims[bottom:])
	p.delims = p.delims[:bottom]
	return finalizeInline(nodes)
}

func finalizeInline(nodes []ast.Node) []ast.Node {
	return resolveLineBreaks(mergeTextNodes(nodes))
}

// processEmphasis is the "process emphasis" procedure of the CommonMark spec
// for one scope. Closers are visited from the bottom of the stack upwards.
// Each looks back for the nearest opener it may match and wraps the nodes in
// between in an Emphasis (one delimiter) or Strong (two) node, repeating while
// it has delimiters left. Unmatched delimiters remain literal text.
func processEmphasis(nodes []ast.Node, scope []*delimiter) []ast.Node {
	stack := append([]*delimiter(nil), scope...)
	for i := 0; i < len(stack); {
		closer := stack[i]
		if !closer.canClose {
			i++
			continue
		}
		o := i - 1
		for o >= 0 && !canMatch(stack[o], closer) {
			o--
		}
		if o < 0 {
			if closer.canOpen {
				i++
			} else {
				stack = removeDelimiter(stack, i)
			}
			continue
		}
		opener := stack[o]
		oi := indexOfNode(nodes, opener.node)
		ci := indexOfNode(nodes, closer.node)
		if oi < 0 || ci <= oi {
			stack = removeDelimiter(stack, i)
			continue
		}

		use := 1
		if opener.length >= 2 && closer.length >= 2 {
			use = 2
		}
		children := finalizeInline(append([]ast.Node(nil), nodes[oi+1:ci]...))
		wrapper := newEmphasisNode(use, opener.node.Position(), children)
		rebuilt := make([]ast.Node, 0, len(nodes)-(ci-oi)+2)
		rebuilt = append(rebuilt, nodes[:oi+1]...)
		rebuilt = append(rebuilt, wrapper)
		nodes = append(rebuilt, nodes[ci:]...)

		opener.length -= use
		closer.length -= use
		opener.node.Value = opener.node.Value[:opener.length]
		closer.node.Value = closer.node.Value[:closer.length]

		// Delimiters between the pair now live inside the new node.
		stack = append(stack[:o+1], stack[i:]...)
		i = o + 1
		if closer.length == 0 {
			nodes = removeNode(nodes, closer.node)
			stack = removeDelimiter(stack, i)
		}
		if opener.length == 0 {
			nodes = removeNode(nodes, opener.node)
			stack = removeDelimiter(stack, o)
			i--
		}
	}
	return nodes
}

// canMatch reports whether opener can pair with closer: same character, the
// opener may open, and the rule of three holds (CommonMark 6.2, rules 9 and
// 10): when either delimiter can both open and close, the sum of the two run
// lengths must not be a multiple of three unless both lengths are.
func canMatch(opener, closer *delimiter) bool {
	if opener.char != closer.char || !opener.canOpen {
		return false
	}
	if opener.canClose || closer.canOpen {
		sum := opener.origLen + closer.origLen
		if sum%3 == 0 && (opener.origLen%3 != 0 || closer.origLen%3 != 0) {
			return false
		}
	}
	return true
}

// newEmphasisNode wraps children in Emphasis (use 1) or Strong (use 2).
// Emphasis around a lone Strong, or Strong around a lone Emphasis, collapses
// into StrongEmphasis so that ***text*** keeps rendering as \textbf{\textit{…}}.
func newEmphasisNode(use int, pos ast.Pos, children []ast.Node) ast.Node {
	if len(children) == 1 {
		switch inner := children[0].(type) {
		case *ast.Strong:
			if use == 1 {
				return withChildren(ast.NewStrongEmphasis(pos), inner.Children())
			}
		case *ast.Emphasis:
			if use == 2 {
				return withChildren(ast.NewStrongEmphasis(pos), inner.Children())
			}
		}
	}
	if use == 2 {
		return withChildren(ast.NewStrong(pos), children)
	}
	return withChildren(ast.NewEmphasis(pos), children)
}

func withChildren(parent ast.Node, children []ast.Node) ast.Node {
	for _, c := range children {
		ast.AppendChild(parent, c)
	}
	return parent
}

func indexOfNode(nodes []ast.Node, target ast.Node) int {
	for i, n := range nodes {
		if n == target {
			return i
		}
	}
	return -1
}

func removeNode(nodes []ast.Node, target ast.Node) []ast.Node {
	if i := indexOfNode(nodes, target); i >= 0 {
		return append(nodes[:i], nodes[i+1:]...)
	}
	return nodes
}

func removeDelimiter(stack []*delimiter, i int) []*delimiter {
	return append(stack[:i], stack[i+1:]...)
}

// parseStrikethrough handles `~~text~~`.
func (p *inlineParser) parseStrikethrough() ([]ast.Node, bool) {
	nodePos := p.posAt(p.pos)
	save := p.pos
	p.pos += 2 // consume `~~`

	// Find closing `~~`
	start := p.pos
	for p.pos < len(p.src) {
		if p.pos+1 < len(p.src) && p.src[p.pos] == '~' && p.src[p.pos+1] == '~' {
			inner := p.src[start:p.pos]
			p.pos += 2
			innerNodes := parseInline(string(inner), nodePos, p.ext, p.citationRefs, p.figureRefs, p.tableRefs)
			st := ast.NewStrikethrough(nodePos)
			for _, n := range innerNodes {
				ast.AppendChild(st, n)
			}
			return []ast.Node{st}, true
		}
		if p.src[p.pos] == '\n' {
			break
		}
		p.pos++
	}
	p.pos = save
	return nil, false
}

// parseEntity handles `&name;`, `&#nnnn;`, `&#xHHHH;`.
func (p *inlineParser) parseEntity() ([]ast.Node, bool) {
	nodePos := p.posAt(p.pos)
	save := p.pos

	end := strings.IndexByte(p.src[p.pos:], ';')
	if end < 0 || end > 32 {
		return nil, false
	}
	seq := p.src[p.pos : p.pos+end+1]
	r, ok := decodeHTMLEntity(seq)
	if !ok {
		p.pos = save
		return nil, false
	}
	p.pos += end + 1
	return []ast.Node{ast.NewEntity(nodePos, seq, r)}, true
}

// parseMath handles Pandoc-style TeX math and keeps the content verbatim.
//
// Display math is `$$…$$` with non-empty content; surrounding whitespace is
// trimmed and the content may span lines. Inline math is `$…$` where the
// opening `$` is followed by a non-space character and the closing `$` is
// preceded by one and is not followed by a digit. A backslash-escaped
// character inside the math never closes it. If the first closing candidate
// violates a rule the whole span is rejected, as in Pandoc, so "$20,000 and
// $30,000" stays literal text. Stricter than Pandoc, inline math must stay on
// one line and may not contain a backtick: otherwise a stray "$5" would run
// on until the next dollar sign, even one inside a later code span.
func (p *inlineParser) parseMath() ([]ast.Node, bool) {
	nodePos := p.posAt(p.pos)
	start := p.pos

	if start+1 < len(p.src) && p.src[start+1] == '$' {
		open := start + 2
		end := indexUnescaped(p.src, open, "$$")
		if end < 0 {
			return nil, false
		}
		literal := strings.TrimSpace(p.src[open:end])
		if literal == "" {
			return nil, false
		}
		p.pos = end + 2
		return []ast.Node{ast.NewMath(nodePos, literal, true)}, true
	}

	open := start + 1
	if open >= len(p.src) || isSpaceByte(p.src[open]) {
		return nil, false
	}
	end := indexUnescaped(p.src, open, "$")
	if end < 0 || isSpaceByte(p.src[end-1]) || strings.ContainsAny(p.src[open:end], "\n`") {
		return nil, false
	}
	if end+1 < len(p.src) && p.src[end+1] >= '0' && p.src[end+1] <= '9' {
		return nil, false
	}
	p.pos = end + 1
	return []ast.Node{ast.NewMath(nodePos, p.src[open:end], false)}, true
}

// indexUnescaped returns the index of the first occurrence of sep in src at or
// after from, skipping any character preceded by a backslash. Returns -1 when
// sep is not found.
func indexUnescaped(src string, from int, sep string) int {
	for i := from; i+len(sep) <= len(src); i++ {
		if src[i] == '\\' {
			i++ // skip the escaped character
			continue
		}
		if src[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

func isSpaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n'
}

// parseLineBreak handles `\n` within inline content. It always emits a
// SoftBreak; resolveLineBreaks later upgrades it to a HardBreak when two or
// more spaces precede the line ending and strips the surrounding whitespace.
func (p *inlineParser) parseLineBreak() []ast.Node {
	nodePos := p.posAt(p.pos)
	p.pos++ // consume '\n'
	return []ast.Node{ast.NewSoftBreak(nodePos)}
}

// posAt returns an ast.Pos for the given byte offset within p.src.
// parseFigureRef handles `[F#key]` inline figure cross-references.
// Resolves to ~\ref{fig:label}.
func (p *inlineParser) parseFigureRef() ([]ast.Node, bool) {
	if p.pos+3 >= len(p.src) || p.src[p.pos+1] != 'F' || p.src[p.pos+2] != '#' {
		return nil, false
	}
	save := p.pos
	p.pos++ // consume '['
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != ']' && p.src[p.pos] != '\n' {
		p.pos++
	}
	if p.pos >= len(p.src) || p.src[p.pos] != ']' {
		p.pos = save
		return nil, false
	}
	key := p.src[start:p.pos]
	p.pos++ // consume ']'

	labelKey, ok := p.figureRefs[key]
	if !ok {
		p.pos = save
		return nil, false
	}
	return []ast.Node{ast.NewFigureRef(p.posAt(save), key, labelKey)}, true
}

// parseFigureImage handles the extended figure syntax:
//
//	![F#key:width:placement][Caption text.](path/to/image.png)
func (p *inlineParser) parseFigureImage() ([]ast.Node, bool) {
	// Quick prefix check: must start with ![F#
	if p.pos+4 >= len(p.src) ||
		p.src[p.pos+1] != '[' ||
		p.src[p.pos+2] != 'F' ||
		p.src[p.pos+3] != '#' {
		return nil, false
	}
	save := p.pos
	p.pos += 2 // consume '!['

	// First bracket: F#key:width:placement
	metaStart := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != ']' && p.src[p.pos] != '\n' {
		p.pos++
	}
	if p.pos >= len(p.src) || p.src[p.pos] != ']' {
		p.pos = save
		return nil, false
	}
	meta := p.src[metaStart:p.pos]
	p.pos++ // consume ']'

	key, width, placement, ok := parseFigureMeta(meta)
	if !ok {
		p.pos = save
		return nil, false
	}

	// Second bracket: caption text
	if p.pos >= len(p.src) || p.src[p.pos] != '[' {
		p.pos = save
		return nil, false
	}
	p.pos++ // consume '['
	// parseCaptionContent calls parseOne() for all characters so [C#01], [F#01],
	// etc. are dispatched to the citation/figure/table ref parsers rather than
	// being depth-counted as bracket pairs (which parseInlineContent does).
	captionNodes, ok2 := p.parseCaptionContent()
	if !ok2 {
		p.pos = save
		return nil, false
	}

	// Parenthesised path
	dest, _, ok3 := p.parseLinkDestinationAndTitle()
	if !ok3 {
		p.pos = save
		return nil, false
	}

	// Resolve label — fall back to key itself if not in registry
	labelKey := p.figureRefs[key]
	if labelKey == "" {
		labelKey = key
	}

	nodePos := p.posAt(save)
	return []ast.Node{ast.NewFigureImage(nodePos, key, labelKey, width, placement, captionNodes, dest)}, true
}

// parseFigureMeta parses the `F#key:width:placement` string from the first
// bracket of an extended figure image. Width must be a valid decimal number.
func parseFigureMeta(s string) (key, width, placement string, ok bool) {
	if !strings.HasPrefix(s, "F#") {
		return "", "", "", false
	}
	// Expect exactly two ':' separators: F#key : width : placement
	first := strings.IndexByte(s, ':')
	if first < 3 { // at minimum "F#x:"
		return "", "", "", false
	}
	second := strings.IndexByte(s[first+1:], ':')
	if second < 0 {
		return "", "", "", false
	}
	second += first + 1 // absolute index

	k := strings.TrimSpace(s[:first])
	w := strings.TrimSpace(s[first+1 : second])
	pl := strings.TrimSpace(s[second+1:])

	if k == "" || w == "" || pl == "" {
		return "", "", "", false
	}
	// Width must be a numeric value (integer or decimal)
	for _, r := range w {
		if r != '.' && (r < '0' || r > '9') {
			return "", "", "", false
		}
	}
	return k, w, pl, true
}

// parseCaptionContent parses inline content until the first unmatched `]`.
// Unlike parseInlineContent it does NOT depth-count `[` — instead it calls
// parseOne() for every character including `[`, so citation refs ([C#key]),
// figure refs ([F#key]), and table refs ([T#key]) inside caption text are
// dispatched to their respective parsers and resolve correctly.
//
// The trade-off: a literal `]` that is not part of a known ref or link will
// terminate the caption early. Authors can escape it as `\]`.
func (p *inlineParser) parseCaptionContent() ([]ast.Node, bool) {
	save := p.pos
	var nodes []ast.Node
	for p.pos < len(p.src) {
		if p.src[p.pos] == ']' {
			p.pos++ // consume caption-closing ']'
			return mergeTextNodes(nodes), true
		}
		if p.src[p.pos] == '\n' {
			break // captions do not span lines
		}
		more := p.parseOne()
		nodes = append(nodes, more...)
	}
	p.pos = save
	return nil, false
}

// parseTableRef handles `[T#key]` inline table cross-references → ~\ref{tab:label}.
func (p *inlineParser) parseTableRef() ([]ast.Node, bool) {
	if p.pos+3 >= len(p.src) || p.src[p.pos+1] != 'T' || p.src[p.pos+2] != '#' {
		return nil, false
	}
	save := p.pos
	p.pos++ // consume '['
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != ']' && p.src[p.pos] != '\n' {
		p.pos++
	}
	if p.pos >= len(p.src) || p.src[p.pos] != ']' {
		p.pos = save
		return nil, false
	}
	key := p.src[start:p.pos]
	p.pos++ // consume ']'

	labelKey, ok := p.tableRefs[key]
	if !ok {
		p.pos = save
		return nil, false
	}
	return []ast.Node{ast.NewTableRef(p.posAt(save), key, labelKey)}, true
}

// parseCitationRef handles `[C#key]` inline citation references.
// It returns a CitationRef node when the key is found in the citation registry,
// and (nil, false) otherwise so the caller can fall through to link parsing.
func (p *inlineParser) parseCitationRef() ([]ast.Node, bool) {
	// Quick prefix check before we commit to scanning forward.
	if p.pos+3 >= len(p.src) || p.src[p.pos+1] != 'C' || p.src[p.pos+2] != '#' {
		return nil, false
	}
	save := p.pos
	p.pos++ // consume '['

	// Read the key until ']' or end-of-line.
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != ']' && p.src[p.pos] != '\n' {
		p.pos++
	}
	if p.pos >= len(p.src) || p.src[p.pos] != ']' {
		p.pos = save
		return nil, false
	}
	key := p.src[start:p.pos]
	p.pos++ // consume ']'

	bibKey, ok := p.citationRefs[key]
	if !ok {
		p.pos = save
		return nil, false
	}

	nodePos := p.posAt(save)
	return []ast.Node{ast.NewCitationRef(nodePos, key, bibKey)}, true
}

func (p *inlineParser) posAt(offset int) ast.Pos {
	return ast.Pos{
		Offset: p.basePos.Offset + offset,
		Line:   p.basePos.Line,
		Col:    p.basePos.Col + offset,
	}
}

// mergeTextNodes collapses adjacent *ast.Text nodes into a single Text node.
// resolveLineBreaks applies the CommonMark line-ending rules (6.7, 6.8) to a
// merged inline sequence: whitespace before and after a line ending is
// removed, two or more spaces before a soft break turn it into a hard break,
// and a break with nothing after it is dropped because a hard break cannot end
// a block (a trailing `\\` would also be a LaTeX error).
func resolveLineBreaks(nodes []ast.Node) []ast.Node {
	out := make([]ast.Node, 0, len(nodes))
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		if !isLineBreak(n) {
			out = append(out, n)
			continue
		}

		trailing := 0
		if len(out) > 0 {
			if t, ok := out[len(out)-1].(*ast.Text); ok {
				trimmed := strings.TrimRight(t.Value, " \t")
				trailing = len(t.Value) - len(trimmed)
				if trimmed == "" {
					out = out[:len(out)-1]
				} else if trailing > 0 {
					out[len(out)-1] = ast.NewText(t.Position(), trimmed)
				}
			}
		}

		// Leading whitespace on the next line is ignored; the text node may
		// disappear entirely.
		next := i + 1
		if next < len(nodes) {
			if t, ok := nodes[next].(*ast.Text); ok {
				trimmed := strings.TrimLeft(t.Value, " \t")
				if trimmed == "" {
					next++
				} else if len(trimmed) != len(t.Value) {
					nodes[next] = ast.NewText(t.Position(), trimmed)
				}
			}
		}
		if next >= len(nodes) {
			break
		}

		if soft, ok := n.(*ast.SoftBreak); ok && trailing >= 2 {
			n = ast.NewHardBreak(soft.Position())
		}
		out = append(out, n)
		i = next - 1
	}
	return out
}

func isLineBreak(n ast.Node) bool {
	switch n.(type) {
	case *ast.SoftBreak, *ast.HardBreak:
		return true
	}
	return false
}

func mergeTextNodes(nodes []ast.Node) []ast.Node {
	if len(nodes) <= 1 {
		return nodes
	}
	out := make([]ast.Node, 0, len(nodes))
	for _, n := range nodes {
		if t, ok := n.(*ast.Text); ok {
			if len(out) > 0 {
				if prev, ok := out[len(out)-1].(*ast.Text); ok {
					out[len(out)-1] = ast.NewText(prev.Position(), prev.Value+t.Value)
					continue
				}
			}
		}
		out = append(out, n)
	}
	return out
}

// decodeHTMLEntity decodes a named or numeric HTML entity string (e.g. "&amp;",
// "&#65;", "&#x41;"). Returns the rune and true on success.
func decodeHTMLEntity(seq string) (rune, bool) {
	if len(seq) < 3 || seq[0] != '&' || seq[len(seq)-1] != ';' {
		return 0, false
	}
	body := seq[1 : len(seq)-1]
	if len(body) == 0 {
		return 0, false
	}
	if body[0] == '#' {
		// Numeric entity
		num := body[1:]
		if len(num) == 0 {
			return 0, false
		}
		base := 10
		if num[0] == 'x' || num[0] == 'X' {
			base = 16
			num = num[1:]
		}
		if len(num) == 0 {
			return 0, false
		}
		var n rune
		for _, ch := range num {
			var digit rune
			switch {
			case ch >= '0' && ch <= '9':
				digit = ch - '0'
			case ch >= 'a' && ch <= 'f':
				digit = ch - 'a' + 10
			case ch >= 'A' && ch <= 'F':
				digit = ch - 'A' + 10
			default:
				return 0, false
			}
			if base == 10 && digit >= 10 {
				return 0, false
			}
			n = n*rune(base) + digit
		}
		if n == 0 || n > unicode.MaxRune {
			return 0xFFFD, true // replacement character
		}
		return n, true
	}
	// Named entity — consult a minimal lookup table
	if r, ok := namedEntities[body]; ok {
		return r, true
	}
	return 0, false
}

// namedEntities is a minimal subset of common HTML entities.
// A production implementation would include the full ~2000-entry table.
var namedEntities = map[string]rune{
	"amp":    '&',
	"lt":     '<',
	"gt":     '>',
	"quot":   '"',
	"apos":   '\'',
	"nbsp":   '\u00A0',
	"copy":   '©',
	"reg":    '®',
	"trade":  '™',
	"mdash":  '—',
	"ndash":  '–',
	"ldquo":  '\u201C',
	"rdquo":  '\u201D',
	"lsquo":  '\u2018',
	"rsquo":  '\u2019',
	"hellip": '…',
	"bull":   '•',
	"laquo":  '«',
	"raquo":  '»',
	"euro":   '€',
	"pound":  '£',
	"yen":    '¥',
	"deg":    '°',
	"plusmn": '±',
	"times":  '×',
	"divide": '÷',
	"frac12": '½',
	"frac14": '¼',
	"frac34": '¾',
	"infin":  '∞',
	"sum":    '∑',
	"prod":   '∏',
	"radic":  '√',
	"pi":     'π',
	"mu":     'μ',
	"alpha":  'α',
	"beta":   'β',
	"gamma":  'γ',
	"delta":  'δ',
	"sigma":  'σ',
	"omega":  'ω',
}
