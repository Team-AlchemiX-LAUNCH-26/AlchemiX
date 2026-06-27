# VS Code Setup Guide — Relic Ring Protocol

## Option A — Use the supplied complete project

1. Extract `relic-ring-protocol.zip`.
2. Open Visual Studio Code.
3. Select **File → Open Folder**.
4. Choose the extracted `relic-ring-protocol` folder.
5. Open **Terminal → New Terminal**.
6. Run:

```powershell
go mod tidy
cd frontend
npm install
cd ..
```

7. Check the development environment:

```powershell
go version
node --version
npm --version
docker --version
docker compose version
wails doctor
```

8. Start the distributed network:

```powershell
docker compose -f deployments/docker/compose.yaml up --build
```

9. Open a second VS Code terminal and start the desktop application:

```powershell
wails dev
```

## Option B — Create the project folders manually

Open an empty folder in VS Code, open the integrated PowerShell terminal, and run:

```powershell
wails init -n relic-ring-protocol -t react-ts
cd relic-ring-protocol
```

Copy `scripts/create-project-folders.ps1` into the root and run:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\scripts\create-project-folders.ps1
```

Then copy the coded files from the supplied project ZIP into the matching folders.

## Recommended VS Code extensions

- Go — `golang.go`
- Docker — `ms-azuretools.vscode-docker`
- ESLint — `dbaeumer.vscode-eslint`
- Prettier — `esbenp.prettier-vscode`

The project includes `.vscode/extensions.json` and `.vscode/tasks.json`. Open **Terminal → Run Task** to start Docker, run Go tests, build the frontend, or launch Wails development mode.

## Verification commands

```powershell
go test ./...
go vet ./...
cd frontend
npm run build
cd ..
```
