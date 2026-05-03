---
Author: Andreas W. Weber
Title: Advances in Markdown Processing
Subtitle: A Comprehensive Survey
Date: 2026-05-03
MakeTitlePage
MakeTOC
MakeLOF
MakeLOT
MakeLOL
C#01:knuth1986texbook
C#02:lamport1994latex
C#03:gruber2004markdown
C#04:commonmark2021spec
F#01:fig_pipeline
F#02:fig_ast
T#01:tab_comparison
T#02:tab_performance
---

# Introduction

Markdown-to-LaTeX conversion has been studied extensively [C#01]. The `TeX` typesetting system [C#02] remains the standard for academic publishing, while Markdown [C#03] has become the dominant lightweight markup language for technical writing. This paper presents MarkTeX, a transpiler that bridges the two.

The CommonMark specification [C#04] defines the authoritative Markdown grammar; our parser targets a strict subset of it, extended with custom syntax for cross-references and document metadata.

## Motivation

Three problems motivate this work:

1. Existing tools [C#01] require **complex configuration** to produce well-structured LaTeX output.
2. Reference management is *manual* in most pipelines.
3. ***No open standard*** exists for embedding LaTeX-specific constructs in Markdown source.

The following sections address each problem in turn.

---

# Architecture

As shown in Figure [F#01], the MarkTeX pipeline consists of three stages: **lexical analysis**, *AST construction*, and ***code generation***.

![F#01:0.9:h][The three-stage transpilation pipeline, adapted from [C#03].](figures/pipeline.png)

## Parser

The parser follows the two-phase approach of [C#04]: a block pass identifies structural elements, and an inline pass handles emphasis, code spans, and custom references. The grammar is implemented without backtracking; ambiguous constructs default to the CommonMark resolution rules.

Block-level elements recognised by the parser:

- **Headings**: ATX (`# H1` through `###### H6`) and setext (`===`, `---`)
- **Lists**: bullet (`-`, `*`, `+`) and ordered (`1.`, `2.`, …) with tight/loose detection
- **Code**: fenced (` ``` `) and indented (4-space)
- **Tables**: GFM pipe tables with `:---`, `---:`, `:---:` alignment
- **Block quotes** and **thematic breaks**

Inline elements recognised:

- Emphasis (`*`, `_`), strong (`**`), strong-emphasis (`***`)
- Code spans (`` ` ``), strikethrough (`~~`)
- Links `[text](url)` and images `![alt](url)`
- Custom refs: `[C#key]`, `[F#key]`, `[T#key]`

## AST

The internal AST is depicted in Figure [F#02]. Every node carries a source position for error reporting.

![F#02:0.7:ht][AST node hierarchy. Block nodes are shown above the dashed line; inline nodes below. See Table [T#01] for node-count statistics.](figures/ast.png)

Each node type implements a `Type() NodeType` method. Traversal is handled by a single `Walk` function; visitors implement `Enter*/Leave*` pairs and embed `BaseVisitor` for default no-op behaviour.

### Extension Mechanism

> Adding a new construct requires five deterministic steps: define the AST node, extend the `Visitor` interface, provide a `BaseVisitor` no-op, register the dispatch case in `Walk`, and implement the generator method. The pattern was validated against [C#02] and [C#04].

---

# Evaluation

## Feature Comparison

Table [T#01] compares MarkTeX against existing tools.

|- T#01:ht -|
|- Feature comparison of Markdown-to-LaTeX tools. Based on criteria from [C#01] and [C#02]. -|
| Tool     | Typed AST | Extensions | Cross-refs | Speed   |
|:---------|:---------:|:----------:|:----------:|--------:|
| MarkTeX  | yes       | yes        | yes        | 12 ms   |
| pandoc   | yes       | yes        | partial    | 85 ms   |
| markdown | no        | no         | no         | 3 ms    |
| kramdown | partial   | limited    | no         | 22 ms   |

## Performance

Table [T#02] reports processing times on a corpus of 1 000 documents.

|- T#02:h -|
|- Processing time by document size (adapted from [C#01]). See Figure [F#01] for pipeline details. -|
| Input size | Parse time | Generate time | Total   |
|:----------:|:----------:|:-------------:|--------:|
| 1 KB       | 0.5 ms     | 0.2 ms        | 0.7 ms  |
| 10 KB      | 4.2 ms     | 1.8 ms        | 6.0 ms  |
| 100 KB     | 38 ms      | 15 ms         | 53 ms   |

---

# Implementation

## Code Example

The public API consists of three functions:

```go
// Parse Markdown with all extensions enabled.
doc := parser.Parse(src, parser.ExtAll)

// Generate LaTeX fragment output.
err := generator.Generate(doc, generator.Options{}, os.Stdout)

// Or use the high-level transpiler.
out, err := transpiler.TranspileString(src, transpiler.Options{
    Standalone: true,
    Extensions: transpiler.ExtAll,
})
```

For standalone documents, `MakeTitlePage` and `MakeTOC` entries in the definition block automatically place `\maketitle` and `\tableofcontents` at the start of the body.

An indented example showing raw AST traversal:

    doc := parser.Parse([]byte("# Hello"), 0)
    visitor.Walk(doc, myVisitor)

## Special Characters

LaTeX special characters in Markdown source are automatically escaped. For example, a 50% discount on the price of $10.00 & free shipping becomes `50\% discount on the price of \$10.00 \& free shipping` in the output. The `_` in identifiers like `my_variable` is escaped to `my\_variable`.

---

# Conclusion

As demonstrated by Figures [F#01] and [F#02] and Tables [T#01] and [T#02], MarkTeX [C#03] achieves competitive performance with a clean, extensible design. The five-step extension pattern [C#04] ensures that new constructs can be added without modifying existing code.

Future work will address the CommonMark delimiter-stack algorithm [C#04] for fully compliant emphasis parsing.

---

See [C#01], [C#02], [C#03], and [C#04] for full bibliographic details.
