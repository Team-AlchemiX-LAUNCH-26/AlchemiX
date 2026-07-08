# Geospatial Intelligence Dashboard — UI Design Template

A reusable UI specification extracted from the supplied dark geospatial analytics dashboard. It is suitable for React, TypeScript, Vite, Wails, Next.js, or a standard web application and can be adapted to the **AlchemiX Chimera Intelligence Dashboard**.

> The reference image is treated as visual direction only. Use original branding, map data, icons, copy, and chart assets in the final product.

---

## 1. Visual Direction

The interface combines:

- A dark charcoal application shell
- A compact left navigation rail
- A large interactive map as the primary canvas
- A contextual detail drawer attached to the map
- A timeline scrubber below the map
- Dense but orderly analytical cards
- Cyan and teal accents for active states and data highlights
- Thin borders, low-contrast separators, and restrained shadows
- Compact technical typography
- Minimal decorative effects so data remains the focus

The intended visual character is:

```text
operational + analytical + geographic + secure + real-time
```

Avoid bright gradients, oversized typography, heavy glassmorphism, and excessive neon. The design should feel like a serious monitoring console rather than a gaming HUD.

---

## 2. High-Level Layout

```text
┌──────────────────────────────────────────────────────────────────────┐
│ Brand / Product Name                                  Alerts  User  │
├──────────────┬───────────────────────────────────────────────────────┤
│              │                                                       │
│ Primary      │               Interactive Map                         │
│ Navigation   │                                      Detail Drawer    │
│              │                                                       │
│              ├───────────────────────────────────────────────────────┤
│              │                    Timeline                           │
│              ├──────────────────────┬────────────────────────────────┤
│              │ Overview             │ Frequency                      │
│              ├──────────────────────┼────────────────────────────────┤
│              │ Contributors / Logs  │ Distribution / Location        │
└──────────────┴──────────────────────┴────────────────────────────────┘
```

### Desktop proportions

- Top bar: `44–52px`
- Left navigation: `68–96px` collapsed or `150–190px` expanded
- Main map area: `45–55vh`
- Context drawer: `220–280px`
- Timeline strip: `48–64px`
- Analytics card grid: remaining viewport height with internal scrolling
- Application outer padding: `8–12px`
- Card gap: `8–12px`

### Recommended CSS grid

```css
.dashboard-shell {
  display: grid;
  grid-template-columns: 168px minmax(0, 1fr);
  grid-template-rows: 48px minmax(0, 1fr);
  min-height: 100vh;
}

.dashboard-main {
  display: grid;
  grid-template-rows: minmax(360px, 52vh) 56px auto;
  gap: 10px;
  min-width: 0;
}
```

---

## 3. Design Tokens

### 3.1 Color palette

```css
:root {
  --bg-app: #121414;
  --bg-topbar: #292b2b;
  --bg-sidebar: #181a1a;
  --bg-map: #111313;
  --bg-panel: #1a1c1c;
  --bg-panel-hover: #202323;
  --bg-panel-strong: #242727;

  --border-subtle: rgba(255, 255, 255, 0.055);
  --border-default: rgba(255, 255, 255, 0.09);
  --border-strong: rgba(255, 255, 255, 0.15);

  --text-primary: #eef3f1;
  --text-secondary: #a5adab;
  --text-muted: #69716f;
  --text-disabled: #484e4c;

  --accent-primary: #3bd6cc;
  --accent-bright: #55f2e7;
  --accent-deep: #159d98;
  --accent-soft: rgba(59, 214, 204, 0.13);
  --accent-glow: rgba(59, 214, 204, 0.32);

  --success: #48d597;
  --warning: #e3bc59;
  --danger: #e56c6c;
  --info: #5daeea;
}
```

### 3.2 Chart palette

```css
:root {
  --chart-1: #55e1d7;
  --chart-2: #119f99;
  --chart-3: #386f6d;
  --chart-4: #96a3a0;
  --chart-grid: rgba(255, 255, 255, 0.06);
  --chart-axis: rgba(255, 255, 255, 0.28);
}
```

Use no more than four colors in one chart. Teal should represent the primary or active metric. Gray tones should represent comparison or inactive values.

### 3.3 Radius, borders, and shadows

```css
:root {
  --radius-xs: 2px;
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-pill: 999px;

  --panel-border: 1px solid var(--border-default);
  --panel-shadow: 0 10px 30px rgba(0, 0, 0, 0.22);
  --focus-ring: 0 0 0 2px rgba(59, 214, 204, 0.35);
}
```

The reference relies on crisp rectangular surfaces. Keep radii small and avoid highly rounded cards.

### 3.4 Spacing scale

```css
:root {
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 20px;
  --space-6: 24px;
  --space-7: 32px;
}
```

---

## 4. Typography

### Recommended font pairing

- UI and body: `Inter`, `IBM Plex Sans`, or `Manrope`
- Technical values: `IBM Plex Mono`, `JetBrains Mono`, or `Roboto Mono`
- Display labels: `Space Grotesk` or `Inter Tight`

### Type hierarchy

```css
.app-brand {
  font: 600 0.78rem/1 "Inter", sans-serif;
  letter-spacing: -0.01em;
}

.nav-item {
  font: 500 0.68rem/1.2 "Inter", sans-serif;
  color: var(--text-secondary);
}

.panel-title {
  font: 600 0.66rem/1.2 "Inter", sans-serif;
  letter-spacing: 0.01em;
  color: var(--text-primary);
}

.metric-value {
  font: 400 clamp(1.8rem, 3vw, 3.4rem)/0.95 "Inter Tight", sans-serif;
  letter-spacing: -0.05em;
}

.metric-label,
.axis-label,
.map-label {
  font: 500 0.56rem/1.3 "IBM Plex Mono", monospace;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
```

Keep text sizes compact, but never below `11px` for interactive content.

---

## 5. Application Anatomy

## 5.1 Top Bar

The top bar includes:

- Compact product mark and application name
- Optional environment indicator
- Settings icon
- Notification icon
- User or account control

Suggested structure:

```text
TopBar
├── BrandMark
├── ProductName
├── EnvironmentBadge
└── TopBarActions
    ├── SettingsButton
    ├── NotificationButton
    └── UserMenu
```

Style rules:

- Height around `48px`
- Darker or slightly lighter than the main background
- Minimal bottom border
- Icons sized `14–16px`
- No large search field unless required by the product

---

## 5.2 Left Navigation

The left navigation acts as the primary mode switch.

Typical items:

```text
Overview
Realtime
Analytics
Traffic
Events
Systems
Reports
Admin
```

### Active state

- Teal text or icon
- Narrow teal indicator line on the left
- Optional soft teal background

```css
.nav-link[aria-current="page"] {
  color: var(--accent-bright);
  background: linear-gradient(
    90deg,
    rgba(59, 214, 204, 0.11),
    transparent
  );
  border-left: 2px solid var(--accent-primary);
}
```

### Navigation behavior

- Expanded on wide desktop
- Collapsible to icon-only mode
- Becomes an off-canvas drawer below `900px`
- Preserve labels through tooltips in collapsed mode

---

## 5.3 Main Map Canvas

The map is the visual focus of the dashboard.

Required layers:

1. Base geography or network topology
2. Region boundaries
3. Active location markers
4. Soft radial activity heat zones
5. Selected location marker
6. Selected location label
7. Map controls
8. Context drawer

### Marker styles

```css
.map-marker {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent-bright);
  box-shadow:
    0 0 0 6px rgba(59, 214, 204, 0.09),
    0 0 18px var(--accent-glow);
}

.map-marker--selected {
  width: 10px;
  height: 10px;
  border: 2px solid #e8fffc;
  box-shadow:
    0 0 0 7px rgba(59, 214, 204, 0.14),
    0 0 24px rgba(59, 214, 204, 0.45);
}
```

### Map controls

Position map controls in a small vertical toolbar:

```text
Locate
Zoom In
Zoom Out
Reset View
Layers
```

Controls should use square buttons, subtle borders, and clear hover states.

### AlchemiX adaptation

Replace world geography with the Zeta-26 network topology:

- Planets become map nodes
- Interplanetary links become edges
- Teal heat zones represent traffic load or targeting probability
- Selected planet or link opens the intelligence drawer
- Link color reflects combined risk or trust
- Animated pulses show current packet transmissions

---

## 5.4 Context Detail Drawer

The drawer in the reference is visually attached to the right side of the map.

Suggested contents:

```text
Region / Link Type
Selected Entity Name
Short Description
Primary Metric
Secondary Metric
Mini Trend Chart
Close Button
```

Example for AlchemiX:

```text
INTERPLANETARY LINK
AEGIS–ELYSIUM
Live intelligence summary for the selected route segment.

TRUST SCORE
0.91

COMBINED COST
44.2 ms

[Mini latency trend]
```

### Drawer layout

```css
.map-drawer {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: clamp(220px, 24vw, 290px);
  padding: 20px;
  background: rgba(27, 29, 29, 0.96);
  border-left: 1px solid var(--border-default);
  box-shadow: -18px 0 40px rgba(0, 0, 0, 0.22);
}
```

### Mobile behavior

Below `720px`, convert the drawer into a bottom sheet rather than overlaying the full map width.

---

## 5.5 Timeline Scrubber

The timeline provides historical navigation or tick selection.

Recommended elements:

- Start label
- End label
- Current tick or year marker
- Selected range track
- Two draggable handles
- Optional play and pause control

```text
1996 ─────────●════════════════●───────── 2006
               2001          2003
```

### AlchemiX usage

Use the timeline to select:

- Historical training ticks
- Live API tick
- Incident windows
- Packet lifecycle stages
- Before and after reroute comparison

### Accessibility

- Handles must be keyboard adjustable
- Show current values in visible labels
- Expose `aria-valuemin`, `aria-valuemax`, and `aria-valuenow`

---

## 5.6 Analytical Card Grid

Use a modular 12-column grid.

```css
.analytics-grid {
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: 10px;
}

.card-overview { grid-column: span 4; }
.card-frequency { grid-column: span 4; }
.card-breakdown { grid-column: span 4; }
.card-table { grid-column: span 8; }
.card-location { grid-column: span 4; }
```

Recommended minimum height:

- Metric card: `160–210px`
- Table card: `180–260px`
- Chart card: `180–220px`

---

## 6. Card Types

## 6.1 Radial KPI Card

The reference uses three compact circular progress gauges.

Recommended usage:

```text
Average Trust
Congestion Safety
Route Entropy
```

Each gauge includes:

- Circular progress ring
- Percentage or score in the center
- Compact label under the ring
- Optional legend or comparison note

### Implementation rule

Use SVG rather than CSS borders for precise arcs.

```tsx
interface RadialMetricProps {
  label: string;
  value: number;
  max?: number;
  suffix?: string;
  status?: "safe" | "warning" | "danger";
}
```

---

## 6.2 Horizontal Distribution Card

Use horizontal bars for categories that must be compared quickly.

Example:

```text
Aegis–Elysium       52%
Boreas–Dawn         43%
Caelum–Fenix        29%
Dawn–Fenix          18%
```

Design rules:

- Category labels on the left
- Values aligned right
- Teal fill over a muted track
- Consistent value scale
- Tooltip on hover or focus

---

## 6.3 Donut Breakdown Card

Use a donut chart for three or four mutually exclusive groups.

Possible AlchemiX categories:

```text
Safe
Elevated Risk
Critical
Unavailable
```

Do not use donut charts for precise comparison between many categories. Provide a textual legend with values.

---

## 6.4 Mini Trend Chart

The mini trend chart belongs in context drawers and summary cards.

Requirements:

- No visible heavy axes
- One main line
- Optional subtle area fill
- Latest value marker
- Tooltip on pointer or keyboard focus

Possible metrics:

- Observed latency
- Predicted penalty
- Trust score
- Traffic share
- Targeting risk

---

## 6.5 Activity or Contributor Table

The reference includes a compact list/table across the lower area.

Suggested columns:

```text
Entity
Event
Status
Time
Action
```

AlchemiX example:

```text
Link             Decision                  Status      Tick
Aegis–Elysium    Retained in route         Safe        1000042
Dawn–Fenix       Avoided due to spoofing   Elevated    1000042
Caelum–Fenix     Marked saturated          Blocked     1000042
```

### Table rules

- Sticky header for long tables
- Zebra striping should be extremely subtle
- Right-align numeric values
- Use status dots plus text, not color alone
- Row hover reveals actions

---

## 6.6 Geographic or Topology Heatmap Card

A compact map card can summarize location concentration.

AlchemiX adaptation:

- Show the universe topology miniature
- Highlight route usage by link
- Display top-risk regions or links
- Add a legend for low, medium, and high exposure

---

## 7. Panel Header Pattern

Every card should share one header pattern:

```text
Panel Title                         Favorite  Pin  More
```

Recommended actions:

- Favorite
- Pin
- Expand
- Export
- More menu

```tsx
interface PanelHeaderProps {
  title: string;
  subtitle?: string;
  actions?: Array<"favorite" | "pin" | "expand" | "export" | "more">;
}
```

Action icons should only appear on hover for low-priority controls, while important actions remain visible.

---

## 8. Interaction Model

### Hover

- Increase border contrast
- Raise panel background slightly
- Reveal secondary controls
- Show chart or map tooltip

### Selection

- Teal outline or left indicator
- Soft teal background
- Clear selected label
- Synchronize map, drawer, cards, and table

### Loading

Use skeletons shaped like actual content:

- Map block skeleton
- Circular gauge skeletons
- Bar skeletons
- Table row skeletons

Do not replace the full dashboard with a centered spinner.

### Empty state

```text
No telemetry available for this tick.
Choose another time window or refresh the live feed.
```

### Error state

Show:

- What failed
- Whether stale data is being displayed
- Retry action
- Last successful update time

---

## 9. Motion Guidelines

Use motion to indicate state change, not decoration.

Recommended durations:

```css
:root {
  --motion-fast: 120ms;
  --motion-standard: 180ms;
  --motion-slow: 280ms;
}
```

Suggested animations:

- Drawer slide: `180–240ms`
- Card hover: `120ms`
- Chart update: `250–400ms`
- Map marker pulse: `1.8–2.4s`, only for live activity
- Timeline handle movement: immediate while dragging

Respect reduced motion:

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## 10. Responsive Behavior

### Large desktop: `≥ 1280px`

- Expanded left navigation
- Map and context drawer visible together
- Three cards per analytics row
- Full table columns

### Compact desktop/tablet: `900–1279px`

- Collapsed icon navigation
- Drawer width reduced
- Analytics cards use two-column layout
- Lower-priority table columns hidden

### Mobile/tablet portrait: `< 900px`

- Navigation becomes drawer
- Map occupies full width
- Detail drawer becomes bottom sheet
- Timeline remains horizontally scrollable if needed
- Cards become one column or a two-column compact grid

```css
@media (max-width: 900px) {
  .dashboard-shell {
    grid-template-columns: 1fr;
  }

  .analytics-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .analytics-grid > * {
    grid-column: span 1;
  }
}

@media (max-width: 640px) {
  .analytics-grid {
    grid-template-columns: 1fr;
  }
}
```

---

## 11. Accessibility Requirements

- Maintain at least WCAG AA contrast for body text and controls
- Do not communicate state through color alone
- Add visible focus rings to all interactive controls
- Provide keyboard navigation for map markers, timeline handles, tabs, and menus
- Include textual summaries for every chart
- Add `aria-live="polite"` to live telemetry updates that matter to the user
- Avoid updating values so frequently that screen-reader users lose context
- Ensure hit areas are at least `40 × 40px` on touch devices
- Give icon-only controls an accessible name

Example:

```tsx
<button aria-label="Expand congestion chart">
  <ExpandIcon aria-hidden="true" />
</button>
```

---

## 12. Recommended Component Architecture

```text
src/
├── app/
│   ├── DashboardPage.tsx
│   └── dashboard.css
├── components/
│   ├── shell/
│   │   ├── TopBar.tsx
│   │   ├── SideNavigation.tsx
│   │   └── AppShell.tsx
│   ├── map/
│   │   ├── IntelligenceMap.tsx
│   │   ├── MapMarker.tsx
│   │   ├── MapControls.tsx
│   │   ├── MapLegend.tsx
│   │   └── EntityDetailDrawer.tsx
│   ├── timeline/
│   │   └── TimelineScrubber.tsx
│   ├── panels/
│   │   ├── AnalyticsPanel.tsx
│   │   ├── PanelHeader.tsx
│   │   └── PanelEmptyState.tsx
│   ├── charts/
│   │   ├── RadialMetric.tsx
│   │   ├── HorizontalBars.tsx
│   │   ├── DonutBreakdown.tsx
│   │   ├── MiniTrendChart.tsx
│   │   └── TopologyHeatmap.tsx
│   └── tables/
│       └── ActivityTable.tsx
├── hooks/
│   ├── useDashboardFilters.ts
│   ├── useLiveTelemetry.ts
│   └── useSelectedEntity.ts
├── services/
│   └── telemetryApi.ts
├── types/
│   └── dashboard.ts
└── styles/
    ├── tokens.css
    ├── globals.css
    └── utilities.css
```

---

## 13. Suggested TypeScript Types

```ts
export interface DashboardEntity {
  id: string;
  name: string;
  type: "planet" | "link" | "region";
  status: "normal" | "warning" | "critical" | "offline";
}

export interface LinkIntelligence {
  linkId: string;
  source: string;
  destination: string;
  trustScore: number;
  congestionPenaltyMs: number;
  targetingRiskScore: number;
  combinedCost: number;
  trafficShare: number;
  status: "ok" | "saturated";
}

export interface TimelineRange {
  startTick: number;
  endTick: number;
  selectedTick: number;
}

export interface MetricDatum {
  label: string;
  value: number;
  unit?: string;
}
```

---

## 14. Example Dashboard Composition

```tsx
export function IntelligenceDashboard() {
  return (
    <AppShell>
      <TopBar />
      <SideNavigation />

      <main className="dashboard-main">
        <section className="map-section">
          <IntelligenceMap />
          <EntityDetailDrawer />
        </section>

        <TimelineScrubber />

        <section className="analytics-grid">
          <AnalyticsPanel title="Overview" className="card-overview">
            <RadialMetric label="Trust" value={91} suffix="%" />
            <RadialMetric label="Safety" value={84} suffix="%" />
            <RadialMetric label="Entropy" value={67} suffix="%" />
          </AnalyticsPanel>

          <AnalyticsPanel title="Link Frequency" className="card-frequency">
            <HorizontalBars data={[]} />
          </AnalyticsPanel>

          <AnalyticsPanel title="Risk Breakdown" className="card-breakdown">
            <DonutBreakdown data={[]} />
          </AnalyticsPanel>

          <AnalyticsPanel title="Decision Activity" className="card-table">
            <ActivityTable rows={[]} />
          </AnalyticsPanel>

          <AnalyticsPanel title="Network Exposure" className="card-location">
            <TopologyHeatmap data={[]} />
          </AnalyticsPanel>
        </section>
      </main>
    </AppShell>
  );
}
```

---

## 15. AlchemiX Dashboard Mapping

The visual template can map directly to the current Phase 2 requirements:

| Reference UI Element | AlchemiX Use |
|---|---|
| World map | Zeta-26 planet and link topology |
| Geographic marker | Planet or link selection |
| Regional detail drawer | Link evaluation and explanation |
| Timeline | Historical tick or live packet timeline |
| Circular overview metrics | Trust, congestion safety, route entropy |
| Frequency bars | Link utilization or route-selection frequency |
| Donut breakdown | Safe, risky, saturated, spoofed link distribution |
| Contributor table | Agent decision audit trail |
| Primary location card | Highest-risk network region or link cluster |

The backend should remain the source of truth for route choice, model output, link scoring, and decision explanations. The frontend should focus on synchronized visualization and inspection.

---

## 16. AlchemiX Page Layout Proposal

```text
┌──────────────────────────────────────────────────────────────────────┐
│ ALCHEMIX · CHIMERA INTELLIGENCE              LIVE · TICK 1000042   │
├──────────────┬───────────────────────────────────────────────────────┤
│ Overview     │                                                       │
│ Live State   │          ZETA-26 NETWORK TOPOLOGY                     │
│ Decisions    │                                      LINK DETAIL      │
│ Incidents    │                                                       │
│ Models       ├───────────────────────────────────────────────────────┤
│ Audit        │          HISTORICAL / LIVE TICK SCRUBBER              │
│ Settings     ├──────────────────────┬────────────────────────────────┤
│              │ Trust / Safety /     │ Link Selection Frequency       │
│              │ Route Entropy        │                                │
│              ├──────────────────────┼────────────────────────────────┤
│              │ Decision Audit Trail │ Link Risk Distribution         │
└──────────────┴──────────────────────┴────────────────────────────────┘
```

---

## 17. Implementation Notes

### Map library options

- Use SVG for a fixed six-planet topology and maximum visual control
- Use React Flow when node dragging, handles, and graph interactions are needed
- Use MapLibre only when real geographic maps are part of the product
- Avoid adding a heavy map dependency for a small static universe graph

### Chart library options

- Use Recharts for fast React integration
- Use ECharts for dense real-time analytics and advanced interactions
- Use D3 only when custom chart behavior is essential
- For the current compact dashboard, SVG components plus Recharts are enough

### Performance

- Memoize chart data transformations
- Debounce timeline changes
- Virtualize long event tables
- Update only changed map nodes or links
- Avoid re-rendering every panel on each live telemetry tick
- Pause marker animations when the page is hidden

---

## 18. Quality Checklist

### Visual

- [ ] Dark neutral surfaces are clearly layered
- [ ] Teal is reserved for active data and selection
- [ ] Cards share consistent headers and spacing
- [ ] Map remains the dominant visual element
- [ ] Chart labels remain readable at actual window size
- [ ] No oversized shadows or glossy gradients

### Interaction

- [ ] Selecting a map entity updates the drawer and cards
- [ ] Timeline changes update all dependent panels
- [ ] Hover and focus states are consistent
- [ ] Loading, empty, stale, and error states are implemented
- [ ] Live updates do not unexpectedly reset user selection

### Accessibility

- [ ] Full keyboard navigation works
- [ ] Charts include textual summaries
- [ ] Status is not indicated by color alone
- [ ] Icon buttons have accessible labels
- [ ] Reduced-motion preferences are respected

### Engineering

- [ ] Design tokens are centralized
- [ ] Dashboard types match backend responses
- [ ] Large tables are virtualized or paginated
- [ ] Components are reusable and data-agnostic
- [ ] Real-time updates are isolated in a dedicated hook or service

---

## 19. Final Design Principle

The dashboard should answer four questions immediately:

1. **What is happening now?**
2. **Where is it happening?**
3. **How severe or trustworthy is it?**
4. **What action did the system take, and why?**

Every map layer, card, table, and interaction should support one of those questions. Keep the interface calm, synchronized, and explainable—even when the underlying network is under pressure.
