# Documented Assumptions

- All physical calculations use seconds internally and are rounded only in the UI.
- Binary framing is a comma-delimited codex-token string converted to UTF-8 bytes and shown as eight-bit groups.
- The minimum chaos requirement reroutes the next packet. Mid-flight rerouting is not enabled in the baseline.
- Calculated physics latency is displayed immediately; the application does not wait for the full simulated duration.
- Links are undirected.
- The origin and final destination each incur one tower processing charge because their entry and exit tower are the same logical tower.
