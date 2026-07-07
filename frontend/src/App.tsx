import { useEffect, useMemo, useState } from "react";
import type { CSSProperties } from "react";
import {
  Activity,
  Network,
  Play,
  RotateCcw,
  Satellite,
  ShieldAlert,
} from "lucide-react";

import {
  DisableLink,
  DisableNode,
  DisableTower,
  EnableLink,
  EnableNode,
  EnableTower,
  GetUniverse,
  ResetNetwork,
  StartTransmission,
} from "../wailsjs/go/main/App";

import { EventsOn } from "../wailsjs/runtime/runtime";

import type { NetworkEvent, Snapshot, TransmissionResult } from "./types";

import { UniverseMap } from "./components/UniverseMap";
import { TowerInspector } from "./components/TowerInspector";
import { LatencyPanel } from "./components/LatencyPanel";
import { HopLogTable } from "./components/HopLogTable";
import { ConversionHistory } from "./components/ConversionHistory";
import { AgentDashboard } from "./components/AgentDashboard";
import "./styles/agent.css";

export default function App() {
  const [snapshot, setSnapshot] = useState<Snapshot>();

  const [origin, setOrigin] = useState("Aegis");
  const [destination, setDestination] = useState("Caelum");
  const [payload, setPayload] = useState("Hello world");

  const [selectedPlanetID, setSelectedPlanetID] = useState<string | null>(null);
  const [isExpanded, setIsExpanded] = useState(false);
  const [isMapExpanded, setIsMapExpanded] = useState(false);
  const [agentLivePath, setAgentLivePath] = useState<string[]>([]);

  const [result, setResult] = useState<TransmissionResult>();

  const [events, setEvents] = useState<NetworkEvent[]>([]);

  const [linkA, setLinkA] = useState("Aegis");
  const [linkB, setLinkB] = useState("Dawn");

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const loadUniverse = async () => {
    try {
      const response = await GetUniverse();

      setSnapshot(response.snapshot);
      setError("");
    } catch (loadError) {
      setError(String(loadError));
    }
  };

  useEffect(() => {
    void loadUniverse();

    const unsubscribe = EventsOn("network:event", (raw: unknown) => {
      const event = raw as NetworkEvent;

      setEvents((currentEvents) => [event, ...currentEvents].slice(0, 100));

      void loadUniverse();
    });

    return unsubscribe;
  }, []);

  const planetStatus = useMemo(
    () =>
      new Map(
        snapshot?.planet_status.map((status) => [status.id, status]) ?? [],
      ),
    [snapshot],
  );

  const openTowerInspector = (planetID: string) => {
    setSelectedPlanetID(planetID);
  };

  const closeTowerInspector = () => {
    setSelectedPlanetID(null);
  };

  const sendPacket = async () => {
    if (busy || origin === destination || !payload.trim()) {
      return;
    }

    setBusy(true);
    setError("");

    try {
      const transmissionResult = await StartTransmission({
        origin_id: origin,
        destination_id: destination,
        payload,
      });

      setResult(transmissionResult as unknown as TransmissionResult);

      closeTowerInspector();
      await loadUniverse();
    } catch (transmissionError) {
      setError(String(transmissionError));
    } finally {
      setBusy(false);
    }
  };

  const toggleNode = async (planetID: string, currentlyAvailable: boolean) => {
    setError("");

    try {
      if (currentlyAvailable) {
        await DisableNode(planetID);
      } else {
        await EnableNode(planetID);
      }

      setResult(undefined);
      closeTowerInspector();

      await loadUniverse();
    } catch (nodeError) {
      setError(String(nodeError));
    }
  };

  const toggleLink = async (disable: boolean) => {
    if (linkA === linkB) {
      return;
    }

    setError("");

    try {
      if (disable) {
        await DisableLink(linkA, linkB);
      } else {
        await EnableLink(linkA, linkB);
      }

      setResult(undefined);
      closeTowerInspector();

      await loadUniverse();
    } catch (linkError) {
      setError(String(linkError));
    }
  };

  const toggleTower = async (
    planetID: string,
    towerIndex: number,
    currentlyDisabled: boolean,
  ) => {
    setError("");

    try {
      if (currentlyDisabled) {
        await EnableTower(planetID, towerIndex);
      } else {
        await DisableTower(planetID, towerIndex);
      }

      setResult(undefined);
      await loadUniverse();
    } catch (towerError) {
      setError(String(towerError));
    }
  };

  const resetNetwork = async () => {
    setError("");

    try {
      await ResetNetwork();

      setResult(undefined);
      setEvents([]);
      closeTowerInspector();

      await loadUniverse();
    } catch (resetError) {
      setError(String(resetError));
    }
  };

  if (!snapshot) {
    return (
      <div className="app-shell">
        <div className="technical-grid" />
        <div className="noise-layer" />

        <main className="loading">
          <Satellite size={46} />
          <h1>Connecting to Zeta-26</h1>
          <p>{error || "Start the orchestrator and planet services."}</p>
        </main>
      </div>
    );
  }

  const onlineNodes = snapshot.planet_status.filter(
    (status) => status.available,
  ).length;

  return (
    <div className="app-shell">
      <div className="technical-grid" />
      <div className="noise-layer" />

      <main className="shell">
        <header className="app-header">
          <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <span className="accent-block">A</span>
            <h1>LCHEMIX</h1>
          </div>

          <div className="system-pill">
            <Activity size={16} />
            <span>
              {snapshot.metadata.system_name}
              {" · "}
              {onlineNodes}/{snapshot.planets.length} nodes online
            </span>
          </div>
        </header>

        {error && <div className="error-banner">{error}</div>}

        <div className={`hero-grid ${isExpanded ? "hero-grid-expanded" : ""}`}>
          <div className="hero-content">
            <p className="hero-eyebrow">Adaptive Interplanetary Intelligence</p>
            <h1 className="hero-title">
              Predict. Verify.
              <br />
              Reroute.
            </h1>
            <p className="hero-copy">
              Predict, verify, and reroute communications before Chimera can
              strike.
            </p>

            <AgentDashboard
              onExpansionChange={setIsExpanded}
              onLivePathUpdate={setAgentLivePath}
            />
          </div>

          <div className={`hero-visual ${isExpanded && !isMapExpanded ? "hero-visual-collapsed" : ""}`}>
            {isExpanded && (
              <div className="hero-visual-header" onClick={() => setIsMapExpanded(!isMapExpanded)}>
                <span>Universe Map & Controls</span>
                <span style={{ fontSize: '0.7rem' }}>{isMapExpanded ? "▼" : "▲"}</span>
              </div>
            )}
            
            <div className="hero-visual-body">
              <div className="vignette" />
              <div className="map-stage" onClick={closeTowerInspector}>
              <UniverseMap
                snapshot={snapshot}
                route={
                  agentLivePath.length > 0
                    ? ({ path: agentLivePath } as any)
                    : result?.route
                }
                selectedPlanetID={selectedPlanetID ?? undefined}
                onSelectPlanet={openTowerInspector}
              />

              <TowerInspector
                snapshot={snapshot}
                result={result}
                selectedPlanetID={selectedPlanetID}
                onClose={closeTowerInspector}
                onToggleTower={toggleTower}
                busy={busy}
              />
            </div>

            <div className="route-strip" style={{ marginTop: "16px" }}>
              <span>Selected route</span>
              <strong>
                {result
                  ? result.route.path.join(" → ")
                  : "No transmission active"}
              </strong>
            </div>

            <div style={{ marginTop: "24px" }}>
              <div className="card-title">
                <ShieldAlert size={18} />
                Chaos controls
              </div>

              <div className="node-list" style={{ marginBottom: "16px" }}>
                {snapshot.planets.map((planet) => {
                  const available =
                    planetStatus.get(planet.id)?.available ?? false;

                  return (
                    <button
                      type="button"
                      key={planet.id}
                      className={
                        available ? "node-button" : "node-button danger"
                      }
                      onClick={() => toggleNode(planet.id, available)}
                    >
                      <strong>{planet.id}</strong>
                      <span>{available ? "Disable" : "Restore"}</span>
                    </button>
                  );
                })}
              </div>

              <div className="link-controls">
                <select
                  value={linkA}
                  onChange={(event) => setLinkA(event.target.value)}
                >
                  {snapshot.planets.map((planet) => (
                    <option key={planet.id} value={planet.id}>
                      {planet.id}
                    </option>
                  ))}
                </select>
                <select
                  value={linkB}
                  onChange={(event) => setLinkB(event.target.value)}
                >
                  {snapshot.planets.map((planet) => (
                    <option key={planet.id} value={planet.id}>
                      {planet.id}
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  className="break-link-button"
                  disabled={linkA === linkB}
                  onClick={() => toggleLink(true)}
                >
                  Break link
                </button>
                <button
                  type="button"
                  className="restore-link-button"
                  disabled={linkA === linkB}
                  onClick={() => toggleLink(false)}
                >
                  Restore link
                </button>
              </div>

              <button
                type="button"
                className="secondary-button"
                style={{ marginTop: "12px" }}
                onClick={resetNetwork}
              >
                <RotateCcw size={15} />
                Reset network
              </button>
            </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
