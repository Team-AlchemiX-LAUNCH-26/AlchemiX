$ErrorActionPreference = "Stop"

Write-Host "Checking Go..."
go version

Write-Host "Downloading Go modules..."
go mod tidy

Write-Host "Installing frontend packages..."
Push-Location frontend
npm install
Pop-Location

Write-Host "Running backend tests..."
go test ./...

Write-Host "Setup completed. Start Docker Compose, then run: wails dev" -ForegroundColor Green
