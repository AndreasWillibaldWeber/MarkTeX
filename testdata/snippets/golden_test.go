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
//
// Per-snippet options are read from an optional <name>.opts sidecar file.
// Each non-empty line is one option token. Supported tokens:
//
//	standalone   — run with Standalone: true
package snippets_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreaswillibaldweber/marktex/pkg/transpiler"
)

var update = flag.Bool("update", false, "overwrite .tex golden files with current output")

// snippetOpts reads the optional <name>.opts sidecar and returns transpiler
// options for the snippet. Falls back to safe defaults when the file is absent.
func snippetOpts(name string) transpiler.Options {
	opts := transpiler.Options{Extensions: transpiler.ExtAll}

	data, err := os.ReadFile(name + ".opts")
	if err != nil {
		return opts // no sidecar — use defaults
	}

	for _, line := range strings.Split(string(data), "\n") {
		switch strings.TrimSpace(strings.ToLower(line)) {
		case "standalone":
			opts.Standalone = true
		}
	}
	return opts
}

func TestGolden(t *testing.T) {
	mdFiles, err := filepath.Glob("*.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(mdFiles) == 0 {
		t.Fatal("no .md snippets found")
	}

	for _, mdPath := range mdFiles {
		mdPath := mdPath // capture
		name := strings.TrimSuffix(mdPath, ".md")
		texPath := name + ".tex"

		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(mdPath)
			if err != nil {
				t.Fatalf("read %s: %v", mdPath, err)
			}

			opts := snippetOpts(name)
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
