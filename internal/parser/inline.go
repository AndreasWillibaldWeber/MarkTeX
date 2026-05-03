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
}

func (p *inlineParser) parse() []ast.Node {
	var nodes []ast.Node
	for p.pos < len(p.src) {
		nodes = append(nodes, p.parseOne()...)
	}
	return mergeTextNodes(nodes)
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
		if nodes, ok := p.parseEmphasis(ch); ok {
			return nodes
		}
	case p.ext.Has(ExtStrikethrough) && ch == '~' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '~':
		if nodes, ok := p.parseStrikethrough(); ok {
			return nodes
		}
	case ch == '&':
		if nodes, ok := p.parseEntity(); ok {
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
				return mergeTextNodes(nodes), true
			}
		} else {
			more := p.parseOne()
			nodes = append(nodes, more...)
		}
	}
	p.pos = save
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

// parseEmphasis handles `*`, `**`, `***`, `_`, `__`, `___`.
// This is a simplified (non-CommonMark-compliant) greedy implementation that
// handles the common cases correctly. The full delimiter-stack algorithm is
// more complex and can be added in a future iteration.
func (p *inlineParser) parseEmphasis(delim byte) ([]ast.Node, bool) {
	nodePos := p.posAt(p.pos)
	save := p.pos

	// Count opening delimiters
	n := 0
	for p.pos < len(p.src) && p.src[p.pos] == delim {
		n++
		p.pos++
	}

	// Underscore delimiters: check left-flanking rule
	if delim == '_' {
		// Must not be followed by an alphanumeric (right-flanking would mean it's
		// inside a word like `foo_bar_baz`)
		if p.pos < len(p.src) {
			r, _ := utf8.DecodeRuneInString(p.src[p.pos:])
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				p.pos = save
				return nil, false
			}
		}
	}

	level := n
	if level > 3 {
		// More than 3 delimiters: treat excess as text, then parse emphasis
		p.pos = save + 3
		level = 3
	}

	// Now consume inner content until we find the matching closing delimiter run
	innerSrc, ok := p.findClosingDelimiters(delim, level)
	if !ok {
		p.pos = save
		return nil, false
	}

	// Parse the inner content recursively
	innerNodes := parseInline(innerSrc, nodePos, p.ext, p.citationRefs, p.figureRefs, p.tableRefs)

	switch level {
	case 1:
		em := ast.NewEmphasis(nodePos)
		for _, n := range innerNodes {
			ast.AppendChild(em, n)
		}
		return []ast.Node{em}, true
	case 2:
		st := ast.NewStrong(nodePos)
		for _, n := range innerNodes {
			ast.AppendChild(st, n)
		}
		return []ast.Node{st}, true
	default: // 3
		se := ast.NewStrongEmphasis(nodePos)
		for _, n := range innerNodes {
			ast.AppendChild(se, n)
		}
		return []ast.Node{se}, true
	}
}

// findClosingDelimiters searches forward for a run of exactly n delimiter
// characters followed by a non-delimiter (or end of line). Returns the
// content between the delimiters and true on success.
func (p *inlineParser) findClosingDelimiters(delim byte, n int) (string, bool) {
	start := p.pos
	for p.pos < len(p.src) {
		if p.src[p.pos] == '\\' && p.pos+1 < len(p.src) {
			p.pos += 2
			continue
		}
		if p.src[p.pos] == '`' {
			// Skip over code span to avoid matching delimiters inside it
			save := p.pos
			tmp := &inlineParser{src: p.src, basePos: p.basePos, ext: p.ext, pos: p.pos}
			nodes := tmp.parseCodeSpan()
			if len(nodes) > 0 {
				if _, ok := nodes[0].(*ast.CodeSpan); ok {
					p.pos = tmp.pos
					continue
				}
			}
			p.pos = save
		}
		if p.src[p.pos] == delim {
			closeStart := p.pos
			m := 0
			for p.pos < len(p.src) && p.src[p.pos] == delim {
				m++
				p.pos++
			}
			if m == n {
				// For underscore: closing run must be right-flanking
				if delim == '_' && p.pos < len(p.src) {
					r, _ := utf8.DecodeRuneInString(p.src[p.pos:])
					if unicode.IsLetter(r) || unicode.IsDigit(r) {
						// Not right-flanking; continue
						continue
					}
				}
				return p.src[start:closeStart], true
			}
			// Wrong number of delimiters; continue
			continue
		}
		if p.src[p.pos] == '\n' {
			// Emphasis doesn't span paragraphs
			break
		}
		p.pos++
	}
	return "", false
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

// parseLineBreak handles `\n` within inline content.
// Two trailing spaces before \n produce a HardBreak; otherwise SoftBreak.
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
