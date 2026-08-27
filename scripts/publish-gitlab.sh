#!/usr/bin/env bash
set -euo pipefail

# Publish CodeMie Terraform Provider & Modules to GitLab Package Registry
# Usage: ./scripts/publish-gitlab.sh [VERSION]

VERSION="${1:-${CI_COMMIT_TAG:-${VERSION:-0.1.0}}}"
VERSION="${VERSION#v}" # Strip leading 'v'

PROVIDER_NAME="codemie"
DIST_DIR="dist"

if [ ! -d "${DIST_DIR}" ]; then
  echo "Error: Directory '${DIST_DIR}' not found. Run ./scripts/package-provider.sh first."
  exit 1
fi

CI_API_V4_URL="${CI_API_V4_URL:-https://git.epam.com/api/v4}"
if [ -z "${CI_PROJECT_ID:-}" ]; then
  echo "Error: CI_PROJECT_ID environment variable is required."
  exit 1
fi

AUTH_HEADER=""
if [ -n "${CI_JOB_TOKEN:-}" ]; then
  AUTH_HEADER="JOB-TOKEN: ${CI_JOB_TOKEN}"
elif [ -n "${GITLAB_TOKEN:-}" ]; then
  AUTH_HEADER="PRIVATE-TOKEN: ${GITLAB_TOKEN}"
else
  echo "Error: CI_JOB_TOKEN or GITLAB_TOKEN environment variable is required."
  exit 1
fi

echo "===================================================================="
echo "==> Publishing to GitLab Package Registry..."
echo "    Project ID: ${CI_PROJECT_ID}"
echo "    API URL:    ${CI_API_V4_URL}"
echo "    Version:    ${VERSION}"
echo "===================================================================="

# 1. Publish Terraform Modules to GitLab Terraform Module Registry
if [ -d "${DIST_DIR}/modules" ]; then
  echo "==> Publishing Terraform Modules to GitLab Module Registry..."
  for TAR_FILE in "${DIST_DIR}/modules"/module-${PROVIDER_NAME}-*-${VERSION}.tar.gz; do
    [ -f "${TAR_FILE}" ] || continue
    FILENAME="$(basename "${TAR_FILE}")"
    # Extract module system/name: module-codemie-assistant-0.1.0.tar.gz -> assistant
    MOD_NAME="${FILENAME#module-${PROVIDER_NAME}-}"
    MOD_NAME="${MOD_NAME%-${VERSION}.tar.gz}"

    MODULE_UPLOAD_URL="${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/packages/terraform/modules/${PROVIDER_NAME}/${MOD_NAME}/${VERSION}/file"
    echo "  --> Uploading module '${PROVIDER_NAME}/${MOD_NAME}' (${FILENAME})..."
    
    RESPONSE_FILE=$(mktemp)
    HTTP_STATUS=$(curl -s -w "%{http_code}" -o "${RESPONSE_FILE}" \
      --header "${AUTH_HEADER}" \
      --upload-file "${TAR_FILE}" \
      "${MODULE_UPLOAD_URL}")
    RESPONSE_BODY=$(cat "${RESPONSE_FILE}")
    rm -f "${RESPONSE_FILE}"

    if [ "${HTTP_STATUS}" -ge 200 ] && [ "${HTTP_STATUS}" -lt 300 ]; then
      echo "      [OK] Module '${PROVIDER_NAME}/${MOD_NAME}' published (HTTP ${HTTP_STATUS})"
    elif [[ "${RESPONSE_BODY}" =~ "already exists" ]]; then
      echo "      [SKIP] Module '${PROVIDER_NAME}/${MOD_NAME}' v${VERSION} already exists in registry."
    else
      echo "      [FAIL] Failed to publish module '${PROVIDER_NAME}/${MOD_NAME}' (HTTP ${HTTP_STATUS}): ${RESPONSE_BODY}"
      exit 1
    fi
  done
fi

# 2. Upload Provider Binaries & Checksums to Generic Package Registry
echo "==> Publishing Provider Binaries to GitLab Package Registry..."
for FILE in "${DIST_DIR}"/*; do
  [ -f "${FILE}" ] || continue
  FILENAME="$(basename "${FILE}")"
  GENERIC_URL="${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/packages/generic/terraform-provider-${PROVIDER_NAME}/${VERSION}/${FILENAME}"
  
  echo "  --> Uploading ${FILENAME}..."
  RESPONSE_FILE=$(mktemp)
  HTTP_STATUS=$(curl -s -w "%{http_code}" -o "${RESPONSE_FILE}" \
    --header "${AUTH_HEADER}" \
    --upload-file "${FILE}" \
    "${GENERIC_URL}")
  RESPONSE_BODY=$(cat "${RESPONSE_FILE}")
  rm -f "${RESPONSE_FILE}"

  if [ "${HTTP_STATUS}" -ge 200 ] && [ "${HTTP_STATUS}" -lt 300 ]; then
    echo "      [OK] Uploaded ${FILENAME} (HTTP ${HTTP_STATUS})"
  elif [[ "${RESPONSE_BODY}" =~ "already exists" ]]; then
    echo "      [SKIP] ${FILENAME} already exists in package registry."
  else
    echo "      [FAIL] Failed to upload ${FILENAME} (HTTP ${HTTP_STATUS}): ${RESPONSE_BODY}"
    exit 1
  fi
done

echo "===================================================================="
echo "==> Successfully published terraform-provider-${PROVIDER_NAME} v${VERSION}!"
echo "===================================================================="

echo "==> Successfully published terraform-provider-${PROVIDER_NAME} v${VERSION} to GitLab Package Registry!"
