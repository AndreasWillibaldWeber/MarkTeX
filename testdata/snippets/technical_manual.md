---
C#01:go2023spec
C#02:latex2023guide
F#01:fig_arch
F#02:fig_flow
F#03:fig_output
T#01:tab_types
T#02:tab_escaping
T#03:tab_extensions
---

# MarkTeX Technical Reference

## Overview

MarkTeX converts Markdown to LaTeX via a typed AST. All source files are UTF-8; output targets `pdflatex`, `xelatex`, or `lualatex` with the packages listed in Table [T#03].

### Design Goals

1. **Correctness**: special characters are always escaped (see Table [T#02]).
2. **Extensibility**: new constructs follow the five-step pattern described in [C#01].
3. **Transparency**: the AST is the only boundary between parsing and generation, as specified in [C#02].

---

## Definition Blocks

A `---…---` block defines all reference keys and document metadata in one place. Multiple entry types may coexist in the same block:

```markdown
---
C#01:knuth1986
F#01:my_figure
T#01:my_table
Author: Ada Lovelace
Title: My Document
MakeTOC
---
```

The block is pre-scanned before parsing begins, so keys defined at the bottom of a file are available at the top. A `---` line that is not followed by a recognised entry type is treated as a thematic break (`\noindent\rule{\linewidth}{0.4pt}`).

---

## AST Node Types

Table [T#01] lists all block and inline node types with their LaTeX counterparts.

|- T#01:H -|
|- Complete node-type reference. Leaf nodes have no children. See Figure [F#01] for the hierarchy. -|
| Node type        | Category | LaTeX output                      | Leaf |
|:-----------------|:--------:|:----------------------------------|:----:|
| Document         | block    | (root)                            | no   |
| Heading          | block    | `\section`, `\subsection`, …      | no   |
| Paragraph        | block    | text + `\n\n`                     | no   |
| BulletList       | block    | `\begin{itemize}`                 | no   |
| OrderedList      | block    | `\begin{enumerate}`               | no   |
| ListItem         | block    | `\item`                           | no   |
| BlockQuote       | block    | `\begin{quote}`                   | no   |
| FencedCode       | block    | `\begin{lstlisting}`              | yes  |
| IndentedCode     | block    | `\begin{verbatim}`                | yes  |
| ThematicBreak    | block    | `\noindent\rule{…}`               | yes  |
| Table            | block    | `\begin{tabular}`                 | no   |
| TableBlock       | block    | `\begin{table}…\end{table}`       | no   |
| DefinitionBlock  | block    | (invisible)                       | yes  |
| FigureBlock      | block    | (invisible)                       | yes  |
| Text             | inline   | escaped text                      | yes  |
| Strong           | inline   | `\textbf{}`                       | no   |
| Emphasis         | inline   | `\textit{}`                       | no   |
| StrongEmphasis   | inline   | `\textbf{\textit{}}`              | no   |
| CodeSpan         | inline   | `\texttt{}`                       | yes  |
| Strikethrough    | inline   | `\sout{}`                         | no   |
| Link             | inline   | `\href{}{}`                       | no   |
| Image            | inline   | `\includegraphics{}`              | no   |
| FigureImage      | inline   | `\begin{figure}…\end{figure}`     | yes  |
| CitationRef      | inline   | `~\cite{}`                        | yes  |
| FigureRef        | inline   | `~\ref{fig:}`                     | yes  |
| TableRef         | inline   | `~\ref{tab:}`                     | yes  |

---

## Special Character Escaping

Table [T#02] shows the ten LaTeX special characters and their escaped forms.

|- T#02:ht -|
|- LaTeX special-character escape table. Escaping is **not** applied inside `lstlisting` or `verbatim` environments, nor inside URL arguments to `\href{}`. See [C#02] for details. -|
| Character | Name          | LaTeX output           |
|:----------|:--------------|:-----------------------|
| `\`       | backslash     | `\textbackslash{}`     |
| `{`       | left brace    | `\{`                   |
| `}`       | right brace   | `\}`                   |
| `$`       | dollar        | `\$`                   |
| `&`       | ampersand     | `\&`                   |
| `#`       | hash          | `\#`                   |
| `%`       | percent       | `\%`                   |
| `_`       | underscore    | `\_`                   |
| `^`       | caret         | `\textasciicircum{}`   |
| `~`       | tilde         | `\textasciitilde{}`    |

---

## Figures

Figure [F#01] shows the overall architecture. Figure [F#02] shows the data flow in detail, and Figure [F#03] shows a sample output page.

![F#01:1.0:ht][System architecture overview. See [C#01] for the Go implementation.](figures/arch.png)

![F#02:0.8:h][Data-flow diagram. The pre-scan phase resolves [C#01], [F#01], and [T#01] references before parsing.](figures/flow.png)

![F#03:0.6:hb][Sample LaTeX output. The document uses packages listed in Table [T#03].](figures/output.png)

---

## Extensions Reference

Table [T#03] lists all available extensions, the LaTeX packages they require, and the `-ext` token used on the command line.

|- T#03:ht -|
|- Extension reference. Packages marked * are only included when actually used. See Figure [F#01] for where each extension fits in the pipeline. -|
| `-ext` token    | Feature                       | Required package  |
|:----------------|:------------------------------|:------------------|
| `tables`        | GFM pipe tables               | `booktabs` *      |
| `strikethrough` | `~~text~~`                    | `ulem` *          |
| `citations`     | `[C#key]` → `\cite{}`        | —                 |
| `figures`       | `[F#key]` → `\ref{fig:}`     | `graphicx` *      |
| `tableFloat`    | `[T#key]` → `\ref{tab:}`     | `booktabs` *      |
| `documentMeta`  | `Author:`, `MakeTOC`, …      | —                 |
| `all`           | All of the above              | as needed         |

---

## Inline Formatting Reference

The following examples show all inline constructs in context. A ~~deprecated~~ feature is shown for completeness. The `\texttt{monospace}` style is used for `inline code`. **Bold text** and *italic text* and ***bold-italic text*** are all supported. Links render as `\href{url}{text}` — for example [the Go specification](https://go.dev/ref/spec).

Hard line break example (two trailing spaces):
First line.
Second line after hard break.

> Blockquotes nest the content in `\begin{quote}…\end{quote}`. A citation [C#02] and a figure reference [F#01] both resolve correctly inside a blockquote.

---

## Standalone Output

When `-standalone` is active, the preamble includes only the packages the document actually needs. For a document with code listings, tables, images, and strikethrough, the generated preamble is:

```latex
\documentclass{article}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{lmodern}
\usepackage[a4paper, margin=2.5cm]{geometry}
\usepackage[hidelinks, unicode]{hyperref}
\usepackage{listings}
\usepackage{enumitem}
\usepackage[normalem]{ulem}
\usepackage{graphicx}
\usepackage{booktabs}
\usepackage{tabularx}
```

Document metadata (`Author:`, `Title:`, etc.) is placed immediately before `\begin{document}`, and front-matter commands (`\maketitle`, `\tableofcontents`, etc.) are placed immediately after it.
