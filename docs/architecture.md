# Architecture

The system is a distributed digital twin:

1. The Wails desktop client calls the Go orchestrator.
2. The orchestrator loads and validates the shared JSON configuration.
3. It builds a valid graph, excludes failed nodes/links and calculates the lowest-latency route.
4. The packet is sent to the origin planet service.
5. Each independent planet decodes, calculates local transit, logs, re-encodes and forwards the packet.
6. Planet telemetry is streamed to the orchestrator and desktop UI.

The React UI contains no official routing or latency formulas.
