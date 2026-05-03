# MarkTeX

A Markdown-to-LaTeX transpiler written in Go. It parses Markdown source with optional extensions like citations, labeled figures, labeled table floats, and document metadata to produce LaTeX source code.

## License

This project is licensed under LGPL-3.0-or-later. See [`LICENSE`](./LICENSE) for details.

## Project name and branding

The project name, logo, and branding are not licensed for use in a way that suggests a modified version is the official project. Modified versions should use a different name unless written permission is granted. See [`NOTICE`](./NOTICE) for details.

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

# Convert a file to a complete standalone LaTeX document
./bin/marktex -standalone -o output.tex input.md
```

**Flags**

| Flag | Default | Description |
|---|---|---|
| `-o <file>` | stdout | Write output to file |
| `-standalone` | off | Emit a complete `\documentclass` document |
| `-strict` | off | Exit non-zero on unsupported constructs (raw HTML, etc.) |
| `-ext <list>` | `all` | Comma-separated extensions to enable (see below) |
| `-no-ext <list>` | — | Comma-separated extensions to disable |
| `-version` | — | Print version and exit |

**Available extensions for `-ext`**

| Name | Description |
|---|---|
| `tables` | GFM pipe tables |
| `strikethrough` | GFM `~~strikethrough~~` |
| `citations` | `C#key` citation references → `\cite{}` |
| `figures` | `F#key` labeled figure floats → `\ref{fig:…}` |
| `tableFloat` | `T#key` labeled table floats → `\ref{tab:…}` |
| `documentMeta` | Document metadata and front-matter commands |
| `all` | All of the above (default) |
| `none` | Base Markdown only |

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
| `![alt](image.png)` (sole paragraph content) | `\begin{figure}[h] \centering \includegraphics{image.png} \caption{alt} \end{figure}` |
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

All extensions share the same **definition block** syntax. A single `---…---` block can contain any mix of `C#`, `F#`, `T#`, and document-metadata entries:

```markdown
---
C#01:doe2023survey
F#01:architecture_diagram
T#01:results_table
Author: Firstname Surname
Title: My Document
MakeTitlePage
MakeTOC
---
```

The block is silently consumed and produces no output in the document. It only populates the reference registry and document metadata. Definition blocks may appear anywhere in the document — references are pre-scanned before parsing begins, so a block at the bottom of the file still resolves references used earlier in the text.

---

### Citations (`-ext citations`)

**Definition:** `C#key:bibtexidentifier`

**Reference:** `[C#key]` → `~\cite{bibtexidentifier}`

```markdown
---
C#01:doe2023survey
C#02:smith2024ml
---

See [C#01] for the original paper and [C#02] for the follow-up.
```

```latex
See~\cite{doe2023survey} for the original paper and~\cite{smith2024ml} for the follow-up.
```

The space immediately before a citation reference is replaced by `~` (non-breaking space).

---

### Labeled figures (`-ext figures`)

**Definition:** `F#key:labelsuffix`

**Reference:** `[F#key]` → `~\ref{fig:labelsuffix}`

**Figure image syntax:** `![F#key:width:placement][Caption.](path)`

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

**Extended image syntax fields**

| Field | Description | Example |
|---|---|---|
| `F#key` | Figure key from the definition block | `F#01` |
| `width` | Width as a fraction of `\linewidth` | `0.8` |
| `placement` | LaTeX float placement specifier | `h`, `ht`, `ht!`, `hb!` |
| `Caption.` | Caption text (may contain inline refs such as `[C#01]`) | `See [C#01].` |
| `path` | Image file path | `figures/arch.png` |

---

### Labeled table floats (`-ext tableFloat`)

**Definition:** `T#key:labelsuffix`

**Reference:** `[T#key]` → `~\ref{tab:labelsuffix}`

**Table float syntax:** one or two `|- … -|` metadata rows placed immediately before the GFM table header with no blank lines in between.

```markdown
---
T#01:comparison_table
---

See Table [T#01] for the comparison.

|- T#01:ht -|
|- Comparison of methods. -|
| Method   | Accuracy | Speed  |
|:---------|:--------:|-------:|
| Ours     | 95.2     | 12 ms  |
| Baseline | 91.0     | 8 ms   |
```

```latex
See Table~\ref{tab:comparison_table} for the comparison.

\begin{table}[ht]
\centering
\caption{Comparison of methods.}
\begin{tabular}{lcr}
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

**Metadata row formats**

| Row | Format | Description |
|---|---|---|
| Key and placement | `\|- T#key:placement -\|` | Required. Links the table to its label and sets the float specifier. |
| Caption | `\|- Caption text. -\|` | Optional. Any row without a `T#` prefix is used as the caption. May contain inline refs. |

**Placement specifiers** follow standard LaTeX float syntax: `h`, `t`, `b`, `p`, `H`, and combinations such as `ht`, `h!`, `ht!`, `hb!`.

---

### Document metadata (`-ext documentMeta`)

Provides document-level metadata and front-matter commands inside the standard definition block. All entries are **silently ignored in fragment mode** and only applied when `-standalone` is active.

```markdown
---
Author: Firstname Middlename Surname
Title: Document Title
Subtitle: A descriptive subtitle
Date: 2025-05-01
MakeTitlePage
MakeTOC
MakeLOF
MakeLOT
MakeLOL
---
```

**Metadata entries** (emitted in the preamble, before `\begin{document}`)

| Entry | LaTeX output | Notes |
|---|---|---|
| `Title: …` | `\title{…}` | |
| `Title: …` + `Subtitle: …` | `\title{…\\[0.5em]{\normalfont\large …}}` | Subtitle encoded inside `\title`; no extra package needed |
| `Author: …` | `\author{…}` | |
| `Date: …` | `\date{…}` | Omit to use `\today` |

**Front-matter commands** (emitted at the start of the document body, always in this fixed order regardless of their order in the definition block)

| Entry | LaTeX output | Notes |
|---|---|---|
| `MakeTitlePage` | `\maketitle` | |
| `MakeTOC` | `\tableofcontents` | |
| `MakeLOF` | `\listoffigures` | |
| `MakeLOT` | `\listoftables` | |
| `MakeLOL` | `\lstlistoflistings` | Also forces inclusion of the `listings` package |

**Example standalone output**

```latex
\documentclass{article}
% … packages …
\title{Document Title\\[0.5em]{\normalfont\large A descriptive subtitle}}
\author{Firstname Middlename Surname}
\date{2025-05-01}

\begin{document}
\maketitle

\tableofcontents

\listoffigures

\listoftables

\lstlistoflistings

% … document content …

\end{document}
```

---

### Inline references in captions

Citation and cross-reference syntax resolves correctly inside figure captions and table captions:

```markdown
---
C#01:doe2023
F#01:fig1
T#01:tab1
---

![F#01:0.8:h][Architecture from [C#01].](figures/arch.png)

|- T#01:ht -|
|- Results reported in [C#01], see also Figure [F#01]. -|
| A | B |
|---|---|
| 1 | 2 |
```

```latex
\begin{figure}[h]
  \centering
  \includegraphics[width=0.8\linewidth]{figures/arch.png}
  \caption{Architecture from~\cite{doe2023}.}
  \label{fig:fig1}
\end{figure}

\begin{table}[ht]
\centering
\caption{Results reported in~\cite{doe2023}, see also Figure~\ref{fig:fig1}.}
…
\label{tab:tab1}
\end{table}
```

---

## Standalone mode

With `-standalone`, the output is wrapped in a complete LaTeX document. Only the packages actually needed for the content present are included in the preamble.

```latex
\documentclass{article}

\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{lmodern}
\usepackage[a4paper, margin=2.5cm]{geometry}
\usepackage[hidelinks, unicode]{hyperref}
% listings  — fenced code blocks or MakeLOL
% enumitem  — lists
% ulem      — strikethrough
% graphicx  — images
% booktabs  — tables

\begin{document}
% … content …
\end{document}
```

---

## Architecture

The transpiler is a three-stage pipeline with a clean AST boundary between stages:

```
Markdown source
      |
      v
  [ Parser ]          internal/parser/
      |                 parser.go   — entry point, extension flags, pre-scan
      |                 block.go    — block-level state machine
      |                 inline.go   — inline parser (emphasis, links, refs)
      v
    [ AST ]           internal/ast/
      |                 node.go     — Node interface, NodeType enum
      |                 block.go    — block node types + DocumentMeta
      |                 inline.go   — inline node types
      v
  [ Generator ]       internal/generator/
                        generator.go  — LaTeX visitor (implements Visitor)
                        escape.go     — special-character escaping
                        templates.go  — standalone preamble
```

Traversal is driven by `visitor.Walk` (`internal/visitor/walk.go`), which dispatches `Enter*`/`Leave*` calls to any type implementing the `Visitor` interface. `BaseVisitor` provides no-op defaults so a new backend only needs to override the nodes it cares about.

**Adding a new Markdown construct** follows a five-step pattern:

1. Add a struct and `NodeType` constant in `internal/ast/`
2. Add `Enter*/Leave*` methods to `visitor.Visitor`
3. Add no-op implementations to `visitor.BaseVisitor`
4. Add dispatch cases to `visitor.Walk`
5. Implement the methods in `internal/generator/generator.go`

**Adding a new definition-block entry type** (like `C#`, `F#`, `T#`) follows an additional pattern:

1. Add a `parseXxxEntryBytes` helper in `internal/parser/parser.go`
2. Register it in `preScan` and `parseDefinitionBlock`
3. Register it in `isAnyDefinitionEntry` in `internal/parser/block.go`
4. Store the resolved value in `defRefs` and thread it down to the inline parser
