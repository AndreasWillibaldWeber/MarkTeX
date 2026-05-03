// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package ast

// ─── Inline nodes ─────────────────────────────────────────────────────────────

// Text is a run of plain characters with no special formatting.
type Text struct {
	inlineBase
	Value string
}

func NewText(pos Pos, value string) *Text  { return &Text{inlineBase{pos: pos}, value} }
func (t *Text) Type() NodeType             { return NodeText }
func (t *Text) Children() []Node           { return nil }

// SoftBreak represents a newline within paragraph source that renders as a space.
type SoftBreak struct{ inlineBase }

func NewSoftBreak(pos Pos) *SoftBreak { return &SoftBreak{inlineBase{pos: pos}} }
func (s *SoftBreak) Type() NodeType   { return NodeSoftBreak }
func (s *SoftBreak) Children() []Node { return nil }

// HardBreak is a forced line break (two trailing spaces or backslash-newline).
type HardBreak struct{ inlineBase }

func NewHardBreak(pos Pos) *HardBreak { return &HardBreak{inlineBase{pos: pos}} }
func (h *HardBreak) Type() NodeType   { return NodeHardBreak }
func (h *HardBreak) Children() []Node { return nil }

// Emphasis wraps italic text (`*text*` or `_text_`).
type Emphasis struct{ inlineBase }

func NewEmphasis(pos Pos) *Emphasis    { return &Emphasis{inlineBase{pos: pos}} }
func (e *Emphasis) Type() NodeType     { return NodeEmphasis }

// Strong wraps bold text (`**text**` or `__text__`).
type Strong struct{ inlineBase }

func NewStrong(pos Pos) *Strong    { return &Strong{inlineBase{pos: pos}} }
func (s *Strong) Type() NodeType   { return NodeStrong }

// StrongEmphasis wraps bold-italic text (`***text***`).
type StrongEmphasis struct{ inlineBase }

func NewStrongEmphasis(pos Pos) *StrongEmphasis { return &StrongEmphasis{inlineBase{pos: pos}} }
func (se *StrongEmphasis) Type() NodeType       { return NodeStrongEmphasis }

// CodeSpan holds a backtick-delimited inline code fragment.
// The content is taken verbatim; no further parsing is performed inside it.
type CodeSpan struct {
	inlineBase
	Literal string
}

func NewCodeSpan(pos Pos, literal string) *CodeSpan {
	return &CodeSpan{inlineBase: inlineBase{pos: pos}, Literal: literal}
}
func (c *CodeSpan) Type() NodeType    { return NodeCodeSpan }
func (c *CodeSpan) Children() []Node  { return nil }

// Link holds an inline or reference-style hyperlink.
// Children contains the link text (parsed as inline).
type Link struct {
	inlineBase
	Destination string
	Title       string
}

func NewLink(pos Pos, dest, title string) *Link {
	return &Link{inlineBase: inlineBase{pos: pos}, Destination: dest, Title: title}
}
func (l *Link) Type() NodeType { return NodeLink }

// Image holds an inline or reference-style image.
// Alt contains the alternative text nodes (used for plain-text extraction).
type Image struct {
	inlineBase
	Destination string
	Title       string
}

func NewImage(pos Pos, dest, title string) *Image {
	return &Image{inlineBase: inlineBase{pos: pos}, Destination: dest, Title: title}
}
func (i *Image) Type() NodeType { return NodeImage }

// RawHTML is an inline HTML tag that the generator passes through or drops.
type RawHTML struct {
	inlineBase
	Literal string
}

func NewRawHTML(pos Pos, literal string) *RawHTML {
	return &RawHTML{inlineBase: inlineBase{pos: pos}, Literal: literal}
}
func (r *RawHTML) Type() NodeType    { return NodeRawHTML }
func (r *RawHTML) Children() []Node  { return nil }

// Escape holds a backslash-escaped ASCII punctuation character.
type Escape struct {
	inlineBase
	Char rune
}

func NewEscape(pos Pos, char rune) *Escape {
	return &Escape{inlineBase: inlineBase{pos: pos}, Char: char}
}
func (e *Escape) Type() NodeType    { return NodeEscape }
func (e *Escape) Children() []Node  { return nil }

// Entity holds a decoded HTML entity. Sequence is the original source form
// (e.g. `&amp;`). Rune is the decoded Unicode code point.
type Entity struct {
	inlineBase
	Sequence string
	Rune     rune
}

func NewEntity(pos Pos, seq string, r rune) *Entity {
	return &Entity{inlineBase: inlineBase{pos: pos}, Sequence: seq, Rune: r}
}
func (e *Entity) Type() NodeType    { return NodeEntity }
func (e *Entity) Children() []Node  { return nil }

// Strikethrough wraps GFM ~~strikethrough~~ text.
type Strikethrough struct{ inlineBase }

func NewStrikethrough(pos Pos) *Strikethrough { return &Strikethrough{inlineBase{pos: pos}} }
func (s *Strikethrough) Type() NodeType       { return NodeStrikethrough }

// ─── Citation extension ───────────────────────────────────────────────────────

// CitationRef is an inline citation reference [C#key] that resolves to
// ~\cite{bibtex} in LaTeX. Key is the short label used in the Markdown source;
// BibKey is the resolved BibTeX identifier from the CitationBlock.
type CitationRef struct {
	inlineBase
	Key    string // e.g. "C#01"
	BibKey string // e.g. "doe2023survey"
}

func NewCitationRef(pos Pos, key, bibKey string) *CitationRef {
	return &CitationRef{inlineBase: inlineBase{pos: pos}, Key: key, BibKey: bibKey}
}
func (c *CitationRef) Type() NodeType    { return NodeCitationRef }
func (c *CitationRef) Children() []Node  { return nil }

// ─── Figure extension ─────────────────────────────────────────────────────────

// FigureRef is an inline figure reference [F#key] that resolves to
// ~\ref{fig:label} in LaTeX.
type FigureRef struct {
	inlineBase
	Key      string // e.g. "F#01"
	LabelKey string // e.g. "figurelabel01"
}

func NewFigureRef(pos Pos, key, labelKey string) *FigureRef {
	return &FigureRef{inlineBase: inlineBase{pos: pos}, Key: key, LabelKey: labelKey}
}
func (f *FigureRef) Type() NodeType    { return NodeFigureRef }
func (f *FigureRef) Children() []Node  { return nil }

// FigureImage is an extended image with figure metadata, produced by:
//
//	![F#01:1.0:ht!][Caption text.](path/to/image.png)
//
// It always renders as a LaTeX figure environment with \label, \caption, and
// \includegraphics[width=Width\linewidth].
type FigureImage struct {
	inlineBase
	Key           string // e.g. "F#01"
	LabelKey      string // e.g. "figurelabel01"
	Width         string // linewidth multiplier, e.g. "1.0" or "0.5"
	Placement     string // LaTeX float specifier, e.g. "h", "ht", "ht!"
	CaptionNodes  []Node // parsed inline nodes; may contain refs like \cite{}
	Path          string // image file path
}

func NewFigureImage(pos Pos, key, labelKey, width, placement string, captionNodes []Node, path string) *FigureImage {
	return &FigureImage{
		inlineBase:   inlineBase{pos: pos},
		Key:          key,
		LabelKey:     labelKey,
		Width:        width,
		Placement:    placement,
		CaptionNodes: captionNodes,
		Path:         path,
	}
}
func (f *FigureImage) Type() NodeType    { return NodeFigureImage }
func (f *FigureImage) Children() []Node  { return nil }

// ─── Table-float extension ────────────────────────────────────────────────────

// TableRef is an inline table cross-reference [T#key] that resolves to
// ~\ref{tab:label} in LaTeX.
type TableRef struct {
	inlineBase
	Key      string // e.g. "T#01"
	LabelKey string // e.g. "tablelabel01"
}

func NewTableRef(pos Pos, key, labelKey string) *TableRef {
	return &TableRef{inlineBase: inlineBase{pos: pos}, Key: key, LabelKey: labelKey}
}
func (t *TableRef) Type() NodeType    { return NodeTableRef }
func (t *TableRef) Children() []Node  { return nil }
