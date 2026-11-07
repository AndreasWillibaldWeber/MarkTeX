// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andreaswillibaldweber/marktex/pkg/transpiler"
)

func TestParseExtensions(t *testing.T) {
	tests := []struct {
		in   string
		want transpiler.Extension
	}{
		{"", 0},
		{"all", transpiler.ExtAll},
		{"none", 0},
		{"math", transpiler.ExtMath},
		{"Math", transpiler.ExtMath},
		{"tables,math", transpiler.ExtTables | transpiler.ExtMath},
		{"tables, strikethrough ,autolinks", transpiler.ExtTables | transpiler.ExtStrikethrough | transpiler.ExtAutolinks},
		{"citations,figures,tableFloat,documentMeta",
			transpiler.ExtCitations | transpiler.ExtFigures | transpiler.ExtTableFloat | transpiler.ExtDocumentMeta},
		{"unknown", 0},
		{"none,math", transpiler.ExtMath},
	}
	for _, tc := range tests {
		if got := parseExtensions(tc.in); got != tc.want {
			t.Errorf("parseExtensions(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// TestRunMathToggle drives the CLI entry point end to end for the math
// extension: on by default, off with -no-ext, and selectable with -ext.
func TestRunMathToggle(t *testing.T) {
	const src = "$2+3=5$ and ~~gone~~\n"
	tests := []struct {
		name    string
		args    []string
		want    string
		notWant string
	}{
		{"default enables math", nil, "$2+3=5$", `\$`},
		{"no-ext math disables it", []string{"-no-ext", "math"}, `\$2+3=5\$`, ""},
		{"ext math enables only math", []string{"-ext", "math"}, "$2+3=5$", `\sout`},
		{"ext tables leaves math off", []string{"-ext", "tables"}, `\$2+3=5\$`, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run(tc.args, strings.NewReader(src), &stdout, &stderr); code != 0 {
				t.Fatalf("exit code %d, stderr: %s", code, stderr.String())
			}
			got := stdout.String()
			if !strings.Contains(got, tc.want) {
				t.Errorf("missing %q\ngot: %q", tc.want, got)
			}
			if tc.notWant != "" && strings.Contains(got, tc.notWant) {
				t.Errorf("unexpected %q\ngot: %q", tc.notWant, got)
			}
		})
	}
}
