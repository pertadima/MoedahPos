# Project Standards & Guard Rails

## Verification Workflow

Every time a task is completed, you MUST run the following commands to verify the code:

### Backend (Go)

- Navigate to `/backend`
- Run `go fmt ./...`
- Run `golangci-lint run`
- **Database Schema:** If any changes are made to `backend/migrations`, you MUST update `backend/database.mmd` to reflect the current schema.
- **Rule:** If linting fails, fix the code and re-run until clean.

### Frontend (Next.js)

- Navigate to `/frontend`
- Run `npm run lint -- --fix`
- Run `npx prettier --write .`
- Run `npm run type-check`
- **Rule:** Do not report the task as "Done" until these commands pass with zero errors.

## Coding Style

- Go: Strict error handling is required (no ignored errors).
- TypeScript: No `any` types; use strict interfaces.

## Documentation Maintenance

- **Architecture & Interfaces:** If any changes are made to the system's architecture, API surfaces, network routing, or domain logic across the stack, you MUST continuously update both `backend/README.md` and `frontend/README.md` to reflect the accurate, up-to-date system architecture and endpoints.
