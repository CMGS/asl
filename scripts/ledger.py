#!/usr/bin/env python3
"""Hygiene ledger: which files a review round still has to read.

ledger.py diff <repo> [ledger.tsv] --lens style|judge   files whose blob changed since the lens last cleared them
ledger.py mark <repo> [ledger.tsv] --lens style|judge --verdict CLEAN|KEPT [--note TEXT] <path>...
ledger.py prune <repo> [ledger.tsv]                     drop rows for files the repo no longer tracks

The ledger defaults to ~/Documents/workspace/cocoonstack/.hygiene/<repo>.tsv (the workspace root, which is
not a repository; cocoon-specs' directory set is closed and every other repo is public).

Rows are `path<TAB>style_blob<TAB>judge_blob<TAB>verdict<TAB>note`; blobs come from `git ls-files -s`,
so an unchanged file keeps its verdict across commits. The header records the rules commit
(the asl checkout HEAD); a rules change invalidates every row.
"""

import argparse
import os
import subprocess
import sys

LENSES = ("style", "judge")


def ledger_path(repo, explicit):
    if explicit:
        return explicit
    url = subprocess.run(["git", "-C", repo, "remote", "get-url", "origin"], capture_output=True, text=True)
    name = url.stdout.strip().rstrip("/").rsplit("/", 1)[-1].removesuffix(".git") if url.returncode == 0 else ""
    name = name or os.path.basename(os.path.abspath(repo))
    return os.path.expanduser(f"~/Documents/workspace/cocoonstack/.hygiene/{name}.tsv")


def tracked(repo):
    out = subprocess.run(["git", "-C", repo, "ls-files", "-s", "*.go"], capture_output=True, text=True, check=True).stdout
    blobs = {}
    for line in out.splitlines():
        meta, path = line.split("\t", 1)
        blobs[path] = meta.split()[1]
    return blobs


def rules_commit():
    asl = os.path.expanduser("~/Documents/workspace/asl")
    return subprocess.run(["git", "-C", asl, "rev-parse", "--short", "HEAD"], capture_output=True, text=True, check=True).stdout.strip()


def load(path):
    rows, rules = {}, ""
    if not os.path.exists(path):
        return rows, rules
    for line in open(path):
        line = line.rstrip("\n")
        if line.startswith("# ledger"):
            rules = line.split("rules=")[-1].strip()
            continue
        if not line or line.startswith("#"):
            continue
        cols = (line.split("\t") + ["", "", "", ""])[:5]
        rows[cols[0]] = cols[1:]
    return rows, rules


def save(path, rows, rules):
    os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
    with open(path, "w") as f:
        f.write(f"# ledger v1 rules={rules}\n")
        for p in sorted(rows):
            f.write("\t".join([p] + rows[p]) + "\n")


def cmd_diff(args):
    rows, rules = load(ledger_path(args.repo, args.ledger))
    blobs = tracked(args.repo)
    col = LENSES.index(args.lens)
    if rules != rules_commit():
        print(f"# rules changed ({rules or 'none'} -> {rules_commit()}): every file is due", file=sys.stderr)
        rows = {}
    due = [p for p, b in blobs.items() if (p not in rows or rows[p][col] != b) and not (args.lens == "judge" and p.endswith("_test.go"))]
    for p in sorted(due):
        print(p)
    scope = len(blobs) if args.lens == "style" else sum(not p.endswith("_test.go") for p in blobs)
    print(f"# {len(due)} of {scope} files due for the {args.lens} lens", file=sys.stderr)


def cmd_mark(args):
    rows, rules = load(ledger_path(args.repo, args.ledger))
    current = rules_commit()
    if rules and rules != current:
        rows = {}
    blobs = tracked(args.repo)
    col = LENSES.index(args.lens)
    for p in args.paths:
        if p not in blobs:
            sys.exit(f"{p}: not tracked in {args.repo}")
        row = rows.get(p, ["", "", "", ""])
        row[col] = blobs[p]
        row[2] = args.verdict
        if args.note:
            row[3] = args.note
        rows[p] = row
    save(ledger_path(args.repo, args.ledger), rows, current)
    print(f"# marked {len(args.paths)} files {args.verdict} for the {args.lens} lens", file=sys.stderr)


def cmd_prune(args):
    rows, rules = load(ledger_path(args.repo, args.ledger))
    blobs = tracked(args.repo)
    gone = [p for p in rows if p not in blobs]
    for p in gone:
        del rows[p]
    save(ledger_path(args.repo, args.ledger), rows, rules)
    print(f"# pruned {len(gone)} rows", file=sys.stderr)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)
    d = sub.add_parser("diff")
    d.add_argument("repo")
    d.add_argument("ledger", nargs="?")
    d.add_argument("--lens", choices=LENSES, default="style")
    d.set_defaults(fn=cmd_diff)
    m = sub.add_parser("mark")
    m.add_argument("repo")
    m.add_argument("ledger", nargs="?")
    m.add_argument("--lens", choices=LENSES, required=True)
    m.add_argument("--verdict", choices=("CLEAN", "KEPT"), required=True)
    m.add_argument("--note", default="")
    m.add_argument("paths", nargs="+")
    m.set_defaults(fn=cmd_mark)
    p = sub.add_parser("prune")
    p.add_argument("repo")
    p.add_argument("ledger", nargs="?")
    p.set_defaults(fn=cmd_prune)
    args = ap.parse_args()
    args.fn(args)


if __name__ == "__main__":
    main()
