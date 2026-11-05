#!/usr/bin/env python3
"""Hygiene ledger: what a review round still has to read, per lens and per flow.

ledger.py diff  <repo> [ledger.tsv] --lens NAME [--flows]   what is due for that lens
ledger.py mark  <repo> [ledger.tsv] --lens NAME --verdict CLEAN|KEPT [--note TEXT] <path|@flow>...
ledger.py lenses <repo> [ledger.tsv]                        which lenses have run, over how much
ledger.py flows <repo> [ledger.tsv]                         the repo's declared flows and their members
ledger.py prune <repo> [ledger.tsv]                         drop rows for files the repo no longer tracks

A lens is any name the round uses (style, judge, loc, comments, billing, ...).
A file carries one blob per lens it was marked for, so a lens that never ran
leaves every file due for it: "not yet read" and "changed since read" are the
same answer, which is the only safe default.

A flow is a named set of paths that only mean something together — admission
through settle, the realtime bridge, one cross-protocol direction. Flows live
in `<ledger dir>/<repo>.flows` as `name<TAB>path [path ...]`, are addressed as
`@name`, and fall due when any member's blob changes, even when every member
is individually clean.

The header pins the ANALYZER sources' hash, not the asl HEAD: editing this
script does not invalidate a verdict, editing a rule does.
"""

import argparse
import hashlib
import os
import subprocess
import sys

ASL = os.path.expanduser("~/Documents/workspace/asl")
SOURCE_EXTS = (".go", ".rs", ".py", ".ts", ".tsx", ".mts", ".mjs", ".js", ".java", ".kt", ".swift")
SOURCE_GLOBS = tuple("*" + e for e in SOURCE_EXTS)
INFRA_GLOBS = ("*.sh", "*Dockerfile", "Makefile", "*.yml", "*.yaml", "*.service", "*.toml", "*.rc", "*platforms", "*.ps1", "*.plist", "*pins")
PROD_LENSES = {"judge", "loc", "perf"}


def ledger_path(repo, explicit):
    if explicit:
        return explicit
    url = subprocess.run(["git", "-C", repo, "remote", "get-url", "origin"], capture_output=True, text=True)
    name = url.stdout.strip().rstrip("/").rsplit("/", 1)[-1].removesuffix(".git") if url.returncode == 0 else ""
    name = name or os.path.basename(os.path.abspath(repo))
    return os.path.expanduser(f"~/Documents/workspace/cocoonstack/.hygiene/{name}.tsv")


def flows_path(ledger):
    return os.path.splitext(ledger)[0] + ".flows"


def is_test(path):
    name = os.path.basename(path)
    return (name.endswith("_test.go") or "/tests/" in path or name.startswith("test_") or name == "conftest.py"
            or ".test." in name or ".spec." in name)


def is_infra(path):
    return not path.endswith(SOURCE_EXTS)


def in_scope(lens, path):
    if lens == "infra" or is_infra(path):
        return lens == "infra" and is_infra(path)
    if lens == "ownership":
        return path.endswith(".rs")
    return not (lens in PROD_LENSES and is_test(path))


def tracked(repo):
    out = subprocess.run(["git", "-C", repo, "ls-files", "-s", *SOURCE_GLOBS, *INFRA_GLOBS],
                         capture_output=True, text=True, check=True).stdout
    blobs = {}
    for line in out.splitlines():
        meta, path = line.split("\t", 1)
        blobs[path] = meta.split()[1]
    return blobs


def rules_key():
    """Hash of the analyzer sources. Tooling (scripts/), tests and testdata are excluded:
    a change there cannot change a verdict, so it must not invalidate one."""
    out = subprocess.run(["git", "-C", ASL, "ls-files", "-s", "*.go"], capture_output=True, text=True)
    if out.returncode != 0:
        return "unknown"
    h = hashlib.sha256()
    for line in sorted(out.stdout.splitlines()):
        meta, path = line.split("\t", 1)
        if path.endswith("_test.go") or "/testdata/" in path or path.startswith("scripts/"):
            continue
        h.update(path.encode() + b" " + meta.split()[1].encode() + b"\n")
    return h.hexdigest()[:12]


def load_flows(ledger):
    flows = {}
    p = flows_path(ledger)
    if not os.path.exists(p):
        return flows
    for line in open(p):
        line = line.rstrip("\n")
        if not line or line.startswith("#"):
            continue
        name, _, members = line.partition("\t")
        flows[name.strip()] = members.split()
    return flows


def flow_blob(repo_blobs, members):
    """One id for the whole set: any member changing, appearing or vanishing changes it."""
    h = hashlib.sha256()
    for m in sorted(members):
        h.update(m.encode() + b" " + repo_blobs.get(m, "missing").encode() + b"\n")
    return h.hexdigest()[:16]


def load(path):
    rows, rules, version = {}, "", 1
    if not os.path.exists(path):
        return rows, rules, 2
    for line in open(path):
        line = line.rstrip("\n")
        if line.startswith("# ledger"):
            version = 2 if "v2" in line.split()[2] else 1
            rules = line.split("rules=")[-1].strip()
            continue
        if not line or line.startswith("#"):
            continue
        cols = line.split("\t")
        if version == 1:
            # v1: path, style_blob, judge_blob, verdict, note -> named lenses
            cols = (cols + ["", "", "", ""])[:5]
            lenses = {}
            if cols[1]:
                lenses["style"] = cols[1]
            if cols[2]:
                lenses["judge"] = cols[2]
            rows[cols[0]] = {"verdict": cols[3] or "CLEAN", "note": cols[4], "lenses": lenses}
        else:
            cols = (cols + ["", "", ""])[:4]
            lenses = dict(p.split(":", 1) for p in cols[3].split(",") if ":" in p)
            rows[cols[0]] = {"verdict": cols[1] or "CLEAN", "note": cols[2], "lenses": lenses}
    return rows, rules, version


def save(path, rows, rules):
    os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
    with open(path, "w") as f:
        f.write(f"# ledger v2 rules={rules}\n")
        f.write("# path<TAB>verdict<TAB>note<TAB>lens:blob,lens:blob,...   (@name = a flow)\n")
        for p in sorted(rows):
            r = rows[p]
            pairs = ",".join(f"{k}:{v}" for k, v in sorted(r["lenses"].items()))
            f.write("\t".join([p, r["verdict"], r["note"], pairs]) + "\n")


def open_ledger(args):
    path = ledger_path(args.repo, args.ledger)
    rows, rules, version = load(path)
    current = rules_key()
    if version == 1:
        # v1 pinned the asl HEAD, v2 pins the analyzer hash: a different namespace,
        # not evidence that a rule moved. Adopt the new key and keep the verdicts.
        print(f"# migrated {len(rows)} v1 rows to v2 (style/judge blobs kept, rules key rebased)", file=sys.stderr)
    elif rules and rules != current:
        print(f"# analyzer rules changed ({rules} -> {current}): every row is due", file=sys.stderr)
        rows = {}
    return path, rows, current


def cmd_diff(args):
    _, rows, _ = open_ledger(args)
    blobs = tracked(args.repo)
    if args.flows:
        flows = load_flows(ledger_path(args.repo, args.ledger))
        due = [n for n, members in flows.items()
               if rows.get("@" + n, {}).get("lenses", {}).get(args.lens) != flow_blob(blobs, members)]
        for n in sorted(due):
            print("@" + n)
        print(f"# {len(due)} of {len(flows)} flows due for the {args.lens} lens", file=sys.stderr)
        return
    due = [p for p, b in blobs.items()
           if in_scope(args.lens, p) and rows.get(p, {}).get("lenses", {}).get(args.lens) != b]
    for p in sorted(due):
        print(p)
    scope = sum(in_scope(args.lens, p) for p in blobs)
    print(f"# {len(due)} of {scope} files due for the {args.lens} lens", file=sys.stderr)


def cmd_mark(args):
    path, rows, current = open_ledger(args)
    blobs = tracked(args.repo)
    flows = load_flows(path)
    for p in args.paths:
        if p.startswith("@"):
            name = p[1:]
            if name not in flows:
                sys.exit(f"{p}: no such flow in {flows_path(path)}")
            blob = flow_blob(blobs, flows[name])
        elif p in blobs:
            blob = blobs[p]
        else:
            sys.exit(f"{p}: not tracked in {args.repo}")
        row = rows.setdefault(p, {"verdict": "CLEAN", "note": "", "lenses": {}})
        row["lenses"][args.lens] = blob
        row["verdict"] = args.verdict
        if args.note:
            row["note"] = args.note
    save(path, rows, current)
    print(f"# marked {len(args.paths)} entries {args.verdict} for the {args.lens} lens", file=sys.stderr)


def cmd_lenses(args):
    path, rows, _ = open_ledger(args)
    blobs = tracked(args.repo)
    flows = load_flows(path)
    seen = {}
    for p, r in rows.items():
        for lens, blob in r["lenses"].items():
            cur = flow_blob(blobs, flows.get(p[1:], [])) if p.startswith("@") else blobs.get(p)
            seen.setdefault(lens, [0, 0])
            seen[lens][1] += 1
            if blob == cur:
                seen[lens][0] += 1
    print(f"{'lens':16} {'current':>8} {'recorded':>9}  scope {len(blobs)} files, {len(flows)} flows")
    for lens in sorted(seen):
        cur, rec = seen[lens]
        print(f"{lens:16} {cur:8} {rec:9}")
    if not seen:
        print("(no lens has been recorded yet: every file is due for every lens)")


def cmd_flows(args):
    path, rows, _ = open_ledger(args)
    blobs = tracked(args.repo)
    flows = load_flows(path)
    if not flows:
        print(f"(no flows declared; create {flows_path(path)} as `name<TAB>path [path ...]`)")
        return
    for n in sorted(flows):
        members = flows[n]
        missing = [m for m in members if m not in blobs]
        lenses = rows.get("@" + n, {}).get("lenses", {})
        fresh = [l for l, b in lenses.items() if b == flow_blob(blobs, members)]
        print(f"@{n}  {len(members)} files  fresh_for={','.join(sorted(fresh)) or '-'}"
              + (f"  MISSING={missing}" if missing else ""))


def cmd_prune(args):
    path, rows, current = open_ledger(args)
    blobs = tracked(args.repo)
    flows = load_flows(path)
    gone = [p for p in rows if not p.startswith("@") and p not in blobs] + \
           [p for p in rows if p.startswith("@") and p[1:] not in flows]
    for p in gone:
        del rows[p]
    save(path, rows, current)
    print(f"# pruned {len(gone)} rows", file=sys.stderr)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)
    for name, fn, extra in (("diff", cmd_diff, "diff"), ("mark", cmd_mark, "mark"),
                            ("lenses", cmd_lenses, ""), ("flows", cmd_flows, ""), ("prune", cmd_prune, "")):
        p = sub.add_parser(name)
        p.add_argument("repo")
        p.add_argument("ledger", nargs="?")
        if extra in ("diff", "mark"):
            p.add_argument("--lens", required=extra == "mark", default="style")
        if extra == "diff":
            p.add_argument("--flows", action="store_true")
        if extra == "mark":
            p.add_argument("--verdict", choices=("CLEAN", "KEPT"), required=True)
            p.add_argument("--note", default="")
            p.add_argument("paths", nargs="+")
        p.set_defaults(fn=fn)
    args = ap.parse_args()
    args.fn(args)


if __name__ == "__main__":
    main()
