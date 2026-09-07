# asl

Structural layout checker for cocoonstack Go projects — the rules of the
code style standard that `go vet` and golangci-lint cannot express, as a
`multichecker` binary.

## Analyzers

- **testorder** — in `_test.go` files, all `Test*`/`Benchmark*`/`Fuzz*`/`Example*`
  funcs come first; every helper (func, fixture type, its methods) sits below
  the last test func. Helper-only test files are exempt.
- **constraintname** — type-parameter constraints must be named types, never
  inline `[T interface{ ... }]` or a bare union `[T ~int | ~string]`; `any`,
  `comparable`, and instantiated constraints (`Store[int]`) pass.
- **functypedup** — a func type spelling the same contract more than once
  (struct fields, parameters, package vars, func results) must be a named
  type. Duplicates that include a func result (a factory) always count; the
  rest count only when the names match across sites — coincidental shape
  twins with different meanings stay inline. Signatures shorter than 40
  characters and `_test.go` files are skipped.
- **topdecl** — `const`/`var` declarations live in single blocks at the top of
  the file, `const` before `var`, never below the first func. Compile-time
  interface checks (`var _ Iface = ...`) are exempt from the blocks but must
  sit immediately above the type they check when that type is declared in the
  same file (other checks may stack between); an untyped `var _ = expr` is a
  data var. A 19-module sweep found seven misplaced checks, two below their
  type.
- **methodpartition** — within one file, a receiver's unexported methods never
  appear above one of its exported methods; a trailing producer (an exported
  method placed below the block of the type it produces) ends no exported
  set. Otherwise the exception-free slice of the exported-above-unexported rule (cocoon 2026-07-28: a whole-repo read found
  23 orderings the human walkthrough had missed, 12 of them this shape).
- **funcpartition** — the standalone-function slice of the same rule: an
  unexported standalone function never appears above an exported one in the
  same file. Constructors/producers (returning a type declared in the file,
  or an interface it implements) are exempt — they belong to their type's
  block — as are `main`, `init`, and test funcs, whose ordering testorder owns
  (cocoon 2026-08-05: a human review caught `armQuota` parked between two
  exported functions; a whole-repo scan showed it was the only instance, and
  the constructor exemption cleanly passes the one lookalike).
- **methodinterleave** — a standalone func or type never sits between two
  methods of the same receiver: utilities trail the method set. The one
  sanctioned adjacency — a type directly above the same-receiver method that
  names it in its signature (`SizeSpec` above `Size.Spec`) — is exempt, and a
  grouped `type ( ... )` declaration is one block: reported once, exempt when
  the method names any member. Another receiver's block inside the method set
  is the producer-trailing shape (`Size.OpenPty` below `Pty`'s block) only
  while every later method of the receiver produces a type declared after
  the split; the first method that does not is reported
  (vk-cocoon 2026-08-09: a human review caught `appendMacosVNCArg` parked
  inside the Provider method set; a whole-repo AST sweep found 8 instances
  across two files, zero false positives against the sanctioned shapes).

- **typeblockgap** — the other half of type-block atomicity: nothing sits between
  a type declaration and that type's first method, and the type is never
  declared below its own methods. Two occupants stay
  exempt — a producer of the type (returning it, or an interface it
  implements) and a result type directly above the owner method
  naming it — and a grouped `type ( ... )` declaration is one block whose members
  never split each other. One finding per split block, reported at the type
  (sandbox 2026-08-12: `resolvedVolume` parked between `catalogVolume` and its
  only method slipped past both the human walkthrough and an all-clean
  `asl ./...`; a 16-repo corpus sweep found 18 instances, 7 of them in repos the
  shipped analyzers call clean, and one false positive — an interface-returning
  constructor — that set the producer exemption).

- **cmpor** — a zero-value fallback written as `if x != zero { return x }; return y`
  (or `if x == zero { x = y }`) is `cmp.Or(x, y)`. Only comparable operands
  count, so slices and maps never match, and the fallback must be free of
  calls and func literals because cmp.Or evaluates every argument (the first
  corpus sweep reported 40 `if err != nil { return err }; return f()` shapes
  before that rule); `x > 0` matches for unsigned types
  and, with `-cmpor.signed`, for signed numbers too — that flag is for the
  advisory pass of a review round, not the commit gate, because a negative x
  means something different under cmp.Or (vk-cocoon 2026-09-07: a `WaitReady`
  fallback survived four rounds; `modernize` has no analyzer for the shape).
- **forwarder** — an unexported func or method whose body is one `return` or
  one call, referenced exactly once in the package (tests included, and the
  one reference must be a call), spanning four or more lines including its
  doc comment: inlining saves at least three lines. Shorter ones are never
  reported — the cocoonstack threshold says a sub-three-line inline is kept,
  not re-judged every round. Methods matching a package-level interface's
  method are exempt; the test-free variant of a package that has `_test.go`
  files stays silent so a test caller is never miscounted, and the one
  statement spans at most two lines so a long literal or func literal never
  qualifies. Advisory, not a gate: the commit hook runs `-forwarder=false`,
  a review round runs it and settles each finding once — inline it, or record
  it as kept in the repo's hygiene ledger (`scripts/ledger.py`).
- **labelenum** — a comment enumerating label values (`result=ok|failed`) in a
  file that imports the Prometheus client. Enumerations drift silently as
  code adds values (vk-cocoon 2026-09-07: three of twelve were stale); the
  comment names the meaning, a test or constants own the values.

Deliberately not covered (prose rules with sanctioned exceptions that make
mechanical checking a false-positive machine): standalone-function placement
relative to method sets beyond the same-receiver interleave slice (that slice
IS covered by methodinterleave, the exported/unexported ordering slice by
funcpartition), functional grouping of utilities ("grouping beats
visibility" — funcpartition reports an unexported helper above a later
exported one even when each pair is its own group; a 2026-08 cocoon sweep
found one instance), vocabulary-type clustering, `const`/`var` blocks placed
after the enum type they belong to (`type VMState string` then its `const`
block is the standard's own example; 14 instances across the corpus), and
producer-method trailing beyond the resume slice methodinterleave covers —
those stay in the human walkthrough.

## Install

```sh
make install   # go install → $GOPATH/bin/asl
```

Every push to `main` also publishes static Linux binaries (amd64 and arm64)
under the rolling [`latest` release](https://github.com/CMGS/asl/releases/tag/latest):

```sh
curl -fsSL -o asl https://github.com/CMGS/asl/releases/download/latest/asl-linux-amd64
chmod +x asl
```

## Usage

```sh
asl ./...                 # standalone, vet-style output and exit code
GOOS=linux asl ./...      # cross-GOOS pass for platform-gated files
go vet -vettool=$(command -v asl) ./...
asl -testorder=false ./...   # disable an analyzer
asl -cmpor.signed ./...      # advisory: also x > 0 fallbacks on signed numbers
scripts/ledger.py diff <repo> <ledger.tsv> --lens style   # files a review round still has to read
```

## Claude Code plugin

The repo doubles as a plugin marketplace shipping the standards this checker
mechanizes a slice of:

```
/plugin marketplace add CMGS/asl
/plugin install cocoon-code-standards@asl-marketplace
```

The plugin carries:

- **code** skill — the full Go style standard: declaration layout, comment
  budget, logging, modern-Go idioms, naming, error handling, build/CI config.
  asl enforces its mechanically checkable layout rules; the rest stays in the
  review walkthrough.
- **loc-justify** skill — whole-repo production-LOC audit: measure and
  attribute code mass with exact commands, then force an item-by-item
  EARNED/CHALLENGED verdict with a ranked cut-list.
- **commit gate hook** (`plugin/hooks/asl-gate.sh`) — a PreToolUse hook on
  `git commit` that resolves the repo the command targets, blocks Go repos
  with asl findings (both GOOS), and blocks `review:`/`fix:`-class commits
  whose staged diff adds more comment lines than it removes. The hook no-ops
  when `asl` is not on PATH — install it per the Install section, built with
  a Go toolchain matching the gated repos' `go` directive. If your
  `settings.json` already wires the same gate, install the plugin without the
  hook copy or drop the settings entry — otherwise the gate runs twice.

Sources live under `plugin/skills/` and `plugin/hooks/`; copying a skill
directory into `~/.claude/skills/` still works without the plugin.
