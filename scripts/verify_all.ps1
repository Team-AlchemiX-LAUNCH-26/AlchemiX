$ErrorActionPreference = "Stop"

Write-Host "== Go formatting =="
gofmt -w agent_app.go internal/models/embed.go tests/agent/*.go

Write-Host "== Go vet =="
go vet ./...

Write-Host "== Go tests with race detector =="
go test -race ./...

Write-Host "== Frontend build =="
Push-Location frontend
try {
    npm run build
}
finally {
    Pop-Location
}

Write-Host "== Wails production build =="
wails build

Write-Host "All verification steps completed."
