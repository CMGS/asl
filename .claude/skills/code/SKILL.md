---
name: code
description: Cocoon code style standards for Go projects — declaration layout (single top-of-file const/var block), test-file layout (tests before helpers), signature naming (named constraints and func types), hard comment budget, mandatory modern Go (1.25/1.26 idioms), import grouping, logging (core/log), context handling, error handling, naming, build/CI config. Apply when reviewing or writing Go code in cocoonstack projects.
---

# Code Review & Refactor Standard

## Style Self-Check (mandatory)

After writing or editing any Go file, re-scan the final file top-to-bottom against this list and fix every violation before presenting or committing. Do this walkthrough yourself — never delegate it to subagents. The rules apply to every file you touch — if an edited file already violates them, fix it in the same change.

1. **Declaration layout** — order: imports → `const` block → `var` block → types & funcs. At most one top-level `const` block and one data `var` block per file, both at the top; never below the first `func`. A single declaration uses the bare form (gofumpt strips single-entry parens). Compile-time interface checks are the exception: they stand alone immediately above the implementing type. Type blocks are atomic: a type is immediately followed by its complete method set — nothing in between (see File Organization). In `_test.go` files, every test func comes before all helpers (see Test File Organization).
2. **Comments** — hard budget (see Comment Style): default is zero. The complete allowance is one-line godoc on exported identifiers and one-line WHY-comments carrying information the code cannot. Everything else is a violation — restated code, section labels, edit narration, multi-line justification narrative. Apply the deletion test to every comment in the file, not only ones you wrote.
3. **Logging** — level matches severity; `Error`/`Errorf`/`Fatalf` take `err` structurally, never embedded via `%v`; no `f` variant without a `%` verb (sole allowance: `Fatalf`, since no non-f `Fatal` exists).
4. **Imports** — three groups: stdlib / external / internal.
5. **Context** — no `context.TODO()`; propagate the caller's `ctx`.
6. **English only** — no Chinese in code, comments, strings, or docs.
7. **Modern Go** — hand-rolling what the Go 1.25/1.26 toolchain already provides is a violation: loops replaceable by `slices`/`maps` helpers (`ContainsFunc`, `IndexFunc`, `SortFunc`, `Collect`), `min`/`max`/`clear` builtins, `for range n`, `cmp.Or`, iterators, `wg.Go`. See Go Modern Features for the catalog.
8. **Signatures** — type-parameter constraints are named interface types, never inline `[T interface{ ... }]`; a func type spelled out more than once gets a named type; a signature that takes 3+ lines to read is a naming failure, not a formatting problem (see Signatures).

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

## Pre-Commit Checklist

Every round of changes must complete these steps before committing:

1. **Style Self-Check** — run the checklist at the top of this document on every changed file
2. **Senior Tech Review** — review all changed code for correctness, edge cases, and consistency
3. **Run `/simplify`** — launch three parallel review agents (reuse, quality, efficiency), fix all actionable findings
4. **`make lint`** — must pass on both `GOOS=darwin` and `GOOS=linux` (dual-platform lint, zero issues)
5. **`asl ./...`** — structural layout gate (test-helper placement, inline constraints, duplicated func types, const/var placement, unexported-method-above-exported partition), zero findings on both GOOS; source lives in `~/Documents/workspace/asl`, install via `go install`

## Cocoon Code Style Reference

All projects under cocoonstack follow these concrete patterns derived from the cocoon codebase.

### Import Grouping (3 groups)

```go
import (
    "context"      // 1. stdlib
    "fmt"

    "github.com/spf13/cobra"           // 2. external (blank line separator)
    "github.com/projecteru2/core/log"

    "github.com/cocoonstack/cocoon/config"  // 3. internal (blank line separator)
    "github.com/cocoonstack/cocoon/types"
)
```

Enforced by goimports with `local-prefixes: github.com/cocoonstack/cocoon`.

### Constants

- At most one top-level `const` block per file, placed at the top (after imports) — never scattered, never below the first `func`
- Use blank lines within the block to separate logical groups
- Type constants: `type VMState string; const ( VMStateRunning VMState = "running" ... )`
- Identity constants: `const typ = "cloud-hypervisor"`

### Global Vars

Same rule as constants: package-level `var` declarations (sentinel
errors, lookup tables, magic byte prefixes, static data) live in a
single `var ( ... )` block at the top of the file, right after the
`const` block (or after imports if there is none). Never scatter
multiple top-level `var` statements. Blank lines separate logical
groups, e.g. sentinel errors → static data.

```go
var (
    ErrNotFound = errors.New("VM not found")

    qcow2Magic = []byte{'Q', 'F', 'I', 0xfb}

    // nonImageSignatures lists magic byte prefixes that qemu-img would
    // otherwise silently misclassify as raw.
    nonImageSignatures = []struct {
        prefix []byte
        desc   string
    }{
        {[]byte("<!"), "content looks like HTML/XML, not a disk image"},
        // ...
    }
)
```

Note the comment density above: `qcow2Magic` needs no comment (the name
and literal carry it); `nonImageSignatures` gets one because the WHY is
invisible in the code.

**Exception — compile-time interface checks.** `var _ Iface = (*Impl)(nil)`
assertions never join the data var block: they stand alone immediately
above the implementing type declaration — a bare statement for one
assertion, a dedicated `var ( ... )` block for several:

```go
var _ images.Images = (*OCI)(nil)

type OCI struct {
    // ...
}
```

**gofumpt exception for single-entry declarations.** gofumpt (enforced
by `make lint`) strips the parentheses off a `var`/`const` block that
contains exactly one declaration, so a lone top-level declaration MUST
use the bare form — "one block per file" means *at most* one. This
applies equally to a lone sentinel error or interface check:

```go
var _ Interface = (*Impl)(nil)       // single compile-time check — bare
const defaultTimeout = 5 * time.Second  // single constant — bare
```

Function-local `var` declarations are exempt: they live where they're used.

### Sentinel Errors

Sentinel errors are global vars and live in the same top-of-file `var`
block, as their own group (see Global Vars). Exported sentinels are
`ErrXxx`, unexported ones `errXxx`. Check with `errors.Is()`, not
string comparison.

### File Organization

- Exported (public) declarations always above unexported (private) ones
- Type blocks are atomic: a type declaration is immediately followed by its
  complete method set — nothing sits between a type and its methods, not
  another type, not another type's methods:
  ```go
  type Foo struct { ... }
  func NewFoo() *Foo { ... }
  func (f *Foo) PublicMethod() { ... }   // public methods first
  func (f *Foo) helper() { ... }         // private methods after
  ```
- Producer methods trail the type they produce: a method on A returning B
  (`Sandbox.OpenPty → *Pty`) goes below B's entire block, even when the
  producer is exported and B's last method is not. A file leads with its
  primary type + methods; producers and auxiliary leaf types follow.
- Method-less vocabulary types (enums, option/error/payload structs) cluster
  ahead of the first method-bearing type they serve (`Cmd`/`ExitError` before
  `Sandbox`). Exception: a result type consumed by exactly one method may sit
  inside the owner's block, directly above that method (`SizeSpec` above
  `Size.Spec`).
- Wire pairs: a request type + its methods first, then its response type
  (`ClaimRequest` + `Key`/`TTL` before `ClaimResponse`)
- Utility/standalone functions: public on top, private below, grouped by functionality (not interleaved)
- Place utility functions below struct methods in the same file
- Precedence: grouping beats visibility — keep a type's methods together and utilities below them; "public above private" applies within each group and yields to type-block atomicity across groups, never across the whole file
- Compile-time interface checks: standalone immediately above the implementing type — bare for one assertion, own `var ( ... )` block for several

### Test File Organization

- In a `_test.go` file the order is: top-of-file `const`/`var` blocks, then
  ALL `Test*`/`Benchmark*`/`Fuzz*`/`Example*` funcs, then everything else —
  helper funcs, fixture/stub types AND their methods. Nothing but
  `const`/`var` declarations may precede the first test func, and no helper
  may sit between two test funcs.
- Helper-only `_test.go` files (no test funcs) follow the normal File
  Organization rules.

### File Naming

- One responsibility per file, named by operation: `create.go`, `start.go`, `stop.go`, `clone.go`
- Shared helpers: `utils.go` — the single canonical name; do not introduce `helper.go`/`helpers.go`, and rename them to `utils.go` when touching those files
- Command definitions: `commands.go`, handler impl: `handler.go`
- Config: `config.go`, DB/index: `db.go`, GC: `gc.go`

### Package Structure

- Interface in parent package, implementation in subpackage:
  - `hypervisor/hypervisor.go` → `Hypervisor` interface
  - `hypervisor/cloudhypervisor/cloudhypervisor.go` → `CloudHypervisor` struct
  - `network/network.go` → `Network` interface
  - `network/cni/cni.go` → `CNI` struct
- Use generics for reusable storage/resolution patterns:
  ```go
  type Store[T any] interface { ... }
  func ResolveRef[T any](items map[string]*T, names map[string]string, ref string, notFound error) (string, error)
  ```

### Struct Field Ordering

Group by concern, in this order: identification → config → runtime →
resources → state → timestamps.

- Small structs: contiguous fields, no separators
- Large structs: blank line between groups; a group-header comment only
  when it carries information beyond the field names, never a bare tag
  like `// runtime`
- Trailing per-field comments only for non-obvious semantics or units
- JSON tags snake_case; `omitempty` where the zero value means absent;
  `json:"-"` for runtime-only fields; embed shared field sets

```go
// VM is the runtime record for a VM, persisted by the hypervisor backend.
type VM struct {
    ID     string   `json:"id"`
    State  VMState  `json:"state"`
    Config VMConfig `json:"config"`

    // Runtime — populated only while State == VMStateRunning.
    PID        int    `json:"pid"`
    SocketPath string `json:"socket_path,omitempty"` // CH API Unix socket

    NetSetup

    CreatedAt time.Time  `json:"created_at"`
    StartedAt *time.Time `json:"started_at,omitempty"`
}
```

### Logging

- Use `github.com/projecteru2/core/log` as the standard logging library across all projects
- Import as `"github.com/projecteru2/core/log"` — never use a redundant alias like `log "github.com/projecteru2/core/log"`
- All entry points (main, cmd root) must initialize the library via `log.SetupLog`:
  ```go
  ctx := context.Background()
  logLevel := pickFirstNonEmpty(os.Getenv("APP_LOG_LEVEL"), "info")
  if err := log.SetupLog(ctx, &types.ServerLogConfig{Level: logLevel}, ""); err != nil {
      log.WithFunc("main").Fatalf(ctx, err, "setup log")
  }
  ```
- Logger access: `logger := log.WithFunc("pkg.Func")` — the string mirrors the enclosing function's real name and case (`images.Import`, `core.buildRecorder`); chain `.WithField(k, v)` for scoped context; always name the local `logger`
- **Log levels must match severity** — never use `Info`/`Infof` to log errors or warnings:
  - Failures → `Error` / `Errorf`
  - Non-fatal issues (bad client input, graceful degradation) → `Warn` / `Warnf`
  - Normal operation → `Info` / `Infof`
  - Diagnostics → `Debug` / `Debugf`
- **`Error`/`Errorf` take a dedicated `err` parameter** — pass the error structurally, never embed it in format strings:
  - `logger.Error(ctx, err, "ws upgrade")` — not `logger.Infof(ctx, "ws upgrade: %v", err)`
  - `logger.Errorf(ctx, err, "connect to %s", host)` — not `logger.Infof(ctx, "connect to %s: %v", host, err)`
- **`Error`/`Errorf` no-op when `err == nil`** — `logger.Error(ctx, err)` right before `return err` is the standard log-and-return idiom, safe on the nil path; the bare form without a message is idiomatic when there is no context to add
- **Never use `f` variants without format interpolation** — if there's no `%` verb, use the plain method:
  - `logger.Info(ctx, "server started")` — not `logger.Infof(ctx, "server started")`
  - `logger.Error(ctx, err, "ws upgrade")` — not `logger.Errorf(ctx, err, "ws upgrade")`
  - `logger.Warn(ctx, "retry exhausted")` — not `logger.Warnf(ctx, "retry exhausted")`
- **`Fatalf` is the only Fatal variant** — the library has no `Fatal` (without `f`):
  - the logger renders the structural `err` itself (zerolog `.Err(err)`), so never embed it again via `%v` — `logger.Fatalf(ctx, err, "setup log")`, not `logger.Fatalf(ctx, err, "setup log: %v", err)`
  - a plain no-verb message is fine here — the only such allowance, since no non-f `Fatal` exists
- Use `f` variants only when interpolation is required:
  - `logger.Infof(ctx, "listening on %s", addr)`
  - `logger.Errorf(ctx, err, "connect to %s failed", host)`
  - `logger.Warnf(ctx, "drop invalid payload from %s: %v", ip, err)` (`Warn`/`Warnf` have no `err` parameter)

### Naming Conventions

- **Functions**: verb-noun prefix — `buildVMConfig()`, `prepareCloudimg()`, `extractBlobIDs()`
- **Check functions**: `IsProcessAlive()`, `VerifyProcess()`, `isCidataDisk()`
- **Receivers**: short, type-derived, identical across all methods of a type — single letter (`b *Backend`, `o *OCI`) or two-letter abbreviation on collision (`ch *CloudHypervisor`, `fc *Firecracker`, `lf *LocalFile`); never `self`/`this`
- **Variables**: short in tight loops (`id`, `ctx`, `err`, `rec`, `f`), full names elsewhere (`vmCfg`, `netProvider`, `bootCfg`)
- **Logging**: `logger := log.WithFunc("pkg.Func")` then `logger.Infof(ctx, "msg", args...)`

### Signatures

- Type-parameter constraints are always named interface types — never inline
  `[T interface{ ... }]`, single-method constraints included:
  ```go
  type typed interface{ Type() string }

  func resolveOwner[T typed](...)   // not: resolveOwner[T interface{ Type() string }](...)
  ```
- A func type spelled out more than once (struct field + return type, var +
  field, ...) gets a named type: `type hypervisorCtor func(context.Context,
  *config.Config) (hypervisor.Hypervisor, error)`.
- A signature that takes 3+ lines to read is a naming failure, not a
  formatting problem — extract named types until it fits.

### Error Handling

- Wrap pattern: `fmt.Errorf("verb noun: %w", err)` — lowercase, no trailing punctuation
- Validation: `fmt.Errorf("--cpu must be at least 1, got %d", cfg.CPU)`
- Config: `fmt.Errorf("root_dir must not be empty")`
- All error messages lowercase, action-oriented
- Aggregate close errors in defers: `defer func() { err = errors.Join(err, f.Close()) }()`

### Comment Style

Hard budget, enforced as Style Self-Check item 2. Default is **zero
comments**: carry intent through naming and structure. The complete
allowance:

- **Godoc on exported identifiers** — one line, starting with the identifier name; more only when the API contract genuinely cannot fit one line. Interface-implementing methods omit godoc — the interface documents the contract.
- **One-line WHY** — a hidden constraint, race condition, external-system quirk, or the reason a non-obvious approach was required. States the constraint itself; never a justification narrative.
- **`//nolint:linter`** on specific lines only, with reason if non-obvious.

Everything else is a violation:

- unexported declarations whose name carries the meaning get no comment at all — comment only magic values, non-obvious units, encoded invariants
- restating the code (`// close the file` above `f.Close()`)
- narrating steps or labeling sections (`// step 2: validate input`, `// --- helpers ---`)
- describing the edit instead of the code (`// now also handles IPv6`, `// changed to use errors.Is`)
- multi-line explanation — if the WHY needs a paragraph, restructure the code instead

**Deletion test**: if removing a comment costs a competent Go reader nothing, remove it — apply this to every comment in code you touch, not only code you wrote.

### Context Handling

- **Never use `context.TODO()`** — always use `context.Background()` or propagate from callers
- **Create `ctx` once, reuse everywhere** — don't scatter `context.Background()` calls:
  ```go
  func main() {
      ctx := context.Background()
      // reuse ctx for all setup calls...
      log.SetupLog(ctx, ...)
      initK8s(ctx)
      // then derive signal context from the same ctx
      ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
      defer cancel()
  }
  ```
- If a function needs context, accept it as first parameter
- **Never create a new `context.Background()` unless truly necessary** — only allowed when a goroutine must outlive the parent ctx (e.g., graceful shutdown that must complete after the signal ctx is cancelled). In all other cases, propagate the existing ctx

### Runtime Discipline

- No `init()` functions — wire dependencies explicitly in constructors and `main` (gochecknoinits is enforced)
- No `panic()` in non-test code — return errors
- Bounded concurrent fan-out via `errgroup.WithContext`; dedup identical in-flight work via `singleflight`

### Go Modern Features

Mandatory, enforced as Style Self-Check item 7. Write against the
toolchain in go.mod (Go 1.25/1.26); hand-rolling what the language or
stdlib already provides is a violation of the same severity as a layout
break. The `modernize` linter catches part of this — the self-check
covers the rest.

Replace on sight:

| Hand-rolled | Modern |
|---|---|
| loop + `if` + `return true` over a slice | `slices.ContainsFunc` / `IndexFunc` |
| `sort.Slice` | `slices.SortFunc` + `cmp.Compare` |
| loop collecting map keys/values | `slices.Collect(maps.Keys(m))` / `slices.Sorted(...)` |
| `if a > b { m = a }` | `min` / `max` builtins |
| `for i := 0; i < n; i++` (counter only) | `for range n` |
| first-non-zero fallback chains | `cmp.Or(x, y, ...)` |
| map-reset loop or realloc | `clear(m)` |
| `interface{}` | `any` |
| near-identical funcs per type | generics (`Store[T]`, `ResolveRef[T]`) |
| callback-style traversal APIs | `iter.Seq` / `iter.Seq2` range-over-func |
| `wg.Add(1)` + `go func` + `defer wg.Done()` | `wg.Go(func() { ... })` |

Example — existence check over config slices:

```go
// hand-rolled: 12 lines
func (c *Config) hasEgressPolicy() bool {
    for _, p := range c.Pools {
        if p.Egress != nil {
            return true
        }
    }
    for _, tn := range c.Tenants {
        if tn.Egress != nil {
            return true
        }
    }
    return false
}

// modern: 4 lines, same short-circuit semantics
func (c *Config) hasEgressPolicy() bool {
    return slices.ContainsFunc(c.Pools, func(p PoolSpec) bool { return p.Egress != nil }) ||
        slices.ContainsFunc(c.Tenants, func(t TenantSpec) bool { return t.Egress != nil })
}
```

Limits: modern means **less code, not different code**. Don't rewrite a
clear loop into a contorted iterator chain, don't force generics where
one concrete type exists, and don't adopt GOEXPERIMENT-gated packages.
When the modern form and the loop tie on length and clarity, either
passes.

### Prefer Native Go Over Subprocess

- Always prefer open-source Go libraries over shelling out to external commands
- Example: use a Go qcow2 library instead of `exec.Command("qemu-img", ...)`; use a Go FAT library instead of `mcopy`
- If a subprocess is truly unavoidable (no viable Go library, or the external tool is the authoritative implementation), document why in a comment and log a debug message so the user knows an external binary is required
- Treat every `exec.Command` / `exec.CommandContext` as tech debt — revisit when a Go-native alternative becomes available

### Code Reuse

- Extract helpers shared across packages into a `utils/` package; helpers private to one package stay in that package's `utils.go`
- Extract generic patterns into reusable generic functions
- If multiple projects share the same utils/generic code, consolidate into one project (do not create a new repo) and have others import via `go mod`

### Test Patterns

- Table-driven tests with `for _, tt := range tests { t.Run(...) }`
- Assertion: `if cond { t.Errorf("got %v, want %v", got, want) }`
- Fatal on setup: `if err != nil { t.Fatalf("setup: %v", err) }`
- Use `t.TempDir()` for temp directories and `t.Context()` instead of `context.Background()` in tests
- Benchmarks iterate with `b.Loop()`, not `for i := 0; i < b.N; i++`
- Timing-dependent concurrency tests use `testing/synctest` instead of real sleeps

## Build & CI Configuration

All projects must have consistent tooling matching cocoon:

### Makefile

Targets: `all`, `build`, `test`, `vet`, `lint`, `fmt`, `fmt-check`, `deps`, `coverage`, `clean`, `help`
- Each target has `## help comment`
- Tool versions as variables: `GOLANGCILINT_VERSION ?= v2.9.0`
- `vet` and `lint` run with GOOS matrix (linux + darwin)
- `fmt` uses gofumpt + goimports

### GitHub Actions

- **`lint.yml`** — golangci-lint on all branches/PRs + gofumpt/goimports check
- **`build.yml`** — build on master push, paths-ignore os-image/
- **`test.yml`** — vet + `go test -race -timeout 120s -count=1 -cover`
- **`goreleaser.yml`** — release on tag push
- All use `go-version-file: 'go.mod'`

### goreleaser

- ldflags: `-X pkg/version.REVISION={{.Commit}} -X pkg/version.VERSION={{.Env.VERSION}} -X pkg/version.BUILTAT={{.Date}}`
- Two builds: debug (amd64, symbols) + release (amd64+arm64, stripped with `-s`)
- Archive naming: `{binary}_{version}_{OS}_{ARCH}`

### golangci-lint

- 20+ linters enabled: asciicheck, bodyclose, errcheck, gocyclo, gosec, govet, misspell, modernize, revive, staticcheck, unparam, unused, etc.
- gocyclo min-complexity: 30
- nestif min-complexity: 5
- goconst min-len: 3, min-occurrences: 3
- Exclude `_test.go` from: errcheck, gosec, unparam, goconst, gocyclo
- Formatters: gofumpt + goimports with local-prefixes

## Documentation

- **README** — rewrite/refactor for each project:
  - Correct project name, structure, features, design
  - Must reflect actual current functionality
  - Clear, concise, easy to understand
  - All English, no Chinese text anywhere in the codebase
- Remove any Chinese comments or strings from code

## Releases

- Projects without a release: create a `v0.1.1` release to verify goreleaser action works
- Goreleaser action should produce a GitHub Release (not just a tag) with binary artifacts and auto-generated release notes
