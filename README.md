# asl

Structural layout checker for cocoonstack Go projects — the rules of the
code style standard that `go vet` and golangci-lint cannot express, as a
`multichecker` binary.

## Analyzers

- **testorder** — in `_test.go` files, all `Test*`/`Benchmark*`/`Fuzz*`/`Example*`
  funcs come first; every helper (func, fixture type, its methods) sits below
  the last test func. Helper-only test files are exempt.
- **constraintname** — type-parameter constraints must be named interface
  types, never inline `[T interface{ ... }]`.
- **functypedup** — a func type spelling the same contract more than once
  (struct fields, parameters, package vars, func results) must be a named
  type. Duplicates that include a func result (a factory) always count; the
  rest count only when the names match across sites — coincidental shape
  twins with different meanings stay inline. Signatures shorter than 40
  characters and `_test.go` files are skipped.
- **topdecl** — `const`/`var` declarations live in single blocks at the top of
  the file, never below the first func. Compile-time interface checks
  (`var _ Iface = ...`) are exempt.
- **methodpartition** — within one file, a receiver's unexported methods never
  appear above one of its exported methods; the exception-free slice of the
  exported-above-unexported rule (cocoon 2026-07-28: a whole-repo read found
  23 orderings the human walkthrough had missed, 12 of them this shape).
- **funcpartition** — the standalone-function slice of the same rule: an
  unexported standalone function never appears above an exported one in the
  same file. Constructors/producers (returning a type declared in the file,
  or an interface it implements) are exempt — they belong to their type's
  block — as are `main` and `init`
  (cocoon 2026-08-05: a human review caught `armQuota` parked between two
  exported functions; a whole-repo scan showed it was the only instance, and
  the constructor exemption cleanly passes the one lookalike).
- **methodinterleave** — a standalone func or type never sits between two
  methods of the same receiver: utilities trail the method set. The one
  sanctioned adjacency — a type directly above the same-receiver method that
  names it in its signature (`SizeSpec` above `Size.Spec`) — is exempt, and a
  grouped `type ( ... )` declaration is one block: reported once, exempt when
  the method names any member
  (vk-cocoon 2026-08-09: a human review caught `appendMacosVNCArg` parked
  inside the Provider method set; a whole-repo AST sweep found 8 instances
  across two files, zero false positives against the sanctioned shapes).

- **typeblockgap** — the other half of type-block atomicity: nothing sits between
  a type declaration and that type's first method either. Two occupants stay
  exempt — a producer of the type (returning it, or an interface it
  implements) and a result type directly above the owner method
  naming it — and a grouped `type ( ... )` declaration is one block whose members
  never split each other. One finding per split block, reported at the type
  (sandbox 2026-08-12: `resolvedVolume` parked between `catalogVolume` and its
  only method slipped past both the human walkthrough and an all-clean
  `asl ./...`; a 16-repo corpus sweep found 18 instances, 7 of them in repos the
  shipped analyzers call clean, and one false positive — an interface-returning
  constructor — that set the producer exemption).

Deliberately not covered (prose rules with sanctioned exceptions that make
mechanical checking a false-positive machine): standalone-function placement
relative to method sets beyond the same-receiver interleave slice (that slice
IS covered by methodinterleave, the exported/unexported ordering slice by
funcpartition), vocabulary-type clustering, and producer-method trailing —
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
```

## Skills

`.claude/skills/` vendors the agent skills this checker mechanizes a slice
of, for reuse in other projects — copy a directory into your own skills
location (project `.claude/skills/` or user-level `~/.claude/skills/`):

- **code** — the full Go style standard: declaration layout, comment budget,
  logging, modern-Go idioms, naming, error handling, build/CI config. asl
  enforces its mechanically checkable layout rules; the rest stays in the
  review walkthrough.
- **loc-justify** — whole-repo production-LOC audit: measure and attribute
  code mass with exact commands, then force an item-by-item EARNED/CHALLENGED
  verdict with a ranked cut-list.
