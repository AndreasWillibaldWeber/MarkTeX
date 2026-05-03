// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package ast

// DocumentMeta holds document-level metadata extracted from a definition block.
// It is only applied when the generator is run in standalone mode; in fragment
// mode every field is silently ignored.
type DocumentMeta struct {
	Author   string // \author{…}
	Title    string // \title{…}
	Subtitle string // appended to \title as \\[0.5em]{\large …}
	Date     string // \date{…}  — empty means LaTeX uses \today

	MakeTitlePage bool // emit \maketitle after \begin{document}
	MakeTOC       bool // emit \tableofcontents
	MakeLOF       bool // emit \listoffigures
	MakeLOT       bool // emit \listoftables
	MakeLOL       bool // emit \lstlistoflistings (requires listings package)
}

// Merge copies non-zero fields from src into dst. Boolean fields are OR-ed;
// string fields from src overwrite dst only when non-empty.
func (dst *DocumentMeta) Merge(src DocumentMeta) {
	if src.Author != "" {
		dst.Author = src.Author
	}
	if src.Title != "" {
		dst.Title = src.Title
	}
	if src.Subtitle != "" {
		dst.Subtitle = src.Subtitle
	}
	if src.Date != "" {
		dst.Date = src.Date
	}
	dst.MakeTitlePage = dst.MakeTitlePage || src.MakeTitlePage
	dst.MakeTOC = dst.MakeTOC || src.MakeTOC
	dst.MakeLOF = dst.MakeLOF || src.MakeLOF
	dst.MakeLOT = dst.MakeLOT || src.MakeLOT
	dst.MakeLOL = dst.MakeLOL || src.MakeLOL
}

// Align represents column alignment in a table.
type Align int

const (
	AlignNone   Align = iota
	AlignLeft         // `:---`
	AlignCenter       // `:---:`
	AlignRight        // `---:`
)

// String returns the LaTeX column specifier character for this alignment.
func (a Align) String() string {
	switch a {
	case AlignLeft:
		return "l"
	case AlignCenter:
		return "c"
	case AlignRight:
		return "r"
	default:
		return "l"
	}
}

// ─── Block nodes ─────────────────────────────────────────────────────────────

// Document is the root node of every parsed tree.
type Document struct{ blockBase }

func NewDocument(pos Pos) *Document { return &Document{blockBase{pos: pos}} }
func (d *Document) Type() NodeType  { return NodeDocument }

// Heading represents ATX (`# Heading`) and setext headings.
// Level is in the range [1, 6].
type Heading struct {
	blockBase
	Level int
}

func NewHeading(pos Pos, level int) *Heading { return &Heading{blockBase{pos: pos}, level} }
func (h *Heading) Type() NodeType            { return NodeHeading }

// Paragraph holds inline content.
type Paragraph struct{ blockBase }

func NewParagraph(pos Pos) *Paragraph { return &Paragraph{blockBase{pos: pos}} }
func (p *Paragraph) Type() NodeType   { return NodeParagraph }

// BlockQuote wraps block children indented by `> `.
type BlockQuote struct{ blockBase }

func NewBlockQuote(pos Pos) *BlockQuote { return &BlockQuote{blockBase{pos: pos}} }
func (b *BlockQuote) Type() NodeType    { return NodeBlockQuote }

// BulletList is an unordered list. Tight is true when no blank lines
// separate list items (CommonMark tight-list semantics).
type BulletList struct {
	blockBase
	Tight  bool
	Marker rune // one of '-', '*', '+'
}

func NewBulletList(pos Pos, marker rune) *BulletList {
	return &BulletList{blockBase: blockBase{pos: pos}, Marker: marker}
}
func (b *BulletList) Type() NodeType { return NodeBulletList }

// OrderedList is a numbered list. Start is the first item number (usually 1).
// Delimiter is '.' or ')'.
type OrderedList struct {
	blockBase
	Tight     bool
	Start     int  // starting number
	Delimiter rune // '.' or ')'
}

func NewOrderedList(pos Pos, start int, delimiter rune) *OrderedList {
	return &OrderedList{blockBase: blockBase{pos: pos}, Start: start, Delimiter: delimiter}
}
func (o *OrderedList) Type() NodeType { return NodeOrderedList }

// ListItem is a single item within a BulletList or OrderedList.
// Its children may be inline nodes (tight list) or block nodes (loose list).
type ListItem struct{ blockBase }

func NewListItem(pos Pos) *ListItem { return &ListItem{blockBase{pos: pos}} }
func (l *ListItem) Type() NodeType  { return NodeListItem }

// FencedCode is a code block delimited by ``` or ~~~.
// Info contains the raw info string (e.g. "go" or "python title").
// The first word of Info is conventionally the language identifier.
type FencedCode struct {
	blockBase
	Info    string
	Literal string
}

func NewFencedCode(pos Pos, info, literal string) *FencedCode {
	return &FencedCode{blockBase: blockBase{pos: pos}, Info: info, Literal: literal}
}
func (f *FencedCode) Type() NodeType { return NodeFencedCode }

// Language extracts the language identifier from the Info string.
func (f *FencedCode) Language() string {
	for i, r := range f.Info {
		if r == ' ' || r == '\t' {
			return f.Info[:i]
		}
	}
	return f.Info
}

// IndentedCode is a code block introduced by four-space (or one-tab) indentation.
type IndentedCode struct {
	blockBase
	Literal string
}

func NewIndentedCode(pos Pos, literal string) *IndentedCode {
	return &IndentedCode{blockBase: blockBase{pos: pos}, Literal: literal}
}
func (i *IndentedCode) Type() NodeType { return NodeIndentedCode }

// HTMLBlock holds raw HTML that appears at the block level.
// The generator typically emits a comment and optionally a warning.
type HTMLBlock struct {
	blockBase
	Literal string
}

func NewHTMLBlock(pos Pos, literal string) *HTMLBlock {
	return &HTMLBlock{blockBase: blockBase{pos: pos}, Literal: literal}
}
func (h *HTMLBlock) Type() NodeType { return NodeHTMLBlock }

// ThematicBreak represents `---`, `***`, or `___`.
type ThematicBreak struct{ blockBase }

func NewThematicBreak(pos Pos) *ThematicBreak { return &ThematicBreak{blockBase{pos: pos}} }
func (t *ThematicBreak) Type() NodeType       { return NodeThematicBreak }

// Table holds a GFM pipe table. Alignments describes each column's alignment
// in order; its length equals the number of columns.
type Table struct {
	blockBase
	Alignments []Align
}

func NewTable(pos Pos, aligns []Align) *Table {
	return &Table{blockBase: blockBase{pos: pos}, Alignments: aligns}
}
func (t *Table) Type() NodeType { return NodeTable }

// TableSection groups header or body rows. IsHeader is true for the header section.
type TableSection struct {
	blockBase
	IsHeader bool
}

func NewTableSection(pos Pos, isHeader bool) *TableSection {
	return &TableSection{blockBase: blockBase{pos: pos}, IsHeader: isHeader}
}
func (ts *TableSection) Type() NodeType { return NodeTableSection }

// TableRow is a single `| cell | cell |` row.
type TableRow struct{ blockBase }

func NewTableRow(pos Pos) *TableRow { return &TableRow{blockBase{pos: pos}} }
func (tr *TableRow) Type() NodeType { return NodeTableRow }

// TableCell holds the inline content of a single table cell.
// Align overrides the column-level alignment if set to a non-None value.
type TableCell struct {
	blockBase
	Align Align
}

func NewTableCell(pos Pos, align Align) *TableCell {
	return &TableCell{blockBase: blockBase{pos: pos}, Align: align}
}
func (tc *TableCell) Type() NodeType { return NodeTableCell }

// ─── Citation extension ───────────────────────────────────────────────────────

// CitationBlock holds the key→bibtex mapping from a citation definition block:
//
//	---
//	C#01:doe2023survey
//	C#02:smith2024ml
//	---
//
// It produces no output in the LaTeX document itself; its purpose is to
// populate the citation registry used by CitationRef nodes.
type CitationBlock struct {
	blockBase
	// Refs maps each short citation key (e.g. "C#01") to its BibTeX identifier.
	Refs map[string]string
}

func NewCitationBlock(pos Pos, refs map[string]string) *CitationBlock {
	return &CitationBlock{blockBase: blockBase{pos: pos}, Refs: refs}
}
func (c *CitationBlock) Type() NodeType { return NodeCitationBlock }

// FigureBlock holds the key→label mapping from a figure definition block:
//
//	---
//	F#01:figurelabel01
//	F#02:figurelabel02
//	---
//
// It produces no output; its purpose is to populate the figure registry used
// by FigureRef and FigureImage nodes.
type FigureBlock struct {
	blockBase
	// Refs maps each short figure key (e.g. "F#01") to its LaTeX label suffix
	// (e.g. "figurelabel01" → \label{fig:figurelabel01}).
	Refs map[string]string
}

func NewFigureBlock(pos Pos, refs map[string]string) *FigureBlock {
	return &FigureBlock{blockBase: blockBase{pos: pos}, Refs: refs}
}
func (f *FigureBlock) Type() NodeType { return NodeFigureBlock }

// ─── Table-float extension ────────────────────────────────────────────────────

// TableBlock wraps a GFM Table inside a LaTeX table-float environment.
// It is produced when the table is preceded by |- ... -| metadata rows:
//
//	|- T#01:ht -|
//	|- This is the caption. -|
//	| col | col |
//	| --- | --- |
//
// Children must contain exactly one Table node.
type TableBlock struct {
	blockBase
	Key          string // short key, e.g. "T#01"
	LabelKey     string // resolved LaTeX label suffix, e.g. "tablelabel01"
	Placement    string // LaTeX float specifier, e.g. "h!", "ht"
	CaptionNodes []Node // parsed inline nodes; empty when no |- caption -| row given
}

func NewTableBlock(pos Pos, key, labelKey, placement string, captionNodes []Node) *TableBlock {
	return &TableBlock{
		blockBase:    blockBase{pos: pos},
		Key:          key,
		LabelKey:     labelKey,
		Placement:    placement,
		CaptionNodes: captionNodes,
	}
}
func (t *TableBlock) Type() NodeType { return NodeTableBlock }

// TableDef holds the T# key→label mapping from a table definition block.
// Like CitationBlock and FigureBlock, it produces no output.
type TableDef struct {
	blockBase
	Refs map[string]string // T#01 → tablelabel01
}

func NewTableDef(pos Pos, refs map[string]string) *TableDef {
	return &TableDef{blockBase: blockBase{pos: pos}, Refs: refs}
}
func (t *TableDef) Type() NodeType { return NodeTableDef }

// DefinitionBlock is the unified form produced when a ---…--- block contains
// any mix of C#, F#, and T# entries. It replaces the separate CitationBlock,
// FigureBlock, and TableDef nodes for mixed blocks. It produces no output.
type DefinitionBlock struct {
	blockBase
	Citations map[string]string // C# key → BibTeX identifier
	Figures   map[string]string // F# key → label suffix
	Tables    map[string]string // T# key → label suffix
	Meta      DocumentMeta      // document metadata (standalone mode only)
}

func NewDefinitionBlock(pos Pos, citations, figures, tables map[string]string, meta DocumentMeta) *DefinitionBlock {
	return &DefinitionBlock{
		blockBase: blockBase{pos: pos},
		Citations: citations,
		Figures:   figures,
		Tables:    tables,
		Meta:      meta,
	}
}
func (d *DefinitionBlock) Type() NodeType { return NodeDefinitionBlock }
