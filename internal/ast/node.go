// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

// Package ast defines the abstract syntax tree produced by the Markdown parser.
// Every node carries source position information and satisfies the Node interface.
// The tree is then consumed by a Visitor (see the visitor package) to produce output.
package ast

import "fmt"

// NodeType identifies the concrete type of an AST node. Using an integer constant
// enables exhaustive type switches and fast dispatch without reflection.
type NodeType int

//go:generate stringer -type=NodeType
const (
	// Block-level nodes
	NodeDocument NodeType = iota
	NodeHeading
	NodeParagraph
	NodeBlockQuote
	NodeBulletList
	NodeOrderedList
	NodeListItem
	NodeFencedCode
	NodeIndentedCode
	NodeHTMLBlock
	NodeThematicBreak
	NodeTable
	NodeTableSection
	NodeTableRow
	NodeTableCell

	// Inline nodes
	NodeText
	NodeSoftBreak
	NodeHardBreak
	NodeEmphasis
	NodeStrong
	NodeStrongEmphasis
	NodeCodeSpan
	NodeLink
	NodeImage
	NodeRawHTML
	NodeEscape
	NodeEntity
	NodeStrikethrough

	// Citation extension nodes
	NodeCitationBlock // block: defines key→bibtex mappings (single-type, kept for compat)
	NodeCitationRef   // inline: [C#key] resolved to \cite{bibtex}

	// Figure extension nodes
	NodeFigureBlock // block: defines key→label mappings (single-type, kept for compat)
	NodeFigureRef   // inline: [F#key] resolved to \ref{fig:label}
	NodeFigureImage // inline: |- F#key:w:pos -| / |- caption -| / ![](path)

	// Table-float extension nodes
	NodeTableDef   // block: defines T# key→label mappings (single-type, kept for compat)
	NodeTableBlock // block: wraps a Table with float metadata (placement, caption, label)
	NodeTableRef   // inline: [T#key] resolved to \ref{tab:label}

	// Unified definition block — handles any mix of C#, F#, T# entries in one ---…--- block.
	NodeDefinitionBlock
)

// Pos records where a node originated in the source text.
// Line and Col are 1-based; Offset is the byte offset from the start of input.
type Pos struct {
	Offset int
	Line   int
	Col    int
}

// Node is the common interface for every element in the AST.
// Block and inline nodes both satisfy this interface.
type Node interface {
	// Type returns the NodeType constant for this node.
	Type() NodeType

	// Children returns the direct child nodes. Leaf nodes return nil.
	Children() []Node

	// Position returns the source location where this node starts.
	Position() Pos
}

// Adder is implemented by all block and inline container nodes.
// Use AppendChild for the ergonomic wrapper.
type Adder interface {
	Node
	AddChild(Node)
}

// AppendChild adds child to parent. It panics if parent does not implement
// Adder (i.e. is a leaf node that cannot have children).
func AppendChild(parent Node, child Node) {
	a, ok := parent.(Adder)
	if !ok {
		panic(fmt.Sprintf("ast.AppendChild: %T does not accept children", parent))
	}
	a.AddChild(child)
}

// ─── blockBase ────────────────────────────────────────────────────────────────

// blockBase is embedded by all block-level nodes to provide the child slice
// and position without repeating those fields everywhere.
type blockBase struct {
	pos      Pos
	children []Node
}

func (b *blockBase) Position() Pos    { return b.pos }
func (b *blockBase) Children() []Node { return b.children }

// AddChild appends a child node. Implements Adder.
func (b *blockBase) AddChild(n Node) { b.children = append(b.children, n) }

// ─── inlineBase ───────────────────────────────────────────────────────────────

// inlineBase is embedded by all inline container nodes.
type inlineBase struct {
	pos      Pos
	children []Node
}

func (i *inlineBase) Position() Pos    { return i.pos }
func (i *inlineBase) Children() []Node { return i.children }

// AddChild appends a child node. Implements Adder.
func (i *inlineBase) AddChild(n Node) { i.children = append(i.children, n) }
