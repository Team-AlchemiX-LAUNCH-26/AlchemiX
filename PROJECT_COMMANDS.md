# AlchemiX Project Commands

Run commands from the repository root unless a section says otherwise:

```powershell
cd C:\Users\Administrator\Desktop\AlchemiX-phase-2-draft
```

## Python ML Environment

Create and activate the virtual environment:

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
```

Install dependencies:

```powershell
python -m pip install --upgrade pip
python -m pip install -r training\requirements.txt
```

Validate datasets and run tests:

```powershell
cd training
python validate_data.py
python -m pytest tests
cd ..
```

Train and export models:

```powershell
python -B training\train_congestion.py
python -B training\train_trust.py
python -B training\train_targeting.py
python -B training\evaluate_all.py
```

Outputs are written to:

```text
internal/models/
training/reports/
```

## FastAPI ML Service

Start the prediction service:

```powershell
python -B -m uvicorn training.service.app:app --host 127.0.0.1 --port 8100
```

Check it:

```powershell
Invoke-RestMethod http://127.0.0.1:8100/health
Invoke-RestMethod http://127.0.0.1:8100/model-info
```

Open the API docs:

```text
http://127.0.0.1:8100/docs
```

If port `8100` is already in use:

```powershell
netstat -ano | findstr :8100
Stop-Process -Id <PID>
```

Or run on another port:

```powershell
python -B -m uvicorn training.service.app:app --host 127.0.0.1 --port 8101
```

## Frontend

Install dependencies and build:

```powershell
cd frontend
npm install
npm.cmd run build
cd ..
```

Use `npm.cmd` on Windows PowerShell if `npm` is blocked by script execution policy.

## Go Backend

Install/check Go dependencies and tests:

```powershell
go mod tidy
go test ./...
go vet ./...
```

## Wails Desktop App

Install Wails if needed:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

Run the desktop app:

```powershell
wails dev
```

## Docker Network

Start the distributed network:

```powershell
docker compose -f deployments/docker/compose.yaml up --build
```

## Recommended Run Order

Terminal 1, ML service:

```powershell
.\.venv\Scripts\Activate.ps1
python -B training\train_congestion.py
python -B training\train_trust.py
python -B training\train_targeting.py
python -B training\evaluate_all.py
python -B -m uvicorn training.service.app:app --host 127.0.0.1 --port 8100
```

Terminal 2, Docker network:

```powershell
docker compose -f deployments/docker/compose.yaml up --build
```

Terminal 3, desktop app:

```powershell
wails dev
```
