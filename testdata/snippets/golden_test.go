// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

// Package snippets_test runs golden-file tests for every .md/.tex pair found
// in this directory.
//
// Normal run (compare against golden files):
//
//	go test ./testdata/snippets/
//
// Regenerate golden files after intentional output changes:
//
//	go test ./testdata/snippets/ -update
package snippets_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"marktex/pkg/transpiler"
)

var update = flag.Bool("update", false, "overwrite .tex golden files with current output")

func TestGolden(t *testing.T) {
	mdFiles, err := filepath.Glob("*.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(mdFiles) == 0 {
		t.Fatal("no .md snippets found")
	}

	opts := transpiler.Options{Extensions: transpiler.ExtAll}

	for _, mdPath := range mdFiles {
		mdPath := mdPath // capture
		name := strings.TrimSuffix(mdPath, ".md")
		texPath := name + ".tex"

		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(mdPath)
			if err != nil {
				t.Fatalf("read %s: %v", mdPath, err)
			}

			got, err := transpiler.TranspileBytes(src, opts)
			if err != nil {
				t.Fatalf("transpile %s: %v", mdPath, err)
			}

			if *update {
				if err := os.WriteFile(texPath, got, 0644); err != nil {
					t.Fatalf("write %s: %v", texPath, err)
				}
				t.Logf("updated %s", texPath)
				return
			}

			want, err := os.ReadFile(texPath)
			if err != nil {
				t.Fatalf("read golden %s: %v (run with -update to create it)", texPath, err)
			}

			if string(got) != string(want) {
				t.Errorf("output mismatch for %s\n\n--- want ---\n%s\n--- got ---\n%s",
					mdPath, want, got)
			}
		})
	}
}
