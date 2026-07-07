import type { LinkEvaluation } from "../types/agent";

interface LinkEvaluationTableProps {
  evaluations: LinkEvaluation[];
}

function score(value: number | undefined): string {
  if (value === undefined || !Number.isFinite(value)) return "—";
  return value.toFixed(3);
}

function latency(value: number | undefined): string {
  if (value === undefined || !Number.isFinite(value)) return "—";
  return `${value.toFixed(2)} ms`;
}

export function LinkEvaluationTable({
  evaluations,
}: LinkEvaluationTableProps) {
  return (
    <section className="agent-card" aria-labelledby="link-evaluations-title">
      <div className="agent-card-header">
        <div>
          <p className="agent-eyebrow">True Cost</p>
          <h2 id="link-evaluations-title">Link evaluations</h2>
        </div>
      </div>

      {evaluations.length === 0 ? (
        <p className="agent-empty">No evaluated links yet.</p>
      ) : (
        <div className="agent-table-wrap">
          <table className="agent-table">
            <thead>
              <tr>
                <th>Link</th>
                <th>Physical</th>
                <th>Congestion</th>
                <th>Trust</th>
                <th>Targeting</th>
                <th>Uncertainty</th>
                <th>True Cost</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {evaluations.map((evaluation, index) => (
                <tr key={`${evaluation.link_id}-${index}`}>
                  <td>
                    <strong>{evaluation.link_id}</strong>
                    {evaluation.reasons && evaluation.reasons.length > 0 && (
                      <small>{evaluation.reasons.join(" · ")}</small>
                    )}
                  </td>
                  <td>{latency(evaluation.physical_latency_ms)}</td>
                  <td>
                    {latency(
                      evaluation.predicted_congestion_penalty_ms,
                    )}
                  </td>
                  <td>{score(evaluation.trust_score)}</td>
                  <td>{score(evaluation.targeting_risk_score)}</td>
                  <td>{score(evaluation.uncertainty_score)}</td>
                  <td>{latency(evaluation.combined_cost)}</td>
                  <td>
                    <span className="agent-status-pill">
                      {evaluation.action ?? "SCORED"}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
