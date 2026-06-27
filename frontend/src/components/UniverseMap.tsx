import type { Route, Snapshot } from "../types";

interface UniverseMapProps {
  snapshot: Snapshot;
  route?: Route;
  selectedPlanetID?: string;
  onSelectPlanet?: (planetID: string) => void;
}

export function UniverseMap({
  snapshot,
  route,
  selectedPlanetID,
  onSelectPlanet,
}: UniverseMapProps) {
  const width = 760;
  const height = 470;
  const pad = 55;

  const xs = snapshot.planets.map((planet) => planet.x);
  const ys = snapshot.planets.map((planet) => planet.y);

  const minX = Math.min(...xs);
  const maxX = Math.max(...xs);
  const minY = Math.min(...ys);
  const maxY = Math.max(...ys);

  const point = (x: number, y: number) => ({
    x: pad + ((x - minX) / (maxX - minX || 1)) * (width - pad * 2),

    y: height - pad - ((y - minY) / (maxY - minY || 1)) * (height - pad * 2),
  });

  const status = new Map(
    snapshot.planet_status.map((planetStatus) => [
      planetStatus.id,
      planetStatus,
    ]),
  );

  const planetPoints = new Map(
    snapshot.planets.map((planet) => [planet.id, point(planet.x, planet.y)]),
  );

  const activeEdges = new Set<string>();

  route?.path.slice(0, -1).forEach((planetID, index) => {
    activeEdges.add([planetID, route.path[index + 1]].sort().join("::"));
  });

  const routePoints =
    route?.path
      .map((planetID) => planetPoints.get(planetID))
      .filter((planetPoint): planetPoint is { x: number; y: number } =>
        Boolean(planetPoint),
      ) ?? [];

  const packetXValues = routePoints
    .map((planetPoint) => planetPoint.x)
    .join(";");

  const packetYValues = routePoints
    .map((planetPoint) => planetPoint.y)
    .join(";");

  const packetKeyTimes = routePoints
    .map((_, index) => (index / Math.max(routePoints.length - 1, 1)).toFixed(4))
    .join(";");

  const packetDuration = Math.max(2.4, (routePoints.length - 1) * 1.15);

  const packetAnimationKey = route?.path.join("--") ?? "idle";

  return (
    <svg
      className="universe-map"
      viewBox={`0 0 ${width} ${height}`}
      role="img"
      aria-label="Zeta-26 universe map"
    >
      <defs>
        <filter id="glow">
          <feGaussianBlur stdDeviation="3" result="blur" />

          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>

        <filter id="packet-glow">
          <feGaussianBlur stdDeviation="5" result="packetBlur" />

          <feMerge>
            <feMergeNode in="packetBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>

      {snapshot.links.map((link) => {
        const firstPoint = planetPoints.get(link.a);
        const secondPoint = planetPoints.get(link.b);

        if (!firstPoint || !secondPoint) {
          return null;
        }

        const active = activeEdges.has([link.a, link.b].sort().join("::"));

        return (
          <line
            key={`${link.a}-${link.b}`}
            x1={firstPoint.x}
            y1={firstPoint.y}
            x2={secondPoint.x}
            y2={secondPoint.y}
            className={active ? "link active-link" : "link"}
          />
        );
      })}

      {snapshot.planets.map((planet, index) => {
        const planetPoint = planetPoints.get(planet.id);

        if (!planetPoint) {
          return null;
        }

        const available = status.get(planet.id)?.available ?? false;

        const inRoute = route?.path.includes(planet.id);
        const selected = selectedPlanetID === planet.id;

        const disabledTowerSet = new Set(
          snapshot.disabled_towers?.[planet.id] ?? [],
        );

        const towerRadius = inRoute ? 34 : 30;

        const towerIndexes = Array.from(
          { length: planet.active_towers },
          (_, towerIndex) => towerIndex,
        );

        return (
          <g
            key={planet.id}
            transform={`translate(${planetPoint.x},${planetPoint.y})`}
            className="planet-group"
            onClick={(event) => {
              event.stopPropagation();
              onSelectPlanet?.(planet.id);
            }}
            style={{
              cursor: onSelectPlanet ? "pointer" : "default",
            }}
          >
            <circle
              r={inRoute ? 21 : 17}
              className={[
                "planet",
                `planet-${index % 6}`,
                available ? "" : "failed",
                inRoute ? "route-planet" : "",
                selected ? "planet-selected" : "",
              ]
                .filter(Boolean)
                .join(" ")}
              filter={inRoute || selected ? "url(#glow)" : undefined}
            />

            <circle
              r={27}
              className="planet-ring"
              strokeDasharray={`${Math.max(4, planet.active_towers)} 5`}
            />

            {towerIndexes.map((towerIndex) => {
              const angle = (towerIndex / planet.active_towers) * Math.PI * 2;

              const towerX = Math.sin(angle) * towerRadius;
              const towerY = -Math.cos(angle) * towerRadius;

              const disabled = disabledTowerSet.has(towerIndex);

              return (
                <circle
                  key={towerIndex}
                  cx={towerX}
                  cy={towerY}
                  r={disabled ? 4.8 : 3.5}
                  className={disabled ? "tower-dot tower-broken" : "tower-dot"}
                >
                  <title>
                    {planet.id} Tower {towerIndex}
                    {disabled ? " inactive" : " active"}
                  </title>
                </circle>
              );
            })}

            <text y={43} textAnchor="middle" className="planet-name">
              {planet.id}
            </text>

            <text y={58} textAnchor="middle" className="planet-meta">
              B{planet.codex} · T{planet.active_towers}
            </text>
          </g>
        );
      })}

      {routePoints.length > 1 && (
        <g
          key={packetAnimationKey}
          className="packet-animation"
          aria-hidden="true"
        >
          <circle r="11" className="packet-wave" filter="url(#packet-glow)">
            <animate
              attributeName="cx"
              values={packetXValues}
              keyTimes={packetKeyTimes}
              dur={`${packetDuration}s`}
              calcMode="linear"
              repeatCount="indefinite"
            />

            <animate
              attributeName="cy"
              values={packetYValues}
              keyTimes={packetKeyTimes}
              dur={`${packetDuration}s`}
              calcMode="linear"
              repeatCount="indefinite"
            />

            <animate
              attributeName="r"
              values="7;13;7"
              dur="0.9s"
              repeatCount="indefinite"
            />
          </circle>

          <circle r="5.5" className="packet-runner">
            <animate
              attributeName="cx"
              values={packetXValues}
              keyTimes={packetKeyTimes}
              dur={`${packetDuration}s`}
              calcMode="linear"
              repeatCount="indefinite"
            />

            <animate
              attributeName="cy"
              values={packetYValues}
              keyTimes={packetKeyTimes}
              dur={`${packetDuration}s`}
              calcMode="linear"
              repeatCount="indefinite"
            />
          </circle>

          <circle r="3.2" className="packet-trail">
            <animate
              attributeName="cx"
              values={packetXValues}
              keyTimes={packetKeyTimes}
              dur={`${packetDuration}s`}
              begin="-0.16s"
              calcMode="linear"
              repeatCount="indefinite"
            />

            <animate
              attributeName="cy"
              values={packetYValues}
              keyTimes={packetKeyTimes}
              dur={`${packetDuration}s`}
              begin="-0.16s"
              calcMode="linear"
              repeatCount="indefinite"
            />
          </circle>

          <circle r="2" className="packet-trail packet-trail-far">
            <animate
              attributeName="cx"
              values={packetXValues}
              keyTimes={packetKeyTimes}
              dur={`${packetDuration}s`}
              begin="-0.3s"
              calcMode="linear"
              repeatCount="indefinite"
            />

            <animate
              attributeName="cy"
              values={packetYValues}
              keyTimes={packetKeyTimes}
              dur={`${packetDuration}s`}
              begin="-0.3s"
              calcMode="linear"
              repeatCount="indefinite"
            />
          </circle>
        </g>
      )}
    </svg>
  );
}
