Adjacent runs: ***testing****without***spaces** render as three spans.

Underscores work too: _italic_, __bold__ and ___both___, while snake_case_name
stays literal inside a word, as does 2 * 3 * 4 with spaces around the stars.

Nesting: *foo **bar** baz*, **foo *bar* baz**, *(*inner*)* and foo***bar***baz.

Emphasis may span a line
*like this one which continues*
onto the next source line, and left-over delimiters such as **foo* keep their
unmatched star. Code spans win: *not `*` closed here* still is emphasis.

- **bold item** with *italic*
- `code` then _emphasis_

| Column | Value |
|--------|-------|
| *a* | **b** |
