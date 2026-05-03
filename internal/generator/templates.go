// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package generator

import "strings"

// standalonePreamble returns the LaTeX preamble for a complete standalone
// document. Only packages that are actually needed (tracked by flags on the
// Generator) are included to keep the output lean.
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

	b.WriteString(`
\begin{document}
`)
	return b.String()
}

const standalonePostamble = `
\end{document}
`

// preambleOpts records which packages the document actually needs so that
// unused packages are not emitted.
type preambleOpts struct {
	needsListings bool
	needsEnumitem bool
	needsUlem     bool
	needsGraphicx bool
	needsBooktabs bool
}
