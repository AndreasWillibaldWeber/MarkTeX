// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

// Package generator contains the LaTeX code generator that walks an AST and
// produces LaTeX output.
package generator

import "strings"

// escapeText escapes the ten LaTeX special characters in regular text content.
// This must NOT be applied to verbatim environments or URL arguments.
func escapeText(s string) string {
	// Process in a single pass to avoid double-escaping.
	var b strings.Builder
	b.Grow(len(s) + len(s)/4)
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\textbackslash{}`)
		case '{':
			b.WriteString(`\{`)
		case '}':
			b.WriteString(`\}`)
		case '$':
			b.WriteString(`\$`)
		case '&':
			b.WriteString(`\&`)
		case '#':
			b.WriteString(`\#`)
		case '%':
			b.WriteString(`\%`)
		case '_':
			b.WriteString(`\_`)
		case '^':
			b.WriteString(`\textasciicircum{}`)
		case '~':
			b.WriteString(`\textasciitilde{}`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// escapeURL escapes characters in a URL for use inside \href{...}.
// hyperref handles most URL characters correctly; we only need to escape
// the small set of characters that break hyperref's argument parsing.
func escapeURL(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '%':
			b.WriteString(`\%`)
		case '#':
			b.WriteString(`\#`)
		case '{':
			b.WriteString(`\{`)
		case '}':
			b.WriteString(`\}`)
		case '\\':
			b.WriteString(`\textbackslash{}`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
