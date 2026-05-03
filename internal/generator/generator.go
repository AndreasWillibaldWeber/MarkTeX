// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package generator

import (
	"fmt"
	"io"
	"marktex/internal/ast"
	"marktex/internal/visitor"
	"strings"
)

// Options configures LaTeX output.
type Options struct {
	// Standalone wraps the output in a complete LaTeX document
	// (\documentclass … \end{document}).
	Standalone bool

	// Strict causes the generator to return an error when it encounters a node
	// type it cannot translate (e.g. raw HTML blocks). In non-strict mode a
	// LaTeX comment is emitted instead.
	Strict bool
}

// Generator walks an AST and produces LaTeX output.
// It satisfies the visitor.Visitor interface via embedded BaseVisitor.
type Generator struct {
	visitor.BaseVisitor
	opts Options
	buf  strings.Builder

	// package-use flags, set as nodes are encountered
	po preambleOpts

	// tableAlignments holds the column alignments for the table currently
	// being rendered. Set in EnterTable, used in LeaveTableSection.
	tableAlignments []ast.Align

	// inTableHeader is true while we are inside the table header section
	inTableHeader bool

	// tableColCount tracks cells in the current row (for separator lines)
	tableColCount int

	// inTableFloat suppresses the trailing \n\n that LeaveTable normally emits
	// so that the TableBlock handler can place \label before \end{table}.
	inTableFloat bool

	// listStack tracks the nesting depth of lists for indentation
	listStack []listFrame
}

type listFrame struct {
	ordered bool
	tight   bool
}

// New creates a Generator with the given options.
func New(opts Options) *Generator {
	return &Generator{opts: opts}
}

// Generate walks the document AST and writes LaTeX to w.
func Generate(doc *ast.Document, opts Options, w io.Writer) error {
	g := New(opts)
	if action := visitor.Walk(doc, g); action == visitor.WalkStop {
		// WalkStop is not used to signal errors in this generator, but keep
		// the check for future use.
	}
	_, err := io.WriteString(w, g.buf.String())
	return err
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// renderInlineNodes walks a slice of inline AST nodes through the generator,
// emitting their LaTeX directly into the output buffer. Used for captions so
// that inline refs ([C#01], [F#01], [T#01]) resolve correctly inside \caption{}.
func (g *Generator) renderInlineNodes(nodes []ast.Node) {
	for _, n := range nodes {
		visitor.Walk(n, g)
	}
}

func (g *Generator) write(s string)         { g.buf.WriteString(s) }
func (g *Generator) writef(f string, a ...interface{}) { fmt.Fprintf(&g.buf, f, a...) }
func (g *Generator) nl()                    { g.buf.WriteByte('\n') }

// ─── Document ─────────────────────────────────────────────────────────────────

func (g *Generator) EnterDocument(_ *ast.Document) visitor.WalkAction {
	// Preamble is written in LeaveDocument after we know which packages are
	// needed. We buffer body content and prepend the preamble at the end.
	return visitor.WalkContinue
}

func (g *Generator) LeaveDocument(_ *ast.Document) visitor.WalkAction {
	if g.opts.Standalone {
		body := g.buf.String()
		g.buf.Reset()
		g.write(standalonePreamble(g.po))
		g.write(body)
		g.write(standalonePostamble)
	}
	return visitor.WalkContinue
}

// ─── Headings ─────────────────────────────────────────────────────────────────

var headingCommands = [7]string{
	0:  "",
	1:  `\section`,
	2:  `\subsection`,
	3:  `\subsubsection`,
	4:  `\paragraph`,
	5:  `\subparagraph`,
	6:  `\textbf`, // no standard level-6 in LaTeX
}

func (g *Generator) EnterHeading(n *ast.Heading) visitor.WalkAction {
	cmd := headingCommands[n.Level]
	g.writef("%s{", cmd)
	return visitor.WalkContinue
}

func (g *Generator) LeaveHeading(_ *ast.Heading) visitor.WalkAction {
	g.write("}\n\n")
	return visitor.WalkContinue
}

// ─── Paragraph ────────────────────────────────────────────────────────────────

func (g *Generator) EnterParagraph(_ *ast.Paragraph) visitor.WalkAction {
	return visitor.WalkContinue
}

func (g *Generator) LeaveParagraph(_ *ast.Paragraph) visitor.WalkAction {
	g.write("\n\n")
	return visitor.WalkContinue
}

// ─── BlockQuote ───────────────────────────────────────────────────────────────

func (g *Generator) EnterBlockQuote(_ *ast.BlockQuote) visitor.WalkAction {
	g.write("\\begin{quote}\n")
	return visitor.WalkContinue
}

func (g *Generator) LeaveBlockQuote(_ *ast.BlockQuote) visitor.WalkAction {
	g.write("\\end{quote}\n\n")
	return visitor.WalkContinue
}

// ─── Lists ────────────────────────────────────────────────────────────────────

func (g *Generator) EnterBulletList(n *ast.BulletList) visitor.WalkAction {
	g.po.needsEnumitem = true
	g.listStack = append(g.listStack, listFrame{ordered: false, tight: n.Tight})
	if n.Tight {
		g.write("\\begin{itemize}[noitemsep]\n")
	} else {
		g.write("\\begin{itemize}\n")
	}
	return visitor.WalkContinue
}

func (g *Generator) LeaveBulletList(_ *ast.BulletList) visitor.WalkAction {
	g.listStack = g.listStack[:len(g.listStack)-1]
	g.write("\\end{itemize}\n\n")
	return visitor.WalkContinue
}

func (g *Generator) EnterOrderedList(n *ast.OrderedList) visitor.WalkAction {
	g.po.needsEnumitem = true
	g.listStack = append(g.listStack, listFrame{ordered: true, tight: n.Tight})
	if n.Start != 1 {
		if n.Tight {
			g.writef("\\begin{enumerate}[noitemsep, start=%d]\n", n.Start)
		} else {
			g.writef("\\begin{enumerate}[start=%d]\n", n.Start)
		}
	} else if n.Tight {
		g.write("\\begin{enumerate}[noitemsep]\n")
	} else {
		g.write("\\begin{enumerate}\n")
	}
	return visitor.WalkContinue
}

func (g *Generator) LeaveOrderedList(_ *ast.OrderedList) visitor.WalkAction {
	g.listStack = g.listStack[:len(g.listStack)-1]
	g.write("\\end{enumerate}\n\n")
	return visitor.WalkContinue
}

func (g *Generator) EnterListItem(_ *ast.ListItem) visitor.WalkAction {
	g.write("  \\item ")
	return visitor.WalkContinue
}

func (g *Generator) LeaveListItem(_ *ast.ListItem) visitor.WalkAction {
	g.nl()
	return visitor.WalkContinue
}

// ─── Code blocks ──────────────────────────────────────────────────────────────

func (g *Generator) EnterFencedCode(n *ast.FencedCode) visitor.WalkAction {
	g.po.needsListings = true
	lang := n.Language()
	if lang != "" {
		g.writef("\\begin{lstlisting}[language=%s]\n", lang)
	} else {
		g.write("\\begin{lstlisting}\n")
	}
	g.write(n.Literal)
	g.write("\\end{lstlisting}\n\n")
	return visitor.WalkSkipChildren
}

func (g *Generator) EnterIndentedCode(n *ast.IndentedCode) visitor.WalkAction {
	g.write("\\begin{verbatim}\n")
	g.write(n.Literal)
	g.write("\\end{verbatim}\n\n")
	return visitor.WalkSkipChildren
}

// ─── HTML blocks (unsupported) ────────────────────────────────────────────────

func (g *Generator) EnterHTMLBlock(n *ast.HTMLBlock) visitor.WalkAction {
	if g.opts.Strict {
		// In a future version this could return an error via a separate channel.
		g.writef("%% [HTML block omitted — strict mode]\n\n")
	} else {
		g.writef("%% [HTML block omitted]\n%% %s\n\n",
			strings.ReplaceAll(strings.TrimSpace(n.Literal), "\n", "\n%% "))
	}
	return visitor.WalkSkipChildren
}

// ─── Thematic break ───────────────────────────────────────────────────────────

func (g *Generator) EnterThematicBreak(_ *ast.ThematicBreak) visitor.WalkAction {
	g.write("\n\\noindent\\rule{\\linewidth}{0.4pt}\n\n")
	return visitor.WalkContinue
}

// ─── Tables ───────────────────────────────────────────────────────────────────

func (g *Generator) EnterTable(n *ast.Table) visitor.WalkAction {
	g.po.needsBooktabs = true
	g.tableAlignments = n.Alignments

	// Build column spec
	cols := make([]string, len(n.Alignments))
	for i, a := range n.Alignments {
		cols[i] = a.String()
	}
	g.writef("\\begin{tabular}{%s}\n", strings.Join(cols, ""))
	g.write("\\toprule\n")
	return visitor.WalkContinue
}

func (g *Generator) LeaveTable(_ *ast.Table) visitor.WalkAction {
	g.write("\\bottomrule\n")
	if g.inTableFloat {
		g.write("\\end{tabular}\n") // TableBlock adds \label + \end{table} after this
	} else {
		g.write("\\end{tabular}\n\n")
	}
	g.tableAlignments = nil
	return visitor.WalkContinue
}

func (g *Generator) EnterTableSection(n *ast.TableSection) visitor.WalkAction {
	g.inTableHeader = n.IsHeader
	return visitor.WalkContinue
}

func (g *Generator) LeaveTableSection(n *ast.TableSection) visitor.WalkAction {
	if n.IsHeader {
		g.write("\\midrule\n")
	}
	return visitor.WalkContinue
}

func (g *Generator) EnterTableRow(_ *ast.TableRow) visitor.WalkAction {
	g.tableColCount = 0
	return visitor.WalkContinue
}

func (g *Generator) LeaveTableRow(_ *ast.TableRow) visitor.WalkAction {
	// Remove trailing " & " added by cells and replace with line ending
	s := g.buf.String()
	if strings.HasSuffix(s, " & ") {
		g.buf.Reset()
		g.buf.WriteString(s[:len(s)-3])
	}
	g.write(" \\\\\n")
	return visitor.WalkContinue
}

func (g *Generator) EnterTableCell(_ *ast.TableCell) visitor.WalkAction {
	if g.tableColCount > 0 {
		g.write(" & ")
	}
	g.tableColCount++
	return visitor.WalkContinue
}

func (g *Generator) LeaveTableCell(_ *ast.TableCell) visitor.WalkAction {
	return visitor.WalkContinue
}

// ─── Inline nodes ─────────────────────────────────────────────────────────────

func (g *Generator) EnterText(n *ast.Text) visitor.WalkAction {
	g.write(escapeText(n.Value))
	return visitor.WalkContinue
}

func (g *Generator) EnterSoftBreak(_ *ast.SoftBreak) visitor.WalkAction {
	g.write(" ")
	return visitor.WalkContinue
}

func (g *Generator) EnterHardBreak(_ *ast.HardBreak) visitor.WalkAction {
	g.write("\\\\\n")
	return visitor.WalkContinue
}

func (g *Generator) EnterEmphasis(_ *ast.Emphasis) visitor.WalkAction {
	g.write(`\textit{`)
	return visitor.WalkContinue
}

func (g *Generator) LeaveEmphasis(_ *ast.Emphasis) visitor.WalkAction {
	g.write("}")
	return visitor.WalkContinue
}

func (g *Generator) EnterStrong(_ *ast.Strong) visitor.WalkAction {
	g.write(`\textbf{`)
	return visitor.WalkContinue
}

func (g *Generator) LeaveStrong(_ *ast.Strong) visitor.WalkAction {
	g.write("}")
	return visitor.WalkContinue
}

func (g *Generator) EnterStrongEmphasis(_ *ast.StrongEmphasis) visitor.WalkAction {
	g.write(`\textbf{\textit{`)
	return visitor.WalkContinue
}

func (g *Generator) LeaveStrongEmphasis(_ *ast.StrongEmphasis) visitor.WalkAction {
	g.write("}}")
	return visitor.WalkContinue
}

func (g *Generator) EnterCodeSpan(n *ast.CodeSpan) visitor.WalkAction {
	g.writef(`\texttt{%s}`, escapeText(n.Literal))
	return visitor.WalkSkipChildren
}

func (g *Generator) EnterLink(n *ast.Link) visitor.WalkAction {
	g.writef(`\href{%s}{`, escapeURL(n.Destination))
	return visitor.WalkContinue
}

func (g *Generator) LeaveLink(_ *ast.Link) visitor.WalkAction {
	g.write("}")
	return visitor.WalkContinue
}

func (g *Generator) EnterImage(n *ast.Image) visitor.WalkAction {
	g.po.needsGraphicx = true
	// Collect alt text for the caption by extracting plain text from children
	alt := plainText(n.Children())
	if alt != "" {
		g.write("\\begin{figure}[h]\n")
		g.write("  \\centering\n")
		g.writef("  \\includegraphics{%s}\n", escapeText(n.Destination))
		g.writef("  \\caption{%s}\n", escapeText(alt))
		g.write("\\end{figure}")
	} else {
		g.writef("\\includegraphics{%s}", escapeText(n.Destination))
	}
	return visitor.WalkSkipChildren
}

func (g *Generator) EnterRawHTML(n *ast.RawHTML) visitor.WalkAction {
	g.writef("%% [inline HTML omitted: %s]\n", strings.TrimSpace(n.Literal))
	return visitor.WalkSkipChildren
}

func (g *Generator) EnterEscape(n *ast.Escape) visitor.WalkAction {
	g.write(escapeText(string(n.Char)))
	return visitor.WalkContinue
}

func (g *Generator) EnterEntity(n *ast.Entity) visitor.WalkAction {
	// Emit the decoded Unicode character directly; LaTeX with utf8 inputenc
	// handles it correctly.
	g.write(escapeText(string(n.Rune)))
	return visitor.WalkContinue
}

func (g *Generator) EnterStrikethrough(_ *ast.Strikethrough) visitor.WalkAction {
	g.po.needsUlem = true
	g.write(`\sout{`)
	return visitor.WalkContinue
}

func (g *Generator) LeaveStrikethrough(_ *ast.Strikethrough) visitor.WalkAction {
	g.write("}")
	return visitor.WalkContinue
}

// ─── Citation extension ───────────────────────────────────────────────────────

// EnterCitationBlock produces no output: the block is metadata only.
func (g *Generator) EnterCitationBlock(_ *ast.CitationBlock) visitor.WalkAction {
	return visitor.WalkSkipChildren
}

// ─── Table-float extension ────────────────────────────────────────────────────

// EnterDefinitionBlock produces no output: it is metadata only.
func (g *Generator) EnterDefinitionBlock(_ *ast.DefinitionBlock) visitor.WalkAction {
	return visitor.WalkSkipChildren
}

// EnterTableDef produces no output: it is metadata only.
func (g *Generator) EnterTableDef(_ *ast.TableDef) visitor.WalkAction {
	return visitor.WalkSkipChildren
}

// EnterTableBlock opens the \begin{table}[placement] float environment and
// sets inTableFloat so LeaveTable omits its trailing blank line.
func (g *Generator) EnterTableBlock(n *ast.TableBlock) visitor.WalkAction {
	g.po.needsBooktabs = true
	g.inTableFloat = true
	g.writef("\\begin{table}[%s]\n", n.Placement)
	g.write("\\centering\n")
	if len(n.CaptionNodes) > 0 {
		g.write("\\caption{")
		g.renderInlineNodes(n.CaptionNodes)
		g.write("}\n")
	}
	return visitor.WalkContinue
}

// LeaveTableBlock closes the float with \label and \end{table}.
func (g *Generator) LeaveTableBlock(n *ast.TableBlock) visitor.WalkAction {
	g.inTableFloat = false
	if n.LabelKey != "" {
		g.writef("\\label{tab:%s}\n", escapeText(n.LabelKey))
	}
	g.write("\\end{table}\n\n")
	return visitor.WalkContinue
}

// EnterTableRef emits `~\ref{tab:label}`, consuming the preceding space.
func (g *Generator) EnterTableRef(n *ast.TableRef) visitor.WalkAction {
	s := g.buf.String()
	if len(s) > 0 && s[len(s)-1] == ' ' {
		g.buf.Reset()
		g.buf.WriteString(s[:len(s)-1])
	}
	g.writef(`~\ref{tab:%s}`, n.LabelKey)
	return visitor.WalkSkipChildren
}

// EnterFigureBlock produces no output: it is metadata only.
func (g *Generator) EnterFigureBlock(_ *ast.FigureBlock) visitor.WalkAction {
	return visitor.WalkSkipChildren
}

// EnterFigureRef emits `~\ref{fig:label}`, consuming a preceding space the
// same way citations consume theirs.
func (g *Generator) EnterFigureRef(n *ast.FigureRef) visitor.WalkAction {
	s := g.buf.String()
	if len(s) > 0 && s[len(s)-1] == ' ' {
		g.buf.Reset()
		g.buf.WriteString(s[:len(s)-1])
	}
	g.writef(`~\ref{fig:%s}`, n.LabelKey)
	return visitor.WalkSkipChildren
}

// EnterFigureImage emits a complete LaTeX figure environment with \label,
// \caption, and \includegraphics[width=…\linewidth].
func (g *Generator) EnterFigureImage(n *ast.FigureImage) visitor.WalkAction {
	g.po.needsGraphicx = true
	g.writef("\\begin{figure}[%s]\n", n.Placement)
	g.write("  \\centering\n")
	g.writef("  \\includegraphics[width=%s\\linewidth]{%s}\n", n.Width, escapeText(n.Path))
	if len(n.CaptionNodes) > 0 {
		g.write("  \\caption{")
		g.renderInlineNodes(n.CaptionNodes)
		g.write("}\n")
	}
	g.writef("  \\label{fig:%s}\n", escapeText(n.LabelKey))
	g.write("\\end{figure}")
	return visitor.WalkSkipChildren
}

// EnterCitationRef emits `~\cite{bibtex}`.
// If the preceding output ends with a space, that space is consumed by the `~`
// (non-breaking space) so we get `text~\cite{key}` rather than `text ~\cite{key}`.
func (g *Generator) EnterCitationRef(n *ast.CitationRef) visitor.WalkAction {
	s := g.buf.String()
	if len(s) > 0 && s[len(s)-1] == ' ' {
		g.buf.Reset()
		g.buf.WriteString(s[:len(s)-1])
	}
	g.writef(`~\cite{%s}`, n.BibKey)
	return visitor.WalkSkipChildren
}

// ─── Utility ──────────────────────────────────────────────────────────────────

// plainText extracts the concatenated text content of a node slice.
// Used to derive alt text for images.
func plainText(nodes []ast.Node) string {
	var b strings.Builder
	var walk func([]ast.Node)
	walk = func(ns []ast.Node) {
		for _, n := range ns {
			if t, ok := n.(*ast.Text); ok {
				b.WriteString(t.Value)
			} else if e, ok := n.(*ast.Escape); ok {
				b.WriteRune(e.Char)
			} else if e, ok := n.(*ast.Entity); ok {
				b.WriteRune(e.Rune)
			}
			walk(n.Children())
		}
	}
	walk(nodes)
	return b.String()
}
