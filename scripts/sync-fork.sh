#!/usr/bin/env bash
# Keep this fork current: rebase local main onto upstream/main, push to the
# fork, rebuild and install.
#
#   scripts/sync-fork.sh [--dry-run]
#
# Expects two remotes: `upstream` (gammons/slk) and `origin` (the fork).
# Run it whenever upstream moves; local commits always stay on top.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

dry=0
if [ "${1:-}" = "--dry-run" ]; then
    dry=1
elif [ -n "${1:-}" ]; then
    echo "usage: scripts/sync-fork.sh [--dry-run]"
    exit 1
fi

git remote get-url upstream >/dev/null || {
    echo "no 'upstream' remote:"
    echo "  git remote add upstream https://github.com/gammons/slk"
    exit 1
}
git remote get-url origin >/dev/null || {
    echo "no 'origin' remote (the fork)"
    exit 1
}

branch=$(git branch --show-current)
[ "$branch" = "main" ] || { echo "on '$branch', not main — switch first"; exit 1; }
[ -z "$(git status --porcelain)" ] || {
    echo "working tree not clean — commit or stash first"
    exit 1
}
gitdir=$(git rev-parse --git-dir)
{ [ ! -e "$gitdir/rebase-merge" ] && [ ! -e "$gitdir/rebase-apply" ]; } || {
    echo "a rebase is in progress — finish or abort it first"
    exit 1
}

git fetch upstream --prune --tags
git fetch origin --prune

ahead=$(git rev-list --count upstream/main..main)
behind=$(git rev-list --count main..upstream/main)
echo "local commits on top of upstream: $ahead; upstream commits to pull: $behind"

if [ "$behind" -gt 0 ]; then
    if [ "$dry" -eq 1 ]; then
        echo "dry-run: would rebase onto upstream/main"
    elif ! git rebase upstream/main; then
        echo
        echo "rebase stopped on conflicts. Resolve, then:"
        echo "  git add <files> && git rebase --continue"
        echo "  scripts/sync-fork.sh        # run again to push"
        echo "If README.md conflicts, keep this fork's block between the"
        echo "FORK:BEGIN / FORK:END markers and merge it into upstream's text."
        exit 1
    fi
fi

if [ "$dry" -eq 1 ]; then
    echo "dry-run: would push main to origin (--force-with-lease)"
else
    git push --force-with-lease origin main
    echo "fork main = upstream/main + $(git rev-list --count upstream/main..main) local commits"
fi

# Reinstall with a stamped version (upstream tag + fork commit), so
# `slk --version` says where this binary came from. go install puts it
# in $(go env GOPATH)/bin (~/go/bin by default), which is on PATH.
if [ "$dry" -eq 1 ]; then
    echo "dry-run: would go install slk"
else
    tag=$(git describe --tags --abbrev=0)
    sha=$(git rev-parse --short HEAD)
    version="${tag}.fork-${sha}"
    date=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    go install -ldflags="-s -w -X main.version=${version} -X main.commit=${sha} -X main.date=${date}" \
        -trimpath ./cmd/slk
    bin="$(go env GOPATH)/bin/slk"
    echo "installed: $bin ($("$bin" --version | head -1))"
fi
