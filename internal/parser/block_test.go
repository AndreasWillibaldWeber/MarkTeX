// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"testing"

	"github.com/andreaswillibaldweber/marktex/internal/ast"
)

func TestSetextHeadingLevel(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"===", 1},
		{"=", 1},
		{"==========", 1},
		{"---", 2},
		{"-", 2},
		{"-------------", 2},
		{"---   ", 2},
		{"---\t", 2},
		{"   ---", 2},
		{"   ===", 1},
		// Four or more spaces of indentation: continuation text, not an underline.
		{"    ---", 0},
		{"    ===", 0},
		{"", 0},
		{"   ", 0},
		{"- - -", 0},
		{"--a", 0},
		{"=-=", 0},
		{"***", 0},
		{"___", 0},
		{"|---|", 0},
		{":---", 0},
		{"---|---", 0},
	}
	for _, tc := range tests {
		if got := setextHeadingLevel([]byte(tc.in)); got != tc.want {
			t.Errorf("setextHeadingLevel(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestIsTableSeparator(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"|---|---|", true},
		{"---|---", true},
		{"| :--- | ---: |", true},
		{":---:", true},
		{"|---|", true},
		{"   |---|", true},
		// A setext underline is never a delimiter row (issue #2).
		{"---", false},
		{"-------------", false},
		{"   ---", false},
		{"--- ", false},
		// Indented four or more spaces: code or continuation text.
		{"    |---|", false},
		{"", false},
		{"   ", false},
		{"| a | b |", false},
		{"--- ---", false},
		{"|--x--|", false},
		{"|", false},
	}
	for _, tc := range tests {
		l := &line{content: []byte(tc.in)}
		if got := isTableSeparator(l); got != tc.want {
			t.Errorf("isTableSeparator(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
	if isTableSeparator(nil) {
		t.Error("isTableSeparator(nil) = true, want false")
	}
}

// TestParseSetextUnderlineYieldsHeading asserts at the AST level that the
// report from issue #2 produces a single level-2 Heading regardless of which
// extensions are active.
func TestParseSetextUnderlineYieldsHeading(t *testing.T) {
	const src = "Also header 2 (not a table!)\n-------------\n"
	for _, ext := range []Extensions{ExtStrikethrough, ExtTables, ExtAll} {
		doc := Parse([]byte(src), ext)
		kids := doc.Children()
		if len(kids) != 1 {
			t.Fatalf("ext=%d: got %d top-level nodes, want 1: %#v", ext, len(kids), kids)
		}
		h, ok := kids[0].(*ast.Heading)
		if !ok {
			t.Fatalf("ext=%d: got %T, want *ast.Heading", ext, kids[0])
		}
		if h.Level != 2 {
			t.Errorf("ext=%d: heading level = %d, want 2", ext, h.Level)
		}
	}
}

// TestParseSetextUnderlineLevels checks both underline characters and that a
// following paragraph is kept separate from the heading.
func TestParseSetextUnderlineLevels(t *testing.T) {
	doc := Parse([]byte("One\n===\n\nTwo\n---\nbody"), ExtAll)
	kids := doc.Children()
	if len(kids) != 3 {
		t.Fatalf("got %d top-level nodes, want 3: %#v", len(kids), kids)
	}
	h1, ok := kids[0].(*ast.Heading)
	if !ok || h1.Level != 1 {
		t.Errorf("node 0 = %#v, want level-1 heading", kids[0])
	}
	h2, ok := kids[1].(*ast.Heading)
	if !ok || h2.Level != 2 {
		t.Errorf("node 1 = %#v, want level-2 heading", kids[1])
	}
	if _, ok := kids[2].(*ast.Paragraph); !ok {
		t.Errorf("node 2 = %T, want *ast.Paragraph", kids[2])
	}
}
