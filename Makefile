.PHONY: test test-cover test-cover-html release publish-release

# Default: same as CI — race detector, no cache reuse surprises.
test:
	go test ./... -race -count=1

# Writes coverage.out; prints total line coverage (last line of go tool cover -func).
test-cover: coverage.out
	@go tool cover -func=coverage.out | tail -1

coverage.out:
	go test ./... -race -count=1 -coverprofile=coverage.out -covermode=atomic

test-cover-html: coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Wrote coverage.html"

release:
	./scripts/build-release.sh

# Usage: make publish-release VERSION=v1.2.3
publish-release:
	@if [ -z "$(VERSION)" ]; then echo "usage: make publish-release VERSION=v1.2.3" >&2; exit 1; fi
	./scripts/publish-release.sh "$(VERSION)"
