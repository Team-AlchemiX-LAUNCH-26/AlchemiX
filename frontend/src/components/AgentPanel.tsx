import { FormEvent, useMemo, useState } from "react";
import type {
  AgentDecisionReport,
  ParsedTransmissionRequest,
} from "../types/agent";

interface AgentPanelProps {
  busy: boolean;
  parsed?: ParsedTransmissionRequest;
  report?: AgentDecisionReport;
  error?: string;
  onParse: (raw: string) => Promise<void>;
  onEvaluate: (raw: string) => Promise<void>;
  onReset: () => Promise<void>;
}

const DEFAULT_REQUEST =
  'Send "Emergency relay check" from Aegis to Caelum';

export function AgentPanel({
  busy,
  parsed,
  report,
  error,
  onParse,
  onEvaluate,
  onReset,
}: AgentPanelProps) {
  const [request, setRequest] = useState(DEFAULT_REQUEST);

  const canSubmit = useMemo(
    () => request.trim().length > 0 && !busy,
    [request, busy],
  );

  async function handleParse(event: FormEvent) {
    event.preventDefault();
    if (!canSubmit) return;
    await onParse(request.trim());
  }

  async function handleEvaluate() {
    if (!canSubmit) return;
    await onEvaluate(request.trim());
  }

  return (
    <section className="agent-card agent-panel" aria-labelledby="agent-title">
      <div className="agent-card-header">
        <div>
          <p className="agent-eyebrow">Analytical Co-Pilot</p>
          <h2 id="agent-title">Natural-language transmission</h2>
        </div>
        <button
          className="agent-button agent-button-secondary"
          type="button"
          disabled={busy}
          onClick={() => void onReset()}
        >
          Reset agent
        </button>
      </div>

      <form onSubmit={handleParse}>
        <label className="agent-label" htmlFor="agent-request">
          Transmission request
        </label>
        <textarea
          id="agent-request"
          className="agent-textarea"
          value={request}
          onChange={(event) => setRequest(event.target.value)}
          rows={4}
          maxLength={4096}
          spellCheck
        />

        <div className="agent-action-row">
          <button
            className="agent-button agent-button-secondary"
            type="submit"
            disabled={!canSubmit}
          >
            Parse only
          </button>
          <button
            className="agent-button agent-button-primary"
            type="button"
            disabled={!canSubmit}
            onClick={() => void handleEvaluate()}
          >
            {busy ? "Evaluating…" : "Evaluate and transmit"}
          </button>
        </div>
      </form>

      {error && <p className="agent-error" role="alert">{error}</p>}

      {parsed && (
        <div className="agent-preview-grid" aria-label="Parsed request">
          <div>
            <span>Origin</span>
            <strong>{parsed.origin_id || "Missing"}</strong>
          </div>
          <div>
            <span>Destination</span>
            <strong>{parsed.destination_id || "Missing"}</strong>
          </div>
          <div>
            <span>Parser confidence</span>
            <strong>{Math.round(parsed.confidence * 100)}%</strong>
          </div>
          <div className="agent-preview-payload">
            <span>Payload</span>
            <strong>{parsed.payload || "Missing"}</strong>
          </div>
          {parsed.ambiguities.length > 0 && (
            <div className="agent-preview-payload">
              <span>Ambiguities</span>
              <strong>{parsed.ambiguities.join("; ")}</strong>
            </div>
          )}
        </div>
      )}

      {report && (
        <div className="agent-summary">
          <div>
            <span>Status</span>
            <strong>{report.status ?? "COMPLETED"}</strong>
          </div>
          <div>
            <span>Chosen route</span>
            <strong>{report.chosen_path.join(" → ")}</strong>
          </div>
          <div>
            <span>Estimated latency</span>
            <strong>{report.final_latency_estimate_ms.toFixed(2)} ms</strong>
          </div>
          <p>{report.explanation}</p>
        </div>
      )}
    </section>
  );
}
