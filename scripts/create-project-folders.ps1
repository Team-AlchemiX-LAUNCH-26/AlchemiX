# Run this only when creating the structure manually instead of extracting the supplied project ZIP.
$folders = @(
  "cmd/orchestrator", "cmd/planet-node",
  "internal/config", "internal/domain", "internal/geometry", "internal/latency",
  "internal/encoding", "internal/routing", "internal/packet", "internal/planet",
  "internal/resilience", "internal/transport", "internal/orchestrator",
  "internal/simulation", "internal/observability", "internal/errors",
  "pkg/protocol", "frontend/src/components", "frontend/src/types",
  "configs", "deployments/docker", "deployments/scripts",
  "tests/unit", "tests/integration", "tests/fixtures", "docs", ".github/workflows"
)

foreach ($folder in $folders) {
  New-Item -ItemType Directory -Path $folder -Force | Out-Null
}

Write-Host "Relic Ring folders created successfully." -ForegroundColor Green
