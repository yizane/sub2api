#!/usr/bin/env bash
# 本地构建并推送镜像，默认发布到当前使用的阿里云镜像仓库。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

REGISTRY_REPO="${REGISTRY_REPO:-registry.cn-hangzhou.aliyuncs.com/data_server/sub2api-ha}"
PLATFORM="${PLATFORM:-linux/amd64}"
VERSION_FILE="${VERSION_FILE:-${REPO_ROOT}/backend/cmd/server/VERSION}"

if [[ ! -f "${VERSION_FILE}" ]]; then
    echo "VERSION file not found: ${VERSION_FILE}" >&2
    exit 1
fi

VERSION="$(tr -d '\r\n' < "${VERSION_FILE}")"
COMMIT="$(git -C "${REPO_ROOT}" rev-parse --short HEAD)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

echo "Building and pushing ${REGISTRY_REPO}:${VERSION}"
echo "Platform: ${PLATFORM}"
echo "Commit: ${COMMIT}"

docker buildx build \
    --platform "${PLATFORM}" \
    -t "${REGISTRY_REPO}:${VERSION}" \
    -t "${REGISTRY_REPO}:latest" \
    --build-arg VERSION="${VERSION}" \
    --build-arg COMMIT="${COMMIT}" \
    --build-arg DATE="${DATE}" \
    --build-arg GOPROXY=https://goproxy.cn,direct \
    --build-arg GOSUMDB=sum.golang.google.cn \
    --push \
    -f "${REPO_ROOT}/Dockerfile" \
    "${REPO_ROOT}"
