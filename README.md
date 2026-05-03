# MarkTeX

A Markdown-to-LaTeX transpiler written in Go. It parses Markdown source into a typed AST and walks that tree to produce LaTeX output, with optional custom extensions for citations, labeled figures, and labeled table floats.

## License

This project is licensed under GPL-3.0-or-later. See [`LICENSE`](./LICENSE) for details.

## Project name and branding

The project name, logo, and branding are not licensed for use in a way that suggests a modified version is the official project.

Modified versions should use a different name unless written permission is granted.

See [`NOTICE`](./NOTICE) for details.

---

## Install MarkTeX as a command-line tool:

1. Check your Go installation: [https://go.dev/doc/tutorial/compile-install](https://go.dev/doc/tutorial/compile-install)
    1. Add binary path to $PATH e.g. add to the file `~/.profile` the command `export PATH=$PATH:~/go/bin`
    2. Set GOBIN path, e.g. with the command `go env -w GOBIN=~/go/bin`
   
2. Check repository structure: [https://go.dev/doc/modules/layout](https://go.dev/doc/modules/layout)
    1. Check the repository for version and release tags. Can be requested by @latest @v1.0.3 suffixes.

```bash
go install github.com/andreaswillibaldweber/marktex/marktex
```

## Build and run

```sh
# Build the binary
make build           # output: bin/marktex

# Run tests
make test

# Read from stdin, write to stdout
echo "# Hello" | ./bin/marktex

# Convert a file to a complete LaTeX document
./bin/marktex -standalone -o output.tex input.md
```

**Flags**

| Flag | Default | Description |
|---|---|---|
| `-o <file>` | stdout | Write output to file |
| `-standalone` | off | Emit a complete `\documentclass` document |
| `-strict` | off | Exit non-zero on unsupported constructs (raw HTML, etc.) |
| `-ext <list>` | `all` | Comma-separated extensions to enable: `tables`, `strikethrough`, `citations`, `figures`, `tableFloat`, `all`, `none` |
| `-no-ext <list>` | — | Comma-separated extensions to disable |
| `-version` | — | Print version and exit |

---

## Markdown syntax and LaTeX output

### Headings

| Markdown | LaTeX |
|---|---|
| `# Heading` | `\section{Heading}` |
| `## Heading` | `\subsection{Heading}` |
| `### Heading` | `\subsubsection{Heading}` |
| `#### Heading` | `\paragraph{Heading}` |
| `##### Heading` | `\subparagraph{Heading}` |
| `###### Heading` | `\textbf{Heading}` |

Setext headings (`===` underline → `\section`, `---` underline → `\subsection`) are also supported.

---

### Inline formatting

| Markdown | LaTeX |
|---|---|
| `**bold**` | `\textbf{bold}` |
| `*italic*` or `_italic_` | `\textit{italic}` |
| `***bold-italic***` | `\textbf{\textit{bold-italic}}` |
| `` `code` `` | `\texttt{code}` |
| `~~strikethrough~~` | `\sout{strikethrough}` (requires `ulem`) |
| `\*escaped\*` | literal `*` |
| `&amp;`, `&#65;`, `&#x41;` | decoded Unicode character |

---

### Links and images

| Markdown | LaTeX |
|---|---|
| `[text](https://example.com)` | `\href{https://example.com}{text}` |
| `![alt](image.png)` (sole content of a paragraph) | `\begin{figure}[h] \centering \includegraphics{image.png} \caption{alt} \end{figure}` |
| `![alt](image.png)` (inline, mixed with text) | `\includegraphics{image.png}` |

---

### Code blocks

| Markdown | LaTeX |
|---|---|
| ` ```python … ``` ` | `\begin{lstlisting}[language=python] … \end{lstlisting}` |
| ` ``` … ``` ` (no language) | `\begin{lstlisting} … \end{lstlisting}` |
| Four-space indented block | `\begin{verbatim} … \end{verbatim}` |

Code block content is emitted verbatim — no character escaping is applied inside listings or verbatim environments.

---

### Block elements

| Markdown | LaTeX |
|---|---|
| Blank-line-separated paragraph | `text\n\n` |
| `> blockquote` | `\begin{quote} … \end{quote}` |
| `---`, `***`, `___` | `\noindent\rule{\linewidth}{0.4pt}` |
| Soft line break (single newline in source) | space |
| Hard line break (two trailing spaces or `\` + newline) | `\\` |

---

### Lists

| Markdown | LaTeX |
|---|---|
| `- item` / `* item` / `+ item` | `\begin{itemize} \item … \end{itemize}` |
| `1. item` | `\begin{enumerate} \item … \end{enumerate}` |
| Ordered list starting at `N` | `\begin{enumerate}[start=N] …` |
| Tight list (no blank lines between items) | `[noitemsep]` option added |

Requires the `enumitem` package (automatically included in standalone mode).

---

### Tables (GFM)

Extension: `tables` (enabled by default).

```markdown
| Name  | Score | Grade |
|:------|------:|:-----:|
| Alice | 95    | A     |
```

Produces:

```latex
\begin{tabular}{lrc}
\toprule
Name & Score & Grade \\
\midrule
Alice & 95 & A \\
\bottomrule
\end{tabular}
```

Column alignment is derived from the separator row (`:---` = left, `---:` = right, `:---:` = center). Output uses `booktabs` style (`\toprule`, `\midrule`, `\bottomrule`).

---

### Special characters

The following characters are automatically escaped in all regular text contexts:

| Character | LaTeX output |
|---|---|
| `\` | `\textbackslash{}` |
| `{` | `\{` |
| `}` | `\}` |
| `$` | `\$` |
| `&` | `\&` |
| `#` | `\#` |
| `%` | `\%` |
| `_` | `\_` |
| `^` | `\textasciicircum{}` |
| `~` | `\textasciitilde{}` |

Characters inside code spans and code blocks are **not** escaped.

---

## Extensions

### Citations (`-ext citations`)

Define short citation keys in a `---` block and reference them inline.

**Markdown:**

```markdown
---
C#01:doe2023survey
C#02:smith2024ml
---

See [C#01] for the original paper and [C#02] for the follow-up.
```

**LaTeX:**

```latex
See~\cite{doe2023survey} for the original paper and~\cite{smith2024ml} for the follow-up.
```

The space immediately before a citation reference is replaced by `~` (non-breaking space).

---

### Labeled figures (`-ext figures`)

Define figure label keys in a `---` block, then use the extended image syntax and cross-reference figures inline.

**Markdown:**

```markdown
---
F#01:architecture_diagram
F#02:results_plot
---

See Figure [F#01] for the architecture.

![F#01:0.8:ht][System architecture.](figures/arch.png)

Results are shown in Figure [F#02].

![F#02:0.5:h][Benchmark results.](figures/results.png)
```

**LaTeX:**

```latex
See Figure~\ref{fig:architecture_diagram} for the architecture.

\begin{figure}[ht]
  \centering
  \includegraphics[width=0.8\linewidth]{figures/arch.png}
  \caption{System architecture.}
  \label{fig:architecture_diagram}
\end{figure}

Results are shown in Figure~\ref{fig:results_plot}.

\begin{figure}[h]
  \centering
  \includegraphics[width=0.5\linewidth]{figures/results.png}
  \caption{Benchmark results.}
  \label{fig:results_plot}
\end{figure}
```

**Extended image syntax:** `![F#key:width:placement][Caption.](path)`

| Part | Description | Example |
|---|---|---|
| `F#key` | Figure key from the definition block | `F#01` |
| `width` | Width as a fraction of `\linewidth` | `0.8` |
| `placement` | LaTeX float placement specifier | `h`, `ht`, `ht!`, `hb!` |
| `Caption.` | Caption text | `System architecture.` |
| `path` | Image file path | `figures/arch.png` |

---

### Labeled table floats (`-ext tableFloat`)

Define table label keys in a `---` block, then precede a GFM table with metadata rows and cross-reference tables inline.

**Markdown:**

```markdown
---
T#01:comparison_table
---

See Table [T#01] for the comparison.

|- T#01:ht -|
|- Comparison of methods. -|
| Method | Accuracy | Speed |
|:-------|:--------:|------:|
| Ours   | 95.2     | 12 ms |
| Baseline | 91.0   | 8 ms  |
```

**LaTeX:**

```latex
See Table~\ref{tab:comparison_table} for the comparison.

\begin{table}[ht]
\centering
\caption{Comparison of methods.}
\begin{tabular}{lrc}
\toprule
Method & Accuracy & Speed \\
\midrule
Ours & 95.2 & 12 ms \\
Baseline & 91.0 & 8 ms \\
\bottomrule
\end{tabular}
\label{tab:comparison_table}
\end{table}
```

**Table metadata rows** (`|- … -|`) must appear immediately before the GFM table header row with no blank lines between them.

| Row | Format | Description |
|---|---|---|
| Key and placement | `\|- T#key:placement -\|` | Required. Associates the table with a label key and sets the float placement. |
| Caption | `\|- Caption text. -\|` | Optional. Any row without a `T#` prefix is used as the caption. |

**Placement specifiers** follow standard LaTeX float syntax: `h`, `t`, `b`, `p`, `H`, and combinations such as `ht`, `h!`, `ht!`, `hb!`.

---

## Standalone mode

With `-standalone`, the output is wrapped in a complete LaTeX document. Only the packages actually needed for the content present are included.

```latex
\documentclass{article}

\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{lmodern}
\usepackage[a4paper, margin=2.5cm]{geometry}
\usepackage[hidelinks, unicode]{hyperref}
% listings, enumitem, ulem, graphicx, booktabs added when needed

\begin{document}
% ... your content ...
\end{document}
```

---

## Architecture

The transpiler is structured as a three-stage pipeline with a clean AST boundary between stages:

```
Markdown source
      |
      v
  [ Parser ]          internal/parser/
      |                 block.go   — block-level state machine
      |                 inline.go  — inline parser (emphasis, links, etc.)
      v
    [ AST ]           internal/ast/
      |                 block.go   — block node types
      |                 inline.go  — inline node types
      v
  [ Generator ]       internal/generator/
                        generator.go — LaTeX visitor (implements Visitor)
                        escape.go    — special-character escaping
                        templates.go — standalone preamble
```

Traversal is driven by `visitor.Walk` (internal/visitor/walk.go), which dispatches to `Enter*`/`Leave*` methods on any type that implements the `Visitor` interface. `BaseVisitor` provides no-op defaults so a new backend only needs to override the node types it cares about.

Adding a new Markdown construct follows a five-step pattern: add AST node, add visitor methods, add no-ops to BaseVisitor, add dispatch cases to Walk, implement in the generator.
