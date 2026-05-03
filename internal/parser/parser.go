// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

// Package parser converts Markdown source bytes into an AST.
// The public entry point is Parse; it runs a block-level pass followed by
// inline parsing of leaf-block content.
//
// Supported Markdown features (base):
//   - ATX and setext headings
//   - Paragraphs with soft/hard breaks
//   - Thematic breaks (---, ***, ___)
//   - Fenced code blocks (``` and ~~~)
//   - Indented code blocks
//   - Block quotes
//   - Bullet and ordered lists (tight and loose)
//   - Backslash escapes and HTML entities
//   - Emphasis (*), Strong (**), StrongEmphasis (***)
//   - Inline code spans
//   - Links and images (inline form)
//
// Optional extensions (enabled via Extensions):
//   - ExtTables:        GFM pipe tables
//   - ExtStrikethrough: GFM ~~strikethrough~~
//   - ExtCitations:     Citation blocks and inline [C#key] → \cite{}
//   - ExtFigures:       Figure blocks and ![F#key:w:pos][cap](path) → figure env
package parser

import (
	"bytes"
	"strings"

	"github.com/andreaswillibaldweber/marktex/internal/ast"
)

// Extensions is a bitmask of optional Markdown extensions to enable.
type Extensions uint32

const (
	// ExtTables enables GFM-style pipe tables.
	ExtTables Extensions = 1 << iota
	// ExtStrikethrough enables GFM ~~strikethrough~~.
	ExtStrikethrough
	// ExtAutolinks enables bare URL autolinks (not yet implemented).
	ExtAutolinks
	// ExtCitations enables the citation block / [C#key] extension.
	ExtCitations
	// ExtFigures enables the figure block / ![F#key:width:pos][caption](path)
	// extension and [F#key] cross-references.
	ExtFigures
	// ExtTableFloat enables table-float blocks (|- T#key:placement -| rows) and
	// [T#key] cross-references that resolve to \ref{tab:label}.
	ExtTableFloat
	// ExtDocumentMeta enables document-metadata entries inside definition blocks:
	//   Author, Title, Subtitle, Date, MakeTitlePage, MakeTOC, MakeLOF, MakeLOT, MakeLOL.
	// Only applied when the generator runs in standalone mode.
	ExtDocumentMeta

	// ExtAll enables all available extensions.
	ExtAll = ExtTables | ExtStrikethrough | ExtAutolinks | ExtCitations | ExtFigures | ExtTableFloat | ExtDocumentMeta
)

// Has reports whether ext includes the given extension flag.
func (e Extensions) Has(flag Extensions) bool { return e&flag != 0 }

// defRefs holds all definition maps produced by the pre-scan pass.
// They are threaded through the parser down to the inline parser.
type defRefs struct {
	citations map[string]string // C# keys → BibTeX identifiers
	figures   map[string]string // F# keys → LaTeX label suffixes
	tables    map[string]string // T# keys → LaTeX label suffixes
	meta      ast.DocumentMeta  // document metadata (standalone mode only)
}

// Parse converts Markdown source bytes into an AST Document.
func Parse(src []byte, ext Extensions) *ast.Document {
	refs := preScan(src, ext)
	p := newBlockParser(src, ext, refs)
	return p.parse()
}

// parseWithRefs is used internally for blockquote re-parsing so that definitions
// from the outer document are still visible.
func parseWithRefs(src []byte, ext Extensions, refs defRefs) *ast.Document {
	p := newBlockParser(src, ext, refs)
	return p.parse()
}

// preScan collects all definition blocks in a single pass so that references
// resolve correctly regardless of where the definition block appears in the source.
func preScan(src []byte, ext Extensions) defRefs {
	refs := defRefs{
		citations: make(map[string]string),
		figures:   make(map[string]string),
		tables:    make(map[string]string),
	}
	if !ext.Has(ExtCitations) && !ext.Has(ExtFigures) && !ext.Has(ExtTableFloat) && !ext.Has(ExtDocumentMeta) {
		return refs
	}

	lines := bytes.Split(src, []byte("\n"))
	i := 0
	for i < len(lines) {
		if !isDefinitionDelimiterLine(lines[i]) {
			i++
			continue
		}
		j := i + 1
		citCollected := map[string]string{}
		figCollected := map[string]string{}
		tabCollected := map[string]string{}
		var metaCollected ast.DocumentMeta
		closingFound := false
		valid := true

		for j < len(lines) {
			if isDefinitionDelimiterLine(lines[j]) {
				closingFound = true
				break
			}
			if ext.Has(ExtCitations) {
				if k, v, ok := parseCitationEntryBytes(lines[j]); ok {
					citCollected[k] = v
					j++
					continue
				}
			}
			if ext.Has(ExtFigures) {
				if k, v, ok := parseFigureEntryBytes(lines[j]); ok {
					figCollected[k] = v
					j++
					continue
				}
			}
			if ext.Has(ExtTableFloat) {
				if k, v, ok := parseTableEntryBytes(lines[j]); ok {
					tabCollected[k] = v
					j++
					continue
				}
			}
			if ext.Has(ExtDocumentMeta) {
				if ok := parseDocumentMetaEntry(lines[j], &metaCollected); ok {
					j++
					continue
				}
			}
			// Line matches no known entry type → not a definition block
			valid = false
			break
		}

		hasEntries := len(citCollected) > 0 || len(figCollected) > 0 ||
			len(tabCollected) > 0 || metaCollected != (ast.DocumentMeta{})
		if closingFound && valid && hasEntries {
			for k, v := range citCollected {
				refs.citations[k] = v
			}
			for k, v := range figCollected {
				refs.figures[k] = v
			}
			for k, v := range tabCollected {
				refs.tables[k] = v
			}
			refs.meta.Merge(metaCollected)
			i = j + 1
		} else {
			i++
		}
	}
	return refs
}

// isDefinitionDelimiterLine returns true if the line is exactly `---`.
func isDefinitionDelimiterLine(line []byte) bool {
	return bytes.Equal(bytes.TrimRight(line, "\r"), []byte("---"))
}

// isCitationDelimiterLine is kept for backward compatibility in block.go.
var isCitationDelimiterLine = isDefinitionDelimiterLine

// parseCitationEntryBytes parses a `C#key:bibtex` line.
func parseCitationEntryBytes(line []byte) (key, bib string, ok bool) {
	s := strings.TrimSpace(string(line))
	if !strings.HasPrefix(s, "C#") {
		return "", "", false
	}
	idx := strings.IndexByte(s, ':')
	if idx < 3 {
		return "", "", false
	}
	k := strings.TrimSpace(s[:idx])
	b := strings.TrimSpace(s[idx+1:])
	if k == "" || b == "" {
		return "", "", false
	}
	return k, b, true
}

// parseFigureEntryBytes parses a `F#key:label` line.
func parseFigureEntryBytes(line []byte) (key, label string, ok bool) {
	s := strings.TrimSpace(string(line))
	if !strings.HasPrefix(s, "F#") {
		return "", "", false
	}
	idx := strings.IndexByte(s, ':')
	if idx < 3 {
		return "", "", false
	}
	k := strings.TrimSpace(s[:idx])
	v := strings.TrimSpace(s[idx+1:])
	if k == "" || v == "" {
		return "", "", false
	}
	return k, v, true
}

// parseTableEntryBytes parses a `T#key:label` line used in definition blocks.
func parseTableEntryBytes(line []byte) (key, label string, ok bool) {
	s := strings.TrimSpace(string(line))
	if !strings.HasPrefix(s, "T#") {
		return "", "", false
	}
	idx := strings.IndexByte(s, ':')
	if idx < 3 {
		return "", "", false
	}
	k := strings.TrimSpace(s[:idx])
	v := strings.TrimSpace(s[idx+1:])
	if k == "" || v == "" {
		return "", "", false
	}
	return k, v, true
}

// parseDocumentMetaEntry attempts to parse a document-metadata line and writes
// the result into meta. It handles both key-value pairs (`Author: Name`) and
// boolean flags (`MakeTOC`). Returns true when the line was recognised.
//
// Key matching is case-insensitive. The supported keys and flags are:
//
//	Author, Title, Subtitle, Date
//	MakeTitlePage, MakeTOC, MakeLOF, MakeLOT, MakeLOL
func parseDocumentMetaEntry(line []byte, meta *ast.DocumentMeta) (ok bool) {
	s := strings.TrimSpace(string(line))
	if s == "" {
		return false
	}

	// Boolean flags (no colon)
	switch strings.ToLower(s) {
	case "maketitlepage":
		meta.MakeTitlePage = true
		return true
	case "maketoc":
		meta.MakeTOC = true
		return true
	case "makelof":
		meta.MakeLOF = true
		return true
	case "makelot":
		meta.MakeLOT = true
		return true
	case "makelol":
		meta.MakeLOL = true
		return true
	}

	// Key: value pairs
	idx := strings.IndexByte(s, ':')
	if idx < 1 {
		return false
	}
	k := strings.TrimSpace(strings.ToLower(s[:idx]))
	v := strings.TrimSpace(s[idx+1:])
	if v == "" {
		return false
	}
	switch k {
	case "author":
		meta.Author = v
	case "title":
		meta.Title = v
	case "subtitle":
		meta.Subtitle = v
	case "date":
		meta.Date = v
	default:
		return false
	}
	return true
}

// preScanCitations is kept for any callers that still use the old API.
func preScanCitations(src []byte) map[string]string {
	return preScan(src, ExtCitations).citations
}
