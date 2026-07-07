# Futuristic AI Landing Page — UI Design Template

A reusable design specification inspired by the supplied dark, cinematic AI interface. The template is intended for a React, TypeScript, Vite, Wails, or standard web frontend and can be adapted to the **AlchemiX** dashboard.

> The reference is treated as visual direction only. Use original artwork, icons, copy, and 3D assets in the final product.

---

## 1. Design Direction

The interface combines:

- A near-black cinematic canvas
- A split-screen hero layout
- Electric blue highlights
- Thin technical grid lines
- Large geometric typography
- A dark AI or robotic focal visual
- Minimal navigation and compact metadata
- Small labels, coordinates, and system annotations
- Subtle glow, grain, and vignette effects

The overall feeling should be:

```text
premium + intelligent + technical + mysterious + controlled
```

Avoid turning the page into a bright neon gaming UI. Most of the screen should remain dark, with blue used only to direct attention.

---

## 2. High-Level Layout

```text
┌─────────────────────────────────────────────────────────────────────┐
│ Logo        Home  About  Work  Services  Contact       Status/Menu │
├───────────────────────────────┬─────────────────────────────────────┤
│                               │                                     │
│ Eyebrow label                 │       Main AI / 3D visual           │
│ Large hero headline           │                                     │
│ Supporting copy               │       Technical annotations         │
│ Primary CTA                   │                                     │
│                               │                                     │
├───────────────────────────────┼─────────────────────────────────────┤
│ Optional telemetry / index    │ Feature 01            Feature 02    │
└───────────────────────────────┴─────────────────────────────────────┘
```

### Desktop proportions

- Header: `64–76px`
- Main content: `calc(100vh - header)`
- Hero split: approximately `52% / 48%`
- Content max-width: `1440px`
- Outer page padding: `24–36px`
- Primary content alignment: left-middle
- Hero visual alignment: center-right

### Recommended grid

```css
.hero-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(420px, 0.92fr);
  min-height: calc(100vh - 72px);
}
```

---

## 3. Design Tokens

### Color palette

```css
:root {
  --color-bg: #070811;
  --color-bg-elevated: #0c0e18;
  --color-panel: #10131f;
  --color-panel-soft: #151827;
  --color-line: rgba(172, 181, 214, 0.12);
  --color-line-strong: rgba(172, 181, 214, 0.22);

  --color-text: #f3f5ff;
  --color-text-muted: #8d93a8;
  --color-text-dim: #5f667c;

  --color-accent: #2647ff;
  --color-accent-bright: #4d63ff;
  --color-accent-soft: rgba(38, 71, 255, 0.2);
  --color-accent-glow: rgba(38, 71, 255, 0.42);

  --color-success: #5ee6a8;
  --color-warning: #ffc45c;
  --color-danger: #ff667d;
}
```

### Surface treatment

```css
:root {
  --panel-border: 1px solid rgba(255, 255, 255, 0.07);
  --panel-shadow: 0 28px 90px rgba(0, 0, 0, 0.5);
  --accent-shadow: 0 0 40px rgba(38, 71, 255, 0.24);
  --page-radius: 20px;
  --control-radius: 4px;
}
```

### Spacing scale

```css
:root {
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 24px;
  --space-6: 32px;
  --space-7: 48px;
  --space-8: 64px;
  --space-9: 96px;
}
```

---

## 4. Typography

### Recommended type pairing

- Display: `Oxanium`, `Orbitron`, or `Space Grotesk`
- Body and UI: `Inter`, `Manrope`, or `IBM Plex Sans`
- Technical metadata: `IBM Plex Mono`, `JetBrains Mono`, or `Space Mono`

### Type hierarchy

```css
.hero-eyebrow {
  font: 600 0.72rem/1.2 "IBM Plex Mono", monospace;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.hero-title {
  font: 300 clamp(3.6rem, 8vw, 8rem)/0.86 "Space Grotesk", sans-serif;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.hero-copy {
  max-width: 38rem;
  font: 400 0.95rem/1.75 "Inter", sans-serif;
  color: var(--color-text-muted);
}

.micro-label {
  font: 500 0.62rem/1.35 "IBM Plex Mono", monospace;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--color-text-dim);
}
```

### Headline treatment

Use one blue block or letter behind the first part of the headline. Keep the rest of the title white or cool gray.

Example:

```text
[ A ] LCHEMIX
    INTELLIGENCE
```

Do not apply glow to the entire heading. Restrict glow to the accent block, cursor, active nav item, or a single keyword.

---

## 5. Page Anatomy

### 5.1 Application shell

Responsibilities:

- Provide the rounded outer frame
- Hold the dark background
- Clip decorative effects
- Apply the global grid
- Maintain a minimum desktop height

```css
.app-shell {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  color: var(--color-text);
  background:
    radial-gradient(circle at 68% 40%, rgba(44, 55, 107, 0.16), transparent 36%),
    linear-gradient(120deg, #090a13 0%, #080912 45%, #06070d 100%);
}
```

### 5.2 Header

Content:

- Wordmark or compact logo
- Five or fewer primary navigation links
- Language, system status, or connection indicator
- Small menu button on narrow layouts

Style rules:

- Keep the header visually quiet
- Use uppercase micro typography
- Add a faint bottom border
- Highlight only the active route
- Avoid large buttons in the header

Suggested component:

```text
Header
├── BrandMark
├── PrimaryNav
├── SystemStatus
└── MenuButton
```

### 5.3 Left hero content

Content order:

1. Eyebrow or section code
2. Main title
3. Supporting paragraph
4. Primary CTA
5. Optional scroll or section index indicator

Suggested AlchemiX copy structure:

```text
ADAPTIVE NETWORK INTELLIGENCE
ALCHEMIX
Predict, verify, and reroute communications before Chimera can strike.
[ Initialize Agent ]
```

### 5.4 Right visual stage

Use one original focal asset:

- 3D robotic bust
- Abstract AI core
- Holographic planetary network
- Neural mesh sculpture
- Wireframe helmet

Layering order:

```text
background glow
→ grid / scan lines
→ main visual
→ annotation lines
→ labels
→ foreground vignette
```

The asset should occupy about `70–90%` of the visual column height and may extend slightly outside its grid cell.

### 5.5 Bottom feature strip

Use two or three small feature panels, for example:

- Predictive Routing
- Telemetry Trust
- Targeting Risk

Each panel contains:

- Tiny uppercase title
- One-line description
- Optional numeric signal or status

Avoid full cards with large shadows. Use borders and spacing to create separation.

### 5.6 Side rail

A compact right-side rail may contain:

- Social links
- Section indicators
- Current slide number
- Vertical navigation controls

Keep icons at `14–16px` and reduce opacity until hover.

---

## 6. Decorative System

### Technical grid

```css
.technical-grid {
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.035) 1px, transparent 1px);
  background-size: 64px 64px;
}
```

### Fine divider lines

Use low-opacity borders to divide the screen into architectural zones. Lines should support alignment rather than becoming decoration by themselves.

### Noise layer

```css
.noise-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  opacity: 0.025;
  mix-blend-mode: screen;
}
```

Use a locally owned noise texture or a CSS-generated alternative.

### Vignette

```css
.vignette {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: radial-gradient(circle, transparent 48%, rgba(0, 0, 0, 0.62) 100%);
}
```

### Accent block

```css
.accent-block {
  display: inline-grid;
  place-items: center;
  min-width: 0.72em;
  background: var(--color-accent);
  box-shadow: var(--accent-shadow);
}
```

---

## 7. Core Components

```text
src/
├── components/
│   ├── AppShell.tsx
│   ├── Header.tsx
│   ├── BrandMark.tsx
│   ├── HeroContent.tsx
│   ├── HeroVisual.tsx
│   ├── TechnicalAnnotation.tsx
│   ├── FeatureStrip.tsx
│   ├── FeatureItem.tsx
│   ├── SideRail.tsx
│   └── PrimaryButton.tsx
├── sections/
│   └── LandingHero.tsx
├── styles/
│   ├── tokens.css
│   ├── globals.css
│   ├── effects.css
│   └── landing.css
└── assets/
    ├── ai-visual.webp
    ├── noise.webp
    └── logo-mark.svg
```

### Component contract example

```ts
export interface FeatureItemProps {
  title: string;
  description: string;
  value?: string;
  status?: "normal" | "warning" | "critical";
}
```

---

## 8. Interaction Design

### Navigation

- Active link uses full text opacity and a small accent marker
- Hover state increases opacity rather than changing layout
- Keyboard focus must be clearly visible

### Primary CTA

- Thin outline or dark translucent fill
- Small arrow or corner marker
- Blue glow only on hover or focus
- Transition duration: `180–240ms`

```css
.primary-cta {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-height: 44px;
  padding: 0 20px;
  border: 1px solid var(--color-line-strong);
  background: rgba(255, 255, 255, 0.01);
  color: var(--color-text);
  text-transform: uppercase;
  letter-spacing: 0.12em;
  transition: border-color 200ms ease, box-shadow 200ms ease, transform 200ms ease;
}

.primary-cta:hover {
  transform: translateY(-1px);
  border-color: rgba(77, 99, 255, 0.72);
  box-shadow: 0 0 28px rgba(38, 71, 255, 0.18);
}
```

### Motion system

Recommended motions:

- Slow visual float: `6–10s`
- Grid or scan-line drift: `10–18s`
- Annotation fade-in: `400–700ms`
- Hero text reveal: staggered by `60–100ms`
- Button hover: under `240ms`

Avoid:

- Constant flashing
- Strong parallax on every element
- Fast rotations
- Heavy motion on text

Respect reduced-motion preferences:

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    scroll-behavior: auto !important;
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## 9. Responsive Behaviour

### Large desktop: `≥ 1280px`

- Maintain two-column hero
- Show all annotations
- Keep feature strip horizontal
- Use full navigation

### Tablet: `768px–1279px`

- Reduce headline size
- Use approximately `55% / 45%` columns
- Hide non-essential annotations
- Keep two feature items per row

### Mobile: `< 768px`

- Convert to a single column
- Place hero copy before the visual
- Use a compact menu button
- Keep the visual at `320–480px` high
- Allow feature cards to stack
- Remove decorative side rail
- Reduce grid density

```css
@media (max-width: 767px) {
  .hero-grid {
    grid-template-columns: 1fr;
  }

  .hero-content {
    padding: 72px 24px 40px;
  }

  .hero-visual {
    min-height: 420px;
  }
}
```

---

## 10. Accessibility Requirements

- Body text contrast should meet WCAG AA
- Do not place critical information only in blue or red
- Give the primary visual an empty `alt` attribute when decorative
- Give informative diagrams descriptive alternative text
- Maintain a minimum interactive size of `44 × 44px`
- Provide visible focus states
- Keep body text at `16px` where practical
- Do not rely on thin low-contrast lines for essential boundaries
- All animations must respect `prefers-reduced-motion`

---

## 11. AlchemiX Adaptation

The visual language maps naturally to the AlchemiX Chimera-defense interface.

### Suggested content mapping

| Reference region | AlchemiX usage |
|---|---|
| Top-left wordmark | AlchemiX logo |
| Primary navigation | Overview, Network, Agent, Audit, Models |
| Hero eyebrow | Adaptive Interplanetary Intelligence |
| Main headline | Predict. Verify. Reroute. |
| Right-side visual | 3D network core or animated universe topology |
| Technical labels | Congestion, Trust, Targeting Risk |
| Bottom feature strip | Model health and current confidence |
| Top-right status | API connected, tick number, agent state |

### Suggested first-screen composition

```text
LEFT
- Agent mission statement
- Natural-language transmission input
- Evaluate / send action

RIGHT
- Zeta-26 network visual
- Current active route
- Chimera risk annotations

BOTTOM
- Congestion model confidence
- Telemetry trust health
- Route entropy / targeting risk
```

### Data-driven status colors

```text
Normal      → muted white / green
Attention   → amber
High risk   → red
Selected    → electric blue
Unavailable → dim gray with strike-through or disabled pattern
```

---

## 12. Content Guidelines

Keep copy short and technical.

Good:

```text
TELEMETRY VERIFIED
Trust confidence: 91%
```

Avoid:

```text
Our advanced artificial intelligence system has verified all available
telemetry data and determined that the connection is probably safe.
```

Use active, precise verbs:

```text
Predict
Verify
Route
Override
Audit
Transmit
```

---

## 13. Visual Asset Guidance

Use original or properly licensed assets.

Recommended formats:

- `WebP` or `AVIF` for raster visuals
- `SVG` for logos and line annotations
- `GLB/GLTF` for optional interactive 3D
- Compressed MP4/WebM for background loops

Performance targets:

- Hero image: preferably under `500 KB`
- Initial page JavaScript: keep as small as practical
- Defer heavy 3D until after first content paint
- Provide a static fallback image for low-power devices

---

## 14. Acceptance Checklist

### Visual

- [ ] Dark cinematic foundation
- [ ] One dominant blue accent
- [ ] Large geometric hero heading
- [ ] Clear split-screen composition
- [ ] Original AI or network focal visual
- [ ] Fine technical grid and annotation details
- [ ] Restrained glow effects
- [ ] Consistent small uppercase labels

### UX

- [ ] Primary action is immediately visible
- [ ] Navigation remains readable
- [ ] Content is usable without animation
- [ ] Mobile layout stacks cleanly
- [ ] Focus states are visible
- [ ] Text remains readable against the visual

### Engineering

- [ ] Tokens stored centrally
- [ ] Components are reusable
- [ ] Assets are optimized
- [ ] Decorative elements do not block pointer events
- [ ] Reduced-motion mode works
- [ ] Layout remains stable during asset loading

---

## 15. Final Design Principle

The interface should feel advanced because of its **hierarchy, spacing, typography, and controlled data presentation**, not because every element glows.

Use darkness as the primary material, blue as the signal, and the hero visual as the emotional focal point.
