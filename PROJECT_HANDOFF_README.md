# AlchemiX — Relic Ring Protocol Project Handoff

This document summarizes the work completed so far so the project can be continued in a new chat without losing context.

It is a **development handoff README**, not only a public repository introduction.

---

## 1. Project Summary

**Project name:** AlchemiX  
**Challenge:** LAUNCH26 — Relic Ring Protocol  
**System:** Zeta-26 star system  
**Application type:** Go + Wails desktop application with a React frontend

The system simulates interplanetary communication after the collapse of the fictional Aether-Net.

Packets travel through:

- Planetary communication towers
- Internal planetary fiber rings
- Atmospheric shells
- Laser-based void links
- Intermediate relay planets

The system must:

- Find the lowest-latency valid route
- Respect the maximum void-hop distance
- Calculate physical latency
- Translate messages between planetary codices
- Store a complete ordered hop log
- Reroute around failed planets, links, and towers
- Visualize the full process in a desktop UI

---

## 2. Technology Stack

### Frontend

- React
- TypeScript
- Vite
- CSS
- SVG-based universe topology
- Lucide React icons
- Wails-generated JavaScript bindings

### Backend

- Go
- Wails v2
- Go modules
- JSON configuration
- In-memory failure state

### Infrastructure

- Docker
- Docker Compose
- Multi-service design for planet services and the orchestrator
- Git and GitHub

The React frontend calls exported Go methods through generated bindings under:

```text
frontend/wailsjs/
```

Wails runtime events are used for live network updates.

---

## 3. Challenge Source Files

The project is based on:

```text
universe-config.json
Equations .pdf
Launch26 - Phase 01 - Challenge.pdf
```

The universe configuration includes:

- System metadata
- Physical constants
- Planet coordinates
- Planet radii
- Planet codices
- Planet tower counts
- Atmosphere thickness values
- Refraction indices

Configured planets:

- Aegis
- Boreas
- Dawn
- Elysium
- Fenix
- Caelum

Important metadata:

```json
{
  "system_name": "Zeta-26",
  "speed_of_light_kms": 300000.0,
  "max_void_hop_distance_km": 50000000.0,
  "coordinate_scale_unit_km": 100000.0,
  "tower_processing_delay_ms": 7.0,
  "fiber_speed_fraction": 0.67
}
```

---

## 4. Core Features Implemented

### Universe Initialization

- Load planets and metadata dynamically from JSON
- Avoid hardcoding values already present in configuration
- Build the universe topology from the loaded data

### Routing

- Build an active graph
- Remove disabled planets
- Remove disabled links
- Enforce the maximum void-hop distance
- Use latency-aware route weights
- Support direct and multi-hop transmission
- Return an undeliverable result when no path exists

### Latency

The system considers:

- Internal fiber transit
- Tower processing delay
- Atmospheric refraction
- Void transmission
- Total route latency

### Codex Translation

The packet translation process is:

```text
Text
→ ASCII decimal
→ receiving planet codex
→ serialized transmission data
→ decode at receiving planet
→ ASCII for internal routing
→ encode for next planet codex
```

### Resilience

The system supports:

- Disable planet
- Restore planet
- Disable link
- Restore link
- Disable tower
- Restore tower
- Reset all failures

---

## 5. Backend Structure

The backend was divided into packages similar to:

```text
internal/
├── config/
├── domain/
├── encoding/
├── geometry/
├── latency/
├── packet/
├── resilience/
├── routing/
└── service/
```

Verify exact names against the current repository.

### Package Responsibilities

#### Config

- Load `universe-config.json`
- Validate metadata
- Validate planet definitions

#### Geometry

- Calculate planet distances
- Calculate tower positions
- Select operational towers
- Handle ring traversal
- Avoid disabled towers

#### Latency

- Calculate internal planet latency
- Calculate void latency
- Add tower processing delay
- Produce detailed latency breakdowns

#### Encoding

- Convert text to ASCII
- Convert ASCII to any planet codex
- Decode codex values back to ASCII
- Serialize transmission data

#### Routing

- Store graph nodes and edges
- Build the graph from the active universe
- Find the lowest-cost valid route

#### Packet

- Store origin and destination
- Store current planet
- Store translated payload
- Append ordered hop logs

#### Resilience

- Store disabled nodes
- Store disabled links
- Store disabled towers
- Reset the failure state

---

## 6. Important Backend Fix

A duplicate declaration problem occurred because the `Graph` type was declared in more than one file.

The intended separation is:

```text
internal/routing/graph.go
```

Contains graph types.

```text
internal/routing/graph_builder.go
```

Contains graph-building logic only.

There should be only one main `Graph` declaration.

---

## 7. Tower Failure Support

Tower failure support was added through the backend layers.

The intended flow is:

1. Disabled tower state is stored in resilience.
2. Geometry excludes disabled towers.
3. Planet ring routing finds an operational path.
4. Graph construction considers tower availability.
5. Transmission services receive disabled tower data.
6. Wails exposes tower disable/restore methods.
7. The frontend refreshes after a tower change.

Expected Wails methods:

```go
DisableTower(planetID string, towerIndex int)
EnableTower(planetID string, towerIndex int)
```

Expected frontend bindings:

```ts
DisableTower(planetID, towerIndex)
EnableTower(planetID, towerIndex)
```

---

## 8. Wails Methods Used by the Frontend

The frontend expects methods similar to:

```ts
GetUniverse()
StartTransmission(request)

DisableNode(planetID)
EnableNode(planetID)

DisableLink(firstPlanetID, secondPlanetID)
EnableLink(firstPlanetID, secondPlanetID)

DisableTower(planetID, towerIndex)
EnableTower(planetID, towerIndex)

ResetNetwork()
```

When Go method signatures change, restart Wails:

```bash
wails dev
```

Do not manually edit generated files under `frontend/wailsjs/`.

---

## 9. Frontend Structure

Current intended structure:

```text
frontend/src/
├── App.tsx
├── styles.css
├── types/
│   └── index.ts
└── components/
    ├── UniverseMap.tsx
    ├── TowerInspector.tsx
    ├── ConversionHistory.tsx
    ├── LatencyPanel.tsx
    └── HopLogTable.tsx
```

---

## 10. Main UI Layout

### Initial View

Before sending a packet:

```text
┌───────────────────────────────┬──────────────────┐
│                               │                  │
│      Universe Topology        │   Transmission   │
│                               │      Controls    │
│                               │                  │
└───────────────────────────────┴──────────────────┘
```

- Universe topology on the left
- Transmission and chaos controls on the right

### After Sending a Packet

Additional cards appear underneath and the desktop window becomes vertically scrollable.

Result sections:

- Latency breakdown
- Delivery summary
- Live event stream
- Delivery conversion history
- Ordered hop log

---

## 11. Universe Map

`UniverseMap.tsx` is responsible for:

- Rendering planets using SVG
- Rendering links
- Highlighting the selected route
- Showing failed planets
- Showing selected planets
- Showing towers around every planet
- Showing active towers in green
- Showing inactive towers in red
- Allowing the user to click a planet

Latest callback:

```ts
onSelectPlanet?: (planetID: string) => void;
```

The old `PlanetPopupAnchor` approach was removed.

---

## 12. Tower Inspector

Clicking a planet opens the tower inspector.

Latest requirement:

> The tower inspector always appears in the center of the Universe Topology section.

Popup position:

```css
.tower-popover {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
}
```

The parent must use:

```css
.map-stage {
  position: relative;
  overflow: hidden;
}
```

The inspector includes:

- Planet name
- Planet codex
- Active tower count
- Inactive tower count
- Green/red legend
- Circular planet illustration
- Tower buttons around the circle
- Close button

Tower colors:

```text
Green = active
Red   = inactive
```

Tower buttons disable or restore individual towers.

---

## 13. Tower Inspector Translation View

The original tower-control list was removed.

The latest design keeps the circular tower map and replaces the list with packet translation information.

When the selected planet belongs to the latest route, show:

```text
Received codex
↓
Decoded text
↓
ASCII decimal values
↓
Next-hop codex
↓
Binary stream, when available
```

When no packet data exists for that planet, show an empty-state message.

---

## 14. `TowerInspectorProps` Type Fix

A TypeScript error occurred because `App.tsx` passed `result={result}` but the prop interface did not contain `result`.

The intended interface is:

```ts
interface TowerInspectorProps {
  snapshot: Snapshot;
  result?: TransmissionResult;
  selectedPlanetID: string | null;
  onClose: () => void;

  onToggleTower: (
    planetID: string,
    towerIndex: number,
    currentlyDisabled: boolean
  ) => Promise<void>;

  busy?: boolean;
}
```

The component must destructure `result`:

```ts
export function TowerInspector({
  snapshot,
  result,
  selectedPlanetID,
  onClose,
  onToggleTower,
  busy = false,
}: TowerInspectorProps) {
```

There must be only one `TowerInspectorProps` declaration.

---

## 15. Delivery Conversion History

A component was introduced:

```text
frontend/src/components/ConversionHistory.tsx
```

Its purpose is to show message conversion for every planet in the route.

Intended props:

```ts
interface ConversionHistoryProps {
  logs: unknown[];
  route: string[];
  planets: {
    id: string;
    codex: number;
  }[];
  originalPayload: string;
}
```

For every planet, it displays:

- Planet name
- Previous planet
- Next planet
- Current codex
- ASCII conversion
- Next-hop codex
- Per-character conversion table
- Complete input sequence
- Complete ASCII sequence
- Complete output sequence
- Optional binary stream

Example:

```text
Character | Received Base | ASCII Decimal | Next Base
-----------------------------------------------------
H         | 242           | 72            | 52
e         | 401           | 101           | 73
```

The component currently uses defensive parsing because the exact final Go hop-log schema was not fully confirmed.

Possible field names checked include:

```text
planet_id
current_id
node_id

decoded_payload
ascii_payload
local_payload

encoded_payload
outgoing_payload
translated_payload

binary_stream
binary_payload
serialized_payload
```

When backend sequences are missing, the frontend calculates display values from the decoded text as a fallback.

The next developer should align this with the actual Go response.

---

## 16. Result Section Integration

### Latency

```tsx
<LatencyPanel latency={result.latency} />
```

### Conversion History

```tsx
<ConversionHistory
  logs={result.packet.hop_log as unknown[]}
  route={result.route.path}
  planets={snapshot.planets}
  originalPayload={result.original_payload}
/>
```

### Hop Log

```tsx
<HopLogTable logs={result.packet.hop_log} />
```

### Tower Inspector

```tsx
<TowerInspector
  snapshot={snapshot}
  result={result}
  selectedPlanetID={selectedPlanetID}
  onClose={closeTowerInspector}
  onToggleTower={toggleTower}
  busy={busy}
/>
```

---

## 17. Main `App.tsx` State

```ts
const [snapshot, setSnapshot] = useState<Snapshot>();

const [origin, setOrigin] = useState("Aegis");
const [destination, setDestination] = useState("Caelum");
const [payload, setPayload] = useState("Hello world");

const [selectedPlanetID, setSelectedPlanetID] =
  useState<string | null>(null);

const [result, setResult] =
  useState<TransmissionResult>();

const [events, setEvents] =
  useState<NetworkEvent[]>([]);

const [linkA, setLinkA] = useState("Aegis");
const [linkB, setLinkB] = useState("Dawn");

const [busy, setBusy] = useState(false);
const [error, setError] = useState("");
```

Main handlers:

```text
loadUniverse
openTowerInspector
closeTowerInspector
sendPacket
toggleNode
toggleLink
toggleTower
resetNetwork
```

Live events use:

```ts
EventsOn("network:event", (raw: unknown) => {
  const event = raw as NetworkEvent;

  setEvents((currentEvents) =>
    [event, ...currentEvents].slice(0, 100)
  );

  void loadUniverse();
});
```

---

## 18. CSS Work

The CSS was redesigned to support:

- Desktop layout
- Scrollable page after results appear
- Fixed-height first screen
- Cyber-futuristic theme
- Interactive SVG map
- Styled transmission controls
- Styled failure controls
- Centered tower inspector
- Tower status colors
- Result cards
- Conversion history tables
- Custom scrollbars
- Smaller desktop fallback

### Important Root Scrolling Rule

Do not use:

```css
html,
body,
#root {
  overflow: hidden;
}
```

Use vertical scrolling:

```css
html,
body {
  overflow-x: hidden;
  overflow-y: auto;
}
```

### First Screen Heights

```css
.map-card,
.controls {
  height: 650px;
}
```

Smaller desktop:

```css
.map-card,
.controls {
  height: 610px;
}
```

### Main Grid

```css
.primary-grid {
  display: grid;
  grid-template-columns: minmax(650px, 1fr) 340px;
  gap: 18px;
}
```

### Conversion History Classes

The consolidated CSS contains styles for:

```text
.conversion-history-card
.conversion-history
.conversion-hop
.conversion-hop-header
.conversion-stage-flow
.conversion-stage
.conversion-table
.conversion-sequences
.character-cell
```

Old tower-control-list styles are no longer needed if the component does not render that list.

---

## 19. Types to Verify

Inspect:

```text
frontend/src/types/index.ts
```

Important types:

```text
Snapshot
TransmissionResult
NetworkEvent
Route
HopLog
Latency
Planet
PlanetStatus
```

A previous strict type issue was fixed by using:

```ts
ring_direction: string;
```

instead of a narrow union when Wails generated a general string.

The next major type task is matching the frontend hop-log type to the exact backend fields.

---

## 20. Recommended Hop-Log Shape

A useful final target is:

```ts
interface HopLog {
  planet_id: string;
  previous_planet_id?: string;
  next_planet_id?: string;

  local_codex: number;
  next_hop_codex?: number;

  received_payload?: string[];
  decoded_payload: string;
  encoded_payload?: string[];

  binary_stream?: string;

  entry_tower?: number;
  exit_tower?: number;
  ring_direction?: string;

  planet_latency_ms?: number;
  void_latency_ms?: number;
}
```

This is a recommended target, not a confirmed exact backend shape.

---

## 21. Docker Context

The project is intended to support Docker and Docker Compose.

Verify actual service names and ports from:

```text
Dockerfile
docker-compose.yml
```

Likely responsibilities:

- Orchestrator service
- Planet services
- Shared configuration
- Health checks
- Docker network

Useful commands:

```bash
docker compose up --build
docker compose ps
docker compose logs -f
docker compose down
```

---

## 22. Running the Project

### Go Dependencies

```bash
go mod download
```

### Frontend Dependencies

```bash
cd frontend
npm install
cd ..
```

### Wails Development

```bash
wails dev
```

### Go Tests

```bash
go test ./...
```

### Frontend Build

```bash
cd frontend
npm run build
```

### Production Build

```bash
wails build
```

---

## 23. Bugs Fixed

### Corrupted TypeScript Interface

A malformed duplicate block caused:

```text
Unterminated string literal
```

Fix: remove the broken duplicate and keep one valid `TowerInspectorProps`.

### Duplicate Graph Declaration

Fix: keep graph types in `graph.go` and builder logic in `graph_builder.go`.

### Tower Popup Position

The popup originally followed the clicked planet.

Fix: center it inside the topology panel.

### Page Could Not Scroll

Root CSS used `overflow: hidden`.

Fix: allow vertical page scrolling so result cards appear below.

### Missing `result` Prop

Fix:

```ts
result?: TransmissionResult;
```

inside `TowerInspectorProps`.

---

## 24. Latest Requirements

1. Initial screen:
   - Universe topology on the left
   - Transmission section on the right

2. Result cards appear only after packet transmission.

3. The page becomes vertically scrollable after results appear.

4. Clicking a planet opens the tower inspector in the center of the topology.

5. Active towers are green.

6. Inactive towers are red.

7. The old tower-control list is replaced by translation details.

8. Delivery status includes a full history showing:
   - Received codex
   - ASCII values
   - Next-hop codex
   - Per-character conversions
   - Every planet in the route

9. The ordered hop log remains visible separately.

---

## 25. Recommended Next Tasks

### 1. Inspect a Real Transmission Result

Log one real value of:

```ts
result.packet.hop_log
```

Confirm exact field names and data types.

### 2. Finalize TypeScript Types

Create accurate interfaces for:

```text
TransmissionResult
Packet
HopLog
LatencyBreakdown
Route
NetworkEvent
```

Remove unnecessary `unknown` casts.

### 3. Fix Conversion History Against Real Data

Use backend-generated conversion values whenever available.

Keep frontend calculation only as a display fallback.

### 4. Verify Origin, Relay, and Destination Behavior

- Origin: text/ASCII → next codex
- Relay: received codex → ASCII → next codex
- Destination: received codex → ASCII → delivered text

### 5. Verify Failures

Test:

- Planet disabled
- Link disabled
- One tower disabled
- Multiple towers disabled
- Alternative route available
- No route available
- Reset restores everything

### 6. Inspect Docker Files

Confirm actual service names, ports, and startup flow.

### 7. Final Cleanup

- Remove unused CSS
- Remove unused imports
- Remove old popup-anchor types
- Run TypeScript build
- Run Go tests
- Run Wails build
- Verify Docker Compose
- Add screenshots
- Finalize public README

---

## 26. Prompt for the Next Chat

```text
I am continuing the AlchemiX Relic Ring Protocol project.

Read the attached PROJECT_HANDOFF_README.md first.

The project is a Go + Wails + React desktop application for the LAUNCH26 Relic Ring Protocol challenge.

Current priorities:

1. Inspect and correct the frontend TypeScript types based on the real Go transmission result.
2. Fix ConversionHistory.tsx to use the exact hop-log fields.
3. Make the delivery history show received codex → ASCII → next-hop codex for every planet.
4. Keep the tower inspector centered inside the universe topology.
5. Keep the initial topology/transmission layout fixed, with result cards appearing below after packet transmission.
6. Preserve node, link, and tower failure controls.
7. Avoid rewriting working backend logic unless an actual mismatch is found.

Ask me to upload:
- App.tsx
- styles.css
- types/index.ts
- TowerInspector.tsx
- ConversionHistory.tsx
- UniverseMap.tsx
- one sample StartTransmission response
- related Go packet/hop-log structs
```

---

## 27. Files to Upload in the Next Chat

```text
frontend/src/App.tsx
frontend/src/styles.css
frontend/src/types/index.ts
frontend/src/components/UniverseMap.tsx
frontend/src/components/TowerInspector.tsx
frontend/src/components/ConversionHistory.tsx
frontend/src/components/LatencyPanel.tsx
frontend/src/components/HopLogTable.tsx

app.go
main.go

internal/routing/*
internal/resilience/*
internal/geometry/*
internal/latency/*
internal/encoding/*
internal/packet/*
internal/service/*

universe-config.json
go.mod
wails.json
Dockerfile
docker-compose.yml
```

Also provide one actual output from:

```text
StartTransmission(...)
```

That response is the most important input for completing the frontend types and conversion history.

---

## 28. Current Status

### Completed or Mostly Completed

- Dynamic universe loading
- Backend package structure
- Routing graph
- Latency logic
- Codex translation concept
- Packet hop logging
- Planet failure state
- Link failure state
- Tower failure state
- Wails methods
- React desktop UI
- SVG universe topology
- Tower status display
- Centered tower inspector
- Scrollable post-transmission results
- Latency panel
- Delivery summary
- Live event stream
- Ordered hop log
- Conversion history design
- Consolidated CSS theme

### Still Needs Verification

- Exact backend hop-log schema
- Exact TypeScript interfaces
- Whether all conversion values come from backend data
- Docker service names and ports
- Final project structure
- End-to-end build after latest frontend changes
- Cleanup of old CSS selectors
- Complete automated tests
- Final screenshots and repository documentation

---

## 29. Development Principle

The backend should remain the source of truth for:

- Route selection
- Tower selection
- Latency
- Codex conversion
- Binary serialization
- Hop order
- Failure handling

The frontend should mainly:

- Display backend results
- Provide controls
- Visualize topology
- Show conversion history
- Show failure state
- Show delivery evidence

Avoid duplicating important routing or physics logic in React unless it is only a display fallback.

---

End of handoff.
