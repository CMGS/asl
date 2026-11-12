---
name: code-rs
description: Cocoon code style standards for Rust projects — module/declaration layout, use-grouping (std/external/crate), comment minimalism, error handling (Result + context, no unwrap in non-test), ownership/borrow discipline (no gratuitous clones), unsafe isolation, async/tokio discipline, naming, cargo/CI config. Apply when reviewing or writing Rust code in cocoonstack projects (gateway, sandbox-init, silkd, and peers).
---

# Rust Code Review & Refactor Standard

The Rust counterpart to `/code` (Go). Same discipline, Rust idioms. Derived
from the cocoonstack Rust crates (`gateway` workspace, `sandbox/boot/init`,
`sandbox/silkd`).

## Style Self-Check (mandatory)

After writing or editing any Rust file, re-scan the final file top-to-bottom
against this list and fix every violation before presenting or committing.
The rules apply to every file you touch — if an edited file already violates
them, fix it in the same change. Pure moves and file splits are not exempt.

Division of labor on any pass wider than one diff: opus/sonnet subagents read
the files — each file in full via `Read`, never grep/sed sampling, which loses
context — against an explicitly enumerated rule list and return a per-file
verdict ledger with line numbers; the session model (Fable) plans, batches,
picks models (sonnet for mechanical scans, opus for judgment-heavy reads),
adjudicates every finding by re-reading the cited source, and walks the final
diff of every touched file through this checklist itself. Subagent output is
unverified text until adjudicated. No linter checks Rust layout (type+impl
atomicity, use grouping, public-above-private) — those rules are
walkthrough-only.

1. **Declaration layout** — order within a file: `//!` module doc → `use`
   groups → `const`/`static` → type + its `impl` blocks (glued) → free fns →
   `#[cfg(test)] mod tests`. Public items above private. A type's `impl` sits
   directly below the type with nothing in between — not another type, not
   another type's `impl`; do not cluster all types first and all impls after.
2. **use grouping** — three blank-line-separated groups: std → external
   crates → this crate (`crate::` / `super::`). Enforced by rustfmt
   (`group_imports`/`imports_granularity` where configured); keep it even
   where rustfmt is lenient.
3. **Comments** — hard budget (see Comment Style): default is zero. The
   complete allowance is one-line `///` docs on public items, one-line
   WHY-comments carrying information the code cannot, and lint-mandated
   `// SAFETY:` lines — all in the STE 101 register. Everything else is a
   violation: restated code, section labels, edit narration, multi-line
   narrative, the same fact at doc and call site, any inline comment inside
   `#[cfg(test)]`/`tests/`. A fix, sweep, or cut commit adds zero comments
   and never raises a file's comment count; the deletion test runs per
   comment and once more on the batch's total density.
4. **Error handling** — functions that can fail return `Result`; propagate
   with `?`; no `unwrap()`/`expect()`/`panic!()` in non-test code except on a
   documented invariant: a per-site `#[allow(clippy::unwrap_used)]` right above
   the line, plus the comment saying why it cannot fire (the ban itself is
   lint-enforced — see Lints under Cargo.toml). `Error`-level logs carry the
   error value, never a pre-formatted string.
5. **unsafe** — isolated in one low-level module (`sys.rs`/`sysutil.rs`), each
   `unsafe` block justified by a `// SAFETY:` comment stating the invariant it
   relies on — that exact prefix (std policy; what
   `clippy::undocumented_unsafe_blocks` recognizes).
6. **English only** — no Chinese in code, comments, strings, or docs.
7. **Ownership** — borrow by default; never clone, `to_string`, `to_vec`, or
   `unwrap` your way past the borrow checker — keep the lifetime flowing.
   Every surviving `.clone()` must justify itself (persistence-row
   construction, crossing a `'static` task boundary, shared write-rare data
   behind `Arc`); watch the hidden copies (`json!` interpolating owned values,
   `to_string().into_bytes()`, clone-before-`unwrap`/`filter`) — see
   Ownership & Borrow Discipline below.
8. **Hot path** — state the per-request cost of the change (target zero);
   a claimed perf win or loss is real only if it survives an A/B bench with
   both arms interleaved in one run and the order swapped.

## Commit Rules

- Never add `Co-Authored-By` or any other trailers to commit messages
- No Claude / AI / Anthropic references anywhere in commits, PR bodies, or issues
- Keep commit messages concise: one-line summary, optional body for why
- Layout-only normalization (reordering declarations to match this standard) rides in its own `review:` commit, never mixed with behavior changes — the message names the rule being applied

## Git Workflow

- Always rebase, never merge — keep linear history with no merge commits
- `git pull --rebase` before pushing
- `git rebase` to integrate upstream changes, not `git merge`
- If conflicts arise during rebase, resolve per-commit, not squash-and-pray
- Never rewrite already-pushed history (rebase/amend/force-push) without explicit per-conversation authorization — rebase freely only on commits that haven't left the machine

## Batch-End Checklist

Runs once at the end of a branch/batch before the final push (the /review-round cadence); mid-batch commits are gated by steps 4–6 only:

1. **Style Self-Check** — run the checklist at the top of this document on every changed file
2. **Senior Tech Review** — review all changed code for correctness, edge cases, ownership/lifetime soundness, cancellation safety, and hot-path cost; no contrived scenarios — if retry already converges, the failure is loud, and the residue is bounded, no mechanism is added
3. **Run `/simplify`** — the four lenses (reuse, simplification, efficiency, altitude) over the whole diff, tests included; fix all actionable findings. The order is fixed: personal walkthrough → `/simplify` → gates
4. **`cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test`** — all three clean. On a workspace, run at the workspace root. If the crate targets Linux only (vsock, etc.), run the Linux-gated tests in a Linux container — a green macOS run that `#[cfg]`-skips them is not evidence. On Apple Silicon the container is `--platform linux/arm64`, never amd64 emulation (`Command::spawn` of a missing binary returns `Ok` under emulation and fakes test failures); reproduce a failure on HEAD before blaming the diff.
5. **Env-gated suites** (`GW_TEST_PG_URL`, `GW_TEST_REDIS_URL`, ...) silently early-return when unset, so a green run without them proves nothing — report them as skipped, or run them and cite server-side artifacts (the created database, the index the migration built) as proof they executed.
6. **Windows-only crates** cannot build on the mac host: `cargo test --target x86_64-pc-windows-gnu --no-run` proves the suite compiles; pure modules are host-tested through a throwaway `#[path]` harness that pulls the real sources in, so judgment logic lives in pure modules, not next to the platform bindings.

## Cocoon Rust Style Reference

Concrete patterns derived from `boot/init` and `silkd`.

### use Grouping (3 groups)

```rust
use std::collections::HashMap;          // 1. std
use std::io;

use serde::{Deserialize, Serialize};    // 2. external crates (blank line)
use tokio::io::{AsyncWrite, AsyncWriteExt};

use crate::proc::Table;                 // 3. this crate (blank line)
use crate::proto::Response;
```

### Constants and Statics

- Group `const`/`static` near the top of the file, after `use`, before the
  first type. Blank lines separate logical groups.
- `SCREAMING_SNAKE_CASE`. A unit or bound belongs in the name or a terse
  comment: `const MAX_FRAME: usize = 8 * 1024 * 1024;` needs none;
  `const LOG_RING_BYTES: usize` is self-documenting.
- Module-private magic values get a one-line WHY only when the number's
  origin is invisible (a protocol cap mirrored from a peer, a measured
  timeout).

### File / Module Organization

- One responsibility per file, named by role: `proto.rs`, `exec.rs`,
  `proc.rs`, `server.rs`, `vsock.rs`, `sysutil.rs`. Shared low-level helpers
  live in `sysutil.rs`/`sys.rs` — the canonical name; do not introduce
  `helpers.rs`.
- `lib.rs` exposes the module surface (`pub mod ...`) so integration tests
  and peers can drive the crate in-process; `main.rs` is a thin wrapper over
  the real entry point.
- Public (`pub`) items above private within a group; a type's `impl` block
  directly below the type, public methods before private, free helpers below
  the impls. Grouping beats visibility — keep a type's methods together.
- Type + `impl` blocks are atomic, mirroring the Go rule: nothing sits between
  a type and its `impl` — declare each type immediately followed by its impls,
  not all types up front with the impls trailing. A free fn or another type's
  method that produces the type follows the produced type's whole block. A
  method-less payload type consumed by exactly one method may sit directly
  above that method inside the owner's block.
- Existing files that predate this rule (e.g. `session.rs`, `proc.rs` cluster
  types ahead of impls) get glued when touched, as a separate layout-only
  `review:` commit.
- A `///` doc or attribute block is part of its item: never land a helper
  between a fn's doc block and the fn, and nothing (no `const`, no helper
  struct) sits between a type and its `impl`.

### Types and Serde

- Derive minimally and intentionally: `#[derive(Debug, Deserialize,
  Serialize)]` on wire types; `Clone`/`Copy` only where cheap and needed.
- Wire enums use an explicit tag and rename convention:
  `#[serde(tag = "op", rename_all = "snake_case")]` — the JSON contract is
  visible at the type, not scattered.
- `#[serde(default)]` on optional inbound fields (forward compat: unknown
  fields are ignored by default, which is the compatibility story);
  `skip_serializing_if = "Option::is_none"` where a null would be noise.
- Field order groups by concern (identification → config → runtime → state),
  like the Go standard.

### Error Handling

- Fallible functions return `Result<T, E>`. For small crates `Result<T,
  String>` (as in `boot/init`) or `io::Result<T>` is fine — a deliberate
  size-budget deviation from the ecosystem's app→`anyhow` / lib→`thiserror`
  rule, not a preference; reach for a `thiserror` enum only when callers must
  match on variants.
- Wrap with lowercase, action-oriented context, no trailing punctuation:
  `format!("mount {src} -> {target}: {e}")`, `io::Error::other(...)`, or a
  typed error frame for protocol code.
- Propagate with `?`. Aggregate cleanup errors rather than dropping them.
- **No `unwrap()`/`expect()`/`panic!()` in non-test code** except on a
  documented invariant — and then `expect("static string contains no NUL")`
  states the invariant under a per-site `#[allow(clippy::unwrap_used)]` /
  `#[allow(clippy::expect_used)]`, so every exception is grep-able.
  `crypto::rand` fill, poisoned-mutex `.lock().unwrap()` on a never-panicking
  critical section, and constant `CString::new` are the usual justified cases;
  comment the non-obvious ones. Tests are exempt via `clippy.toml`, not
  attributes.

### Ownership & Borrow Discipline

Clone-一把梭 (cloning your way past the borrow checker) is banned. Gratuitous
clones are a hot-path perf tax and a signal the ownership story wasn't thought
through. Derived from the gateway /simplify rounds, which found a whole class
of them.

- **Borrow-first signatures**: return `&str`/`&T`/`Option<&T>` and take
  references unless the caller genuinely stores the value
  (`fn account_name(&self) -> &str`, `fn model_provider(..) -> Option<&str>`,
  not owned `String`s the caller immediately re-borrows).
- **Consume owned inputs**: a request body you own is moved into the internal
  struct — `into_iter()`, field moves, `std::mem::take` — never `.iter()` +
  per-field `.clone()`. Compute borrowing accessors (`m.text()`) before moving
  the field they read.
- **Accumulate in place**: `push_str`/`extend` on the accumulator; rebuilding a
  string per fragment is quadratic. Decode borrowed (`from_utf8_lossy(&buf[..n])`
  then `drain`), don't copy out first.
- **References across trait seams for big payloads**: `ledger_add(&BillingRecord)`
  pattern — the one impl that stores clones internally; the others bind fields
  by reference.
- **Share, don't copy**: `Arc<T>` for write-rare shared data cloned per request
  (`AkInfo.model_quotas`); `Cow<'static, str>` for constant wire fields; move a
  value out of an owned chunk (`c.tool_calls.take()`) instead of cloning it.
- **Justified clones only**: constructing a persistence row, crossing a
  `'static`/spawn boundary, or a measured-and-accepted cost (the gateway's
  CallEngine `GatewayRequest` clone is tracked pending bench evidence — an
  explicit decision, not a habit). In review, every `.clone()` is guilty until
  proven necessary.
- Avoid borrow-checker workarounds that reintroduce copies: to log after a
  mutation, format the message from the borrow first, then mutate, then log —
  don't clone the field to sidestep NLL.
- **Hidden copies**: `json!` interpolates every expression through
  `to_value(&expr)`, so an owned `Value`/`Vec<Value>`/`String` placed in the
  macro is deep-copied and the original dropped — build the envelope with
  `json!`, then move owned values in (`body["k"] = Value::Array(v)`,
  `body["k"] = s.into()`). `Value::to_string().into_bytes()` →
  `serde_json::to_vec(&v)`. Big blobs (base64 media, prompts) leave the owned
  body through `take_*` helpers, never `.clone()`.
- **Shape the Option chain**: `as_ref().filter(..).cloned()` — never clone
  before `unwrap`/`filter`; `Option::is_none_or` over `map_or(true, ..)`.

### Comment Style

Default is **no comment**. Carry intent through naming and structure.

- **`///` doc** on public types, fns, and trait methods, starting with the
  identifier's role; **`//!`** for module-level docs explaining the file's
  responsibility.
- Unexported items: comment only when the name cannot carry the meaning
  (protocol invariants, ring-buffer trimming rules, race notes).
- **Inline** comments state only what code cannot: cancellation safety, why a
  reborrow/`fill_buf` loop instead of the obvious combinator, external-system
  quirks (VMM handshake, libc buffer aliasing).
- **Register — STE 101** (Simplified Technical English): one short sentence
  per comment, present tense, one fact per sentence, no hedging or narrative;
  inline comments start lowercase, `///` docs keep the identifier's case. PR
  inline replies use the same register.
- Never restate code, narrate steps (`// step 2`), label sections
  (`// --- helpers ---`), or describe the edit (`// now also handles X`), or
  state the same fact at both the doc and the call site — pick the call site.
- A fix, sweep, or cut commit adds zero comments — the rationale goes in the
  commit message; code that had no comment gets none, and the file's comment
  count never rises in such a commit.
- Tests carry zero inline comments: the test name and assertion messages are
  the documentation.
- `// SAFETY:` lines are lint-mandated and exempt from every tightening pass.
- **Deletion test**: if removing a comment costs a competent Rust reader
  nothing, remove it — apply to every comment in code you touch, then once
  more to the batch's total density.

### unsafe Discipline

- Concentrate `unsafe` in one module (`sys.rs`/`sysutil.rs`); the rest of the
  crate stays safe.
- Every `unsafe` block gets a `// SAFETY:` comment stating the invariant that
  makes it sound (pointer validity window, that a libc static buffer is read
  before the next call can clobber it, that a raw fd is owned). The prefix is
  load-bearing: it is the std-library policy spelling and the one
  `clippy::undocumented_unsafe_blocks` enforces.
- Edition 2024's `unsafe_op_in_unsafe_fn` applies: inside an `unsafe fn`, each
  unsafe operation still sits in its own justified `unsafe` block.
- Prefer a safe wrapper fn exposing a `Result` over sprinkling `unsafe` at
  call sites.

### Async / tokio Discipline

- No blocking calls on the async runtime; use tokio's async fs/process/io.
- Spawned tasks either are awaited/joined or are deliberate detached workers
  with a comment saying so; prefer `JoinSet` over collecting `JoinHandle`s in
  a `Vec`; use `kill_on_drop` intentionally. A daemon that needs coordinated
  shutdown uses tokio_util's `CancellationToken` + `TaskTracker`, not ad-hoc
  flags (a fatal-exit daemon like silkd legitimately needs neither).
- Bound fan-out (semaphore/`mpsc` capacity); never an unbounded channel on a
  hostile-input path.
- Cancellation safety: a select arm or dropped future must not leave shared
  state half-updated — flip state under the same lock as the removal.
- Enforce frame/size caps before allocating unbounded (`MAX_FRAME`), so a
  malformed peer cannot OOM the daemon.

### Naming Conventions

- Types/traits `CamelCase`; fns/vars/modules `snake_case`; consts
  `SCREAMING_SNAKE_CASE`.
- Functions verb-noun: `read_frame`, `apply_user`, `build_cmdline`,
  `synth_pid`. Predicates `is_*`/`has_*`.
- Receivers are `self`/`&self`/`&mut self` (Rust has no receiver-naming
  choice — never emulate Go's short receiver by aliasing).
- Short names in tight scopes (`r`, `w`, `n`, `buf`, `e`), full names in
  signatures and fields (`vsock_socket`, `started_at_epoch_secs`).

### Prefer std / crates Over Subprocess

- Prefer std or a vetted crate over shelling out. If a subprocess is genuinely
  the authoritative implementation (git binary, the cocoon CLI as a stable
  contract), document why in a comment.
- Treat every `Command::new` of an external tool as a deliberate contract
  choice, not a shortcut.

### Test Patterns

- `#[cfg(test)] mod tests` at the bottom of the file for unit tests;
  `tests/*.rs` for integration tests that drive the public surface.
- `#[tokio::test]` for async; drive I/O over `tokio::io::duplex` in-process
  instead of standing up real sockets where possible.
- Assertions with a message on setup failures; prefer explicit
  `assert_eq!(got, want)`.
- Use `tempfile`/`TempDir` for filesystem tests; never touch a shared path.
- **Shared protocol fixtures**: when a wire protocol is implemented in two
  languages (silkd Rust + SDK Go), both sides parse/emit a shared golden
  corpus (`protocol/fixtures/`) in their tests; a frame only one side can
  round-trip is a drift bug.
- OS-specific expectations (`/root` vs `/var/root`) must not be asserted
  exactly when the test runs cross-platform — assert the invariant, not the
  platform value.

## Build & CI Configuration

### Cargo.toml

- `edition = "2024"` for new crates; migrate older ones with `cargo fix
  --edition` when touched. The edition also buys `unsafe_op_in_unsafe_fn`
  by default and the MSRV-aware resolver v3.
- Declare `rust-version` (MSRV); set `publish = false` for internal crates; a
  terse `description`.
- Lint levels live in the `[lints]` table — one source of truth for local
  builds, editors, and CI (the CI `-D warnings` flag becomes a backstop, not
  the config):
  ```toml
  [lints.clippy]
  unwrap_used = "deny"
  expect_used = "deny"
  undocumented_unsafe_blocks = "deny"
  ```
  with the test exemption in `clippy.toml`:
  ```toml
  allow-unwrap-in-tests = true
  allow-expect-in-tests = true
  ```
  Documented-invariant exceptions are per-site `#[allow(...)]` (see Error
  Handling), never file- or crate-wide.
- Binary/release profile is size- and startup-conscious where the binary
  ships into a guest image:
  ```toml
  [profile.release]
  opt-level = 3
  lto = true
  codegen-units = 1
  strip = true
  # panic = "abort" for a PID-1/initramfs binary; a long-lived daemon that
  # must survive a panicking task keeps unwinding.
  ```
- Dependencies minimal and feature-scoped (`tokio = { features = [...] }`,
  not `full`); platform-only deps under `[target.'cfg(...)'.dependencies]`.

### Toolchain

- `rust-toolchain.toml` pins an exact version (`channel = "1.98.0"` today,
  the same across gateway, silkd and boot/init) with
  `components = ["rustfmt", "clippy"]` — never `stable`; every gate and every
  shipped binary compile with the same rustc. Bump it deliberately, in its own
  commit, with fmt/clippy/test green on the new version.
- Every Dockerfile builder stage `COPY rust-toolchain.toml ./` before any
  `rustup`/`cargo` call (before `rustup target add`, or the target installs
  against the image's default toolchain). Without it the pin governs the gates
  while the image's rustc builds the artifact — silkd shipped a silently
  swapped compiler that way (2026-08-24); the toolchain, not the base image,
  determines the bytes.
- A digest bump of a `rust:*` builder image is a compiler bump in disguise:
  read the rustc version it carries before merging.
- Release and image builds run `cargo build --locked` so the artifact matches
  the committed `Cargo.lock`.

### GitHub Actions

- A Rust job runs, on the crate/workspace: `rustup component add rustfmt
  clippy`, then `cargo fmt --check`, `cargo clippy --all-targets -- -D
  warnings`, `cargo test`. Mirror the existing `build-boot.yml` shape.
- `cargo deny check` (RustSec advisories + license/source policy) rides in the
  same job once a `deny.toml` exists; adopt it before the dependency tree
  grows past hand-audit size.
- Path-filter to the crate's directory so unrelated changes don't rebuild.
- A guest-shipped binary builds from source in the image build stage (pinning
  the version to the commit), not a release download, and builds with
  `--locked` so the shipped binary matches the committed Cargo.lock.

## Documentation

- **README** per crate: role, module map, protocol/contract, build. English
  only, reflecting actual current functionality.
- Design docs for anything non-trivial live in `cocoon-specs/design/`, not in
  the code repo.
