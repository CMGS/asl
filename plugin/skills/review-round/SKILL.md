---
name: review-round
description: The cocoonstack batch-end quality convergence loop — style walkthroughs per /code, /code-rs, /code-py and /code-ts (subagents read files in full, Fable adjudicates and runs the final Self-Check), multi-agent correctness review with model tiering, flow readers over crash windows, /simplify lenses, comment tightening, the single platform-symmetric gate list, hardware evidence rules, a coverage table, and convergence to 闭环 (zero new findings, all gates green) without manufacturing churn. Invoke at the end of a branch/batch or for a whole-repo audit, never per commit.
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
  root, which is not a repository; cocoon-specs' directory set is closed and
  every other repo is public — `path, verdict, note, lens:blob,...`, keyed by
  git blob id so an unchanged file keeps its verdict across commits. The rules
  key in the header is a hash of the asl ANALYZER sources, so editing the
  ledger tooling does not invalidate a verdict; a rule change does.
- **A lens is any name the round uses** (style, judge, loc, comments, billing,
  fleet, ...), and a file carries one blob per lens it was marked for. A lens
  that never ran leaves every file due for it: "not yet read" and "changed
  since read" are the same answer, which is the only safe default. Open the
  round with `ledger.py diff <repo> --lens <name>`: the output IS the reader
  file list. `ledger.py lenses <repo>` prints which lenses have actually run
  and over how much of the repo — read it before claiming a round is closed,
  because a clean `diff` for two lenses says nothing about the other eight.
  It lists only lenses that were ever recorded, and the Go lenses never see a
  non-Go file, so also run `ledger.py diff <repo> --lens infra` (scripts,
  Dockerfiles, workflows, Makefile, release config). cocoon 2026-11-05: "all
  lenses 0 due" was reported while 56 infra files had never been read.
- **Flows are the cross-file axis a per-file lens cannot see.** Declare them in
  `<ledger dir>/<repo>.flows` as `name<TAB>path [path ...]` — admission through
  settle, the realtime turn, one cross-protocol direction, the fleet's shared
  state. Address them as `@name`; a flow falls due when any member's blob
  changes even when every member is individually clean. `ledger.py diff <repo>
  --lens <name> --flows` lists them, and `ledger.py flows <repo>` shows their
  members and freshness. Every bug the gateway's 2026-09-20 round found lived
  on a flow, not in a file.
- **A whole-repo audit runs one flow reader per due flow** (opus): walk the
  call order, list each pair of adjacent side effects, and answer for the
  window between them: does retry or gc converge, is the failure loud, is the
  residue bounded. Return a window table. cocoon 2026-11-04: per-file readers
  and codex missed both HIGH bugs; only flow readers found them. A flow
  re-read for changed windows only is marked with a note that says so.
- Close the round with `ledger.py mark … --lens <lens> --verdict CLEAN|KEPT
  [--note "<kept items and why>"] <path|@flow>...` for every file a reader
  cleared or the adjudicator kept; a round whose ledger has not been marked is
  not closed, and the round report names the ledger's due count before and
  after, per lens.
A full re-read happens only when the analyzer rules changed or CMGS asks.

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
  SendMessage for its final verdict. Agent reports also arrive truncated: ask
  for the remainder before adjudicating, never adjudicate a cut-off report.
- **Lenses every whole-repo brief carries** (sandbox 2026-11-08: a solo
  whole-repo read reported zero bugs; a 14-reader pass the same day found an
  egress allow-list bypass, an SDK stream reuse, an adapter leak, a pty owner
  bug and nine doc drifts — each one a lens the solo read never ran):
  - *Check vs use.* For every guest or tenant input that a policy, auth or
    ACL check passes, follow the same value to every later parser (resolver,
    dialer, `http.Transport`, path join, shell) and prove both read it alike:
    IPv6 zones, IDNA and Unicode case folding, percent-encoding, trailing
    dots, mapped and embedded IPv4. Reproduce a suspected differential with
    a ten-line program before ruling on it.
  - *Error matching on control flow.* For every `errors.Is`/`As`/`AsType` that
    decides a branch, list the wrapper types that can reach it; a wrapper
    with `Unwrap` makes `As` match where a type assertion would not.
  - *Sibling sweep.* For each fix commit since the last audit, check its
    siblings: the other SDK languages, the other adapters, the other doors
    of the same proxy.
  - *Feature → docs.* For each feature commit since the last audit, find
    every doc, godoc and tool description that states the rule it changed.
    Reading docs against code catches wrong text, not stale text.
  - *Runtime semantics.* Behaviour decided by the kernel or a runtime
    (devpts ownership, cgroup, netfilter) goes on the hardware list (§6b),
    never on the CLEAN list.
  - *Compatibility direction.* "No backward compatibility" covers breaks that
    fail loud. A cut whose old reader (rollback, mixed-version cell) fails
    open, granting access or skipping a check, is a security regression.
- **No unevidenced CLEAN.** "Accurate", "matches" or "no finding" names the
  file:line read. A subagent's claim is unverified until the adjudicator
  re-reads the source; the 2026-11-08 pass also shipped four wrong claims
  (an IPv4 variant, a path net/http already refuses, a load-bearing pre-filter
  called redundant, a field the API never serves).
- **Regression tests fail on the old code.** Run each new regression test
  against the pre-fix source (a saved copy, restored by full path) and paste
  the failure; a test that passes on both proves nothing.
- **Re-read the final diff.** A fix that changed shape after review (altitude,
  codex, hardware) gets a fresh style reader and flow reader on the final diff
  before the PR goes final. cocoon 2026-11-05: the last flow read caught a
  shutdown budget cut from 40 s to 10 s that every earlier reader had passed.

## 3. /simplify lenses

Run the four lenses (reuse, simplification, efficiency, altitude). Every finding
is either applied or skipped WITH A RECORDED REASON — one line per finding in
the round report: `file:line — lens — finding — applied <sha> | skipped: <reason>`.
Silent drops are not allowed. Take altitude findings seriously: when a state
exists only because our own code creates it, deleting the cause beats adding a
convergence path for it (cocoon 2026-11-05: removing one `vm.shutdown` call
replaced a start-time recovery plus a hibernate special case).

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
- CI is evidence only for the steps it reached. A job that died before lint
  (a tool download 504) proves nothing about lint or tests: run the gates
  locally and say CI did not reach them.
- `asl` gate mode is `-forwarder=false`; the advisory `forwarder` findings are
  listed separately and settled once, not counted as gate failures.

## 6b. Hardware evidence — when the round changes behaviour

Book hosts through the coordinator session; run in parallel where the hosts
have room, never serialize work that can overlap. A case that closes no listed
unknown is not run. A subagent's message to the coordinator goes out under the
parent session's address and the reply lands in the parent: the parent relays
it to the subagent by name, in full.

- **Shared hosts:** never claim against another stack's sandboxd. Run a private
  kit (own binary, cocoon root, slice, ports, net_scope, data_dir) under the
  coordinator's grant; a bridged kit sweeps foreign `sandbox_egress_*` tables,
  so list them read-only first. Send START before the first leg and the raw
  TEARDOWN after the last.
- **Acceptance per fix:** the fixed arm, the old arm as negative control, and
  for a guest-side fix the stock image as a control on the new host binary.

- **Crash windows:** deterministic injection (a kill-point build or a
  `ch_binary` wrapper that kills its parent) with the old version as the
  negative control. Random kills that hit nothing prove nothing (cocoon:
  0 hits in 180). Killing the CLI does not reach a window that needs the child
  dead: kill the child.
- **A/B:** arms interleaved in one run in shuffled order, labels swapped once,
  a control pair of the same code on separate inodes, each binary copied onto
  the measured NUMA node and verified with `numa_maps` from a pinned process.
  Page-cache placement faked 0.5 ms; a cycled permutation schedule faked an
  11 ms position effect. Sample neighbour affinity before calling a node quiet.
- **Budgets and defaults:** a change to a timeout, budget or default is
  measured at the boundary scale, not on small samples. cocoon 2026-11-05:
  144 small-guest stops passed and stops failed from 117 GiB RSS.
- **Teardown:** every kit removes its VMs, network, cgroups and dirs, checks
  for leftovers including poll loops it left on the host, logs it to the
  host's ops audit, and reports it to the coordinator.

## 7. Convergence = 闭环

A round with zero newly applied findings and all gates green closes the loop —
then STOP. When `ledger.py diff` lists no file for any recorded lens, nor for
`infra`, no flow is due, and `asl` is clean on both GOOS, the round is closed
before any reader is spawned: report that state as the outcome. Closed means
no known high-severity defect, not no bugs: never claim "no unknowns", and
never present 0 due as proof that every file was read. When the bar is already met, "no change needed" backed by the audit
evidence is the correct outcome; manufacturing churn to look productive is a
judgment failure ("if not necessary, don't change/push/release").

## 8. Report

Lead with the verdict. State what ran (agents, lenses, gates with real output), findings applied,
findings skipped with reasons, the round's net prod/test LOC (with per-item
justification for any prod increase), and the convergence verdict. A declared step that
did not actually run is a violation with no exceptions — report skipped as
skipped.

- **Coverage table** (CMGS: "bugs I can accept, unknowns I cannot"): one row
  per flow or window, marked injected on hardware, probed, read only, or not
  covered with the reason. Name what only CMGS can unlock (production canary,
  hardware the testbed lacks) as not covered rather than asking again.
- **PR:** the final PR carries the real numbers, the coverage table and the
  known issues kept without code; English, sanitized, no attribution. Work
  autonomously to an open final PR; merge when it meets the bar, otherwise say
  why it is left to CMGS.
- **Split by cost:** a change that triggers an image rebuild or publish goes
  in its own PR with that cost stated. An optimization lead that is not
  measured end to end becomes an issue, not part of the round.
- Close with the decisions only CMGS can make, each with a recommendation.
