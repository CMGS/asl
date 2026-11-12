---
name: code-ts
description: TypeScript/React code review & refactor standard for cocoonstack projects — the TS counterpart to /code (Go), /code-rs (Rust) and /code-py (Python). Load for any TypeScript or React work in cocoonstack repos; subagents read files in full, the session model adjudicates and runs the final Style Self-Check on every touched file.
---

# TypeScript Code Review & Refactor Standard

The TypeScript counterpart to `/code` (Go), `/code-rs` (Rust) and
`/code-py` (Python). Same discipline, TS/React idioms. Derived from the
cocoonstack web code (`gateway/control-plane/web`) plus the Google
TypeScript Style Guide and React-19-era community practice (compiler-first,
`useEffect` as a last resort).

## Style Self-Check (mandatory)

After writing or editing any `.ts`/`.tsx` file, re-scan the final file
top-to-bottom against this list and fix every violation before presenting
or committing. The rules apply to every file you touch — if an edited file
already violates them, fix it in the same change. Pure moves and file splits
are not exempt.

Division of labor on any pass wider than one diff: opus/sonnet subagents read
the files — each file in full via `Read`, never grep/sed sampling, which loses
context — against an explicitly enumerated rule list and return a per-file
verdict ledger with line numbers; the session model (Fable) plans, batches,
picks models (sonnet for mechanical scans, opus for judgment-heavy reads),
adjudicates every finding by re-reading the cited source, and walks the final
diff of every touched file through this checklist itself. Subagent output is
unverified text until adjudicated. Until the app adopts eslint, module layout,
floating promises and accessibility have no mechanical gate — they are
walkthrough-only.

1. **Module layout** — order: imports (three groups: framework/external
   packages → internal modules → styles; type-only imports use `import
   type` or the inline `type` modifier, never a bare value import of a
   type) → module constants → the module's primary export (page/component/
   fn) → private subcomponents and free helpers below it. One route page
   per file (`pages/XxxPage.tsx`, default export); shared presentational
   pieces in `components/`, named exports. Nothing sits between a
   component and the helpers only it uses.
2. **Comments** — same hard budget as the sibling standards: default zero.
   The complete allowance is a one-line JSDoc on an exported item whose
   contract the name cannot carry and one-line WHY-comments carrying
   information the code cannot (protocol quirks, browser behavior, fallback
   rationale), both in the STE 101 register (Simplified Technical English:
   short, present tense, one fact per sentence, lowercase-first inline; PR
   inline replies use the same register). Everything else is a violation —
   restated code, section labels, edit narration, the same fact at JSDoc and
   call site, any comment in test code. A fix, sweep, or cut commit adds
   zero comments (rationale goes in the commit message) and never raises a
   file's comment count. Apply the deletion test to every comment in the
   file, then once more to the batch's total density.
3. **Types** — `strict` plus `noUncheckedIndexedAccess` are the floor,
   never weakened per-file. No `any`; hostile/unknown input is `unknown`
   narrowed by `instanceof`/discriminants. No `@ts-ignore`; a justified
   `@ts-expect-error` carries a reason. Wire types mirror the server JSON
   byte-for-byte (snake_case fields; Go pointer fields become
   `?: T | null`) and live in `types.ts`. Enum-like values are unions of
   string literals, not `enum`. Exported functions declare their return
   type; internal code lets inference work.
4. **React discipline** — function components only. Derive during render —
   never mirror props or other state into `useState` when it can be
   computed. `useEffect` is a last resort, reserved for syncing with
   something outside React (network via the shared hooks, timers, scroll,
   globals); every effect that subscribes or schedules returns a cleanup,
   and async effects guard against post-unmount `setState` (the `active`
   flag pattern). State lives in the closest common parent. No manual
   `useMemo`/`useCallback`/`memo` unless a measurement demands it — the
   compiler era made speculative memoization noise. List keys are stable
   domain ids, never array indexes when rows can move.
5. **English only** — no Chinese in code, comments, strings, or
   identifiers. (CJK *test fixtures* that exercise multi-byte handling are
   data, not prose — exempt.)
6. **Errors and promises** — no floating promises: a fire-and-forget call
   is explicitly `void`-marked. `catch (err: unknown)` narrows before use;
   user-visible failures render a message (the `useAction`/`useAPI` error
   channel), never `console.log`. One typed error class per API seam
   (`APIError` with `status`), checked via `instanceof`, never message
   sniffing.
7. **Modern platform** — use what the toolchain already provides: `??`
   and `?.` over `||` chains guarding falsy-but-valid values, hoisted
   `Intl.NumberFormat`/`DateTimeFormat` instances over per-render
   construction or hand-rolled formatting, template literals, `satisfies`
   for typed literal maps, native `fetch` — no axios/lodash-style
   dependencies for stdlib work. Accessibility is part of the standard:
   icon-only buttons carry `aria-label`, decorations `aria-hidden`,
   interactive semantics use real elements (`button`, `label`, roles).

## Commit Rules

- Never add `Co-Authored-By` or any other trailers to commit messages
- No Claude / AI / Anthropic references anywhere in commits, PR bodies, or issues
- Keep commit messages concise: one-line summary, optional body for why
- Layout-only normalization rides in its own `review:` commit, never mixed
  with behavior changes — the message names the rule being applied
- A bug found during a style/audit round is fixed in that round, in its own
  commit with a regression test where one is cheap — never parked in a
  report, never folded into a `review:` commit. Gate the fix on re-reading
  the source and proving the failure reachable from a real call site

## Git Workflow

- Always rebase, never merge — keep linear history with no merge commits
- `git pull --rebase` before pushing
- `git rebase` to integrate upstream changes, not `git merge`
- If conflicts arise during rebase, resolve per-commit, not squash-and-pray
- Never rewrite already-pushed history (rebase/amend/force-push) without
  explicit per-conversation authorization — rebase freely only on commits
  that haven't left the machine

## Batch-End Checklist

Runs once at the end of a branch/batch before the final push (the
/review-round cadence); mid-batch commits are gated by step 4 only:

1. **Style Self-Check** — walk every changed file, tests included, against
   the checklist yourself — the final pass after the subagent ledgers, never
   a substitute for them. For a fix, sweep, or cut commit count comment
   lines in the staged diff: `git diff --cached -U0 -- '*.ts' '*.tsx' |
   grep -cE '^\+[[:space:]]*//'` must not exceed the `^-` count (the commit
   hook checks the same)
2. **Senior Tech Review** — review all changed code for correctness, edge
   cases, and render/data-flow consistency; no contrived scenarios
3. **Run `/simplify`** — the four lenses (reuse, simplification,
   efficiency, altitude) over the whole diff, tests included; fix all
   actionable findings. The order is fixed: personal walkthrough →
   `/simplify` → gates
4. **Gates** — `npm run typecheck && npm test && npm run build` clean, and
   `grep -rnE '@ts-ignore|: any\b' src/` returns nothing; run the Playwright
   e2e (`npm run e2e`) when routing, auth, or page flows changed. Import
   grouping, floating promises and accessibility have no gate until the app
   adopts eslint (`typescript-eslint` `no-floating-promises`, `jsx-a11y`) —
   the next batch that touches the web app adds `npm run lint` here and in
   CI; until then those rules are walkthrough-only and the report says so

## Cocoon TS Style Reference

Concrete patterns derived from `control-plane/web`.

### Project Shape

```
src/
  api.ts            # ONE fetch wrapper: CSRF, error envelope, 401 handling
  types.ts          # wire types, mirror the server JSON exactly
  hooks.ts          # useAPI (GET + reload), useAction (mutation + busy/error)
  format.ts         # hoisted Intl formatters + tiny label helpers
  components/UI.tsx # shared presentational components (named exports)
  pages/XxxPage.tsx # one route per file, default export
  App.tsx           # auth context + shell + role-gated route table
  main.tsx          # thin: createRoot + providers only
```

- Pages never call `fetch` directly — all traffic goes through `api<T>()`,
  which owns CSRF headers, the error envelope (`{error:{message}}`),
  `APIError`, and the global-401 → forced re-login hook.
- Data GETs go through `useAPI<T>(path | null)` — `null` path means "not
  entitled, don't fetch". Mutations go through `useAction()`; both expose a
  single error string the page renders. New pages must not invent parallel
  loading/error scaffolding.
- Role gating is data, not branches: the `navItems` table carries `minRole`
  and both the sidebar and the route table filter on `roleRank` — extend
  the table, don't add ad-hoc conditionals. The server re-checks every
  privileged call; UI gating is UX, not security.

### Types and Wire Contracts

- `types.ts` is the single mirror of the server API: field names snake_case
  as on the wire; optional-and-nullable Go fields are `?: T | null`;
  state enums are literal unions (`"active" | "banned" | ...`) so `switch`
  narrowing is exhaustive.
- Request payloads are shaped at the call site with `jsonBody(...)`;
  response envelopes are typed inline (`useAPI<{ keys: AccessKey[] }>`) —
  no wrapper-interface zoo for one-shot envelopes.
- Props: inline type literal for one-off component props; a named
  `interface` only when the shape is reused or exported.

### State and Effects

- Derive, don't store: filtered lists, totals, flags computed in render
  from `data` and props. `useState` is for genuine inputs (form fields,
  open/closed, selected tab).
- The debounce pattern: `useState` raw value + `useEffect` with
  `setTimeout`/`clearTimeout` cleanup writing the debounced value.
- Async effect cleanup: `let active = true; ... return () => { active =
  false; }` before any post-await `setState`.
- Context: one `createContext` + a `useX()` accessor that throws when the
  provider is absent (`useAuth`), so consumers never null-check.

### Tests

- Vitest + Testing Library; queries by accessible role/name
  (`findByRole("link", { name: /usage/i })`), not test ids or CSS.
- Network is mocked at the boundary: `vi.stubGlobal("fetch", ...)`
  returning real `Response` objects; `cleanup()` +
  `vi.unstubAllGlobals()` in `afterEach`.
- Playwright e2e drives real role flows against the dev-seeded backend;
  assertions use the same accessible queries.
- Zero comments in test code.

### Build & CI

- Vite + `tsc -b` (build runs typecheck); `npm run typecheck` for the fast
  gate. CI: `npm ci && npm test && npm run build` then Playwright with
  installed browsers, path-filtered to the web directory; `npm run lint`
  joins the chain once eslint lands.
- Dependencies stay minimal: react, react-dom, react-router-dom at
  runtime; everything else dev. Adding a runtime dependency needs a reason
  a stdlib/platform API can't cover.

## Documentation

- README per app: role, stack, dev/build/test commands. English only,
  reflecting actual current behavior.
