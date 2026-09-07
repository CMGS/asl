#!/bin/bash
# PreToolUse gate on `git commit`: asl layout findings block Go repos, and
# review/fix/lint/cut/simplify/style/tidy commits must not add comment lines.
# The target repo comes from the command (`cd <dir> &&`, `git -C <dir>`), not
# from the session cwd; heredoc bodies are ignored so a script that merely
# mentions git commit is not gated.
input=$(cat)
cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty')
[ -n "$cmd" ] || exit 0
cmd=$(printf '%s\n' "$cmd" | perl -0pe 's/<<-?\s*["\x27]?(\w+)["\x27]?[^\n]*\n.*?\n[ \t]*\1[ \t]*(\n|\z)/\n/sg')
printf '%s' "$cmd" | grep -qE '(^|[^[:alnum:]_-])git[[:space:]]' || exit 0

resolve() {
	local d=$1
	d=${d//\"/}
	d=${d//\'/}
	d=${d/#\~/$HOME}
	[[ $d = /* ]] || d="$2/$d"
	printf '%s' "$d"
}

gate() {
	local seg=$1 dir=$2 root fail="" findings prefix range diff added removed
	root=$(git -C "$dir" rev-parse --show-toplevel 2>/dev/null) || return 0
	if [ -f "$root/go.mod" ] && command -v asl >/dev/null 2>&1; then
		findings=$(cd "$root" && { asl -forwarder=false ./... 2>&1; GOOS=linux asl -forwarder=false ./... 2>&1; } | grep -v '^#' | grep -v '^$' | sort -u)
		[ -n "$findings" ] && fail+="asl gate: layout findings block this commit (fix, then retry):"$'\n'"$findings"$'\n'
	fi
	if [[ $seg =~ -m[[:space:]]*[\"\']?(review|fix|lint|cut|simplify|style|tidy): ]]; then
		prefix=${BASH_REMATCH[1]}
		range="--cached"
		[[ $seg =~ [[:space:]](-a|--all|-am|-qa|-am)([[:space:]]|$) ]] && range="HEAD"
		diff=$(git -C "$root" diff $range -U0 -- '*.go' '*.rs' '*.ts' '*.tsx' 2>/dev/null | grep -E '^[-+][[:space:]]*//')
		diff+=$'\n'$(git -C "$root" diff $range -U0 -- '*.py' 2>/dev/null | grep -E '^[-+][[:space:]]*#[^!]')
		added=$(printf '%s\n' "$diff" | grep -c '^+')
		removed=$(printf '%s\n' "$diff" | grep -c '^-')
		if [ "$added" -gt "$removed" ]; then
			fail+="comment gate: a ${prefix}: commit adds comment lines (+$added/-$removed); move the rationale to the commit message:"$'\n'
			fail+=$(printf '%s\n' "$diff" | grep '^+' | head -10)$'\n'
		fi
	fi
	if [ -n "$fail" ]; then
		printf '%s' "$fail" >&2
		exit 2
	fi
}

dir=$cwd
while IFS= read -r seg; do
	seg="${seg#"${seg%%[![:space:]]*}"}"
	[ -n "$seg" ] || continue
	if [[ $seg =~ ^(cd|pushd)[[:space:]]+([^[:space:]]+) ]]; then
		dir=$(resolve "${BASH_REMATCH[2]}" "$dir")
		continue
	fi
	[[ $seg =~ (^|[^[:alnum:]_-])git([[:space:]]+[^[:space:]]+)*[[:space:]]+commit([[:space:]]|$) ]] || continue
	target=$dir
	[[ $seg =~ git[[:space:]]+-C[[:space:]]+([^[:space:]]+) ]] && target=$(resolve "${BASH_REMATCH[1]}" "$dir")
	gate "$seg" "$target"
done < <(printf '%s\n' "$cmd" | perl -pe 's/\s*(&&|\|\||;)\s*/\n/g')
exit 0
