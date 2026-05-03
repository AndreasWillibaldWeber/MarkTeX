// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package transpiler

import (
	"bytes"
	"io"
	"marktex/internal/generator"
	"marktex/internal/parser"
)

// Transpiler combines the parser and LaTeX generator. Create one with New and
// reuse it to transpile many documents.
type Transpiler struct {
	opts Options
}

// New creates a Transpiler configured with opts.
// If opts.Extensions is 0 it defaults to ExtAll.
func New(opts Options) *Transpiler {
	opts.defaults()
	return &Transpiler{opts: opts}
}

// Transpile reads Markdown from r and writes LaTeX to w.
func (t *Transpiler) Transpile(r io.Reader, w io.Writer) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	doc := parser.Parse(src, t.opts.Extensions)
	return generator.Generate(doc, generator.Options{
		Standalone: t.opts.Standalone,
		Strict:     t.opts.Strict,
	}, w)
}

// TranspileBytes transpiles src and returns the LaTeX output as a byte slice.
func (t *Transpiler) TranspileBytes(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := t.Transpile(bytes.NewReader(src), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// TranspileString transpiles src and returns the LaTeX output as a string.
func (t *Transpiler) TranspileString(src string) (string, error) {
	out, err := t.TranspileBytes([]byte(src))
	return string(out), err
}

// ─── Package-level convenience functions ──────────────────────────────────────

// Transpile reads Markdown from r and writes LaTeX to w using the given options.
func Transpile(r io.Reader, w io.Writer, opts Options) error {
	return New(opts).Transpile(r, w)
}

// TranspileBytes transpiles src with the given options.
func TranspileBytes(src []byte, opts Options) ([]byte, error) {
	return New(opts).TranspileBytes(src)
}

// TranspileString transpiles src with the given options.
func TranspileString(src string, opts Options) (string, error) {
	return New(opts).TranspileString(src)
}
