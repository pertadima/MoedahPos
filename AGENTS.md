# Agent Instructions — Moedah POS

## Verification Workflow

Every time a task is completed, you MUST run the following verification commands before reporting "Done".

---

### Backend (Go) — run from `backend/`

```bash
go fmt ./...
golangci-lint run --timeout=5m
go test ./...
```

- If `golangci-lint` fails, fix the reported issues and re-run until clean.
- If `go test` fails, fix the failing tests before continuing.
- **Database Schema:** Any change to `backend/migrations/` MUST be followed by an update to `backend/database.mmd` to keep the ERD accurate.

---

### Frontend (Next.js) — run from `frontend/`

```bash
npm run lint -- --fix
npx prettier --write .
npm run type-check
```

- Do not report the task as "Done" until all three commands exit with zero errors.
- `type-check` (`tsc --noEmit`) must pass with no TypeScript errors.

---

## Coding Standards

### Go
- Strict error handling — no ignored errors (`_ = err` is forbidden unless explicitly justified).
- Follow existing package structure: `domain → repository → service → handler`.
- All new public methods on interfaces must have a corresponding mock updated in `internal/*/mocks/`.

### TypeScript / Next.js
- No `any` types — use strict interfaces.
- All API response types must be defined in `frontend/src/types/`.
- This project may use a non-standard Next.js version. Read `node_modules/next/dist/docs/` before writing routing or server-component code.

---

## Documentation Maintenance

- **API changes:** Update `backend/README.md` and `frontend/README.md` when endpoints are added, removed, or modified.
- **Architecture changes:** Update both READMEs when domain logic, service interfaces, or routing changes.
- **ERD:** Keep `backend/database.mmd` in sync with every migration.
