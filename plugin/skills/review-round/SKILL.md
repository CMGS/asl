---
name: review-round
description: The cocoonstack batch-end quality convergence loop — style walkthroughs per /code, /code-rs, /code-py and /code-ts (subagents read files in full, Fable adjudicates and runs the final Self-Check), multi-agent correctness review with model tiering, /simplify lenses, comment tightening, the single platform-symmetric gate list, and convergence to 闭环 (zero new findings, all gates green) without manufacturing churn. Invoke at the end of a branch/batch or for a whole-repo audit, never per commit.
---

# Review Round — batch-end convergence loop

Cadence rule (CMGS 2026-07-07): this loop runs ONCE at the end of a branch/batch,
before the final push. Mid-batch commits are gated by build/test/lint only.
Hardware evidence rules are unchanged — each feature still closes its own loop.

## 0. Scope and the hygiene ledger

- Batch end: review the accumulated diff (`git diff <base>..HEAD` + working tree).
- Whole-repo audit: every source file is in scope; the bar is repo-level, the
  diff is just the means — but "in scope" is decided by the ledger, not by
  re-reading everything. Each repo has a ledger at
  `~/Documents/workspace/cocoonstack/.hygiene/<repo>.tsv` — the workspace
  root, which is not a repository: cocoon-specs' directory set is closed and
  every other repo is public — (`path, style_blob, judge_blob, verdict,
  note`, keyed by git blob id so an
  unchanged file keeps its verdict across commits; the header pins the rules
  commit — the asl checkout HEAD — and a rules change invalidates every row).
  Open the round with `~/Documents/workspace/asl/scripts/ledger.py diff <repo>
  <ledger> --lens style|judge`: the output IS the reader file list. Files
  that are not due are not read (why rounds kept re-mining the same files,
  CMGS 2026-09-07: readers sample, and only rejections were recorded, never
  the clean verdicts). Docs pages are read in full every round — a doc goes
  stale when the code moves, not when the page does.
- Close the round with `ledger.py mark … --lens <lens> --verdict CLEAN|KEPT
  [--note "<kept items and why>"] <files>` for every file a reader cleared
  or the adjudicator kept; a round whose ledger has not been marked is not
  closed, and the round report names the ledger's due count before and after.
  A full re-read happens only when the rules commit changed or CMGS asks.

## 1. Style walkthrough — subagents read, Fable adjudicates

Load /code (Go), /code-rs (Rust), /code-py (Python), /code-ts (TypeScript) —
reload them here even if loaded earlier in the session; a compacted context
keeps an impression of the rules, not the rules. For Go, also run the
`use-modern-go` CLI (`list --file-path` on a representative file) and paste
its output into the reader briefs — it is the version-correct idiom catalog;
the /code table alone misses post-1.25 entries. Opus/sonnet subagents read
every file in scope in full via `Read` — never grep/sed sampling, which loses
the context a verdict needs — against the enumerated Style Self-Check and
return a per-file ledger with line numbers. Fable plans the batches, picks
models (sonnet for mechanical scans, opus for judgment-heavy reads),
adjudicates every finding against the source, and walks the final diff of
every touched file through the Self-Check itself — that last pass is never
delegated: layout misses slipped through twice when it was (CMGS caught both).
Specific trap: a type's declaration must be immediately followed by its
complete method set; producer methods of another type must not wedge in between.

## 2. Correctness review — adversarial multi-agent

- Spawn 2-3 independent review agents over the diff with distinct lenses
  (correctness/edge cases, concurrency/locks, protocol/API contract). Same
  reading discipline as §1: full `Read`, per-file ledger
  (`file:line — lens — claim — failure scenario`).
- Model tiering (CMGS rule): sonnet for mechanical scans and full-file reads,
  opus for judgment-heavy reads; only adversarial verify/judge earns the top
  model. Never run every agent on the top model.
- Verify each finding adversarially before accepting it — reviewers must prove
  a failure scenario, not pattern-match. Track record: this loop found ~27 real
  bugs across silkd's five rounds; it also produces plausible-but-wrong claims
  that die under verification.
- Agents sometimes deliver a report then idle without sending — nudge each via
  SendMessage for its final verdict.

## 3. /simplify lenses

Run the four lenses (reuse, simplification, efficiency, altitude). Every finding
is either applied or skipped WITH A RECORDED REASON — one line per finding in
the round report: `file:line — lens — finding — applied <sha> | skipped: <reason>`.
Silent drops are not allowed.

## 4. Comment tightening

The /code hard budget (Style Self-Check item 2) applies as written: default NO
comment; a genuinely load-bearing WHY gets one line in the STE 101 register; no
justification narrative, no perf-benefit sentences in godoc; `// SAFETY:` lines
stay. Fix, sweep, and cut commits add zero comments and never raise a file's
comment count — count it and put both numbers in the report:
`git diff -U0 <base>..HEAD -- '*.go' '*.rs' '*.ts' '*.tsx' | grep -cE '^\+[[:space:]]*//'`
against the same count for `^-`.

## 5. Prod-LOC budget — quality rounds shrink, not grow

Rule (CMGS 2026-07-17): a review/audit round defaults to net-zero or net-negative
production lines. Evidence behind the rule: by July 2026 cocoon's comments were
only 5% of prod code, but review passes contributed 34% of the month's prod
growth (+930 of +2,733) — the bloat lever is round-generated additions, not
comments or tests.

- Measure, don't estimate: split the round's own diff prod vs test via
  `git diff --numstat` and put both net numbers in the round report.
- Tests are exempt — regression pinning is the paid cost of 闭环. A bug-fix
  commit is justified by the bug it fixes (name it); it still appears in the
  item-by-item list.
- Every net-positive prod contribution the round itself generates (new mechanism,
  dedup helper, defensive branch) must be justified item-by-item in the report.
  Precedent: PR #117's +863 prod survived only because all 6 mechanisms + 11
  helpers were individually adjudicated. Unjustified additions are churn
  (section 7's judgment failure), not thoroughness.

## 6. Gates — platform-symmetric

This is the single gate list; the language skills, /loc-justify and
/codexreview point here. Paste each gate's real output in the report (§8).

- Go: `GOWORK=off make lint` and `GOWORK=off make fmt-check` on BOTH
  `GOOS=linux` and `GOOS=darwin` (the workspace `go.work` resolves sibling
  checkouts and diverges from CI on every repo, not just sandbox);
  `asl ./...` and `GOOS=linux asl ./...` with zero findings;
  `GOWORK=off go test -race -count=1 ./...`. Multi-module repos count the
  `0 issues.` lines — modules × 2 GOOS, one short means lint died in a module.
- Rust: `cargo fmt --all -- --check && cargo clippy --workspace --all-targets
  -- -D warnings && cargo test --workspace` on mac AND in a `rust:1` Linux
  container (`--platform linux/arm64` on Apple Silicon — amd64 emulation makes
  `Command::spawn` of a missing binary return `Ok` and fakes results). A
  Linux-only clippy lint and a mac-green/Linux-broken PTY have both slipped a
  mac-only check. Env-gated suites (`GW_TEST_*`, `CP_TEST_*`) silently
  early-return when unset — report them as skipped, or run them and cite
  server-side artifacts as proof.
- Python: `ruff format --check && ruff check && python -m pytest` from the
  package root (bare `pytest` leaves the cwd off `sys.path`).
- TypeScript: `npm run typecheck && npm test && npm run build`;
  `grep -rnE '@ts-ignore|: any\b' src/` returns nothing; Playwright e2e when
  routing, auth, or page flows changed.
- fmt authority for cocoon repo: CI golangci-lint fmt --diff @pinned version;
  do NOT trust local gofumpt@latest or run repo-wide `make fmt`.

## 7. Convergence = 闭环

A round with zero newly applied findings and all gates green closes the loop —
then STOP. When `ledger.py diff` lists no file for either lens and `asl` is
clean on both GOOS, the round is closed before any reader is spawned: report
that state as the outcome. When the bar is already met, "no change needed" backed by the audit
evidence is the correct outcome; manufacturing churn to look productive is a
judgment failure ("if not necessary, don't change/push/release").

## 8. Report

State what ran (agents, lenses, gates with real output), findings applied,
findings skipped with reasons, the round's net prod/test LOC (with per-item
justification for any prod increase), and the convergence verdict. A declared step that
did not actually run is a violation with no exceptions — report skipped as
skipped.
