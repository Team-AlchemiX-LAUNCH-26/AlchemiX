.PHONY: fmt vet test run-orchestrator run-aegis frontend-install frontend-build wails-dev docker-up docker-down

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

run-orchestrator:
	go run ./cmd/orchestrator

run-aegis:
	go run ./cmd/planet-node --planet Aegis --port 8101

frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

wails-dev:
	wails dev

docker-up:
	docker compose -f deployments/docker/compose.yaml up --build

docker-down:
	docker compose -f deployments/docker/compose.yaml down
