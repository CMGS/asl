---
name: loc-justify
description: Whole-repo production-code (non-test) volume audit — measure prod LOC composition and growth with exact commands, attribute it to packages/commits/buckets, then force an item-by-item justify verdict (EARNED / CHALLENGED with a ranked cut-list) via multi-agent lenses and adversarial verification. Report-only — a cut-list is applied only on the user's explicit instruction; bugs the scan surfaces are fixed in the round. Invoke when the codebase feels bloated, after a heavy batch, or on a cadence.
---

# LOC Justify — production code mass audit

Goal: every block of logic code must earn its lines. Output is a justify
ledger + ranked cut-list backed by evidence. This skill never applies cuts on
its own: the user states explicitly whether a cut-list gets applied (§6). A
real bug the scan surfaces is different — it is fixed in the round, in its own
commit, after re-reading the source proves the failure reachable.

Reading discipline: opus/sonnet lens agents read every file in scope in full
via `Read` — never grep/sed sampling, which loses the context a verdict needs;
the session model plans the batches, picks models (sonnet for mechanical
lenses, opus for judgment-heavy ones), adjudicates every finding against the
source itself, and does the final review. Before scanning, read the previous
round's memory for this repo and carry its standing rejections forward (§4).

## 1. Measure — exact commands, never estimate

All counting from `git ls-files` / `git archive` into the scratchpad; never
checkout old revs into the worktree.

- Totals (Go shown; same split for .rs/.py with their test conventions):
  - total: `git ls-files '*.go' | xargs wc -l | tail -1`
  - prod:  `git ls-files '*.go' | grep -v '_test\.go$' | xargs wc -l | tail -1`
  - test:  `git ls-files '*_test.go' | xargs wc -l | tail -1`
  - prod comments: `... | xargs grep -h '^\s*//' | wc -l`; blanks: `grep -hc '^$'`
- Rust: prod/test split inside a file at the column-0 `#[cfg(test)]` line
  (everything from it to EOF is test), plus `tests/*.rs` and `benches/` as
  test; comments `^\s*//`. Python: `tests/`, `test_*.py`, `conftest.py` are
  test; comments `^\s*#`. Use the same five numbers.
- Test scaffolding that lacks the test suffix — `contracttest/`, `test/*bench`,
  `e2e/`, `bench/`, `*/capture` harnesses imported only by tests — is counted
  as prod by the commands above; list those dirs with their line counts and
  report effective prod both with and without them (cocoon 2026-07: 496
  lines of `meta/contracttest` inflated the meta package).
- Effective prod code = prod − blanks − comments. Report all five numbers.
- Comment share sanity: if comments are <10% of prod, say explicitly that
  comments are NOT the story (cocoon 2026-07: 5%).
- Per-package table + delta vs a base rev (user-provided, else ~1 month back
  via `git rev-list -1 --before=<date> master`): `git archive <rev> | tar -C
  <scratchpad>/snap -x`, count the same way, diff the tables.
- Top line-contributing commits since base: `git log --numstat` aggregated
  per commit, prod/test split, ranked by net adds.

## 2. Attribute

Bucket the growth since base: feature / fix+hardening mechanism /
review-round additions / tests (exempt) / comments — plus an initial-import
bucket for young repos (gateway 2026-08: 71% of prod was the day-one import).
Name the top commits per bucket with their net prod adds. Watch the
review-round bucket specifically — cocoon 2026-07 data: review passes were 34%
of monthly prod growth; /review-round §5 now caps them at net-zero prod by
default.

The standing answer to "why does LOC grow while we keep cleaning" is these
two numbers side by side: feature commits' net prod vs review/simplify
commits' net prod. Review rounds net-negative while features drive the total
is healthy (cocoon 2026-07-24: features +8855, every review commit negative);
review rounds net-positive is the finding.

## 3. Justify scan — multi-agent lenses, prod code only

Mechanical scans on haiku/sonnet, one lens per agent:

- **dead/vestigial**: exported funcs with no callers, seams nothing injects,
  flags nothing sets, error branches no caller can trigger (cross-check
  golangci `unused` + grep call sites). The repo is not the universe of
  callers: before any dead verdict on a `pkg/`/`api/`/exported symbol, grep
  the whole `~/Documents/workspace/cocoonstack` workspace (`go.work` siblings
  — vk-sandbox, vk-cocoon, sandbox, gateway, cocoon-macos consuming cocoon
  types), and after touching an exported type run `go build ./...` in each
  consumer. Test-only ≠ dead (constructors and fixtures the tests drive stay).
  Write-only observability trails (decision logs, weight fields) are a
  judgment row, not an automatic cut. Every grep a lens agent cites is re-run
  by the adjudicator before it enters the ledger — agents have fabricated
  symbols.
- **duplication**: structurally-same logic across files/packages. Per CMGS
  rule, dedup bias applies — a cut that adds a shared helper still counts as
  a cut if net-negative.
- **over-abstraction**: generics/interfaces/option patterns with exactly one
  concrete user; layers that only forward.
- **hand-rolled stdlib**: loops/plumbing the Go 1.25/1.26 toolchain already
  provides (see /code Modern Features table); 50+ line funcs reducible to a
  utils/stdlib call.
- **unreachable defense**: guards whose failure mode cannot occur given the
  actual callers — the agent must enumerate call sites, not pattern-match.
- **over-design / contrived-case protection** (CMGS 2026-07-17): mechanisms
  (recovery paths, extra lock layers, disambiguation reads, heal loops) whose
  justifying scenario was manufactured, not observed. Mechanical part
  (sonnet): inventory every mechanism with its provenance — `git log -S`,
  linked issue/PR, incident evidence, and whether a test encodes a REAL
  trigger sequence. Judgment part (top model only): per mechanism ask
  (a) incident-backed or reviewer-hypothesized? (b) would 幂等收敛三问 already
  hold WITHOUT it (retry/GC converges, failure loud, residue bounded)? if yes
  it is redundant protection; (c) probability × damage: near-unreachable
  trigger + benign failure + heavy mechanism = over-protection; (d) the 构陷
  test: does the scenario require a sequence no real deployment produces
  (adversary inside the process boundary, impossible interleaving)? Review-
  round provenance + no incident + no realistic test ⇒ CHALLENGED by default.

Every candidate then goes to adversarial verify on the top model: prove the
cut is behavior-preserving and the lines are truly unearned, against real
call sites and tests. 幂等收敛三问 applies in reverse: if the "defensive" code
is what makes retry/GC converge, it is EARNED — record why and drop the finding.

## 4. Ledger + verdict

Per package: prod lines, delta since base, verdict —
- **EARNED**: one line on what the mass buys (mechanism, backend parity, recovery path).
- **CHALLENGED**: cut-list entries — `file:symbol`, estimated −lines, why
  behavior-preserving, verification evidence.

Mechanisms from the over-design lens get their own ledger rows: mechanism,
origin commit, incident/hypothesis provenance, verdict.

Cut-list ranked by net lines saved. "All EARNED" backed by the scan is a valid
outcome — never manufacture cuts to look productive, and never cut tests or
load-bearing WHY comments to move the number.

**Standing rejections.** Every round ends with an adjudicated-rejected list
(candidate, reason) written to the repo's memory, and every round begins by
reading the previous one: a rejected candidate is not re-flagged unless the
code or the callers changed. Typical entries: single-consumer helpers kept for
sibling symmetry, guards at an exported-API boundary, deliberate scaffolding
for a rollout phase.

## 5. Report discipline

Same bar as /review-round §8: what ran (agents, lenses), real command outputs,
skipped-as-skipped. Section 1's numbers appear verbatim; "feels big" without
numbers is a violation. The report closes with the prod/test net delta of the
range audited and of the round's own commits — the user must never have to
discover a `+459/−61` diff on their own.

Agent ops: a lens agent that drifts into "waiting for the other agents" has
left its job — one SendMessage ("YOU are the scanner, report now") recovers
it; an agent that returns a report without findings for files it did not
list has not read them.

## 6. Apply — only on the user's explicit instruction

Cuts land only after the user says so, as their own commit(s), separate from
any bug-fix commit:

- Behavior preservation is proven by the existing tests, never by adapting a
  test to the cut — a shared fixture that breaks under `slices.DeleteFunc`'s
  in-place tail zeroing is the contract speaking (cocoon 2026-08-05 rollback).
- Comment count is monotonically non-increasing; a cut never adds a comment.
- `git status` before `git add -A`; no build artifacts in the index.
- A cut on a hot path gets an A/B bench with both arms interleaved in one run
  and the order swapped — arms measured in separate runs lie.
- Gates as in /code or /code-rs, both platforms, before push.
