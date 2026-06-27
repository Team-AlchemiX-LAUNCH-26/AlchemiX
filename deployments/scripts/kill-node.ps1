param([Parameter(Mandatory=$true)][string]$Node)
docker compose -f deployments/docker/compose.yaml stop $Node.ToLower()
