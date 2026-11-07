// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package transpiler_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andreaswillibaldweber/marktex/pkg/transpiler"
)

// requireLaTeXTools skips the test unless pdflatex and pdftotext are
// installed. A compile takes about a second, so -short skips it as well.
func requireLaTeXTools(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("LaTeX compilation skipped in -short mode")
	}
	for _, tool := range []string{"pdflatex", "pdftotext"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not installed", tool)
		}
	}
}

// renderedLines transpiles src in standalone mode, compiles it with pdflatex
// and returns the non-empty text lines of the resulting PDF.
func renderedLines(t *testing.T, src string) []string {
	t.Helper()
	tex, err := transpiler.TranspileString(src, transpiler.Options{Standalone: true})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "doc.tex"), []byte(tex), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	compile := exec.CommandContext(ctx, "pdflatex", "-interaction=nonstopmode", "-halt-on-error", "doc.tex")
	compile.Dir = dir
	if out, err := compile.CombinedOutput(); err != nil {
		t.Fatalf("pdflatex failed: %v\n%s\n--- LaTeX source ---\n%s", err, lastLines(string(out), 30), tex)
	}

	extract := exec.CommandContext(ctx, "pdftotext", "-layout", "doc.pdf", "-")
	extract.Dir = dir
	out, err := extract.Output()
	if err != nil {
		t.Fatalf("pdftotext failed: %v", err)
	}
	var lines []string
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// TestRenderedHeadingsOnSeparateLines is the end-to-end check for issue #1:
// in the compiled PDF every heading of levels 4 to 6 must occupy a line of
// its own, in source order, with the following text on the next line.
func TestRenderedHeadingsOnSeparateLines(t *testing.T) {
	requireLaTeXTools(t)

	src := "#### Header 4\nBody after four.\n\n##### Header 5\n###### Header 6\nBody after six.\n"
	phrases := []string{"Header 4", "Body after four.", "Header 5", "Header 6", "Body after six."}
	lines := renderedLines(t, src)

	lineOf := map[string]int{}
	for i, line := range lines {
		var hits []string
		for _, p := range phrases {
			if strings.Contains(line, p) {
				hits = append(hits, p)
				lineOf[p] = i
			}
		}
		if len(hits) > 1 {
			t.Errorf("rendered on one line: %q\nline: %q", hits, line)
		}
	}
	prev := -1
	for _, p := range phrases {
		i, ok := lineOf[p]
		if !ok {
			t.Errorf("%q missing from rendered text\nlines: %q", p, lines)
			continue
		}
		if i <= prev {
			t.Errorf("%q rendered out of order\nlines: %q", p, lines)
		}
		prev = i
	}
}
