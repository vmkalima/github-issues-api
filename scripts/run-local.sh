#!/usr/bin/env bash
set -euo pipefail

if [ -z "${API_TOKEN:-}" ]; then
	echo "ERROR: API_TOKEN is not set."
	echo "  export API_TOKEN=\"choose-a-secret\""
	exit 1
fi

if [ -z "${GITHUB_TOKEN:-}" ]; then
	echo "ERROR: GITHUB_TOKEN is not set."
	echo "  export GITHUB_TOKEN=\"a-fine-grained-PAT-scoped-to-Issues-only\""
	exit 1
fi

echo "==> Starting github-issues-api on :8080..."
go run .