# Project Standards & Guard Rails

## Verification Workflow

Every time a task is completed, you MUST run the following commands to verify the code:

### Backend (Go)

- Navigate to `/backend`
- Run `go fmt ./...`
- Run `golangci-lint run`
- **Rule:** If linting fails, fix the code and re-run until clean.

### Frontend (Next.js)

- Navigate to `/frontend`
- Run `npm run lint -- --fix`
- Run `npx prettier --write .`
- **Rule:** Do not report the task as "Done" until these commands pass with zero errors.

## Coding Style

- Go: Strict error handling is required (no ignored errors).
- TypeScript: No `any` types; use strict interfaces.
