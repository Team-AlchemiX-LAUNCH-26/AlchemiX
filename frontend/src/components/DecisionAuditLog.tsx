import type { AuditRecord } from "../types/agent";

interface DecisionAuditLogProps {
  records: AuditRecord[];
}

export function DecisionAuditLog({ records }: DecisionAuditLogProps) {
  return (
    <section className="agent-card" aria-labelledby="audit-title">
      <div className="agent-card-header">
        <div>
          <p className="agent-eyebrow">Explainability</p>
          <h2 id="audit-title">Decision audit</h2>
        </div>
        <span className="agent-count">{records.length} records</span>
      </div>

      {records.length === 0 ? (
        <p className="agent-empty">No decisions have been recorded.</p>
      ) : (
        <div className="agent-audit-list">
          {[...records].reverse().map((record) => (
            <details className="agent-audit-item" key={record.audit_id}>
              <summary>
                <span className="agent-status-pill">{record.action}</span>
                <strong>{record.link_id}</strong>
                <span>Tick {record.tick}</span>
                <span>{record.combined_cost.toFixed(2)} ms</span>
              </summary>

              <div className="agent-audit-body">
                <dl>
                  <div>
                    <dt>Current planet</dt>
                    <dd>{record.current_planet}</dd>
                  </div>
                  <div>
                    <dt>Physical latency</dt>
                    <dd>{record.physical_latency_ms.toFixed(2)} ms</dd>
                  </div>
                  <div>
                    <dt>Congestion penalty</dt>
                    <dd>
                      {record.predicted_congestion_penalty_ms.toFixed(2)} ms
                    </dd>
                  </div>
                  <div>
                    <dt>Trust</dt>
                    <dd>{record.trust_score.toFixed(3)}</dd>
                  </div>
                  <div>
                    <dt>Targeting risk</dt>
                    <dd>{record.targeting_risk_score.toFixed(3)}</dd>
                  </div>
                  <div>
                    <dt>Uncertainty</dt>
                    <dd>{record.uncertainty_score.toFixed(3)}</dd>
                  </div>
                </dl>

                {record.reasons.length > 0 && (
                  <div>
                    <h3>Reasons</h3>
                    <ul>
                      {record.reasons.map((reason) => (
                        <li key={reason}>{reason}</li>
                      ))}
                    </ul>
                  </div>
                )}

                {record.alternative_path &&
                  record.alternative_path.length > 0 && (
                    <p>
                      <strong>Alternative:</strong>{" "}
                      {record.alternative_path.join(" → ")}
                    </p>
                  )}
              </div>
            </details>
          ))}
        </div>
      )}
    </section>
  );
}
