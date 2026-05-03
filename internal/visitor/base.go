// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package visitor

import "github.com/andreaswillibaldweber/marktex/internal/ast"

// BaseVisitor provides no-op implementations for every method in the Visitor
// interface. Embed it in your concrete visitor and override only the methods
// you care about:
//
//	type MyVisitor struct {
//	    visitor.BaseVisitor
//	    // your fields
//	}
//
//	func (v *MyVisitor) EnterHeading(n *ast.Heading) visitor.WalkAction {
//	    fmt.Printf("heading level %d\n", n.Level)
//	    return visitor.WalkContinue
//	}
//
// The Go compiler will verify at compile time that MyVisitor satisfies Visitor.
type BaseVisitor struct{}

func (BaseVisitor) EnterDocument(_ *ast.Document) WalkAction               { return WalkContinue }
func (BaseVisitor) LeaveDocument(_ *ast.Document) WalkAction               { return WalkContinue }
func (BaseVisitor) EnterHeading(_ *ast.Heading) WalkAction                 { return WalkContinue }
func (BaseVisitor) LeaveHeading(_ *ast.Heading) WalkAction                 { return WalkContinue }
func (BaseVisitor) EnterParagraph(_ *ast.Paragraph) WalkAction             { return WalkContinue }
func (BaseVisitor) LeaveParagraph(_ *ast.Paragraph) WalkAction             { return WalkContinue }
func (BaseVisitor) EnterBlockQuote(_ *ast.BlockQuote) WalkAction           { return WalkContinue }
func (BaseVisitor) LeaveBlockQuote(_ *ast.BlockQuote) WalkAction           { return WalkContinue }
func (BaseVisitor) EnterBulletList(_ *ast.BulletList) WalkAction           { return WalkContinue }
func (BaseVisitor) LeaveBulletList(_ *ast.BulletList) WalkAction           { return WalkContinue }
func (BaseVisitor) EnterOrderedList(_ *ast.OrderedList) WalkAction         { return WalkContinue }
func (BaseVisitor) LeaveOrderedList(_ *ast.OrderedList) WalkAction         { return WalkContinue }
func (BaseVisitor) EnterListItem(_ *ast.ListItem) WalkAction               { return WalkContinue }
func (BaseVisitor) LeaveListItem(_ *ast.ListItem) WalkAction               { return WalkContinue }
func (BaseVisitor) EnterFencedCode(_ *ast.FencedCode) WalkAction           { return WalkContinue }
func (BaseVisitor) LeaveFencedCode(_ *ast.FencedCode) WalkAction           { return WalkContinue }
func (BaseVisitor) EnterIndentedCode(_ *ast.IndentedCode) WalkAction       { return WalkContinue }
func (BaseVisitor) LeaveIndentedCode(_ *ast.IndentedCode) WalkAction       { return WalkContinue }
func (BaseVisitor) EnterHTMLBlock(_ *ast.HTMLBlock) WalkAction             { return WalkContinue }
func (BaseVisitor) LeaveHTMLBlock(_ *ast.HTMLBlock) WalkAction             { return WalkContinue }
func (BaseVisitor) EnterThematicBreak(_ *ast.ThematicBreak) WalkAction     { return WalkContinue }
func (BaseVisitor) LeaveThematicBreak(_ *ast.ThematicBreak) WalkAction     { return WalkContinue }
func (BaseVisitor) EnterTable(_ *ast.Table) WalkAction                     { return WalkContinue }
func (BaseVisitor) LeaveTable(_ *ast.Table) WalkAction                     { return WalkContinue }
func (BaseVisitor) EnterTableSection(_ *ast.TableSection) WalkAction       { return WalkContinue }
func (BaseVisitor) LeaveTableSection(_ *ast.TableSection) WalkAction       { return WalkContinue }
func (BaseVisitor) EnterTableRow(_ *ast.TableRow) WalkAction               { return WalkContinue }
func (BaseVisitor) LeaveTableRow(_ *ast.TableRow) WalkAction               { return WalkContinue }
func (BaseVisitor) EnterTableCell(_ *ast.TableCell) WalkAction             { return WalkContinue }
func (BaseVisitor) LeaveTableCell(_ *ast.TableCell) WalkAction             { return WalkContinue }
func (BaseVisitor) EnterText(_ *ast.Text) WalkAction                       { return WalkContinue }
func (BaseVisitor) LeaveText(_ *ast.Text) WalkAction                       { return WalkContinue }
func (BaseVisitor) EnterSoftBreak(_ *ast.SoftBreak) WalkAction             { return WalkContinue }
func (BaseVisitor) LeaveSoftBreak(_ *ast.SoftBreak) WalkAction             { return WalkContinue }
func (BaseVisitor) EnterHardBreak(_ *ast.HardBreak) WalkAction             { return WalkContinue }
func (BaseVisitor) LeaveHardBreak(_ *ast.HardBreak) WalkAction             { return WalkContinue }
func (BaseVisitor) EnterEmphasis(_ *ast.Emphasis) WalkAction               { return WalkContinue }
func (BaseVisitor) LeaveEmphasis(_ *ast.Emphasis) WalkAction               { return WalkContinue }
func (BaseVisitor) EnterStrong(_ *ast.Strong) WalkAction                   { return WalkContinue }
func (BaseVisitor) LeaveStrong(_ *ast.Strong) WalkAction                   { return WalkContinue }
func (BaseVisitor) EnterStrongEmphasis(_ *ast.StrongEmphasis) WalkAction   { return WalkContinue }
func (BaseVisitor) LeaveStrongEmphasis(_ *ast.StrongEmphasis) WalkAction   { return WalkContinue }
func (BaseVisitor) EnterCodeSpan(_ *ast.CodeSpan) WalkAction               { return WalkContinue }
func (BaseVisitor) LeaveCodeSpan(_ *ast.CodeSpan) WalkAction               { return WalkContinue }
func (BaseVisitor) EnterLink(_ *ast.Link) WalkAction                       { return WalkContinue }
func (BaseVisitor) LeaveLink(_ *ast.Link) WalkAction                       { return WalkContinue }
func (BaseVisitor) EnterImage(_ *ast.Image) WalkAction                     { return WalkContinue }
func (BaseVisitor) LeaveImage(_ *ast.Image) WalkAction                     { return WalkContinue }
func (BaseVisitor) EnterRawHTML(_ *ast.RawHTML) WalkAction                 { return WalkContinue }
func (BaseVisitor) LeaveRawHTML(_ *ast.RawHTML) WalkAction                 { return WalkContinue }
func (BaseVisitor) EnterEscape(_ *ast.Escape) WalkAction                   { return WalkContinue }
func (BaseVisitor) LeaveEscape(_ *ast.Escape) WalkAction                   { return WalkContinue }
func (BaseVisitor) EnterEntity(_ *ast.Entity) WalkAction                   { return WalkContinue }
func (BaseVisitor) LeaveEntity(_ *ast.Entity) WalkAction                   { return WalkContinue }
func (BaseVisitor) EnterStrikethrough(_ *ast.Strikethrough) WalkAction     { return WalkContinue }
func (BaseVisitor) LeaveStrikethrough(_ *ast.Strikethrough) WalkAction     { return WalkContinue }
func (BaseVisitor) EnterCitationBlock(_ *ast.CitationBlock) WalkAction     { return WalkContinue }
func (BaseVisitor) LeaveCitationBlock(_ *ast.CitationBlock) WalkAction     { return WalkContinue }
func (BaseVisitor) EnterCitationRef(_ *ast.CitationRef) WalkAction         { return WalkContinue }
func (BaseVisitor) LeaveCitationRef(_ *ast.CitationRef) WalkAction         { return WalkContinue }
func (BaseVisitor) EnterFigureBlock(_ *ast.FigureBlock) WalkAction         { return WalkContinue }
func (BaseVisitor) LeaveFigureBlock(_ *ast.FigureBlock) WalkAction         { return WalkContinue }
func (BaseVisitor) EnterFigureRef(_ *ast.FigureRef) WalkAction             { return WalkContinue }
func (BaseVisitor) LeaveFigureRef(_ *ast.FigureRef) WalkAction             { return WalkContinue }
func (BaseVisitor) EnterFigureImage(_ *ast.FigureImage) WalkAction         { return WalkContinue }
func (BaseVisitor) LeaveFigureImage(_ *ast.FigureImage) WalkAction         { return WalkContinue }
func (BaseVisitor) EnterTableDef(_ *ast.TableDef) WalkAction               { return WalkContinue }
func (BaseVisitor) LeaveTableDef(_ *ast.TableDef) WalkAction               { return WalkContinue }
func (BaseVisitor) EnterTableBlock(_ *ast.TableBlock) WalkAction           { return WalkContinue }
func (BaseVisitor) LeaveTableBlock(_ *ast.TableBlock) WalkAction           { return WalkContinue }
func (BaseVisitor) EnterTableRef(_ *ast.TableRef) WalkAction               { return WalkContinue }
func (BaseVisitor) LeaveTableRef(_ *ast.TableRef) WalkAction               { return WalkContinue }
func (BaseVisitor) EnterDefinitionBlock(_ *ast.DefinitionBlock) WalkAction { return WalkContinue }
func (BaseVisitor) LeaveDefinitionBlock(_ *ast.DefinitionBlock) WalkAction { return WalkContinue }
