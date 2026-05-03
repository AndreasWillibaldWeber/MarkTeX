# Contributing to MarkTeX

Thank you for taking the time to contribute. This document explains how to report bugs, propose features, submit code, and add new Markdown constructs following the project's extension pattern.

---

## Table of contents

1. [Code of conduct](#code-of-conduct)
2. [Reporting bugs](#reporting-bugs)
3. [Requesting features](#requesting-features)
4. [Development setup](#development-setup)
5. [Coding standards](#coding-standards)
6. [Commit messages](#commit-messages)
7. [Branch naming](#branch-naming)
8. [Submitting a pull request](#submitting-a-pull-request)
9. [Adding a new Markdown construct](#adding-a-new-markdown-construct)
10. [Updating golden files](#updating-golden-files)
11. [Sign-off requirement](#sign-off-requirement)

---

## Code of conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating you agree to abide by its terms. Report unacceptable behaviour to the maintainers via a private issue or email.

---

## Reporting bugs

Use the [bug report issue template](.github/ISSUE_TEMPLATE/bug_report.md). Before filing:

- Search existing issues to avoid duplicates.
- Confirm the bug is reproducible on the latest commit of `main`.
- Include the minimal Markdown snippet that triggers the problem and the actual vs expected LaTeX output.

---

## Requesting features

Use the [feature request issue template](.github/ISSUE_TEMPLATE/feature_request.md). Describe the Markdown syntax you want to support, the LaTeX output it should produce, and the use case that motivates it.

---

## Development setup

Requirements: Go 1.21 or later.

```sh
git clone https://github.com/<owner>/marktex.git
cd marktex
make build    # compiles bin/marktex
make test     # runs all unit and golden-file tests
make lint     # runs go vet (and golangci-lint if installed)
```

No external dependencies are required beyond the Go standard library.

---

## Coding standards

- Format all code with `gofmt` before committing. The CI will reject unformatted files.
- Pass `go vet ./...` with zero warnings.
- Keep packages at their intended layer boundaries:
  - `internal/ast` — data only, no imports from other internal packages.
  - `internal/visitor` — imports `ast` only.
  - `internal/parser` — imports `ast` and `visitor`.
  - `internal/generator` — imports `ast` and `visitor`.
  - `pkg/transpiler` — public API, imports `parser` and `generator`.
- Do not add comments that describe *what* the code does — only *why* when the reason is non-obvious.
- Do not introduce new external module dependencies without prior discussion in an issue.

---

## Commit messages

Use the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <short summary>

<optional body>

Signed-off-by: Your Name <you@example.com>
```

**Types:** `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

**Scope** (optional): the package or subsystem affected, e.g. `parser`, `generator`, `ast`, `cli`

**Examples:**

```
feat(parser): add strikethrough delimiter detection
fix(generator): escape caret in text nodes
docs: document the 5-step extension pattern in CONTRIBUTING
test(golden): add snippet for labeled table floats
```

The summary line must be 72 characters or fewer and written in the imperative mood ("add", not "added" or "adds").

---

## Branch naming

| Purpose | Pattern | Example |
|---|---|---|
| New feature | `feat/<short-description>` | `feat/footnote-support` |
| Bug fix | `fix/<short-description>` | `fix/setext-heading-detection` |
| Documentation | `docs/<short-description>` | `docs/extension-guide` |
| Refactor | `refactor/<short-description>` | `refactor/inline-parser` |
| Tests | `test/<short-description>` | `test/golden-table-float` |

Branch off `main`. Do not commit directly to `main`.

---

## Submitting a pull request

1. Fork the repository and create a branch from `main`.
2. Make your changes following the coding standards above.
3. Add or update tests. Every new feature must include at least one unit test in `pkg/transpiler/transpiler_test.go` and a golden-file snippet pair in `testdata/snippets/`.
4. Run `make test` and confirm all tests pass.
5. Run `make lint` and resolve any warnings.
6. Open a pull request against `main` using the [PR template](.github/PULL_REQUEST_TEMPLATE.md).
7. Address review feedback with new commits — do not force-push a branch that is under review.
8. A maintainer will squash-merge once the PR is approved.

---

## Adding a new Markdown construct

The architecture enforces a strict five-step process. Every new node type must pass through all five steps to compile.

### Step 1 — Define the AST node

Add a `NodeType` constant to `internal/ast/node.go`. Add the concrete struct to `internal/ast/block.go` (for block nodes) or `internal/ast/inline.go` (for inline nodes). Embed `blockBase` or `inlineBase` to inherit `Children()`, `Position()`, and `AddChild()`. Leaf nodes (no children) must override `Children()` to return `nil`.

### Step 2 — Extend the Visitor interface

Add `EnterXxx` and `LeaveXxx` methods to the `Visitor` interface in `internal/visitor/visitor.go`.

### Step 3 — Add no-op defaults to BaseVisitor

Add the two corresponding methods to `BaseVisitor` in `internal/visitor/base.go`, both returning `WalkContinue`. This ensures existing visitors continue to compile without changes.

### Step 4 — Add dispatch cases to Walk

Add `case ast.NodeXxx:` entries to both the `enter` and `leave` functions in `internal/visitor/walk.go`.

### Step 5 — Implement in the generator (and parser)

Add the parser logic that produces the new node in `internal/parser/block.go` or `internal/parser/inline.go`. Add the `EnterXxx`/`LeaveXxx` methods to the `Generator` in `internal/generator/generator.go`.

If the node requires a new LaTeX package, set the corresponding flag in `generator.po` inside the Enter method so the package is included in standalone mode.

### Add tests

- One or more unit tests in `pkg/transpiler/transpiler_test.go`.
- A snippet pair `testdata/snippets/<name>.md` and `testdata/snippets/<name>.tex`. Generate the golden file with:

```sh
./bin/marktex testdata/snippets/<name>.md > testdata/snippets/<name>.tex
```

Then verify the content is correct before committing it.

---

## Updating golden files

When an intentional output change affects existing golden files, regenerate them with:

```sh
# Regenerate all golden files
go test ./testdata/snippets/ -update

# Regenerate via the shell script
./scripts/test-snippets.sh --update
```

Review every changed `.tex` file in the diff before committing. Golden files are the source of truth for output correctness.

---

## Sign-off requirement

All commits must include a `Signed-off-by` trailer. This constitutes your agreement to the [Developer Certificate of Origin](https://developercertificate.org/):

```sh
git commit -s -m "feat(parser): add footnote support"
```

Pull requests containing unsigned commits will not be merged.
