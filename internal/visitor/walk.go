// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package visitor

import "marktex/internal/ast"

// Walk traverses the AST rooted at n in depth-first order, calling the
// appropriate Enter and Leave methods on v for each node.
//
// Walk returns WalkStop if traversal was aborted by a visitor method; it
// returns WalkContinue otherwise. Callers can generally ignore the return value
// unless they are themselves inside a larger Walk.
func Walk(n ast.Node, v Visitor) WalkAction {
	return walk(n, v)
}

func walk(n ast.Node, v Visitor) WalkAction {
	action := enter(n, v)
	if action == WalkStop {
		return WalkStop
	}
	if action != WalkSkipChildren {
		for _, child := range n.Children() {
			if walk(child, v) == WalkStop {
				return WalkStop
			}
		}
	}
	return leave(n, v)
}

// enter dispatches to the correct EnterXxx method.
func enter(n ast.Node, v Visitor) WalkAction {
	switch n.Type() {
	case ast.NodeDocument:
		return v.EnterDocument(n.(*ast.Document))
	case ast.NodeHeading:
		return v.EnterHeading(n.(*ast.Heading))
	case ast.NodeParagraph:
		return v.EnterParagraph(n.(*ast.Paragraph))
	case ast.NodeBlockQuote:
		return v.EnterBlockQuote(n.(*ast.BlockQuote))
	case ast.NodeBulletList:
		return v.EnterBulletList(n.(*ast.BulletList))
	case ast.NodeOrderedList:
		return v.EnterOrderedList(n.(*ast.OrderedList))
	case ast.NodeListItem:
		return v.EnterListItem(n.(*ast.ListItem))
	case ast.NodeFencedCode:
		return v.EnterFencedCode(n.(*ast.FencedCode))
	case ast.NodeIndentedCode:
		return v.EnterIndentedCode(n.(*ast.IndentedCode))
	case ast.NodeHTMLBlock:
		return v.EnterHTMLBlock(n.(*ast.HTMLBlock))
	case ast.NodeThematicBreak:
		return v.EnterThematicBreak(n.(*ast.ThematicBreak))
	case ast.NodeTable:
		return v.EnterTable(n.(*ast.Table))
	case ast.NodeTableSection:
		return v.EnterTableSection(n.(*ast.TableSection))
	case ast.NodeTableRow:
		return v.EnterTableRow(n.(*ast.TableRow))
	case ast.NodeTableCell:
		return v.EnterTableCell(n.(*ast.TableCell))
	case ast.NodeText:
		return v.EnterText(n.(*ast.Text))
	case ast.NodeSoftBreak:
		return v.EnterSoftBreak(n.(*ast.SoftBreak))
	case ast.NodeHardBreak:
		return v.EnterHardBreak(n.(*ast.HardBreak))
	case ast.NodeEmphasis:
		return v.EnterEmphasis(n.(*ast.Emphasis))
	case ast.NodeStrong:
		return v.EnterStrong(n.(*ast.Strong))
	case ast.NodeStrongEmphasis:
		return v.EnterStrongEmphasis(n.(*ast.StrongEmphasis))
	case ast.NodeCodeSpan:
		return v.EnterCodeSpan(n.(*ast.CodeSpan))
	case ast.NodeLink:
		return v.EnterLink(n.(*ast.Link))
	case ast.NodeImage:
		return v.EnterImage(n.(*ast.Image))
	case ast.NodeRawHTML:
		return v.EnterRawHTML(n.(*ast.RawHTML))
	case ast.NodeEscape:
		return v.EnterEscape(n.(*ast.Escape))
	case ast.NodeEntity:
		return v.EnterEntity(n.(*ast.Entity))
	case ast.NodeStrikethrough:
		return v.EnterStrikethrough(n.(*ast.Strikethrough))
	case ast.NodeCitationBlock:
		return v.EnterCitationBlock(n.(*ast.CitationBlock))
	case ast.NodeCitationRef:
		return v.EnterCitationRef(n.(*ast.CitationRef))
	case ast.NodeFigureBlock:
		return v.EnterFigureBlock(n.(*ast.FigureBlock))
	case ast.NodeFigureRef:
		return v.EnterFigureRef(n.(*ast.FigureRef))
	case ast.NodeFigureImage:
		return v.EnterFigureImage(n.(*ast.FigureImage))
	case ast.NodeTableDef:
		return v.EnterTableDef(n.(*ast.TableDef))
	case ast.NodeTableBlock:
		return v.EnterTableBlock(n.(*ast.TableBlock))
	case ast.NodeTableRef:
		return v.EnterTableRef(n.(*ast.TableRef))
	}
	return WalkContinue
}

// leave dispatches to the correct LeaveXxx method.
func leave(n ast.Node, v Visitor) WalkAction {
	switch n.Type() {
	case ast.NodeDocument:
		return v.LeaveDocument(n.(*ast.Document))
	case ast.NodeHeading:
		return v.LeaveHeading(n.(*ast.Heading))
	case ast.NodeParagraph:
		return v.LeaveParagraph(n.(*ast.Paragraph))
	case ast.NodeBlockQuote:
		return v.LeaveBlockQuote(n.(*ast.BlockQuote))
	case ast.NodeBulletList:
		return v.LeaveBulletList(n.(*ast.BulletList))
	case ast.NodeOrderedList:
		return v.LeaveOrderedList(n.(*ast.OrderedList))
	case ast.NodeListItem:
		return v.LeaveListItem(n.(*ast.ListItem))
	case ast.NodeFencedCode:
		return v.LeaveFencedCode(n.(*ast.FencedCode))
	case ast.NodeIndentedCode:
		return v.LeaveIndentedCode(n.(*ast.IndentedCode))
	case ast.NodeHTMLBlock:
		return v.LeaveHTMLBlock(n.(*ast.HTMLBlock))
	case ast.NodeThematicBreak:
		return v.LeaveThematicBreak(n.(*ast.ThematicBreak))
	case ast.NodeTable:
		return v.LeaveTable(n.(*ast.Table))
	case ast.NodeTableSection:
		return v.LeaveTableSection(n.(*ast.TableSection))
	case ast.NodeTableRow:
		return v.LeaveTableRow(n.(*ast.TableRow))
	case ast.NodeTableCell:
		return v.LeaveTableCell(n.(*ast.TableCell))
	case ast.NodeText:
		return v.LeaveText(n.(*ast.Text))
	case ast.NodeSoftBreak:
		return v.LeaveSoftBreak(n.(*ast.SoftBreak))
	case ast.NodeHardBreak:
		return v.LeaveHardBreak(n.(*ast.HardBreak))
	case ast.NodeEmphasis:
		return v.LeaveEmphasis(n.(*ast.Emphasis))
	case ast.NodeStrong:
		return v.LeaveStrong(n.(*ast.Strong))
	case ast.NodeStrongEmphasis:
		return v.LeaveStrongEmphasis(n.(*ast.StrongEmphasis))
	case ast.NodeCodeSpan:
		return v.LeaveCodeSpan(n.(*ast.CodeSpan))
	case ast.NodeLink:
		return v.LeaveLink(n.(*ast.Link))
	case ast.NodeImage:
		return v.LeaveImage(n.(*ast.Image))
	case ast.NodeRawHTML:
		return v.LeaveRawHTML(n.(*ast.RawHTML))
	case ast.NodeEscape:
		return v.LeaveEscape(n.(*ast.Escape))
	case ast.NodeEntity:
		return v.LeaveEntity(n.(*ast.Entity))
	case ast.NodeStrikethrough:
		return v.LeaveStrikethrough(n.(*ast.Strikethrough))
	case ast.NodeCitationBlock:
		return v.LeaveCitationBlock(n.(*ast.CitationBlock))
	case ast.NodeCitationRef:
		return v.LeaveCitationRef(n.(*ast.CitationRef))
	case ast.NodeFigureBlock:
		return v.LeaveFigureBlock(n.(*ast.FigureBlock))
	case ast.NodeFigureRef:
		return v.LeaveFigureRef(n.(*ast.FigureRef))
	case ast.NodeFigureImage:
		return v.LeaveFigureImage(n.(*ast.FigureImage))
	case ast.NodeTableDef:
		return v.LeaveTableDef(n.(*ast.TableDef))
	case ast.NodeTableBlock:
		return v.LeaveTableBlock(n.(*ast.TableBlock))
	case ast.NodeTableRef:
		return v.LeaveTableRef(n.(*ast.TableRef))
	}
	return WalkContinue
}
