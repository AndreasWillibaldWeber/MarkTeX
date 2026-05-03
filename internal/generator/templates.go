// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package generator

import "strings"

// preambleOpts records which packages the document needs and carries optional
// document-metadata strings used in standalone mode.
type preambleOpts struct {
	// Package flags — only include packages actually used by the content.
	needsListings bool
	needsEnumitem bool
	needsUlem     bool
	needsGraphicx bool
	needsBooktabs bool

	// Document metadata (standalone mode only).
	author   string
	title    string
	subtitle string // appended to \title as \\[0.5em]{\large …}
	date     string // empty → LaTeX default (\today)
}

// standalonePreamble returns a complete LaTeX preamble up to and including
// \begin{document}. Only packages that are actually needed are included.
// Document-metadata commands (\author, \title, \date) are emitted when set.
func standalonePreamble(opts preambleOpts) string {
	b := &strings.Builder{}

	b.WriteString(`\documentclass{article}

% ── Encoding & fonts ─────────────────────────────────────────────────────────
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{lmodern}

% ── Page geometry ─────────────────────────────────────────────────────────────
\usepackage[a4paper, margin=2.5cm]{geometry}

% ── Hyperlinks ────────────────────────────────────────────────────────────────
\usepackage[hidelinks, unicode]{hyperref}
`)

	if opts.needsListings {
		b.WriteString(`
% ── Source code listings ─────────────────────────────────────────────────────
\usepackage{listings}
\lstset{
  basicstyle=\ttfamily\small,
  breaklines=true,
  keepspaces=true,
  columns=flexible,
  frame=single,
  numbers=left,
  numberstyle=\tiny,
}
`)
	}

	if opts.needsEnumitem {
		b.WriteString(`
% ── List customisation ───────────────────────────────────────────────────────
\usepackage{enumitem}
`)
	}

	if opts.needsUlem {
		b.WriteString(`
% ── Strikethrough ────────────────────────────────────────────────────────────
\usepackage[normalem]{ulem}
`)
	}

	if opts.needsGraphicx {
		b.WriteString(`
% ── Graphics ─────────────────────────────────────────────────────────────────
\usepackage{graphicx}
`)
	}

	if opts.needsBooktabs {
		b.WriteString(`
% ── Tables ───────────────────────────────────────────────────────────────────
\usepackage{booktabs}
\usepackage{tabularx}
`)
	}

	// Document-metadata commands go in the preamble (before \begin{document}).
	if opts.title != "" {
		b.WriteString("\n% ── Document metadata ────────────────────────────────────────────────────────\n")
		if opts.subtitle != "" {
			b.WriteString(`\title{` + escapeText(opts.title) +
				`\\[0.5em]{\normalfont\large ` + escapeText(opts.subtitle) + `}}` + "\n")
		} else {
			b.WriteString(`\title{` + escapeText(opts.title) + "}\n")
		}
	}
	if opts.author != "" {
		b.WriteString(`\author{` + escapeText(opts.author) + "}\n")
	}
	if opts.date != "" {
		b.WriteString(`\date{` + escapeText(opts.date) + "}\n")
	}

	b.WriteString(`
\begin{document}
`)
	return b.String()
}

const standalonePostamble = `
\end{document}
`
