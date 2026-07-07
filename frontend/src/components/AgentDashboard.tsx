import { useCallback, useEffect, useState } from "react";
import { agentApi } from "../services/agentApi";
import type {
  AgentDecisionReport,
  AgentState,
  AuditRecord,
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

export function AgentDashboard() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [parsed, setParsed] = useState<ParsedTransmissionRequest>();
  const [report, setReport] = useState<AgentDecisionReport>();
  const [state, setState] = useState<AgentState>(EMPTY_STATE);
  const [audit, setAudit] = useState<AuditRecord[]>([]);
  const [timeline, setTimeline] = useState<PacketTimelineEntry[]>([]);

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
      await refreshDiagnostics();
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
          <p>
            {state.current_path && state.current_path.length > 0
              ? state.current_path.join(" → ")
              : "No active path."}
          </p>

          <h3>Quarantined links</h3>
          {state.quarantined_links &&
          state.quarantined_links.length > 0 ? (
            <ul>
              {state.quarantined_links.map((link) => (
                <li key={link}>{link}</li>
              ))}
            </ul>
          ) : (
            <p className="agent-empty">None</p>
          )}

          {state.last_error && (
            <>
              <h3>Last error</h3>
              <p className="agent-error">{state.last_error}</p>
            </>
          )}
        </div>
      </div>

      <LinkEvaluationTable
        evaluations={report?.link_evaluations ?? []}
      />

      <div className="agent-dashboard-grid agent-dashboard-grid-equal">
        <DecisionAuditLog records={audit} />
        <PacketTimeline entries={timeline} />
      </div>
    </section>
  );
}
