#!/usr/bin/env bash
set -euo pipefail

echo "== Go formatting =="
gofmt -w agent_app.go internal/models/embed.go tests/agent/*.go

echo "== Go vet =="
go vet ./...

echo "== Go tests with race detector =="
go test -race ./...

echo "== Frontend build =="
(
  cd frontend
  npm run build
)

echo "== Wails production build =="
wails build

echo "All verification steps completed."
