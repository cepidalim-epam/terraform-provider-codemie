#!/usr/bin/env bash
set -euo pipefail

# Package script for CodeMie Terraform Provider and Modules
# Usage: ./scripts/package-provider.sh [VERSION]

VERSION="${1:-${CI_COMMIT_TAG:-${VERSION:-0.1.0}}}"
VERSION="${VERSION#v}" # Strip leading 'v' if present

PROVIDER_NAME="codemie"
DIST_DIR="dist"
TARGETS=(
  "linux_amd64"
  "linux_arm64"
  "darwin_amd64"
  "darwin_arm64"
  "windows_amd64"
)

echo "==> Building terraform-provider-${PROVIDER_NAME} v${VERSION}..."
rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}/modules"

# 1. Compile and package provider binaries
for TARGET in "${TARGETS[@]}"; do
  OS="${TARGET%_*}"
  ARCH="${TARGET#*_}"
  STAGE_DIR="${DIST_DIR}/${TARGET}"
  mkdir -p "${STAGE_DIR}"

  BINARY_NAME="terraform-provider-${PROVIDER_NAME}_v${VERSION}"
  if [ "${OS}" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
  fi

  echo "  --> Compiling for ${OS}/${ARCH}..."
  CGO_ENABLED=0 GOOS="${OS}" GOARCH="${ARCH}" go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "${STAGE_DIR}/${BINARY_NAME}" \
    .

  # Include terraform-registry-manifest.json if present
  if [ -f "terraform-registry-manifest.json" ]; then
    cp "terraform-registry-manifest.json" "${STAGE_DIR}/"
  fi

  # Include LICENSE if present
  if [ -f "LICENSE" ]; then
    cp "LICENSE" "${STAGE_DIR}/"
  fi

  ZIP_NAME="terraform-provider-${PROVIDER_NAME}_${VERSION}_${OS}_${ARCH}.zip"
  echo "  --> Packaging ${ZIP_NAME}..."
  (cd "${STAGE_DIR}" && zip -q -9 "../../${DIST_DIR}/${ZIP_NAME}" *)
  rm -rf "${STAGE_DIR}"
done

# 2. Generate SHA256 checksums
echo "==> Generating SHA256 checksums..."
SHASUMS_NAME="terraform-provider-${PROVIDER_NAME}_${VERSION}_SHA256SUMS"
(
  cd "${DIST_DIR}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum terraform-provider-${PROVIDER_NAME}_${VERSION}_*.zip > "${SHASUMS_NAME}"
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 terraform-provider-${PROVIDER_NAME}_${VERSION}_*.zip > "${SHASUMS_NAME}"
  fi
)

# 3. Package Terraform Modules into tar.gz for GitLab Terraform Module Registry
echo "==> Packaging Terraform Modules for GitLab Module Registry..."
if [ -d "modules" ]; then
  for MOD_DIR in modules/*; do
    [ -d "${MOD_DIR}" ] || continue
    MOD_NAME="$(basename "${MOD_DIR}")"
    TAR_NAME="module-${PROVIDER_NAME}-${MOD_NAME}-${VERSION}.tar.gz"
    echo "  --> Packaging module ${MOD_NAME} (${TAR_NAME})..."
    tar -czf "${DIST_DIR}/modules/${TAR_NAME}" -C "${MOD_DIR}" .
  done
fi

echo "==> Build complete. Artifacts created in ${DIST_DIR}/:"
ls -lh "${DIST_DIR}"
ls -lh "${DIST_DIR}/modules"
