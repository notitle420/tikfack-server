#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
USER_ID="$(id -u)"
GROUP_ID="$(id -g)"
BUF_CACHE_DIR="${PROJECT_ROOT}/.buf-cache"

mkdir -p "${BUF_CACHE_DIR}"

# Use bufbuild/buf image so no local installation is needed.
docker run --rm \
  -u "${USER_ID}:${GROUP_ID}" \
  -v "${PROJECT_ROOT}:/workspace" \
  -v "${BUF_CACHE_DIR}:/tmp/buf-cache" \
  -w /workspace \
  -e XDG_CACHE_HOME=/tmp/buf-cache \
  --env BUF_TOKEN \
  bufbuild/buf:latest \
  generate "$@"
