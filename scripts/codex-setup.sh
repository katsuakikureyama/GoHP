#!/bin/sh
# Reproduce the repository's pinned development environment and validate it.
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir"

if ! command -v docker >/dev/null 2>&1; then
  echo 'Docker Engine and Docker Compose v2 are required; see docs/codex.md.' >&2
  exit 1
fi
docker info >/dev/null
docker compose version

# Codex can provide a read-only home directory. Keep Buildx state in the
# existing ignored cache directory without replacing Docker authentication.
export BUILDX_CONFIG="${BUILDX_CONFIG:-$repo_dir/.cache/docker-buildx}"
mkdir -p "$BUILDX_CONFIG"

# Building again is intentional: Dockerfile copies the current source tree.
# Keep these images/containers separate from other Compose workflows.
docker compose --project-name gohp-codex --profile legacy build test php74
for service in test php74; do
  docker compose --project-name gohp-codex --profile legacy run --rm --no-deps -T "$service"
done

printf '%s\n' 'GoHP Codex setup PASS: PHP 8.3 and PHP 7.4 validation completed.'
