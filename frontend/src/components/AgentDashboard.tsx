import { useCallback, useEffect, useState } from "react";
import { agentApi } from "../services/agentApi";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import type {
  AgentDecisionReport,
  AgentState,
  AuditRecord,
  HopDecision,
  PacketTimelineEntry,
  ParsedTransmissionRequest,
} from "../types/agent";
import { AgentPanel } from "./AgentPanel";
import { DecisionAuditLog } from "./DecisionAuditLog";
import { LinkEvaluationTable } from "./LinkEvaluationTable";
import { PacketTimeline } from "./PacketTimeline";

const EMPTY_STATE: AgentState = {
  status: "IDLE",
  current_tick: 0,
  queued_packets: 0,
};

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === "string") return error;
  return "An unexpected agent error occurred.";
}

export interface AgentDashboardProps {
  onExpansionChange?: (expanded: boolean) => void;
  onLivePathUpdate?: (path: string[]) => void;
}

export function AgentDashboard({ onExpansionChange, onLivePathUpdate }: AgentDashboardProps) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [parsed, setParsed] = useState<ParsedTransmissionRequest>();
  const [report, setReport] = useState<AgentDecisionReport>();
  const [state, setState] = useState<AgentState>(EMPTY_STATE);
  const [audit, setAudit] = useState<AuditRecord[]>([]);
  const [timeline, setTimeline] = useState<PacketTimelineEntry[]>([]);
  const [liveHops, setLiveHops] = useState<HopDecision[]>([]);
  const [livePath, setLivePath] = useState<string[]>([]);
  const [expandedHopIndex, setExpandedHopIndex] = useState<number | null>(null);

  const refreshDiagnostics = useCallback(async () => {
    const [nextState, nextAudit, nextTimeline] = await Promise.all([
      agentApi.state(),
      agentApi.audit(),
      agentApi.timeline(),
    ]);

    setState(nextState);
    setAudit(nextAudit);
    setTimeline(nextTimeline);
  }, []);

  useEffect(() => {
    void refreshDiagnostics().catch((reason) => {
      setError(errorMessage(reason));
    });

    const unsubscribe = EventsOn("agent:hop", (raw: unknown) => {
      const hop = raw as HopDecision;
      setLiveHops((current) => {
        setExpandedHopIndex(current.length);
        return [...current, hop];
      });
      setLivePath((current) => {
        let newPath = current;
        if (current.length === 0) newPath = [hop.current_planet, hop.next_planet];
        else if (current[current.length - 1] === hop.current_planet) {
          newPath = [...current, hop.next_planet];
        }
        if (onLivePathUpdate) onLivePathUpdate(newPath);
        return newPath;
      });
    });

    return unsubscribe;
  }, [refreshDiagnostics]);

  async function parse(raw: string) {
    setBusy(true);
    setError("");
    try {
      setParsed(await agentApi.parse(raw));
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  }

  async function evaluate(raw: string) {
    setBusy(true);
    setError("");
    try {
      const parsedRequest = await agentApi.parse(raw);
      setParsed(parsedRequest);
      setLiveHops([]);
      setLivePath([]);
      if (onLivePathUpdate) onLivePathUpdate([]);
      if (onExpansionChange) onExpansionChange(true);

      const decision = await agentApi.evaluate(raw);
      setReport(decision);

      await refreshDiagnostics();
    } catch (reason) {
      setError(errorMessage(reason));
      await refreshDiagnostics().catch(() => undefined);
    } finally {
      setBusy(false);
    }
  }

  async function reset() {
    setBusy(true);
    setError("");
    try {
      await agentApi.reset();
      setParsed(undefined);
      setReport(undefined);
      setLiveHops([]);
      setLivePath([]);
      if (onLivePathUpdate) onLivePathUpdate([]);
      setExpandedHopIndex(null);
      await refreshDiagnostics();
      if (onExpansionChange) onExpansionChange(false);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="agent-dashboard">
      <div className="agent-state-strip">
        <div>
          <span>Agent state</span>
          <strong>{state.status}</strong>
        </div>
        <div>
          <span>Tick</span>
          <strong>{state.current_tick}</strong>
        </div>
        <div>
          <span>Current planet</span>
          <strong>{state.current_planet ?? "—"}</strong>
        </div>
        <div>
          <span>Last checkpoint</span>
          <strong>{state.last_confirmed_planet ?? "—"}</strong>
        </div>
        <div>
          <span>Queued packets</span>
          <strong>{state.queued_packets}</strong>
        </div>
      </div>

      <div className="agent-dashboard-grid">
        <AgentPanel
          busy={busy}
          parsed={parsed}
          report={report}
          error={error}
          onParse={parse}
          onEvaluate={evaluate}
          onReset={reset}
        />

        <div className="agent-side-card agent-card">
          <p className="agent-eyebrow">Live route</p>
          <h2>Current execution</h2>
          <p className="agent-live-path">
            {livePath.length > 0
              ? livePath.join(" → ")
              : state.current_path && state.current_path.length > 0
                ? state.current_path.join(" → ")
                : "No active path."}
          </p>

          <div className="agent-live-hops">
            {liveHops.length === 0 && <p className="agent-empty">Waiting for execution stream...</p>}
            {liveHops.map((hop, index) => {
              const isExpanded = expandedHopIndex === index;
              return (
                <div key={`${hop.tick}-${index}`} className={`agent-live-hop-item live-enter ${isExpanded ? "is-expanded" : ""}`}>
                  <div 
                    className="agent-live-hop-header" 
                    onClick={() => report && setExpandedHopIndex(isExpanded ? null : index)}
                    style={{ cursor: report ? "pointer" : "default" }}
                  >
                    <strong>{hop.current_planet}</strong> 
                    <span className="arrow">→</span> 
                    <strong>{hop.next_planet || "?"}</strong>
                    <span className={`agent-status-pill ${hop.action.toLowerCase()}`}>{hop.action}</span>
                    {report && (
                      <span className="agent-chevron">{isExpanded ? "▲" : "▼"}</span>
                    )}
                  </div>
                  <div className="agent-live-hop-body">
                    <div className="agent-live-hop-metrics">
                      <div>
                        <span>Cost</span>
                        <strong>{hop.evaluation.combined_cost.toFixed(1)}</strong>
                      </div>
                      <div>
                        <span>Trust</span>
                        <strong>{hop.evaluation.trust_score.toFixed(2)}</strong>
                      </div>
                      <div>
                        <span>Risk</span>
                        <strong>{hop.evaluation.targeting_risk_score.toFixed(2)}</strong>
                      </div>
                    </div>
                    {hop.reasons && hop.reasons.length > 0 && (
                      <div className="agent-live-hop-context">
                        <span className="agent-live-hop-context-label">Decision context:</span>
                        <ul className="agent-live-hop-reasons">
                          {hop.reasons.map((r, i) => <li key={i}>{r}</li>)}
                        </ul>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {report && (
        <div className="agent-dashboard-results">
          <LinkEvaluationTable
            evaluations={report.link_evaluations ?? []}
          />

          <div className="agent-dashboard-grid agent-dashboard-grid-equal">
            <DecisionAuditLog records={audit} />
            <PacketTimeline entries={timeline} />
          </div>
        </div>
      )}
    </section>
  );
}
