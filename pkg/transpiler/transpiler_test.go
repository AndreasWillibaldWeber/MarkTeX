// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package transpiler_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/andreaswillibaldweber/marktex/pkg/transpiler"
)

func TestHeadings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"h1", "# Heading", `\section{Heading}`},
		{"h2", "## Heading", `\subsection{Heading}`},
		{"h3", "### Heading", `\subsubsection{Heading}`},
		{"h4", "#### Heading", `\paragraph{Heading}\mbox{}`},
		{"h5", "##### Heading", `\subparagraph{Heading}\mbox{}`},
		{"h6", "###### Heading", `\noindent\textbf{Heading}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := transpiler.TranspileString(tc.input, transpiler.Options{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(got, tc.want) {
				t.Errorf("output does not contain %q\ngot:\n%s", tc.want, got)
			}
		})
	}
}

func TestEmphasis(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"**bold**", `\textbf{bold}`},
		{"*italic*", `\textit{italic}`},
		{"***both***", `\textbf{\textit{both}}`},
	}
	for _, tc := range tests {
		got, err := transpiler.TranspileString(tc.input, transpiler.Options{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, tc.want) {
			t.Errorf("input %q: output does not contain %q\ngot:\n%s", tc.input, tc.want, got)
		}
	}
}

func TestCodeSpan(t *testing.T) {
	got, err := transpiler.TranspileString("Use `fmt.Println` here.", transpiler.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\texttt{fmt.Println}`) {
		t.Errorf("got:\n%s", got)
	}
}

func TestFencedCode(t *testing.T) {
	src := "```go\nfmt.Println(\"hello\")\n```"
	got, err := transpiler.TranspileString(src, transpiler.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\begin{lstlisting}[language=go]`) {
		t.Errorf("missing lstlisting[language=go]\ngot:\n%s", got)
	}
	if !strings.Contains(got, `\end{lstlisting}`) {
		t.Errorf("missing \\end{lstlisting}\ngot:\n%s", got)
	}
}

func TestBulletList(t *testing.T) {
	src := "- alpha\n- beta\n- gamma"
	got, err := transpiler.TranspileString(src, transpiler.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\begin{itemize}`) {
		t.Errorf("missing \\begin{itemize}\ngot:\n%s", got)
	}
	if !strings.Contains(got, `\item alpha`) {
		t.Errorf("missing \\item alpha\ngot:\n%s", got)
	}
}

func TestOrderedList(t *testing.T) {
	src := "1. one\n2. two\n3. three"
	got, err := transpiler.TranspileString(src, transpiler.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\begin{enumerate}`) {
		t.Errorf("missing \\begin{enumerate}\ngot:\n%s", got)
	}
	if !strings.Contains(got, `\item one`) {
		t.Errorf("missing \\item one\ngot:\n%s", got)
	}
}

func TestOrderedListCustomStart(t *testing.T) {
	src := "3. three\n4. four"
	got, err := transpiler.TranspileString(src, transpiler.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `[start=3]`) {
		t.Errorf("expected [start=3] for list starting at 3\ngot:\n%s", got)
	}
}

func TestBlockQuote(t *testing.T) {
	got, err := transpiler.TranspileString("> A quoted line.", transpiler.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\begin{quote}`) {
		t.Errorf("missing \\begin{quote}\ngot:\n%s", got)
	}
}

func TestThematicBreak(t *testing.T) {
	got, err := transpiler.TranspileString("---", transpiler.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\noindent\rule`) {
		t.Errorf("missing \\noindent\\rule\ngot:\n%s", got)
	}
}

func TestLink(t *testing.T) {
	got, err := transpiler.TranspileString("[click here](https://example.com)", transpiler.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\href{https://example.com}{click here}`) {
		t.Errorf("missing \\href\ngot:\n%s", got)
	}
}

func TestSpecialCharEscaping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"50% off", `50\% off`},
		{"$10.00", `\$10.00`},
		{"a & b", `a \& b`},
		{"x_y", `x\_y`},
	}
	for _, tc := range tests {
		got, err := transpiler.TranspileString(tc.input, transpiler.Options{})
		if err != nil {
			t.Fatalf("input %q: %v", tc.input, err)
		}
		if !strings.Contains(got, tc.want) {
			t.Errorf("input %q: expected %q in output\ngot:\n%s", tc.input, tc.want, got)
		}
	}
}

func TestTable(t *testing.T) {
	src := "| A | B |\n|---|---|\n| 1 | 2 |"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtTables})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\begin{tabular}`) {
		t.Errorf("missing \\begin{tabular}\ngot:\n%s", got)
	}
	if !strings.Contains(got, `\toprule`) {
		t.Errorf("missing \\toprule\ngot:\n%s", got)
	}
}

func TestStrikethrough(t *testing.T) {
	got, err := transpiler.TranspileString("~~deleted~~", transpiler.Options{Extensions: transpiler.ExtStrikethrough})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\sout{deleted}`) {
		t.Errorf("missing \\sout\ngot:\n%s", got)
	}
}

func TestStandalone(t *testing.T) {
	got, err := transpiler.TranspileString("# Title", transpiler.Options{Standalone: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\documentclass{article}`) {
		t.Errorf("missing \\documentclass in standalone mode\ngot:\n%s", got)
	}
	if !strings.Contains(got, `\begin{document}`) {
		t.Errorf("missing \\begin{document}\ngot:\n%s", got)
	}
	if !strings.Contains(got, `\end{document}`) {
		t.Errorf("missing \\end{document}\ngot:\n%s", got)
	}
}

func TestFragmentMode(t *testing.T) {
	got, err := transpiler.TranspileString("# Title", transpiler.Options{Standalone: false})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, `\documentclass`) {
		t.Errorf("\\documentclass should not appear in fragment mode\ngot:\n%s", got)
	}
}

func TestCitationBasic(t *testing.T) {
	src := "---\nC#01:doe2023\nC#02:smith2024\n---\n\nSee [C#01] and [C#02]."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtCitations})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `~\cite{doe2023}`) {
		t.Errorf("missing \\cite{doe2023}\ngot:\n%s", got)
	}
	if !strings.Contains(got, `~\cite{smith2024}`) {
		t.Errorf("missing \\cite{smith2024}\ngot:\n%s", got)
	}
	// Citation block itself must not appear in output
	if strings.Contains(got, "C#01") || strings.Contains(got, "C#02") {
		t.Errorf("citation keys leaked into output\ngot:\n%s", got)
	}
}

func TestCitationTildeReplacesPrecedingSpace(t *testing.T) {
	// The space before [C#01] must be consumed by ~, not left as a separate space.
	src := "---\nC#01:ref\n---\n\nSome text [C#01] here."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtCitations})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, ` ~\cite`) {
		t.Errorf("unexpected space before ~\\cite (should be consumed)\ngot:\n%s", got)
	}
	if !strings.Contains(got, `text~\cite{ref}`) {
		t.Errorf("expected 'text~\\cite{ref}'\ngot:\n%s", got)
	}
}

func TestCitationBlockDoesNotBreakThematicBreak(t *testing.T) {
	// A plain --- without citation entries must still produce a thematic break.
	src := "Before\n\n---\n\nAfter"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtCitations})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\noindent\rule`) {
		t.Errorf("thematic break missing when --- is not a citation block\ngot:\n%s", got)
	}
}

func TestCitationExactUserExample(t *testing.T) {
	src := "---\nC#01:bibtexidentifier01\nC#02:bibtexidentifier02\n---\n\nSome text [C#01] some more text [C#02]."
	want := `Some text~\cite{bibtexidentifier01} some more text~\cite{bibtexidentifier02}.`
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtCitations})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, want) {
		t.Errorf("exact user example failed\nwant (substring): %s\ngot:\n%s", want, got)
	}
}

// ── Figure extension ──────────────────────────────────────────────────────────

func TestFigureRef(t *testing.T) {
	src := "---\nF#01:myfig\n---\n\nSee Figure [F#01] here."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtFigures})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `~\ref{fig:myfig}`) {
		t.Errorf("missing \\ref{fig:myfig}\ngot:\n%s", got)
	}
}

func TestFigureRefTildeReplacesPrecedingSpace(t *testing.T) {
	src := "---\nF#01:lbl\n---\n\nSee Figure [F#01] here."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtFigures})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, ` ~\ref`) {
		t.Errorf("space before ~\\ref should be consumed\ngot:\n%s", got)
	}
	if !strings.Contains(got, `Figure~\ref{fig:lbl}`) {
		t.Errorf("expected 'Figure~\\ref{fig:lbl}'\ngot:\n%s", got)
	}
}

func TestFigureImageFull(t *testing.T) {
	src := "---\nF#01:figurelabel01\n---\n\n![F#01:1.0:h][My Caption.](img/photo.png)"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtFigures})
	if err != nil {
		t.Fatal(err)
	}
	checks := []string{
		`\begin{figure}[h]`,
		`\includegraphics[width=1.0\linewidth]{img/photo.png}`,
		`\caption{My Caption.}`,
		`\label{fig:figurelabel01}`,
		`\end{figure}`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestFigureImagePlacementVariants(t *testing.T) {
	cases := []struct{ placement, src string }{
		{"ht!", "---\nF#01:l\n---\n\n![F#01:0.5:ht!][Cap.](a.png)"},
		{"hb", "---\nF#01:l\n---\n\n![F#01:0.8:hb][Cap.](a.png)"},
	}
	for _, tc := range cases {
		got, err := transpiler.TranspileString(tc.src, transpiler.Options{Extensions: transpiler.ExtFigures})
		if err != nil {
			t.Fatalf("placement %q: %v", tc.placement, err)
		}
		want := `\begin{figure}[` + tc.placement + `]`
		if !strings.Contains(got, want) {
			t.Errorf("placement %q: missing %q\ngot:\n%s", tc.placement, want, got)
		}
	}
}

func TestFigureWidthInOutput(t *testing.T) {
	src := "---\nF#01:fig\n---\n\n![F#01:0.5:h][Cap.](img.png)"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtFigures})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `width=0.5\linewidth`) {
		t.Errorf("expected width=0.5\\linewidth\ngot:\n%s", got)
	}
}

func TestFigureBlockInvisibleInOutput(t *testing.T) {
	src := "---\nF#01:lbl\n---\n\nNo figure referenced."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtFigures})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "F#01") || strings.Contains(got, "lbl") {
		t.Errorf("figure block leaked into output\ngot:\n%s", got)
	}
}

func TestFigureExactUserExample(t *testing.T) {
	src := "---\nF#01:figurelabel01\nF#02:figurelabel02\n---\n\n" +
		"See Figure [F#01] to see the figure.\n\n" +
		"![F#01:1.0:h][Project Logo 1.](assets/logo1.png)\n\n" +
		"See Figure [F#02] to see the figure.\n\n" +
		"![F#02:0.5:ht!][Project Logo 2.](assets/logo2.png)"

	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtFigures})
	if err != nil {
		t.Fatal(err)
	}
	wants := []string{
		`See Figure~\ref{fig:figurelabel01} to see the figure.`,
		`\begin{figure}[h]`,
		`\includegraphics[width=1.0\linewidth]{assets/logo1.png}`,
		`\caption{Project Logo 1.}`,
		`\label{fig:figurelabel01}`,
		`See Figure~\ref{fig:figurelabel02} to see the figure.`,
		`\begin{figure}[ht!]`,
		`\includegraphics[width=0.5\linewidth]{assets/logo2.png}`,
		`\label{fig:figurelabel02}`,
	}
	for _, w := range wants {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q\ngot:\n%s", w, got)
		}
	}
}

// ── Table-float extension ─────────────────────────────────────────────────────

func TestTableRef(t *testing.T) {
	src := "---\nT#01:mytable\n---\n\nSee Table [T#01] here."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtTableFloat})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `~\ref{tab:mytable}`) {
		t.Errorf("missing \\ref{tab:mytable}\ngot:\n%s", got)
	}
}

func TestTableRefTildeReplacesPrecedingSpace(t *testing.T) {
	src := "---\nT#01:lbl\n---\n\nSee Table [T#01] here."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtTableFloat})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, ` ~\ref`) {
		t.Errorf("space before ~\\ref should be consumed\ngot:\n%s", got)
	}
	if !strings.Contains(got, `Table~\ref{tab:lbl}`) {
		t.Errorf("expected 'Table~\\ref{tab:lbl}'\ngot:\n%s", got)
	}
}

func TestTableFloat(t *testing.T) {
	src := "---\nT#01:mytab\n---\n\n" +
		"|- T#01:ht -|\n|- My Caption. -|\n" +
		"| A | B |\n|---|---|\n| 1 | 2 |"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtTableFloat})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`\begin{table}[ht]`,
		`\centering`,
		`\caption{My Caption.}`,
		`\begin{tabular}`,
		`\label{tab:mytab}`,
		`\end{table}`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestTableFloatPlacementVariants(t *testing.T) {
	for _, placement := range []string{"h!", "ht", "hb!", "H"} {
		src := "---\nT#01:l\n---\n\n|- T#01:" + placement + " -|\n| A |\n|---|\n| 1 |"
		got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtTableFloat})
		if err != nil {
			t.Fatalf("placement %q: %v", placement, err)
		}
		want := `\begin{table}[` + placement + `]`
		if !strings.Contains(got, want) {
			t.Errorf("placement %q: missing %q\ngot:\n%s", placement, want, got)
		}
	}
}

func TestTableDefBlockInvisible(t *testing.T) {
	src := "---\nT#01:lbl\n---\n\nNo table referenced."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtTableFloat})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "T#01") || strings.Contains(got, "lbl") {
		t.Errorf("table definition block leaked into output\ngot:\n%s", got)
	}
}

func TestTableFloatNoCaptionRow(t *testing.T) {
	src := "---\nT#01:tab\n---\n\n|- T#01:h -|\n| A |\n|---|\n| 1 |"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtTableFloat})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\begin{table}[h]`) {
		t.Errorf("missing \\begin{table}[h]\ngot:\n%s", got)
	}
	if strings.Contains(got, `\caption`) {
		t.Errorf("\\caption should not appear without a caption row\ngot:\n%s", got)
	}
}

func TestPlainTableUnaffected(t *testing.T) {
	src := "| A | B |\n|---|---|\n| 1 | 2 |"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtAll})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, `\begin{table}`) {
		t.Errorf("plain table should not be wrapped in a float\ngot:\n%s", got)
	}
	if !strings.Contains(got, `\begin{tabular}`) {
		t.Errorf("missing \\begin{tabular}\ngot:\n%s", got)
	}
}

func TestTableFloatExactUserExample(t *testing.T) {
	src := "---\nT#01:tablelabel01\nT#02:tablelabel02\n---\n\n" +
		"See Table [T#01] to see the table.\n\n" +
		"|- T#01:h! -|\n|- This is table 1. -|\n" +
		"| Name    | Score | Grade |\n|:--------|------:|:-----:|\n" +
		"| Alice   | 95    | A     |\n\n" +
		"See Table [T#02] to see the table.\n\n" +
		"|- T#02:ht -|\n|- This is table 2. -|\n" +
		"| Name    | Score | Grade |\n|:--------|------:|:-----:|\n" +
		"| Alice   | 95    | A     |"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtTableFloat})
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{
		`See Table~\ref{tab:tablelabel01} to see the table.`,
		`\begin{table}[h!]`,
		`\caption{This is table 1.}`,
		`\label{tab:tablelabel01}`,
		`See Table~\ref{tab:tablelabel02} to see the table.`,
		`\begin{table}[ht]`,
		`\caption{This is table 2.}`,
		`\label{tab:tablelabel02}`,
	} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q\ngot:\n%s", w, got)
		}
	}
}

// ── Integration: mixed definition blocks and cross-extension combinations ────

// TestMixedDefinitionBlock verifies that C#, F#, and T# entries can coexist
// in a single ---…--- block.
func TestMixedDefinitionBlock(t *testing.T) {
	src := "---\n" +
		"C#01:doe2023\n" +
		"F#01:figarch\n" +
		"T#01:tabresults\n" +
		"---\n\n" +
		"See [C#01] and Figure [F#01] and Table [T#01]."

	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtAll})
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"C#01", "F#01", "T#01", `\noindent\rule`} {
		if strings.Contains(got, bad) {
			t.Errorf("definition block leaked or became thematic break; found %q\ngot:\n%s", bad, got)
		}
	}
	for _, want := range []string{`~\cite{doe2023}`, `~\ref{fig:figarch}`, `~\ref{tab:tabresults}`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q\ngot:\n%s", want, got)
		}
	}
}

// TestAllRefTypesInOneParagraph ensures all three ref types resolve in one sentence.
func TestAllRefTypesInOneParagraph(t *testing.T) {
	src := "---\nC#01:smith2024\nF#01:fig_overview\nT#01:tab_summary\n---\n\n" +
		"As shown by [C#01], see Figure [F#01] and Table [T#01] for details."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtAll})
	if err != nil {
		t.Fatal(err)
	}
	want := `As shown by~\cite{smith2024}, see Figure~\ref{fig:fig_overview} and Table~\ref{tab:tab_summary} for details.`
	if !strings.Contains(got, want) {
		t.Errorf("expected %q\ngot:\n%s", want, got)
	}
}

// TestCitationInFigureCaption verifies [C#key] resolves inside a figure caption.
func TestCitationInFigureCaption(t *testing.T) {
	src := "---\nC#01:doe2023\nF#01:fig1\n---\n\n" +
		"![F#01:0.8:h][Architecture from [C#01].](figures/arch.png)"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtAll})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\caption{Architecture from~\cite{doe2023}.}`) {
		t.Errorf("citation not resolved inside figure caption\ngot:\n%s", got)
	}
}

// TestCitationInTableCaption verifies [C#key] resolves inside a table caption.
func TestCitationInTableCaption(t *testing.T) {
	src := "---\nC#01:smith2024\nT#01:tab1\n---\n\n" +
		"|- T#01:ht -|\n|- Results from [C#01]. -|\n| A | B |\n|---|---|\n| 1 | 2 |"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtAll})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\caption{Results from~\cite{smith2024}.}`) {
		t.Errorf("citation not resolved inside table caption\ngot:\n%s", got)
	}
}

// TestFigureRefInTableCaption verifies [F#key] resolves inside a table caption.
func TestFigureRefInTableCaption(t *testing.T) {
	src := "---\nF#01:fig_overview\nT#01:tab1\n---\n\n" +
		"|- T#01:h -|\n|- Corresponds to Figure [F#01]. -|\n| A |\n|---|\n| 1 |"
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtAll})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\caption{Corresponds to Figure~\ref{fig:fig_overview}.}`) {
		t.Errorf("figure ref not resolved inside table caption\ngot:\n%s", got)
	}
}

// TestMixedBlockDoesNotBreakThematicBreak ensures a plain --- still renders as
// a thematic break even when all extensions are active.
func TestMixedBlockDoesNotBreakThematicBreak(t *testing.T) {
	src := "Before.\n\n---\n\nAfter."
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: transpiler.ExtAll})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\noindent\rule`) {
		t.Errorf("plain --- should still produce a thematic break\ngot:\n%s", got)
	}
}

// ── Document metadata extension ───────────────────────────────────────────────

func TestDocumentMetaFragmentModeIgnored(t *testing.T) {
	src := "---\nAuthor: Ada Lovelace\nTitle: My Doc\nMakeTitlePage\nMakeTOC\n---\n\n# Hello"
	got, err := transpiler.TranspileString(src, transpiler.Options{
		Extensions: transpiler.ExtDocumentMeta,
		Standalone: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{`\documentclass`, `\author`, `\title`, `\maketitle`, `\tableofcontents`} {
		if strings.Contains(got, bad) {
			t.Errorf("metadata command %q must not appear in fragment mode\ngot:\n%s", bad, got)
		}
	}
}

func TestDocumentMetaTitleAuthorDate(t *testing.T) {
	src := "---\nTitle: My Document\nAuthor: Ada Lovelace\nDate: 2025-01-01\n---\n\n# Hello"
	got, err := transpiler.TranspileString(src, transpiler.Options{
		Extensions: transpiler.ExtDocumentMeta,
		Standalone: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`\title{My Document}`,
		`\author{Ada Lovelace}`,
		`\date{2025-01-01}`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestDocumentMetaSubtitle(t *testing.T) {
	src := "---\nTitle: Main\nSubtitle: Secondary\n---\n\n# Hello"
	got, err := transpiler.TranspileString(src, transpiler.Options{
		Extensions: transpiler.ExtDocumentMeta,
		Standalone: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\title{Main\\[0.5em]{\normalfont\large Secondary}}`) {
		t.Errorf("subtitle not encoded in \\title\ngot:\n%s", got)
	}
}

func TestDocumentMetaFrontMatterOrder(t *testing.T) {
	src := "---\nTitle: T\nMakeLOL\nMakeLOT\nMakeLOF\nMakeTOC\nMakeTitlePage\n---\n\n# Body"
	got, err := transpiler.TranspileString(src, transpiler.Options{
		Extensions: transpiler.ExtDocumentMeta,
		Standalone: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	commands := []string{`\maketitle`, `\tableofcontents`, `\listoffigures`, `\listoftables`, `\lstlistoflistings`}
	prev := 0
	for _, cmd := range commands {
		pos := strings.Index(got, cmd)
		if pos < 0 {
			t.Errorf("missing %q\ngot:\n%s", cmd, got)
			continue
		}
		if pos < prev {
			t.Errorf("command order wrong: %q appeared before previous command\ngot:\n%s", cmd, got)
		}
		prev = pos
	}
}

func TestDocumentMetaMakeLOLAddsListings(t *testing.T) {
	// MakeLOL must trigger inclusion of the listings package even when no
	// code block is present in the document.
	src := "---\nMakeLOL\n---\n\n# Hello"
	got, err := transpiler.TranspileString(src, transpiler.Options{
		Extensions: transpiler.ExtDocumentMeta,
		Standalone: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `\usepackage{listings}`) {
		t.Errorf("\\usepackage{listings} missing when MakeLOL is set\ngot:\n%s", got)
	}
	if !strings.Contains(got, `\lstlistoflistings`) {
		t.Errorf("\\lstlistoflistings missing\ngot:\n%s", got)
	}
}

func TestDocumentMetaMixedWithRefs(t *testing.T) {
	src := "---\nAuthor: Ada\nTitle: Doc\nMakeTitlePage\nC#01:doe2023\nF#01:fig1\nT#01:tab1\n---\n\n" +
		"See [C#01], Figure [F#01], Table [T#01]."
	got, err := transpiler.TranspileString(src, transpiler.Options{
		Extensions: transpiler.ExtAll,
		Standalone: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`\author{Ada}`,
		`\title{Doc}`,
		`\maketitle`,
		`~\cite{doe2023}`,
		`~\ref{fig:fig1}`,
		`~\ref{tab:tab1}`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q\ngot:\n%s", want, got)
		}
	}
}

// ── Setext headings (issue #2) ────────────────────────────────────────────────

// extensionSets covers the configurations under which parser bugs have shown
// different symptoms (issue #2: without tables the `---` underline became a
// thematic break, with tables it became a one-column table). Options treats an
// Extensions value of 0 as ExtAll, so "no-tables" uses a non-zero mask.
var extensionSets = map[string]transpiler.Extension{
	"no-tables": transpiler.ExtStrikethrough,
	"tables":    transpiler.ExtTables,
	"all":       transpiler.ExtAll,
}

func transpile(t *testing.T, src string, ext transpiler.Extension) string {
	t.Helper()
	got, err := transpiler.TranspileString(src, transpiler.Options{Extensions: ext})
	if err != nil {
		t.Fatalf("TranspileString(%q): %v", src, err)
	}
	return got
}

// assertInOrder checks that every substring is present and that they appear in
// the given order.
func assertInOrder(t *testing.T, got string, subs ...string) {
	t.Helper()
	prev := -1
	for _, s := range subs {
		idx := strings.Index(got, s)
		if idx < 0 {
			t.Errorf("missing %q\ngot:\n%s", s, got)
			return
		}
		if idx < prev {
			t.Errorf("%q appears out of order\ngot:\n%s", s, got)
			return
		}
		prev = idx
	}
}

// TestSetextHeadingIssue02 reproduces the exact report from issue #2: a `---`
// underline directly below a paragraph line is a level-2 heading, not a table.
func TestSetextHeadingIssue02(t *testing.T) {
	const src = "Also header 2 (not a table!)\n-------------\n"
	const want = "\\subsection{Also header 2 (not a table!)}\n\n"
	for name, ext := range extensionSets {
		t.Run(name, func(t *testing.T) {
			if got := transpile(t, src, ext); got != want {
				t.Errorf("want exactly %q\ngot:\n%s", want, got)
			}
		})
	}
}

func TestSetextHeadings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"h1 long underline", "Title\n=========", `\section{Title}`},
		{"h1 minimal underline", "Title\n=", `\section{Title}`},
		{"h2 long underline", "Title\n---------", `\subsection{Title}`},
		{"h2 minimal underline", "Title\n-", `\subsection{Title}`},
		{"h2 exactly three dashes", "Title\n---", `\subsection{Title}`},
		{"trailing whitespace on underline", "Title\n---  \t", `\subsection{Title}`},
		{"underline indented three spaces", "Title\n   ---", `\subsection{Title}`},
		{"multi-line heading content", "Line one\nLine two\n---", `\subsection{Line one Line two}`},
		{"inline markup in heading", "*Foo* and `bar`\n---", `\subsection{\textit{Foo} and \texttt{bar}}`},
		{"trailing spaces on text are trimmed", "Title  \n---", `\subsection{Title}`},
		{"pipes in text stay text", "| a | b |\n---", `\subsection{| a | b |}`},
		{"heading after a paragraph", "intro\n\nTitle\n---", `\subsection{Title}`},
		{"crlf line endings", "Title\r\n---\r\n", `\subsection{Title}`},
	}
	for name, ext := range extensionSets {
		for _, tc := range tests {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				got := transpile(t, tc.input, ext)
				if !strings.Contains(got, tc.want) {
					t.Errorf("missing %q\ngot:\n%s", tc.want, got)
				}
				for _, bad := range []string{`\begin{tabular}`, `\noindent\rule`} {
					if strings.Contains(got, bad) {
						t.Errorf("unexpected %q\ngot:\n%s", bad, got)
					}
				}
			})
		}
	}
}

// TestSetextUnderlineNotRecognised lists lines that must never be read as a
// setext underline. None of them may yield a heading or a table either.
func TestSetextUnderlineNotRecognised(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // optional positive expectation
	}{
		{"blank line before dashes", "Text\n\n---", `\noindent\rule`},
		{"spaced dashes are a thematic break", "Text\n- - -", `\noindent\rule`},
		{"mixed characters are paragraph text", "Text\n--a", "Text --a"},
		// Four spaces of indentation disqualify an underline; the exact
		// fallback rendering is not pinned here.
		{"four-space indent", "Text\n    ---", ""},
		{"no open paragraph after list item", "- item\n===", `\item item`},
		{"underline-like text at start of input", "---\nText", "Text"},
	}
	for name, ext := range extensionSets {
		for _, tc := range tests {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				got := transpile(t, tc.input, ext)
				if tc.want != "" && !strings.Contains(got, tc.want) {
					t.Errorf("missing %q\ngot:\n%s", tc.want, got)
				}
				for _, bad := range []string{`\section{`, `\subsection{`, `\begin{tabular}`} {
					if strings.Contains(got, bad) {
						t.Errorf("unexpected %q\ngot:\n%s", bad, got)
					}
				}
			})
		}
	}
}

// TestThematicBreaksUnaffectedBySetextPrecedence guards every thematic-break
// position that does not follow an open paragraph.
func TestThematicBreaksUnaffectedBySetextPrecedence(t *testing.T) {
	tests := []struct {
		name  string
		input string
		also  string // additional expected substring
	}{
		{"dashes alone", "---", ""},
		{"asterisks alone", "***", ""},
		{"underscores alone", "___", ""},
		{"spaced dashes alone", "- - -", ""},
		{"after blank-line separated paragraph", "Text\n\n---", "Text"},
		{"directly after ATX heading", "# Title\n---", `\section{Title}`},
		{"directly after fenced code", "```\ncode\n```\n---", `\end{lstlisting}`},
		{"directly after indented code", "    code\n---", `\begin{verbatim}`},
		{"directly after bullet list", "- item\n---", `\item item`},
		{"directly after ordered list", "1. item\n---", `\item item`},
		{"between paragraphs", "Before\n\n---\n\nAfter", "After"},
	}
	for name, ext := range extensionSets {
		for _, tc := range tests {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				got := transpile(t, tc.input, ext)
				if !strings.Contains(got, `\noindent\rule`) {
					t.Errorf("missing thematic break\ngot:\n%s", got)
				}
				if tc.also != "" && !strings.Contains(got, tc.also) {
					t.Errorf("missing %q\ngot:\n%s", tc.also, got)
				}
				for _, bad := range []string{`\section{`, `\subsection{`, `\begin{tabular}`} {
					if strings.HasPrefix(tc.also, bad) {
						continue // the ATX case legitimately contains \section{
					}
					if strings.Contains(got, bad) {
						t.Errorf("unexpected %q\ngot:\n%s", bad, got)
					}
				}
			})
		}
	}
}

// TestTablesUnaffectedBySetextPrecedence verifies that real GFM tables, whose
// separator rows contain pipes or colons, are still recognised everywhere a
// table may appear.
func TestTablesUnaffectedBySetextPrecedence(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wants []string
	}{
		{"two columns", "| A | B |\n|---|---|\n| 1 | 2 |",
			[]string{`\begin{tabular}{ll}`, `A & B \\`, `1 & 2 \\`}},
		{"single column with pipes", "| A |\n|---|\n| 1 |",
			[]string{`\begin{tabular}{l}`, `A \\`, `1 \\`}},
		{"alignment separator without outer pipes", "A | B\n:---|---:\n1 | 2",
			[]string{`\begin{tabular}{lr}`, `A & B \\`, `1 & 2 \\`}},
		{"separator indented three spaces", "| A |\n   |---|\n| 1 |",
			[]string{`\begin{tabular}{l}`, `A \\`}},
		{"table interrupting a paragraph", "Intro text\n| A | B |\n|---|---|\n| 1 | 2 |",
			[]string{"Intro text", `\begin{tabular}{ll}`, `A & B \\`}},
		{"table after setext heading", "Results\n---\n\n| A | B |\n|---|---|\n| 1 | 2 |",
			[]string{`\subsection{Results}`, `\begin{tabular}{ll}`, `1 & 2 \\`}},
		{"table before setext heading", "| A |\n|---|\n| 1 |\n\nNext\n---",
			[]string{`\begin{tabular}{l}`, `\subsection{Next}`}},
	}
	for _, ext := range []transpiler.Extension{transpiler.ExtTables, transpiler.ExtAll} {
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got := transpile(t, tc.input, ext)
				assertInOrder(t, got, tc.wants...)
				if strings.Contains(got, `\subsection{Intro text}`) {
					t.Errorf("paragraph text before a table became a heading\ngot:\n%s", got)
				}
			})
		}
	}
}

func TestSetextHeadingInsideBlockQuote(t *testing.T) {
	got := transpile(t, "> Title\n> ---\n>\n> Body.", transpiler.ExtAll)
	assertInOrder(t, got, `\begin{quote}`, `\subsection{Title}`, "Body.", `\end{quote}`)
	if strings.Contains(got, `\begin{tabular}`) {
		t.Errorf("quoted setext heading became a table\ngot:\n%s", got)
	}
}

// TestThematicBreakClosesBlockQuote covers CommonMark 4.3: a setext underline
// cannot be a lazy continuation line, so `---` below a quoted paragraph ends
// the quote and renders as a rule outside it.
func TestThematicBreakClosesBlockQuote(t *testing.T) {
	got := transpile(t, "> Quoted\n---", transpiler.ExtAll)
	assertInOrder(t, got, `\begin{quote}`, "Quoted", `\end{quote}`, `\noindent\rule`)
	for _, bad := range []string{`\subsection{`, `\begin{tabular}`} {
		if strings.Contains(got, bad) {
			t.Errorf("unexpected %q\ngot:\n%s", bad, got)
		}
	}
}

// TestDefinitionBlockTakesPrecedenceOverSetextUnderline documents a deliberate
// precedence choice: an exact `---` that opens a valid definition block is the
// block delimiter even directly below a paragraph line, so documents that place
// their definition block right after text keep working. Only an exact `---`
// followed by recognised entries and a closing `---` qualifies.
func TestDefinitionBlockTakesPrecedenceOverSetextUnderline(t *testing.T) {
	src := "Intro line\n---\nC#01:doe2023\n---\n\nSee [C#01]."
	got := transpile(t, src, transpiler.ExtCitations)
	assertInOrder(t, got, "Intro line", `See~\cite{doe2023}.`)
	for _, bad := range []string{`\subsection{`, "C#01", `\noindent\rule`} {
		if strings.Contains(got, bad) {
			t.Errorf("unexpected %q\ngot:\n%s", bad, got)
		}
	}
}

func TestSetextHeadingSequence(t *testing.T) {
	src := "Document\n========\n\nIntro paragraph.\n\nSection\n-------\n\nBody paragraph."
	got := transpile(t, src, transpiler.ExtAll)
	assertInOrder(t, got,
		`\section{Document}`,
		"Intro paragraph.",
		`\subsection{Section}`,
		"Body paragraph.",
	)
}

// ── Hard and soft line breaks (issues #4 and #5) ─────────────────────────────

// TestHardLineBreakIssue04 reproduces the exact report from issue #4: two
// trailing spaces before a newline inside a paragraph produce `\\`.
func TestHardLineBreakIssue04(t *testing.T) {
	const src = "This is line 1 with two trailing spaces.  \nThis is line 2.\n"
	const want = "This is line 1 with two trailing spaces.\\\\\nThis is line 2.\n\n"
	for name, ext := range extensionSets {
		t.Run(name, func(t *testing.T) {
			if got := transpile(t, src, ext); got != want {
				t.Errorf("want exactly %q\ngot:        %q", want, got)
			}
		})
	}
}

// TestHardLineBreakInBlockQuoteIssue05 reproduces the exact report from
// issue #5: the same rule applies to paragraphs inside block quotes.
func TestHardLineBreakInBlockQuoteIssue05(t *testing.T) {
	const src = "> And this one  \n> has multiple lines  \n> in it.\n"
	const want = "\\begin{quote}\nAnd this one\\\\\nhas multiple lines\\\\\nin it.\n\n\\end{quote}\n\n"
	for name, ext := range extensionSets {
		t.Run(name, func(t *testing.T) {
			if got := transpile(t, src, ext); got != want {
				t.Errorf("want exactly %q\ngot:        %q", want, got)
			}
		})
	}
}

func TestHardLineBreaks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"two trailing spaces", "a  \nb", "a\\\\\nb"},
		{"three trailing spaces", "a   \nb", "a\\\\\nb"},
		{"many trailing spaces", "a        \nb", "a\\\\\nb"},
		{"backslash before newline", "a\\\nb", "a\\\\\nb"},
		{"leading spaces on next line dropped", "a  \n   b", "a\\\\\nb"},
		{"after strong", "**bold**  \nnext", "\\textbf{bold}\\\\\nnext"},
		{"after code span", "`code`  \nnext", "\\texttt{code}\\\\\nnext"},
		{"after escape", "x\\*  \nnext", "x*\\\\\nnext"},
		{"after link", "[t](http://x)  \nnext", "\\href{http://x}{t}\\\\\nnext"},
		{"inside link text", "[link  \ntext](http://x)", "\\href{http://x}{link\\\\\ntext}"},
		{"consecutive breaks", "l1  \nl2  \nl3", "l1\\\\\nl2\\\\\nl3"},
		{"mixed soft and hard", "l1\nl2  \nl3", "l1 l2\\\\\nl3"},
		{"crlf input", "a  \r\nb\r\n", "a\\\\\nb"},
		{"trailing tab after spaces", "a  \t\nb", "a\\\\\nb"},
	}
	for name, ext := range extensionSets {
		for _, tc := range tests {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				got := transpile(t, tc.input, ext)
				if !strings.Contains(got, tc.want) {
					t.Errorf("missing %q\ngot: %q", tc.want, got)
				}
				if strings.Contains(got, " \\\\") {
					t.Errorf("whitespace left before line break\ngot: %q", got)
				}
			})
		}
	}
}

func TestSoftLineBreaks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain newline", "a\nb", "a b"},
		{"one trailing space", "a \nb", "a b"},
		{"leading spaces on next line", "a\n   b", "a b"},
		{"trailing and leading spaces", "a \n  b", "a b"},
		{"across three lines", "a\nb\nc", "a b c"},
	}
	for name, ext := range extensionSets {
		for _, tc := range tests {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				got := transpile(t, tc.input, ext)
				if !strings.Contains(got, tc.want) {
					t.Errorf("missing %q\ngot: %q", tc.want, got)
				}
				if strings.Contains(got, `\\`) {
					t.Errorf("soft break rendered as hard break\ngot: %q", got)
				}
			})
		}
	}
}

// TestNoHardBreakAtBlockEnd covers CommonMark 6.7: a hard line break never
// occurs at the end of a block. A trailing `\\` would also be a LaTeX error.
func TestNoHardBreakAtBlockEnd(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wants []string
	}{
		{"paragraph before blank line", "end.  \n\nnext", []string{"end.\n\nnext"}},
		{"last paragraph of document", "end.  ", []string{"end.\n\n"}},
		{"last line of block quote", "> q  \n\nafter", []string{"q\n\n\\end{quote}", "after"}},
		{"list items", "- item  \n- two  ", []string{"\\item item\n", "\\item two\n"}},
		{"backslash at paragraph end is literal", "a\\", []string{"a\\textbackslash{}"}},
		{"leading spaces of paragraph dropped", "   a\nb", []string{"\na b\n"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := "\n" + transpile(t, tc.input, transpiler.ExtAll)
			for _, w := range tc.wants {
				if !strings.Contains(got, w) {
					t.Errorf("missing %q\ngot: %q", w, got)
				}
			}
			if strings.Contains(got, `\\`) {
				t.Errorf("unexpected hard break\ngot: %q", got)
			}
		})
	}
}

// TestHardBreakInsideHeadingRendersAsSpace guards the generator: `\\` inside
// a sectioning command breaks the table of contents, so heading content joins
// its lines with a space instead.
func TestHardBreakInsideHeadingRendersAsSpace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"two spaces in setext h2", "Line one  \nline two\n---", `\subsection{Line one line two}`},
		{"backslash in setext h1", "Line one\\\nline two\n===", `\section{Line one line two}`},
		{"atx trailing spaces", "# ATX heading  ", `\section{ATX heading}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := transpile(t, tc.input, transpiler.ExtAll)
			if !strings.Contains(got, tc.want) {
				t.Errorf("missing %q\ngot: %q", tc.want, got)
			}
			if strings.Contains(got, `\\`) {
				t.Errorf("hard break leaked into heading\ngot: %q", got)
			}
		})
	}
}

// TestCodeBlocksKeepTrailingSpaces makes sure line-break handling never
// touches verbatim content.
func TestCodeBlocksKeepTrailingSpaces(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"fenced", "```\nfoo  \nbar\n```", "foo  \nbar"},
		{"indented", "    foo  \n    bar", "foo  \nbar"},
		{"code span with inner spaces", "`a  b`", `\texttt{a  b}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := transpile(t, tc.input, transpiler.ExtAll)
			if !strings.Contains(got, tc.want) {
				t.Errorf("missing %q\ngot: %q", tc.want, got)
			}
			if strings.Contains(got, `\\`) {
				t.Errorf("hard break inside code\ngot: %q", got)
			}
		})
	}
}

// TestTableCellsUnaffectedByLineBreakHandling: cell padding is trimmed and
// row terminators are the only `\\` in the output.
func TestTableCellsUnaffectedByLineBreakHandling(t *testing.T) {
	got := transpile(t, "| a  | b |\n|---|---|\n| x  | y |", transpiler.ExtTables)
	assertInOrder(t, got, "a & b \\\\\n", `\midrule`, "x & y \\\\\n", `\bottomrule`)
	if n := strings.Count(got, `\\`); n != 2 {
		t.Errorf("want exactly 2 row terminators, got %d\ngot: %q", n, got)
	}
}

// ── Math extension (issue #6) ────────────────────────────────────────────────

// TestInlineMathIssue06 reproduces the exact report from issue #6.
func TestInlineMathIssue06(t *testing.T) {
	got := transpile(t, "$2+3=5$\n", transpiler.ExtAll)
	if want := "$2+3=5$\n\n"; got != want {
		t.Errorf("want exactly %q\ngot:        %q", want, got)
	}
}

func TestExtAllIncludesMath(t *testing.T) {
	if transpiler.ExtAll&transpiler.ExtMath == 0 {
		t.Fatal("ExtAll must include ExtMath so the CLI default renders math")
	}
}

func TestInlineMath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"in sentence", "The sum $2+3=5$ holds.", "The sum $2+3=5$ holds."},
		{"latex commands kept verbatim", `$e^{i\pi}+1=0$`, `$e^{i\pi}+1=0$`},
		{"special characters not escaped", `$x_1 \le y^2 \& z$`, `$x_1 \le y^2 \& z$`},
		{"percent kept verbatim", `$50\%$`, `$50\%$`},
		{"two spans in one paragraph", "$x$ and $y$", "$x$ and $y$"},
		{"escaped dollar inside math", `$a \$ b$`, `$a \$ b$`},
		{"mid-word", "a$b$c", "a$b$c"},
		{"inside emphasis", "*emph $x$ inside*", `\textit{emph $x$ inside}`},
		{"inside strong", "**$x$**", `\textbf{$x$}`},
		{"inside link text", "[link $x$](http://u)", `\href{http://u}{link $x$}`},
		{"inside heading", "# Energy $E=mc^2$", `\section{Energy $E=mc^2$}`},
		{"inside list item", "- item $x^2$", `\item item $x^2$`},
		{"inside block quote", "> quote $a+b$", "quote $a+b$"},
		{"inside table cell", "| a | b |\n|---|---|\n| $x$ | $y$ |", `$x$ & $y$ \\`},
		{"text around math still escaped", "100% of $x$ & more", `100\% of $x$ \& more`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := transpile(t, tc.input, transpiler.ExtAll)
			if !strings.Contains(got, tc.want) {
				t.Errorf("missing %q\ngot: %q", tc.want, got)
			}
			// Every escaped dollar in the output must be one the expectation asks for.
			if strings.Count(got, `\$`) != strings.Count(tc.want, `\$`) {
				t.Errorf("math delimiter was escaped\ngot: %q", got)
			}
		})
	}
}

func TestDisplayMath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"basic", "$$E=mc^2$$", `\[E=mc^2\]`},
		{"surrounding whitespace trimmed", `$$ \frac{a}{b} $$`, `\[\frac{a}{b}\]`},
		{"within a sentence", "Display: $$E=mc^2$$ here.", `Display: \[E=mc^2\] here.`},
		{"spanning lines", "$$\n\\sum_i x_i\n$$", `\[\sum_i x_i\]`},
		{"escaped dollar inside", `$$ a \$ b $$`, `\[a \$ b\]`},
		{"followed by inline math", "$$a$$ and $b$", `\[a\] and $b$`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := transpile(t, tc.input, transpiler.ExtAll)
			if !strings.Contains(got, tc.want) {
				t.Errorf("missing %q\ngot: %q", tc.want, got)
			}
			if strings.Contains(got, "$$") {
				t.Errorf("display delimiters leaked into output\ngot: %q", got)
			}
		})
	}
}

// TestNotMath lists dollar signs that must stay literal, escaped text under
// the Pandoc rules the extension follows.
func TestNotMath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single price", "Price is $10.00 today.", `Price is \$10.00 today.`},
		{"two prices", "$20,000 and $30,000", `\$20,000 and \$30,000`},
		{"price in parentheses", "Costs $5 (about $4.50).", `Costs \$5 (about \$4.50).`},
		{"space after opener", "$ x$", `\$ x\$`},
		{"space before closer", "$x $", `\$x \$`},
		{"newline before closer", "$x\n$", `\$x \$`},
		{"digit after closer", "$x$5", `\$x\$5`},
		{"escaped delimiters", `\$2+3=5\$`, `\$2+3=5\$`},
		{"empty display", "Only $$ alone", `Only \$\$ alone`},
		{"unterminated display", "$$x", `\$\$x`},
		{"lone dollar", "Lone $ sign", `Lone \$ sign`},
		{"unterminated inline", "$x and more", `\$x and more`},
		{"code span", "`$x$`", `\texttt{\$x\$}`},
		// Stricter than Pandoc: inline math stays on one line and never
		// swallows a code span, so a stray price cannot run on.
		{"spanning a line ending", "$x\ny$", `\$x y\$`},
		{"closer inside a later code span", "costs $5 and `$x`", `costs \$5 and \texttt{\$x}`},
		{"price then code span on next line", "or $5 (about $4.50).\nSee `$x$` here.", `or \$5 (about \$4.50). See \texttt{\$x\$} here.`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := transpile(t, tc.input, transpiler.ExtAll)
			if !strings.Contains(got, tc.want) {
				t.Errorf("missing %q\ngot: %q", tc.want, got)
			}
			if strings.Count(got, "$") != strings.Count(got, `\$`) {
				t.Errorf("unescaped dollar sign in output\ngot: %q", got)
			}
		})
	}
}

func TestMathInsideCodeBlocksIsVerbatim(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"fenced", "```\n$x$ and $$y$$\n```"},
		{"indented", "    $x$ and $$y$$"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := transpile(t, tc.input, transpiler.ExtAll)
			if !strings.Contains(got, "$x$ and $$y$$\n") {
				t.Errorf("code content altered\ngot: %q", got)
			}
			if strings.Contains(got, `\[`) {
				t.Errorf("display math produced inside code\ngot: %q", got)
			}
		})
	}
}

func TestMathExtensionDisabled(t *testing.T) {
	for name, ext := range map[string]transpiler.Extension{
		"all-but-math": transpiler.ExtAll &^ transpiler.ExtMath,
		"tables-only":  transpiler.ExtTables,
	} {
		t.Run(name, func(t *testing.T) {
			got := transpile(t, "$2+3=5$ and $$E$$", ext)
			if want := `\$2+3=5\$ and \$\$E\$\$`; !strings.Contains(got, want) {
				t.Errorf("missing %q\ngot: %q", want, got)
			}
		})
	}
}

func TestMathStandalonePreamble(t *testing.T) {
	for _, src := range []string{"inline $x$", "display $$x$$"} {
		got, err := transpiler.TranspileString(src, transpiler.Options{Standalone: true})
		if err != nil {
			t.Fatal(err)
		}
		assertInOrder(t, got, `\usepackage{amsmath}`, `\usepackage{amssymb}`, `\begin{document}`)
	}
	got, err := transpiler.TranspileString("no math here", transpiler.Options{Standalone: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "amsmath") || strings.Contains(got, "amssymb") {
		t.Errorf("math packages included without math\ngot:\n%s", got)
	}
}

// ── Heading levels 4-6 (issue #1) ────────────────────────────────────────────

// TestHeadingLevels5And6Issue01 reproduces the exact report from issue #1.
// \subparagraph is a run-in heading; without the \mbox{} the level-6 heading
// is typeset on the same line.
func TestHeadingLevels5And6Issue01(t *testing.T) {
	got := transpile(t, "##### Header 5\n###### Header 6\n", transpiler.ExtAll)
	want := "\\subparagraph{Header 5}\\mbox{}\n\n\\noindent\\textbf{Header 6}\n\n"
	if got != want {
		t.Errorf("want exactly %q\ngot:        %q", want, got)
	}
}

// TestHeadingCommandsPerLevel pins the full output of a heading followed by a
// paragraph at every level: the command, the run-in guard where needed, and
// the blank line that separates the heading block from the body.
func TestHeadingCommandsPerLevel(t *testing.T) {
	tests := []struct {
		level int
		want  string
	}{
		{1, "\\section{Title}\n\n"},
		{2, "\\subsection{Title}\n\n"},
		{3, "\\subsubsection{Title}\n\n"},
		{4, "\\paragraph{Title}\\mbox{}\n\n"},
		{5, "\\subparagraph{Title}\\mbox{}\n\n"},
		{6, "\\noindent\\textbf{Title}\n\n"},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("level%d", tc.level), func(t *testing.T) {
			src := strings.Repeat("#", tc.level) + " Title\nBody text."
			got := transpile(t, src, transpiler.ExtAll)
			if want := tc.want + "Body text.\n\n"; got != want {
				t.Errorf("want exactly %q\ngot:        %q", want, got)
			}
		})
	}
}

// TestDeepHeadingFollowedByBlock checks the guard in front of every kind of
// block that can directly follow a run-in heading.
func TestDeepHeadingFollowedByBlock(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"paragraph", "##### H\nText", "\\subparagraph{H}\\mbox{}\n\nText"},
		{"bullet list", "#### H\n- a", "\\paragraph{H}\\mbox{}\n\n\\begin{itemize}"},
		{"ordered list", "#### H\n1. a", "\\paragraph{H}\\mbox{}\n\n\\begin{enumerate}"},
		{"block quote", "#### H\n> q", "\\paragraph{H}\\mbox{}\n\n\\begin{quote}"},
		{"fenced code", "#### H\n```\nx\n```", "\\paragraph{H}\\mbox{}\n\n\\begin{lstlisting}"},
		{"table", "#### H\n| a |\n|---|", "\\paragraph{H}\\mbox{}\n\n\\begin{tabular}"},
		{"thematic break", "#### H\n---", "\\paragraph{H}\\mbox{}\n\n"},
		{"chain of deep headings", "#### H4\n##### H5\n###### H6",
			"\\paragraph{H4}\\mbox{}\n\n\\subparagraph{H5}\\mbox{}\n\n\\noindent\\textbf{H6}\n\n"},
		{"inside block quote", "> ##### H\n> text",
			"\\begin{quote}\n\\subparagraph{H}\\mbox{}\n\ntext\n\n\\end{quote}"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := transpile(t, tc.input, transpiler.ExtAll)
			if !strings.Contains(got, tc.want) {
				t.Errorf("missing %q\ngot: %q", tc.want, got)
			}
		})
	}
}

func TestDeepHeadingInlineContent(t *testing.T) {
	got := transpile(t, "#### *x* and $y$ and `c`\n\n###### **b** only", transpiler.ExtAll)
	for _, want := range []string{
		"\\paragraph{\\textit{x} and $y$ and \\texttt{c}}\\mbox{}\n\n",
		"\\noindent\\textbf{\\textbf{b} only}\n\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q\ngot: %q", want, got)
		}
	}
}

// TestShallowHeadingsHaveNoRunInGuard: \section through \subsubsection and
// both setext forms are display headings and must stay untouched.
func TestShallowHeadingsHaveNoRunInGuard(t *testing.T) {
	got := transpile(t, "# A\n## B\n### C\nD\n===\n\nE\n---\ntext", transpiler.ExtAll)
	assertInOrder(t, got, "\\section{A}\n\n", "\\subsection{B}\n\n", "\\subsubsection{C}\n\n", "\\section{D}\n\n", "\\subsection{E}\n\ntext")
	if strings.Contains(got, `\mbox`) || strings.Contains(got, `\noindent`) {
		t.Errorf("run-in guard applied to a display heading\ngot: %q", got)
	}
}

// ── Emphasis delimiter runs (issue #3) ───────────────────────────────────────

// TestEmphasisDelimiterRunsIssue03 reproduces the exact report from issue #3.
func TestEmphasisDelimiterRunsIssue03(t *testing.T) {
	got := transpile(t, "***testing****without***spaces**\n", transpiler.ExtAll)
	want := "\\textbf{\\textit{testing}}\\textit{without}\\textbf{spaces}\n\n"
	if got != want {
		t.Errorf("want exactly %q\ngot:        %q", want, got)
	}
}

// TestEmphasisCommonMarkSpec runs the examples of CommonMark 6.2 (emphasis
// and strong emphasis, rules 1 to 17) through the transpiler and compares the
// full output. Expected values are the spec's HTML translated to LaTeX, with
// `***x***` collapsed to \textbf{\textit{x}} as documented in the README.
func TestEmphasisCommonMarkSpec(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		// Rule 1: * opens when left-flanking.
		{"*foo bar*", `\textit{foo bar}`},
		{"a * foo bar*", `a * foo bar*`},
		{`a*"foo"*`, `a*"foo"*`},
		{"x * a *", `x * a *`},
		{"foo*bar*", `foo\textit{bar}`},
		{"5*6*78", `5\textit{6}78`},
		// Rule 2: _ opens when left-flanking and not inside a word.
		{"_foo bar_", `\textit{foo bar}`},
		{"_ foo bar_", `\_ foo bar\_`},
		{`a_"foo"_`, `a\_"foo"\_`},
		{"foo_bar_", `foo\_bar\_`},
		{"5_6_78", `5\_6\_78`},
		{"пристаням_стремятся_", `пристаням\_стремятся\_`},
		{`aa_"bb"_cc`, `aa\_"bb"\_cc`},
		{"foo-_(bar)_", `foo-\textit{(bar)}`},
		// Rule 3: * closes when right-flanking.
		{"_foo*", `\_foo*`},
		{"*foo bar *", `*foo bar *`},
		{"*(*foo)", `*(*foo)`},
		{"*(*foo*)*", `\textit{(\textit{foo})}`},
		{"*foo*bar", `\textit{foo}bar`},
		// Rule 4: _ closes when right-flanking and not inside a word.
		{"_foo bar _", `\_foo bar \_`},
		{"_(_foo)", `\_(\_foo)`},
		{"_(_foo_)_", `\textit{(\textit{foo})}`},
		{"_foo_bar", `\_foo\_bar`},
		{"_пристаням_стремятся", `\_пристаням\_стремятся`},
		{"_foo_bar_baz_", `\textit{foo\_bar\_baz}`},
		{"_(bar)_.", `\textit{(bar)}.`},
		// Rules 5 and 6: ** and __ open.
		{"**foo bar**", `\textbf{foo bar}`},
		{"** foo bar**", `** foo bar**`},
		{`a**"foo"**`, `a**"foo"**`},
		{"foo**bar**", `foo\textbf{bar}`},
		{"__foo bar__", `\textbf{foo bar}`},
		{"__ foo bar__", `\_\_ foo bar\_\_`},
		{`a__"foo"__`, `a\_\_"foo"\_\_`},
		{"foo__bar__", `foo\_\_bar\_\_`},
		{"5__6__78", `5\_\_6\_\_78`},
		{"__foo, __bar__, baz__", `\textbf{foo, \textbf{bar}, baz}`},
		{"foo-__(bar)__", `foo-\textbf{(bar)}`},
		// Rules 7 and 8: ** and __ close.
		{"**foo bar **", `**foo bar **`},
		{"**(**foo)", `**(**foo)`},
		{"*(**foo**)*", `\textit{(\textbf{foo})}`},
		{"**Gomphocarpus (*Gomphocarpus physocarpus*, syn.\n*Asclepias physocarpa*)**",
			`\textbf{Gomphocarpus (\textit{Gomphocarpus physocarpus}, syn. \textit{Asclepias physocarpa})}`},
		{`**foo "*bar*" foo**`, `\textbf{foo "\textit{bar}" foo}`},
		{"**foo**bar", `\textbf{foo}bar`},
		{"__foo bar __", `\_\_foo bar \_\_`},
		{"__(__foo)", `\_\_(\_\_foo)`},
		{"_(__foo__)_", `\textit{(\textbf{foo})}`},
		{"__foo__bar", `\_\_foo\_\_bar`},
		{"__foo__bar__baz__", `\textbf{foo\_\_bar\_\_baz}`},
		{"__(bar)__.", `\textbf{(bar)}.`},
		// Rule 9: nesting inside emphasis.
		{"*foo [bar](/url)*", `\textit{foo \href{/url}{bar}}`},
		{"*foo\nbar*", `\textit{foo bar}`},
		{"_foo __bar__ baz_", `\textit{foo \textbf{bar} baz}`},
		{"_foo _bar_ baz_", `\textit{foo \textit{bar} baz}`},
		{"__foo_ bar_", `\textit{\textit{foo} bar}`},
		{"*foo *bar**", `\textit{foo \textit{bar}}`},
		{"*foo **bar** baz*", `\textit{foo \textbf{bar} baz}`},
		{"*foo**bar**baz*", `\textit{foo\textbf{bar}baz}`},
		{"*foo**bar*", `\textit{foo**bar}`},
		{"***foo** bar*", `\textit{\textbf{foo} bar}`},
		{"*foo **bar***", `\textit{foo \textbf{bar}}`},
		{"*foo**bar***", `\textit{foo\textbf{bar}}`},
		{"foo***bar***baz", `foo\textbf{\textit{bar}}baz`},
		{"foo******bar*********baz", `foo\textbf{\textbf{\textbf{bar}}}***baz`},
		{"*foo **bar *baz* bim** bop*", `\textit{foo \textbf{bar \textit{baz} bim} bop}`},
		{"*foo [*bar*](/url)*", `\textit{foo \href{/url}{\textit{bar}}}`},
		{"** is not an empty emphasis", `** is not an empty emphasis`},
		{"**** is not an empty strong emphasis", `**** is not an empty strong emphasis`},
		// Rule 10: nesting inside strong emphasis.
		{"**foo [bar](/url)**", `\textbf{foo \href{/url}{bar}}`},
		{"**foo\nbar**", `\textbf{foo bar}`},
		{"__foo _bar_ baz__", `\textbf{foo \textit{bar} baz}`},
		{"__foo __bar__ baz__", `\textbf{foo \textbf{bar} baz}`},
		{"____foo__ bar__", `\textbf{\textbf{foo} bar}`},
		{"**foo **bar****", `\textbf{foo \textbf{bar}}`},
		{"**foo *bar* baz**", `\textbf{foo \textit{bar} baz}`},
		{"**foo*bar*baz**", `\textbf{foo\textit{bar}baz}`},
		{"***foo* bar**", `\textbf{\textit{foo} bar}`},
		{"**foo *bar***", `\textbf{foo \textit{bar}}`},
		{"**foo *bar **baz**\nbim* bop**", `\textbf{foo \textit{bar \textbf{baz} bim} bop}`},
		{"**foo [*bar*](/url)**", `\textbf{foo \href{/url}{\textit{bar}}}`},
		{"__ is not an empty emphasis", `\_\_ is not an empty emphasis`},
		{"____ is not an empty strong emphasis", `\_\_\_\_ is not an empty strong emphasis`},
		// Rules 11 and 12: unmatched delimiters stay literal.
		{"foo ***", `foo ***`},
		{`foo *\**`, `foo \textit{*}`},
		{"foo *_*", `foo \textit{\_}`},
		{"foo *****", `foo *****`},
		{`foo **\***`, `foo \textbf{*}`},
		{"foo **_**", `foo \textbf{\_}`},
		{"**foo*", `*\textit{foo}`},
		{"*foo**", `\textit{foo}*`},
		{"***foo**", `*\textbf{foo}`},
		{"****foo*", `***\textit{foo}`},
		{"**foo***", `\textbf{foo}*`},
		{"*foo****", `\textit{foo}***`},
		{"foo ___", `foo \_\_\_`},
		{`foo _\__`, `foo \textit{\_}`},
		{"foo _*_", `foo \textit{*}`},
		{"foo _____", `foo \_\_\_\_\_`},
		{`foo __\___`, `foo \textbf{\_}`},
		{"foo __*__", `foo \textbf{*}`},
		{"__foo_", `\_\textit{foo}`},
		{"_foo__", `\textit{foo}\_`},
		{"___foo__", `\_\textbf{foo}`},
		{"____foo_", `\_\_\_\textit{foo}`},
		{"__foo___", `\textbf{foo}\_`},
		{"_foo____", `\textit{foo}\_\_\_`},
		// Rules 13 and 14: nested strong and emphasis.
		{"**foo**", `\textbf{foo}`},
		{"*_foo_*", `\textit{\textit{foo}}`},
		{"__foo__", `\textbf{foo}`},
		{"_*foo*_", `\textit{\textit{foo}}`},
		{"****foo****", `\textbf{\textbf{foo}}`},
		{"____foo____", `\textbf{\textbf{foo}}`},
		{"******foo******", `\textbf{\textbf{\textbf{foo}}}`},
		{"***foo***", `\textbf{\textit{foo}}`},
		{"_____foo_____", `\textbf{\textit{\textbf{foo}}}`},
		// Rules 15 and 16: overlapping and unbalanced runs.
		{"*foo _bar* baz_", `\textit{foo \_bar} baz\_`},
		{"*foo __bar *baz bim__ bam*", `\textit{foo \textbf{bar *baz bim} bam}`},
		{"**foo **bar baz**", `**foo \textbf{bar baz}`},
		{"*foo *bar baz*", `*foo \textit{bar baz}`},
		// Rule 17: code spans and links take precedence.
		{"*[bar*](/url)", `*\href{/url}{bar*}`},
		{"_foo [bar_](/url)", `\_foo \href{/url}{bar\_}`},
		{"*a `*`*", `\textit{a \texttt{*}}`},
		{"_a `_`_", `\textit{a \texttt{\_}}`},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got := transpile(t, tc.in, transpiler.ExtAll)
			if want := tc.want + "\n\n"; got != want {
				t.Errorf("input %q\nwant %q\ngot  %q", tc.in, want, got)
			}
		})
	}
}

// TestEmphasisWithOtherInlineConstructs checks emphasis around and inside the
// other inline nodes, where delimiter scopes and line-break handling meet.
func TestEmphasisWithOtherInlineConstructs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"hard break inside emphasis", "*a  \nb*", "\\textit{a\\\\\nb}"},
		{"strikethrough inside emphasis", "*~~x~~*", `\textit{\sout{x}}`},
		{"emphasis inside strikethrough", "~~*x*~~", `\sout{\textit{x}}`},
		{"math inside emphasis", "*see $x^2$ here*", `\textit{see $x^2$ here}`},
		{"escaped delimiter inside", `*foo\*bar*`, `\textit{foo*bar}`},
		{"entity next to delimiter", "*&amp;*", `\textit{\&}`},
		{"emphasis in heading", "## *x* and **y**", `\subsection{\textit{x} and \textbf{y}}`},
		{"emphasis in list item", "- **b** and _i_", `\item \textbf{b} and \textit{i}`},
		{"emphasis in table cell", "| a | b |\n|---|---|\n| *x* | __y__ |", `\textit{x} & \textbf{y} \\`},
		{"emphasis in block quote", "> *q*", "\\begin{quote}\n\\textit{q}\n"},
		{"image alt text not emphasised", "![*alt*](i.png)", `\includegraphics{i.png}`},
		{"link text scope", "*foo [bar*](/u) baz*", `\textit{foo \href{/u}{bar*} baz}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := transpile(t, tc.input, transpiler.ExtAll)
			if !strings.Contains(got, tc.want) {
				t.Errorf("missing %q\ngot: %q", tc.want, got)
			}
		})
	}
}
