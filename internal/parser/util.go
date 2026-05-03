// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"strings"
	"unicode"
)

// isASCIIPunctuation returns true for ASCII punctuation characters per the
// CommonMark spec definition (used for left/right flanking run rules).
func isASCIIPunctuation(r rune) bool {
	return (r >= '!' && r <= '/') ||
		(r >= ':' && r <= '@') ||
		(r >= '[' && r <= '`') ||
		(r >= '{' && r <= '~')
}

// isUnicodePunctuation returns true if r is ASCII punctuation or Unicode
// category P (punctuation) or S (symbol).
func isUnicodePunctuation(r rune) bool {
	if isASCIIPunctuation(r) {
		return true
	}
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

// trimLeftSpaces returns s with up to n leading spaces removed.
func trimLeftSpaces(s []byte, n int) []byte {
	for i := 0; i < n && i < len(s) && s[i] == ' '; i++ {
		s = s[1:]
	}
	return s
}

// countLeadingChar counts the number of times ch appears at the start of s.
func countLeadingChar(s []byte, ch byte) int {
	for i, b := range s {
		if b != ch {
			return i
		}
	}
	return len(s)
}

// isBlankLine returns true if s consists entirely of space/tab characters.
func isBlankLine(s []byte) bool {
	for _, b := range s {
		if b != ' ' && b != '\t' {
			return false
		}
	}
	return true
}

// isThematicBreak returns true if the line content matches `---`, `***`, or
// `___` (with optional internal spaces, and at least 3 of the marker character).
func isThematicBreak(content []byte) bool {
	s := strings.TrimSpace(string(content))
	if len(s) < 3 {
		return false
	}
	ch := s[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}
	count := 0
	for _, b := range []byte(s) {
		if b == ' ' || b == '\t' {
			continue
		}
		if b != ch {
			return false
		}
		count++
	}
	return count >= 3
}

// parseATXHeading attempts to parse an ATX heading from content.
// Returns the level (1-6) and the trimmed heading text.
// Returns 0, "" if the line is not an ATX heading.
func parseATXHeading(content []byte) (level int, text string) {
	s := content
	// Count leading '#' characters
	n := countLeadingChar(s, '#')
	if n == 0 || n > 6 {
		return 0, ""
	}
	rest := s[n:]
	// Must be followed by a space or be empty (for level-1 empty heading "# ")
	if len(rest) > 0 && rest[0] != ' ' && rest[0] != '\t' {
		return 0, ""
	}
	text = strings.TrimSpace(string(rest))
	// Strip trailing '#' characters (closing sequence)
	if i := strings.LastIndexFunc(text, func(r rune) bool { return r != '#' }); i >= 0 {
		// Check that the character before the trailing hashes is a space
		tail := text[i+1:]
		if len(tail) > 0 {
			prefix := text[:i+1]
			if len(prefix) == 0 || prefix[len(prefix)-1] == ' ' {
				text = strings.TrimRight(prefix, " \t")
			}
		}
	} else {
		text = ""
	}
	return n, text
}

// parseOrderedListMarker attempts to parse an ordered list marker at the start
// of content. Returns number, delimiter rune, and the width of the marker
// (number of bytes consumed including the required trailing space).
// Returns 0, 0, 0 on failure. Note: a list starting at item 0 is valid
// Markdown; callers should treat width == 0 as the failure indicator.
func parseOrderedListMarker(content []byte) (num int, delimiter rune, width int) {
	i := 0
	for i < len(content) && content[i] >= '0' && content[i] <= '9' {
		i++
	}
	if i == 0 || i > 9 { // CommonMark: at most 9 digits
		return 0, 0, 0
	}
	// Parse the numeric value
	n := 0
	for _, ch := range content[:i] {
		n = n*10 + int(ch-'0')
	}
	if i >= len(content) || (content[i] != '.' && content[i] != ')') {
		return 0, 0, 0
	}
	del := rune(content[i])
	i++ // consume delimiter
	// Must be followed by at least one space
	if i >= len(content) {
		return n, del, i
	}
	if content[i] != ' ' && content[i] != '\t' {
		return 0, 0, 0
	}
	i++ // consume required space
	return n, del, i
}

// parseBulletListMarker attempts to parse a bullet list marker at the start
// of content. Returns the marker rune and the number of bytes consumed
// (including the required trailing space). Returns 0, 0 on failure.
func parseBulletListMarker(content []byte) (marker rune, width int) {
	if len(content) == 0 {
		return 0, 0
	}
	ch := content[0]
	if ch != '-' && ch != '*' && ch != '+' {
		return 0, 0
	}
	if len(content) == 1 {
		return rune(ch), 1
	}
	if content[1] != ' ' && content[1] != '\t' {
		return 0, 0
	}
	return rune(ch), 2
}

// stripBlockQuoteMarker strips the leading `> ` (or `>`) from a block-quote
// line. Returns the remaining bytes and true on success.
func stripBlockQuoteMarker(content []byte) ([]byte, bool) {
	if len(content) == 0 || content[0] != '>' {
		return nil, false
	}
	rest := content[1:]
	if len(rest) > 0 && rest[0] == ' ' {
		return rest[1:], true
	}
	return rest, true
}
