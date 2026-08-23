#!/usr/bin/env bash
set -euo pipefail

echo "** Validating YML..."
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"

echo "** Building..."
go build ./...

echo "** Running go vet..."
go vet ./...

echo "** Running tests..."
go test ./...

echo "** Running lint..."
golangci-lint run

echo "** Running gosec..."
gosec ./...

echo "** Running govulncheck..."
govulncheck ./...

echo "** Simulating full CI pipeline locally (act)..."
/usr/local/bin/act

echo "** All checks passed! **"