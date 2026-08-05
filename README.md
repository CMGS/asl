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
  (struct fields, func results) must be a named type. Duplicates that include
  a func result (a factory) always count; field-only duplicates count only
  when the field names match across structs — coincidental shape twins with
  different meanings stay inline. Signatures shorter than 40 characters and
  `_test.go` files are skipped.
- **topdecl** — `const`/`var` declarations live in single blocks at the top of
  the file, never below the first func. Compile-time interface checks
  (`var _ Iface = ...`) are exempt.
- **methodpartition** — within one file, a receiver's unexported methods never
  appear above one of its exported methods; the exception-free slice of the
  exported-above-unexported rule (cocoon 2026-07-28: a whole-repo read found
  23 orderings the human walkthrough had missed, 12 of them this shape).
- **funcpartition** — the standalone-function slice of the same rule: an
  unexported standalone function never appears above an exported one in the
  same file. Constructors/producers (returning a type declared in the file)
  are exempt — they belong to their type's block — as are `main` and `init`
  (cocoon 2026-08-05: a human review caught `armQuota` parked between two
  exported functions; a whole-repo scan showed it was the only instance, and
  the constructor exemption cleanly passes the one lookalike).

Deliberately not covered (prose rules with sanctioned exceptions that make
mechanical checking a false-positive machine): standalone-function placement
relative to method sets (the exported/unexported ordering slice IS covered by
funcpartition), vocabulary-type clustering, producer-method trailing,
and type-block atomicity — those stay in the human walkthrough.

## Install

```sh
make install   # go install → $GOPATH/bin/asl
```

## Usage

```sh
asl ./...                 # standalone, vet-style output and exit code
GOOS=linux asl ./...      # cross-GOOS pass for platform-gated files
go vet -vettool=$(command -v asl) ./...
asl -testorder=false ./...   # disable an analyzer
```
