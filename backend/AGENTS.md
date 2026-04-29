# Agent Instructions — Backend (Go)

## Verification (run from this directory)

After every change, run all three in order:

```bash
go fmt ./...
golangci-lint run --timeout=5m
go test ./...
```

Do not report "Done" until all three exit clean.

- `go fmt` — must produce no diff.
- `golangci-lint` — fix every reported issue; re-run until zero warnings.
- `go test ./...` — all tests must pass; no skipped failures.

## Schema Changes

Any modification to `migrations/` **must** be followed by an update to `database.mmd`.

## Coding Standards

- **Error handling:** Never ignore errors. `_ = err` is forbidden unless explicitly justified in a comment.
- **Package structure:** `domain → repository → service → handler`. Do not skip layers.
- **Mocks:** Every new public method added to an interface must have a corresponding mock in `internal/*/mocks/`. Hand-edit or regenerate with mockery.
- **Goose migrations:** All SQL migration files must begin with `-- +goose Up` / `-- +goose Down` markers.
- **Linting:** `golangci-lint` config lives in `.golangci.yml`. Do not suppress warnings with `//nolint` unless the lint rule is genuinely inapplicable — add a reason comment when you do.
