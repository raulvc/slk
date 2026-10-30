#!/usr/bin/env bash
# Sync this fork with upstream and reinstall the slk binary.
#
# Branch layout:
#   main  pristine upstream (github.com/gammons/slk), fast-forward only
#   raul  this fork's working branch: upstream + personal patches,
#         kept in sync by rebasing onto main
#
# Usage: scripts/sync.sh
#
# The build is stamped with the nearest upstream tag plus the fork
# commit (e.g. v0.29.0-63.raul-1a2b3c4) and installed to ~/.local/bin,
# which shadows any AUR-installed /usr/bin/slk.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

git fetch origin --tags
git checkout main
git merge --ff-only origin/main
git checkout raul
git rebase main

tag=$(git describe --tags --abbrev=0)
sha=$(git rev-parse --short HEAD)
version="${tag}.raul-${sha}"
date=$(date -u +%Y-%m-%dT%H:%M:%SZ)

mkdir -p bin
go build -ldflags="-s -w -X main.version=${version} -X main.commit=${sha} -X main.date=${date}" \
	-trimpath -o bin/slk ./cmd/slk
install -m 0755 bin/slk "$HOME/.local/bin/slk"

echo "installed: $HOME/.local/bin/slk ($("$HOME/.local/bin/slk" --version))"
