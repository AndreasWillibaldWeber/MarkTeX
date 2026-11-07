// SPDX-FileCopyrightText: 2026 Andreas W. Weber
// SPDX-License-Identifier: GPL-3.0-or-later

package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/andreaswillibaldweber/marktex/internal/ast"
)

// kinds renders an inline node sequence compactly, e.g. `T(a) HB T(b)`.
func kinds(nodes []ast.Node) string {
	var b strings.Builder
	for i, n := range nodes {
		if i > 0 {
			b.WriteByte(' ')
		}
		switch n := n.(type) {
		case *ast.Text:
			fmt.Fprintf(&b, "T(%s)", n.Value)
		case *ast.SoftBreak:
			b.WriteString("SB")
		case *ast.HardBreak:
			b.WriteString("HB")
		case *ast.Escape:
			b.WriteString("ESC")
		case *ast.CodeSpan:
			fmt.Fprintf(&b, "C(%s)", n.Literal)
		case *ast.Math:
			if n.Display {
				fmt.Fprintf(&b, "DM(%s)", n.Literal)
			} else {
				fmt.Fprintf(&b, "M(%s)", n.Literal)
			}
		default:
			fmt.Fprintf(&b, "%T", n)
		}
	}
	return b.String()
}

func TestResolveLineBreaks(t *testing.T) {
	var pos ast.Pos
	txt := func(s string) ast.Node { return ast.NewText(pos, s) }
	sb := func() ast.Node { return ast.NewSoftBreak(pos) }
	hb := func() ast.Node { return ast.NewHardBreak(pos) }
	esc := func() ast.Node { return ast.NewEscape(pos, '*') }

	tests := []struct {
		name string
		in   []ast.Node
		want string
	}{
		{"plain soft break", []ast.Node{txt("a"), sb(), txt("b")}, "T(a) SB T(b)"},
		{"one trailing space stays soft", []ast.Node{txt("a "), sb(), txt("b")}, "T(a) SB T(b)"},
		{"two trailing spaces become hard", []ast.Node{txt("a  "), sb(), txt("b")}, "T(a) HB T(b)"},
		{"many trailing spaces become hard", []ast.Node{txt("a     "), sb(), txt("b")}, "T(a) HB T(b)"},
		{"trailing tab counts as whitespace", []ast.Node{txt("a \t"), sb(), txt("b")}, "T(a) HB T(b)"},
		{"leading spaces after break dropped", []ast.Node{txt("a"), sb(), txt("   b")}, "T(a) SB T(b)"},
		{"whitespace-only text before break removed", []ast.Node{esc(), txt("  "), sb(), txt("b")}, "ESC HB T(b)"},
		{"whitespace-only text after break removed", []ast.Node{txt("a"), sb(), txt("   "), esc()}, "T(a) SB ESC"},
		{"explicit hard break keeps type and trims", []ast.Node{txt("a "), hb(), txt(" b")}, "T(a) HB T(b)"},
		{"break at end dropped", []ast.Node{txt("a  "), sb()}, "T(a)"},
		{"break followed only by whitespace dropped", []ast.Node{txt("a  "), sb(), txt("  ")}, "T(a)"},
		{"hard break at end dropped", []ast.Node{txt("a"), hb()}, "T(a)"},
		{"break as first node kept when content follows", []ast.Node{sb(), txt("b")}, "SB T(b)"},
		{"no breaks untouched", []ast.Node{txt("a  b")}, "T(a  b)"},
		{"empty", nil, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := kinds(resolveLineBreaks(tc.in)); got != tc.want {
				t.Errorf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}

// TestParseInlineLineBreaks checks the rules end to end through parseInline,
// including the backslash form that parseEscape produces directly.
func TestParseInlineLineBreaks(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"a\nb", "T(a) SB T(b)"},
		{"a \nb", "T(a) SB T(b)"},
		{"a  \nb", "T(a) HB T(b)"},
		{"a  \n   b", "T(a) HB T(b)"},
		{"a\\\nb", "T(a) HB T(b)"},
		{"a \\\n b", "T(a) HB T(b)"},
		{"a  \nb  \nc", "T(a) HB T(b) HB T(c)"},
		{"a\nb  \nc", "T(a) SB T(b) HB T(c)"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got := kinds(parseInline(tc.in, ast.Pos{}, ExtAll, nil, nil, nil))
			if got != tc.want {
				t.Errorf("parseInline(%q)\ngot  %s\nwant %s", tc.in, got, tc.want)
			}
		})
	}
}

// TestParseParagraphTrimsEdges checks the block parser's contribution: the
// paragraph's initial and final whitespace is not content (CommonMark 4.8).
func TestParseParagraphTrimsEdges(t *testing.T) {
	doc := Parse([]byte("   a  \nb  "), ExtAll)
	kids := doc.Children()
	if len(kids) != 1 {
		t.Fatalf("got %d top-level nodes, want 1", len(kids))
	}
	para, ok := kids[0].(*ast.Paragraph)
	if !ok {
		t.Fatalf("got %T, want *ast.Paragraph", kids[0])
	}
	if got, want := kinds(para.Children()), "T(a) HB T(b)"; got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

// TestParseInlineMath checks the Pandoc tex_math_dollars rules at the AST
// level: inline `$…$`, display `$$…$$`, and every rejection condition.
func TestParseInlineMath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"$x$", "M(x)"},
		{"a $x$ b", "T(a ) M(x) T( b)"},
		{"a$b$c", "T(a) M(b) T(c)"},
		{"$x$ $y$", "M(x) T( ) M(y)"},
		{`$a \$ b$`, `M(a \$ b)`},
		{"$$x$$", "DM(x)"},
		{"$$ x $$", "DM(x)"},
		{"$$\nx\n$$", "DM(x)"},
		{"$$a$$ $b$", "DM(a) T( ) M(b)"},
		// Rejections: the dollar signs stay text.
		{"$20,000 and $30,000", "T($20,000 and $30,000)"},
		{"$ x$", "T($ x$)"},
		{"$x $", "T($x $)"},
		{"$x\n$", "T($x) SB T($)"},
		{"$x\ny$", "T($x) SB T(y$)"},
		{"$5 and `$x`", "T($5 and ) C($x)"},
		{"$a `b` c$", "T($a ) C(b) T( c$)"},
		{"$x$5", "T($x$5)"},
		{"$$", "T($$)"},
		{"$$x", "T($$x)"},
		{"$", "T($)"},
		{"$x", "T($x)"},
		{`\$x\$`, "ESC T(x) ESC"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got := kinds(parseInline(tc.in, ast.Pos{}, ExtAll, nil, nil, nil))
			if got != tc.want {
				t.Errorf("parseInline(%q)\ngot  %s\nwant %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseInlineMathDisabled(t *testing.T) {
	got := kinds(parseInline("$x$ and $$y$$", ast.Pos{}, ExtAll&^ExtMath, nil, nil, nil))
	if want := "T($x$ and $$y$$)"; got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestIndexUnescaped(t *testing.T) {
	tests := []struct {
		src  string
		from int
		sep  string
		want int
	}{
		{"a$b", 0, "$", 1},
		{`a\$b$`, 0, "$", 4},
		{"ab", 0, "$", -1},
		{"$$", 0, "$$", 0},
		{"x$$", 1, "$$", 1},
		{`\$$x$$`, 0, "$$", 4},
		{`a\`, 0, "$", -1},
		{"$", 0, "$$", -1},
		{"a$b", 2, "$", -1},
	}
	for _, tc := range tests {
		if got := indexUnescaped(tc.src, tc.from, tc.sep); got != tc.want {
			t.Errorf("indexUnescaped(%q, %d, %q) = %d, want %d", tc.src, tc.from, tc.sep, got, tc.want)
		}
	}
}
