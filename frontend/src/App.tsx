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

type SpaceStyle = CSSProperties & {
  "--star-drift-x"?: string;
  "--star-drift-y"?: string;
  "--satellite-scale"?: string;
};

const SPACE_STARS = Array.from({ length: 72 }, (_, index) => ({
  left: `${(index * 37 + 11) % 100}%`,
  top: `${(index * 61 + 7) % 100}%`,
  size: 1 + (index % 3),
  opacity: 0.22 + (index % 5) * 0.12,
  duration: 12 + (index % 9) * 2.1,
  delay: -(index % 13) * 1.35,
  driftX: `${((index % 7) - 3) * 9}px`,
  driftY: `${18 + (index % 6) * 8}px`,
}));

const SPACE_SATELLITES = Array.from({ length: 5 }, (_, index) => ({
  top: `${18 + index * 16}%`,
  duration: 24 + index * 7,
  delay: -(index * 8 + 3),
  scale: 0.72 + index * 0.12,
}));

function SpaceBackdrop() {
  return (
    <div className="space-backdrop" aria-hidden="true">
      <div className="space-nebula space-nebula-one" />
      <div className="space-nebula space-nebula-two" />

      <div className="space-star-field">
        {SPACE_STARS.map((star, index) => {
          const style: SpaceStyle = {
            left: star.left,
            top: star.top,
            width: star.size,
            height: star.size,
            opacity: star.opacity,
            animationDuration: `${star.duration}s`,
            animationDelay: `${star.delay}s`,
            "--star-drift-x": star.driftX,
            "--star-drift-y": star.driftY,
          };

          return (
            <span
              key={`star-${index}`}
              className={
                index % 9 === 0 ? "space-star space-star-node" : "space-star"
              }
              style={style}
            />
          );
        })}
      </div>

      <div className="space-satellite-field">
        {SPACE_SATELLITES.map((satellite, index) => {
          const style: SpaceStyle = {
            top: satellite.top,
            animationDuration: `${satellite.duration}s`,
            animationDelay: `${satellite.delay}s`,
            "--satellite-scale": String(satellite.scale),
          };

          return (
            <span
              key={`satellite-${index}`}
              className="space-satellite"
              style={style}
            />
          );
        })}
      </div>
    </div>
  );
}

export default function App() {
  const [snapshot, setSnapshot] = useState<Snapshot>();

  const [origin, setOrigin] = useState("Aegis");
  const [destination, setDestination] = useState("Caelum");
  const [payload, setPayload] = useState("Hello world");

  const [selectedPlanetID, setSelectedPlanetID] = useState<string | null>(null);

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
      <>
        <SpaceBackdrop />

        <main className="loading">
          <Satellite size={46} />

          <h1>Connecting to Zeta-26</h1>

          <p>{error || "Start the orchestrator and planet services."}</p>
        </main>
      </>
    );
  }

  const onlineNodes = snapshot.planet_status.filter(
    (status) => status.available,
  ).length;

  return (
    <>
      <SpaceBackdrop />

      <main className="shell">
        <header className="app-header">
          <div>
            <span className="eyebrow">LAUNCH26 · DISTRIBUTED DIGITAL TWIN</span>

            <h1>Relic Ring Control Centre</h1>
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

        <section className="primary-grid">
          <div className="map-card">
            <div className="card-title">
              <Network size={18} />
              Universe topology
            </div>

            <div className="map-stage" onClick={closeTowerInspector}>
              <UniverseMap
                snapshot={snapshot}
                route={result?.route}
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

            <div className="route-strip">
              <span>Selected route</span>

              <strong>
                {result?.route.path.join(" → ") || "No route calculated"}
              </strong>
            </div>
          </div>

          <aside className="card controls">
            <div className="card-title">
              <Play size={18} />
              Transmission
            </div>

            <label>
              Origin
              <select
                value={origin}
                onChange={(event) => setOrigin(event.target.value)}
              >
                {snapshot.planets.map((planet) => (
                  <option key={planet.id} value={planet.id}>
                    {planet.id}
                  </option>
                ))}
              </select>
            </label>

            <label>
              Destination
              <select
                value={destination}
                onChange={(event) => setDestination(event.target.value)}
              >
                {snapshot.planets.map((planet) => (
                  <option key={planet.id} value={planet.id}>
                    {planet.id}
                  </option>
                ))}
              </select>
            </label>

            <label>
              Payload
              <textarea
                value={payload}
                onChange={(event) => setPayload(event.target.value)}
                rows={4}
              />
            </label>

            <button
              type="button"
              className="primary-button"
              disabled={busy || origin === destination || payload.trim() === ""}
              onClick={sendPacket}
            >
              {busy ? "Transmitting…" : "Send packet"}
            </button>

            <div className="divider" />

            <div className="card-title">
              <ShieldAlert size={18} />
              Chaos controls
            </div>

            <div className="node-list">
              {snapshot.planets.map((planet) => {
                const available =
                  planetStatus.get(planet.id)?.available ?? false;

                return (
                  <button
                    type="button"
                    key={planet.id}
                    className={available ? "node-button" : "node-button danger"}
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
              onClick={resetNetwork}
            >
              <RotateCcw size={15} />
              Reset network
            </button>
          </aside>
        </section>

        {result && (
          <section className="results-section">
            <div className="results-grid results-enter">
              <article className="card result-panel">
                <div className="card-title">Latency breakdown</div>

                <LatencyPanel latency={result.latency} />
              </article>

              <article className="card result-panel">
                <div className="card-title">Delivery summary</div>

                <div className="delivery-summary">
                  <div>
                    <span>Status</span>

                    <strong className="delivery-status">{result.status}</strong>
                  </div>

                  <div>
                    <span>Original message</span>
                    <strong>{result.original_payload}</strong>
                  </div>

                  <div>
                    <span>Delivered message</span>
                    <strong>{result.decoded_payload}</strong>
                  </div>

                  <div>
                    <span>Route</span>

                    <strong>{result.route.path.join(" → ")}</strong>
                  </div>
                </div>

                <div className="binary">Packet ID: {result.packet.id}</div>
              </article>

              <article className="card result-panel">
                <div className="card-title">Live event stream</div>

                <div className="events">
                  {events.length > 0 ? (
                    events.map((event, index) => (
                      <div key={`${event.timestamp}-${index}`}>
                        <span>
                          {new Date(event.timestamp).toLocaleTimeString()}
                        </span>

                        <strong>{event.type}</strong>

                        <small>{event.planet_id || event.message || ""}</small>
                      </div>
                    ))
                  ) : (
                    <div className="empty">No network events recorded.</div>
                  )}
                </div>
              </article>
            </div>
            <article className="card hop-card results-enter">
              <div className="card-title">Ordered hop log</div>

              <HopLogTable logs={result.packet.hop_log} />
            </article>
            <article className="card conversion-history-card results-enter">
              <div className="card-title">Delivery conversion history</div>

              <ConversionHistory
                logs={result.packet.hop_log as unknown[]}
                route={result.route.path}
                planets={snapshot.planets}
                originalPayload={result.original_payload}
              />
            </article>
          </section>
        )}

        <AgentDashboard />
      </main>
    </>
  );
}
