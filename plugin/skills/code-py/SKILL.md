---
name: code-py
description: Python code review & refactor standard for cocoonstack projects — the Python counterpart to /code (Go) and /code-rs (Rust). Load for any Python code work in cocoonstack repos; subagents read files in full, the session model adjudicates and runs the final Style Self-Check on every touched file.
---

# Python Code Review & Refactor Standard

The Python counterpart to `/code` (Go) and `/code-rs` (Rust). Same
discipline, Python idioms. Derived from the cocoonstack Python code
(`sandbox/sdk/python`) and CMGS's Python habits (typed signatures, named
loggers, argparse CLIs, module docstrings).

## Style Self-Check (mandatory)

After writing or editing any Python file, re-scan the final file
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
unverified text until adjudicated. No linter checks module layout or class
atomicity (ruff `I` covers import order only) — those rules are
walkthrough-only.

1. **Module layout** — order: module docstring → `from __future__ import
   annotations` (when any modern annotation syntax is used and the floor is
   <3.10) → imports (three groups: stdlib / external / local, blank-line
   separated, alphabetical within groups — ruff `I` enforces) → module
   constants → classes (each class with its complete method set) → free
   functions → `if __name__ == "__main__":` guard last. Public API above
   private helpers; class blocks are atomic — nothing sits between a class
   and its methods.
2. **Comments and docstrings** — hard budget: default zero. The complete
   allowance is a one-line docstring on the public surface (module, public
   classes, public methods whose behavior isn't obvious from the signature —
   stating contracts such as units, atomicity, blocking, never a parameter
   list repeating the signature) and one-line WHY-comments carrying
   information the code cannot, both in the STE 101 register (Simplified
   Technical English: short, present tense, one fact per sentence,
   lowercase-first inline; PR inline replies use the same register).
   Everything else is a violation — restated code, section labels, edit
   narration, the same fact at docstring and call site, any inline comment
   under `tests/` (the test name and assertion message carry it). A fix,
   sweep, or cut commit adds zero comments (rationale goes in the commit
   message) and never raises a file's comment count. Apply the deletion
   test to every comment in the file, then once more to the batch's total
   density.
3. **Errors** — a small typed exception hierarchy rooted in one base class
   per package; raise with context, never `raise Exception(...)`. No bare
   `except:`; catch the narrowest type that is actually handled.
   Intentional suppression is `contextlib.suppress(X)`, never
   `try/except/pass`. Error messages lowercase, action-oriented, no
   trailing punctuation.
4. **Typing** — public signatures fully annotated (params + return).
   Library code stays compatible with the floor declared in
   `requires-python` (use `from __future__ import annotations` rather than
   raising the floor for syntax). Internal helpers may omit annotations
   where they'd restate the obvious.
5. **English only** — no Chinese in code, comments, docstrings, or
   identifiers, regardless of what older personal scripts did.
6. **Resources** — anything holding a socket, file, or process is a
   context manager (`__enter__`/`__exit__`) and safe to close twice;
   `with` at call sites. Never leak on the error path of a constructor —
   wrap post-acquisition setup in try/except that releases.

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

1. **Style Self-Check** — walk every changed file, `tests/` included,
   against the checklist yourself — the final pass after the subagent
   ledgers, never a substitute for them. For a fix, sweep, or cut commit
   count comment lines in the staged diff: `git diff --cached -U0 -- '*.py'
   | grep -cE '^\+[[:space:]]*#[^!]'` must not exceed the `^-` count (the
   commit hook checks the same)
2. **Senior Tech Review** — correctness, edge cases, resource lifetimes,
   exception-path behavior; no contrived scenarios — if retry already
   converges, the failure is loud, and the residue is bounded, no mechanism
   is added
3. **Run `/simplify`** — the four lenses (reuse, simplification,
   efficiency, altitude) over the whole diff, `tests/` included; fix all
   actionable findings. The order is fixed: personal walkthrough →
   `/simplify` → gates
4. **`ruff format --check && ruff check && python -m pytest`** — all clean,
   from the package root, with the package importable the way CI imports it
   (`pip install -e .`; bare `pytest` does not put the cwd on `sys.path`).
   Typing (item 4 above) has no gate until the package declares
   `[tool.mypy]` with `strict = true` — the next batch that touches a
   Python package adds it and runs `mypy` here; until then the rule is
   walkthrough-only and the report says so

## Cocoonstack Python Style Reference

Concrete patterns derived from `sandbox/sdk/python`.

### Dependencies

- **Library code is stdlib-only** unless the project explicitly decides
  otherwise — an SDK must be `pip install`-able with zero transitive risk.
  Dev tooling (ruff, pytest) is exempt; it never ships.
- When a dependency is genuinely warranted, pin a floor (`>=`), never an
  exact version, and justify it in the PR body.

### Package Layout

- One responsibility per module, named by role, mirroring the Go SDK's
  file split where a sibling exists: `client.py`, `sandbox.py`, `conn.py`,
  `frames.py`, `errors.py`, `template.py`, `checkpoint.py`.
- `__init__.py` re-exports the public surface and defines `__all__`
  (sorted); consumers import from the package, not submodules.
- `errors.py` holds the exception hierarchy; every other module imports
  from it, never defines ad-hoc exceptions.
- Avoid circular imports structurally (low-level modules import nothing
  from high-level ones); a function-level import is a last resort and gets
  a comment naming the cycle it breaks.
- Tests in `tests/` (no `__init__.py`), hardware/e2e drivers as top-level
  scripts with a usage line in the module docstring.

### Classes

- A class is declared with its complete method set: public methods first
  (dunder lifecycle — `__init__`, `__enter__`/`__exit__`, `__iter__` —
  counts as public and leads), private `_helpers` after. Module-level free
  functions below the classes they support.
- Prefer plain classes with explicit `__init__` assignment for handles
  that hold live resources or a client reference; `@dataclass` for pure
  data records.
- No `@staticmethod` indirection where a free function reads better; no
  `@property` for anything that does I/O.
- Constructor takes what the object needs, not a grab-bag: handles receive
  `(client, ...identity fields)`; parse wire records in one place, not at
  every call site.

### Wire / Protocol Code

- Mirror the wire contract exactly — field names come from the shared
  fixture corpus (`protocol/fixtures/`), never from memory. When a
  protocol is implemented in multiple languages, the Python side
  round-trips the same golden corpus in its tests; a frame only one side
  can round-trip is a drift bug.
- Encode/decode live in one module (`frames.py`); size caps enforced
  before allocation; binary payloads' base64 handling is invisible outside
  that module.
- Auth semantics (which token rides which header) are part of the wire
  contract — copy them from the server handler, not from intuition.

### Naming

- Modules, functions, variables `snake_case`; classes `CamelCase`;
  constants `SCREAMING_SNAKE_CASE` at module top.
- Functions verb-noun (`encode_request`, `dial_agent`); predicates
  `is_*`/`has_*`. Short names in tight scopes (`sb`, `conn`, `rec`), full
  names in signatures.
- Private helpers and internal-contract methods prefixed `_`; a
  cross-module private helper stays in the module that owns the concept.

### Logging & CLIs

- Library code raises; it does not log or print. Only entry points
  (scripts, daemons) configure logging: `logging.basicConfig` once, named
  loggers (`logging.getLogger("rproxy")`), level matches severity.
- CLIs use `argparse` with defaults visible in the parser definition;
  scripts exit via `sys.exit(main())` and keep `main()` testable.
- e2e/step drivers print progress lines and never mask a step failure
  with a cleanup failure (`contextlib.suppress` around cleanup, re-raise
  the original).

### Tests

- pytest, plain `assert` with a message payload where the value helps
  (`assert x == y, x`). `@pytest.mark.parametrize` over copy-paste;
  fixtures for servers/temp state; `http.server`-based fakes for HTTP
  peers, in-process, bound to port 0.
- Test names state the behavior: `test_claim_follows_redirect_with_no_redirect`.
- Never assert on wall-clock or real network; bound retries by deadline.

### Tooling (pyproject)

```toml
[tool.ruff]
line-length = 120
target-version = "py39"        # match requires-python

[tool.ruff.lint]
select = ["E", "F", "W", "I", "UP", "B", "SIM"]
```

- `ruff format --check` + `ruff check` are the lint gate (CI and local);
  autofix with `--fix` is fine but re-read what it changed.
- CI: `pip install -e <pkg> && ruff format --check . && ruff check . && python -m pytest -q`.
