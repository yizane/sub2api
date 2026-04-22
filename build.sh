#!/usr/bin/env bash
# build.sh — Build and push sub2api Docker image to Alibaba Cloud registry.
#
# Usage:
#   ./build.sh              # auto-detect version from backend/cmd/server/VERSION
#   ./build.sh 0.1.116      # explicit version tag
#
# Requirements: docker buildx (Docker Desktop default)

set -euo pipefail

REGISTRY="registry.cn-hangzhou.aliyuncs.com/data_server/sub2api-ha"
VERSION="${1:-$(tr -d '[:space:]' < backend/cmd/server/VERSION)}"

echo "Building ${REGISTRY}:${VERSION} (linux/amd64) ..."

docker buildx build \
  --platform linux/amd64 \
  -f deploy/Dockerfile \
  -t "${REGISTRY}:${VERSION}" \
  -t "${REGISTRY}:latest" \
  --push \
  .

echo ""
echo "Done: ${REGISTRY}:${VERSION}"
echo ""
echo "Verifying manifest platform (expect linux/amd64):"
docker buildx imagetools inspect "${REGISTRY}:${VERSION}" | grep -E "^  Platform:" | grep -v "unknown/unknown"
