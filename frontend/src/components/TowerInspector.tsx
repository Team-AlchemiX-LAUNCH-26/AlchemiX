import type { Snapshot,TransmissionResult,} from "../types";

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

export function TowerInspector({
  snapshot,
  selectedPlanetID,
  onClose,
  onToggleTower,
  busy = false,
}: TowerInspectorProps) {
  if (!selectedPlanetID) {
    return null;
  }

  const selectedPlanet = snapshot.planets.find(
    (planet) => planet.id === selectedPlanetID
  );

  if (!selectedPlanet) {
    return null;
  }

  const disabledSet = new Set(
    snapshot.disabled_towers?.[selectedPlanet.id] ?? []
  );

  const towerIndexes = Array.from(
    { length: selectedPlanet.active_towers },
    (_, index) => index
  );

  return (
    <div
      className="tower-popover"
      onClick={(event) => event.stopPropagation()}
    >
      <div className="tower-popover-header">
        <div>
          <span className="tower-popover-kicker">
            Signal Tower Network
          </span>

          <h3>{selectedPlanet.id}</h3>

          <p>
            <span className="popup-active-count">
              {selectedPlanet.active_towers - disabledSet.size} active
            </span>

            {" · "}

            <span className="popup-inactive-count">
              {disabledSet.size} inactive
            </span>
          </p>
        </div>

        <button
          type="button"
          className="icon-button"
          onClick={onClose}
          aria-label="Close tower inspector"
        >
          ×
        </button>
      </div>

      <div className="tower-legend">
        <span className="tower-legend-active">
          <i className="tower-status-dot active" />
          Active
        </span>

        <span className="tower-legend-inactive">
          <i className="tower-status-dot inactive" />
          Inactive
        </span>
      </div>

      <div className="tower-map-preview">
        <div className="tower-planet-core">
          <strong>{selectedPlanet.id}</strong>
          <small>Base {selectedPlanet.codex}</small>
        </div>

        {towerIndexes.map((towerIndex) => {
          const angle =
            (towerIndex / selectedPlanet.active_towers) *
              Math.PI *
              2 -
            Math.PI / 2;

          const radius = 43;
          const x = 50 + Math.cos(angle) * radius;
          const y = 50 + Math.sin(angle) * radius;

          const disabled = disabledSet.has(towerIndex);

          return (
            <button
              type="button"
              key={towerIndex}
              disabled={busy}
              className={
                disabled
                  ? "tower-map-node inactive"
                  : "tower-map-node active"
              }
              style={{
                left: `${x}%`,
                top: `${y}%`,
              }}
              onClick={() =>
                onToggleTower(
                  selectedPlanet.id,
                  towerIndex,
                  disabled
                )
              }
              title={
                disabled
                  ? `Restore Tower ${towerIndex}`
                  : `Disable Tower ${towerIndex}`
              }
            >
              {towerIndex}
            </button>
          );
        })}
      </div>

      <div 
  className="tower-list-heading" 
  style={{
    display: "flex",
    alignItems: "flex-end",
    justifyContent: "space-between",
    gap: "10px",
    marginBottom: "7px",
    paddingBottom: "7px",
    borderBottom: "1px solid rgba(71, 85, 105, 0.68)"
  }}
>
  <span 
    style={{
      color: "#e0f2fe",
      fontSize: "11px",
      fontWeight: 800
    }}
  >
    Tower controls
  </span>
  <small 
    style={{
      color: "#64748b",
      fontSize: "8px"
    }}
  >
    Click a tower to change its state
  </small>
</div>

     
    </div>
  );
}