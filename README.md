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
- **functypedup** — a func type spelled out more than once (struct fields,
  return types) must be a named type. Signatures shorter than 40 characters
  and `_test.go` files are skipped.
- **topdecl** — `const`/`var` declarations live in single blocks at the top of
  the file, never below the first func. Compile-time interface checks
  (`var _ Iface = ...`) are exempt.

Deliberately not covered (prose rules with sanctioned exceptions that make
mechanical checking a false-positive machine): exported-above-unexported
ordering and type-block atomicity — those stay in the human walkthrough.

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
