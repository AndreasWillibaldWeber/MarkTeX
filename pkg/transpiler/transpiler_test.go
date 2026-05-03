// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package transpiler_test

import (
	"marktex/pkg/transpiler"
	"strings"
	"testing"
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
		{"h4", "#### Heading", `\paragraph{Heading}`},
		{"h5", "##### Heading", `\subparagraph{Heading}`},
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
		{"hb",  "---\nF#01:l\n---\n\n![F#01:0.8:hb][Cap.](a.png)"},
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
