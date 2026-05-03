// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

// Command marktex converts Markdown files to LaTeX.
//
// Usage:
//
//	marktex [flags] [file ...]
//
// If no file arguments are provided, Markdown is read from stdin.
// Output is written to stdout unless -o is specified.
//
// Flags:
//
//	-o <file>       Write output to file instead of stdout
//	-standalone     Emit a complete LaTeX document (\documentclass…\end{document})
//	-strict         Exit non-zero when an unsupported construct is encountered
//	-ext <list>     Comma-separated extensions to enable: tables,strikethrough,autolinks (default: all)
//	-no-ext <list>  Comma-separated extensions to disable
//	-version        Print version and exit
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/andreaswillibaldweber/marktex/pkg/transpiler"
)

const version = "0.1.0"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("marktex", flag.ContinueOnError)
	fs.SetOutput(stderr)

	outputFile := fs.String("o", "", "write output to `file` (default: stdout)")
	standalone := fs.Bool("standalone", false, "emit a complete LaTeX document")
	strict := fs.Bool("strict", false, "exit non-zero on unsupported constructs")
	extStr := fs.String("ext", "all", "comma-separated extensions to enable: tables,strikethrough,autolinks,all,none")
	noExtStr := fs.String("no-ext", "", "comma-separated extensions to disable")
	showVersion := fs.Bool("version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		// flag.ContinueOnError prints the error; just return usage exit code
		return 2
	}

	if *showVersion {
		fmt.Fprintf(stdout, "marktex %s\n", version)
		return 0
	}

	ext := parseExtensions(*extStr)
	ext &^= parseExtensions(*noExtStr) // clear disabled extensions

	opts := transpiler.Options{
		Standalone: *standalone,
		Strict:     *strict,
		Extensions: ext,
	}

	// Determine output writer
	var w io.Writer = stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(stderr, "marktex: %v\n", err)
			return 1
		}
		defer f.Close()
		w = f
	}

	files := fs.Args()
	if len(files) == 0 {
		// Read from stdin
		if err := transpiler.Transpile(stdin, w, opts); err != nil {
			fmt.Fprintf(stderr, "marktex: %v\n", err)
			return 1
		}
		return 0
	}

	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(stderr, "marktex: %v\n", err)
			return 1
		}
		if err := transpiler.Transpile(f, w, opts); err != nil {
			f.Close()
			fmt.Fprintf(stderr, "marktex: %s: %v\n", path, err)
			return 1
		}
		f.Close()
	}
	return 0
}

// parseExtensions converts a comma-separated extension list to an Extension
// bitmask. Recognised tokens: "all", "none", "tables", "strikethrough",
// "autolinks". Unknown tokens are silently ignored.
func parseExtensions(s string) transpiler.Extension {
	if s == "" {
		return 0
	}
	var ext transpiler.Extension
	for _, token := range strings.Split(s, ",") {
		switch strings.TrimSpace(strings.ToLower(token)) {
		case "all":
			ext |= transpiler.ExtAll
		case "none":
			// keep ext as 0 but don't break the loop (other tokens may follow)
		case "tables":
			ext |= transpiler.ExtTables
		case "strikethrough":
			ext |= transpiler.ExtStrikethrough
		case "autolinks":
			ext |= transpiler.ExtAutolinks
		}
	}
	return ext
}
