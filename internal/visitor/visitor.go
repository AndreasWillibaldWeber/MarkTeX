// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

// Package visitor defines the Visitor interface used to traverse an AST produced
// by the parser. The Walk function in the ast package drives traversal and calls
// the appropriate Enter*/Leave* method for each node.
//
// Adding a new node type requires:
//  1. Adding the struct to the ast package.
//  2. Adding Enter/Leave methods here.
//  3. Adding no-op implementations to BaseVisitor.
//  4. Handling the new case in ast.Walk.
//  5. Optionally implementing the methods in any concrete visitor (e.g. the
//     LaTeX generator) that cares about the new node type.
package visitor

import "marktex/internal/ast"

// WalkAction controls traversal after an Enter or Leave method returns.
type WalkAction int

const (
	// WalkContinue recurses into children (on Enter) or continues (on Leave).
	WalkContinue WalkAction = iota

	// WalkSkipChildren skips the children of the current node but still calls
	// the Leave method. Only meaningful when returned from an Enter method.
	WalkSkipChildren

	// WalkStop aborts the entire traversal immediately.
	WalkStop
)

// Visitor is implemented by any type that wants to process an AST.
// An Enter method is called before the children are visited; the corresponding
// Leave method is called after. Returning WalkStop from any method aborts the
// walk immediately. Returning WalkSkipChildren from an Enter method skips the
// node's children but still calls the Leave method.
//
// Leaf nodes (Text, SoftBreak, HardBreak, CodeSpan, RawHTML, Escape, Entity,
// ThematicBreak, FencedCode, IndentedCode, HTMLBlock) have no meaningful Leave
// call; the BaseVisitor provides no-op Leave implementations for consistency.
type Visitor interface {
	// Block nodes
	EnterDocument(n *ast.Document) WalkAction
	LeaveDocument(n *ast.Document) WalkAction
	EnterHeading(n *ast.Heading) WalkAction
	LeaveHeading(n *ast.Heading) WalkAction
	EnterParagraph(n *ast.Paragraph) WalkAction
	LeaveParagraph(n *ast.Paragraph) WalkAction
	EnterBlockQuote(n *ast.BlockQuote) WalkAction
	LeaveBlockQuote(n *ast.BlockQuote) WalkAction
	EnterBulletList(n *ast.BulletList) WalkAction
	LeaveBulletList(n *ast.BulletList) WalkAction
	EnterOrderedList(n *ast.OrderedList) WalkAction
	LeaveOrderedList(n *ast.OrderedList) WalkAction
	EnterListItem(n *ast.ListItem) WalkAction
	LeaveListItem(n *ast.ListItem) WalkAction
	EnterFencedCode(n *ast.FencedCode) WalkAction
	LeaveFencedCode(n *ast.FencedCode) WalkAction
	EnterIndentedCode(n *ast.IndentedCode) WalkAction
	LeaveIndentedCode(n *ast.IndentedCode) WalkAction
	EnterHTMLBlock(n *ast.HTMLBlock) WalkAction
	LeaveHTMLBlock(n *ast.HTMLBlock) WalkAction
	EnterThematicBreak(n *ast.ThematicBreak) WalkAction
	LeaveThematicBreak(n *ast.ThematicBreak) WalkAction
	EnterTable(n *ast.Table) WalkAction
	LeaveTable(n *ast.Table) WalkAction
	EnterTableSection(n *ast.TableSection) WalkAction
	LeaveTableSection(n *ast.TableSection) WalkAction
	EnterTableRow(n *ast.TableRow) WalkAction
	LeaveTableRow(n *ast.TableRow) WalkAction
	EnterTableCell(n *ast.TableCell) WalkAction
	LeaveTableCell(n *ast.TableCell) WalkAction

	// Inline nodes
	EnterText(n *ast.Text) WalkAction
	LeaveText(n *ast.Text) WalkAction
	EnterSoftBreak(n *ast.SoftBreak) WalkAction
	LeaveSoftBreak(n *ast.SoftBreak) WalkAction
	EnterHardBreak(n *ast.HardBreak) WalkAction
	LeaveHardBreak(n *ast.HardBreak) WalkAction
	EnterEmphasis(n *ast.Emphasis) WalkAction
	LeaveEmphasis(n *ast.Emphasis) WalkAction
	EnterStrong(n *ast.Strong) WalkAction
	LeaveStrong(n *ast.Strong) WalkAction
	EnterStrongEmphasis(n *ast.StrongEmphasis) WalkAction
	LeaveStrongEmphasis(n *ast.StrongEmphasis) WalkAction
	EnterCodeSpan(n *ast.CodeSpan) WalkAction
	LeaveCodeSpan(n *ast.CodeSpan) WalkAction
	EnterLink(n *ast.Link) WalkAction
	LeaveLink(n *ast.Link) WalkAction
	EnterImage(n *ast.Image) WalkAction
	LeaveImage(n *ast.Image) WalkAction
	EnterRawHTML(n *ast.RawHTML) WalkAction
	LeaveRawHTML(n *ast.RawHTML) WalkAction
	EnterEscape(n *ast.Escape) WalkAction
	LeaveEscape(n *ast.Escape) WalkAction
	EnterEntity(n *ast.Entity) WalkAction
	LeaveEntity(n *ast.Entity) WalkAction
	EnterStrikethrough(n *ast.Strikethrough) WalkAction
	LeaveStrikethrough(n *ast.Strikethrough) WalkAction

	// Citation extension
	EnterCitationBlock(n *ast.CitationBlock) WalkAction
	LeaveCitationBlock(n *ast.CitationBlock) WalkAction
	EnterCitationRef(n *ast.CitationRef) WalkAction
	LeaveCitationRef(n *ast.CitationRef) WalkAction

	// Figure extension
	EnterFigureBlock(n *ast.FigureBlock) WalkAction
	LeaveFigureBlock(n *ast.FigureBlock) WalkAction
	EnterFigureRef(n *ast.FigureRef) WalkAction
	LeaveFigureRef(n *ast.FigureRef) WalkAction
	EnterFigureImage(n *ast.FigureImage) WalkAction
	LeaveFigureImage(n *ast.FigureImage) WalkAction

	// Table-float extension
	EnterTableDef(n *ast.TableDef) WalkAction
	LeaveTableDef(n *ast.TableDef) WalkAction
	EnterTableBlock(n *ast.TableBlock) WalkAction
	LeaveTableBlock(n *ast.TableBlock) WalkAction
	EnterTableRef(n *ast.TableRef) WalkAction
	LeaveTableRef(n *ast.TableRef) WalkAction
}
