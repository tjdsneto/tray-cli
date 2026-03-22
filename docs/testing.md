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

Coverage is a **signal**, not a goal: prioritize tests for pure logic, decoding, and HTTP adapters (httptest), and integration-style tests only where they add confidence without flaking.
