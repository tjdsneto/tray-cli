# Unit testing

Run the full suite (matches CI):

```bash
make test
# or: go test ./... -race -count=1
```

## Coverage

Generate a merged profile for all packages and print the **total** statement coverage:

```bash
make test-cover
```

Open an HTML report (writes `coverage.html`; both `coverage.out` and `coverage.html` are gitignored):

```bash
make test-cover-html
open coverage.html   # macOS
```

Coverage is a **signal**, not a goal: prioritize tests for **pure helpers** (URL/query builders, body maps, tray resolution without I/O, format flags without cobra), decoding, and HTTP adapters (`httptest`). Integration-style tests only where they add confidence without flaking.

Related project rules: [`CLAUDE.md`](../CLAUDE.md) and [`.cursor/rules/pure-functions-testing.mdc`](../.cursor/rules/pure-functions-testing.mdc). When you change those, keep **Cursor rules and `CLAUDE.md` in sync** (see [`.cursor/rules/cursor-claude-rules-parity.mdc`](../.cursor/rules/cursor-claude-rules-parity.mdc)).
