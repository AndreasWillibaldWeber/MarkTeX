// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

// Package transpiler is the public API for MarkTeX.
// It combines the parser and generator into a single Transpile call.
package transpiler

import "github.com/andreaswillibaldweber/marktex/internal/parser"

// Extension is a named Markdown extension flag.
type Extension = parser.Extensions

const (
	// ExtTables enables GFM pipe tables.
	ExtTables = parser.ExtTables

	// ExtStrikethrough enables GFM ~~strikethrough~~ syntax.
	ExtStrikethrough = parser.ExtStrikethrough

	// ExtAutolinks enables bare URL autolinks.
	ExtAutolinks = parser.ExtAutolinks

	// ExtCitations enables the citation block / inline reference extension:
	//
	//   ---
	//   C#01:doe2023survey
	//   ---
	//
	//   See [C#01] for details.   →   See~\cite{doe2023survey} for details.
	ExtCitations = parser.ExtCitations

	// ExtFigures enables the figure block, extended image syntax, and cross-reference:
	//
	//   ---
	//   F#01:myfig
	//   ---
	//
	//   ![F#01:0.8:ht][Caption.](img.png)   →   \begin{figure}[ht] … \end{figure}
	//   See Figure [F#01].                  →   See Figure~\ref{fig:myfig}.
	ExtFigures = parser.ExtFigures

	// ExtTableFloat enables table-float blocks and [T#key] cross-references:
	//
	//   ---
	//   T#01:mytable
	//   ---
	//
	//   |- T#01:ht -|
	//   |- Caption text. -|
	//   | col | col |
	//   | --- | --- |
	//
	//   See Table [T#01].   →   See Table~\ref{tab:mytable}.
	ExtTableFloat = parser.ExtTableFloat

	// ExtDocumentMeta enables document-metadata entries in definition blocks:
	//
	//   ---
	//   Author: Firstname Surname
	//   Title: My Document
	//   Subtitle: A deeper look
	//   Date: 2025-05-01
	//   MakeTitlePage
	//   MakeTOC
	//   MakeLOF
	//   MakeLOT
	//   MakeLOL
	//   ---
	//
	// Metadata is only applied when -standalone is active; silently ignored
	// in fragment mode.
	ExtDocumentMeta = parser.ExtDocumentMeta

	// ExtAll enables all available extensions.
	ExtAll = parser.ExtAll
)

// Options configures the transpilation pipeline.
type Options struct {
	// Standalone wraps the LaTeX output in a complete \documentclass document.
	// When false (the default) only the document body is emitted.
	Standalone bool

	// Strict causes Transpile to return an error for any Markdown construct
	// that cannot be faithfully represented in LaTeX (e.g. raw HTML blocks).
	// In non-strict mode a LaTeX comment is emitted in place of the construct.
	Strict bool

	// Extensions selects optional Markdown syntax extensions. Combine flags
	// with |. Defaults to ExtAll when left at the zero value.
	Extensions Extension
}

// defaults fills zero-value fields with sensible defaults.
func (o *Options) defaults() {
	if o.Extensions == 0 {
		o.Extensions = ExtAll
	}
}
