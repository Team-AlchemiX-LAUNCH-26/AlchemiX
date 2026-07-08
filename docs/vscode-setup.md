# VS Code Setup

1. Install Go, Node.js, Git, Docker Desktop, VS Code and Wails.
2. Open VS Code, press **Ctrl+Shift+P**, run **Git: Clone**, and open the cloned folder.
3. Install the official **Go**, **Docker**, **ESLint** and **Prettier** extensions.
4. Copy `.env.example` to `.env` if you need custom values.
5. Run `go mod tidy` in the integrated terminal.
6. Run `cd frontend && npm install && cd ..`.
7. Start the network with `docker compose -f deployments/docker/compose.yaml up --build`.
8. In another terminal, run `wails dev`.
