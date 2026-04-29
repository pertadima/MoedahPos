<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.

<!-- END:nextjs-agent-rules -->

---

# Agent Instructions — Frontend (Next.js)

## Verification (run from this directory)

After every change, run all three in order:

```bash
npm run lint -- --fix
npx prettier --write .
npm run type-check
```

Do not report "Done" until all three exit with zero errors.

- `lint --fix` — auto-fixes what it can; any remaining ESLint errors must be resolved manually.
- `prettier --write` — formats all files in place; re-check that no files are left dirty.
- `type-check` (`tsc --noEmit`) — zero TypeScript errors required.

## Coding Standards

- **No `any`** — use strict interfaces. All API response shapes must be defined in `src/types/`.
- **API calls** — all backend calls go through `src/lib/api/`. Never call `fetch` directly from a page or component.
- **Next.js version** — this project may use a non-standard version. Always read `node_modules/next/dist/docs/` before using routing APIs, server components, or middleware patterns.
- **Imports** — prefer named imports. Do not add barrel-file re-exports unless they already exist.

## Documentation Maintenance

Update `README.md` when:
- A new page or route is added.
- A new API client function is added to `src/lib/api/`.
- A new type is added that affects the public API contract.

