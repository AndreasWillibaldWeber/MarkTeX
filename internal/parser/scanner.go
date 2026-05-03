// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"bytes"
)

// line represents a single source line after tab expansion.
type line struct {
	raw     []byte // original bytes (including newline if present)
	content []byte // tab-expanded, trailing newline stripped
	lineNum int    // 1-based
	offset  int    // byte offset of the first byte of this line in the original input
}

// indent returns the number of leading spaces in the expanded content.
func (l *line) indent() int {
	for i, b := range l.content {
		if b != ' ' {
			return i
		}
	}
	return len(l.content)
}

// isEmpty returns true if the line contains only whitespace.
func (l *line) isEmpty() bool {
	return len(bytes.TrimSpace(l.content)) == 0
}

// scanner splits source bytes into lines, expanding tabs to the nearest 4-space
// boundary (per CommonMark spec). It provides one-line lookahead.
type scanner struct {
	src     []byte
	pos     int // current read position in src
	lineNum int // next line number to assign (1-based)
	peeked  *line
}

func newScanner(src []byte) *scanner {
	return &scanner{src: src, lineNum: 1}
}

// peek returns the next line without consuming it. Returns nil at EOF.
func (s *scanner) peek() *line {
	if s.peeked != nil {
		return s.peeked
	}
	s.peeked = s.readLine()
	return s.peeked
}

// next consumes and returns the next line. Returns nil at EOF.
func (s *scanner) next() *line {
	if s.peeked != nil {
		l := s.peeked
		s.peeked = nil
		return l
	}
	return s.readLine()
}

// readLine reads the next raw line from src, expands its tabs, and returns a
// line struct. Returns nil when the input is exhausted.
func (s *scanner) readLine() *line {
	if s.pos >= len(s.src) {
		return nil
	}
	start := s.pos
	end := bytes.IndexByte(s.src[s.pos:], '\n')
	var raw []byte
	if end < 0 {
		raw = s.src[s.pos:]
		s.pos = len(s.src)
	} else {
		end = s.pos + end + 1 // include the newline
		raw = s.src[s.pos:end]
		s.pos = end
	}
	lineNum := s.lineNum
	s.lineNum++

	// Strip trailing newline (and CR for CRLF inputs) for content.
	content := raw
	if len(content) > 0 && content[len(content)-1] == '\n' {
		content = content[:len(content)-1]
	}
	if len(content) > 0 && content[len(content)-1] == '\r' {
		content = content[:len(content)-1]
	}

	return &line{
		raw:     raw,
		content: expandTabs(content),
		lineNum: lineNum,
		offset:  start,
	}
}

// expandTabs replaces each tab with spaces to the nearest 4-column boundary.
func expandTabs(b []byte) []byte {
	if !bytes.ContainsRune(b, '\t') {
		return b
	}
	var out []byte
	col := 0
	for _, ch := range b {
		if ch == '\t' {
			spaces := 4 - (col % 4)
			for i := 0; i < spaces; i++ {
				out = append(out, ' ')
			}
			col += spaces
		} else {
			out = append(out, ch)
			col++
		}
	}
	return out
}
