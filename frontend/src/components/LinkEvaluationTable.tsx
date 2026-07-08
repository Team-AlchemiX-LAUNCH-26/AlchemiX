import type { LinkEvaluation } from "../types/agent";

interface LinkEvaluationTableProps {
  evaluations: LinkEvaluation[];
  visitedPlanets?: string[];
}

interface PlanetLinkGroup {
  planetID: string;
  evaluations: LinkEvaluation[];
}

function score(value: number | undefined): string {
  if (value === undefined || !Number.isFinite(value)) return "-";
  return value.toFixed(3);
}

function latency(value: number | undefined): string {
  if (value === undefined || !Number.isFinite(value)) return "-";
  return `${value.toFixed(2)} ms`;
}

function linkEndpoints(linkID: string): string[] {
  if (linkID.includes("-")) return linkID.split("-").filter(Boolean);
  if (linkID.includes("::")) return linkID.split("::").filter(Boolean);
  return [linkID];
}

function buildPlanetGroups(
  evaluations: LinkEvaluation[],
  visitedPlanets: string[] = [],
): PlanetLinkGroup[] {
  const allPlanets = new Set<string>();
  const byPlanet = new Map<string, LinkEvaluation[]>();

  evaluations.forEach((evaluation) => {
    linkEndpoints(evaluation.link_id).forEach((planetID) => {
      allPlanets.add(planetID);
      const existing = byPlanet.get(planetID) ?? [];
      existing.push(evaluation);
      byPlanet.set(planetID, existing);
    });
  });

  const orderedPlanets =
    visitedPlanets.length > 0
      ? visitedPlanets
      : [...allPlanets].sort((left, right) => left.localeCompare(right));

  return orderedPlanets.map((planetID) => ({
    planetID,
    evaluations: byPlanet.get(planetID) ?? [],
  }));
}

export function LinkEvaluationTable({
  evaluations,
  visitedPlanets = [],
}: LinkEvaluationTableProps) {
  const planetGroups = buildPlanetGroups(evaluations, visitedPlanets);

  return (
    <section className="agent-card" aria-labelledby="link-evaluations-title">
      <div className="agent-card-header">
        <div>
          <p className="agent-eyebrow">True Cost</p>
          <h2 id="link-evaluations-title">Visited planet diagnostics</h2>
        </div>
      </div>

      {evaluations.length === 0 ? (
        <p className="agent-empty">No evaluated links yet.</p>
      ) : (
        <div className="agent-planet-accordion">
          {planetGroups.map((group, index) => (
            <details
              className="agent-planet-diagnostics"
              key={`${group.planetID}-${index}`}
              open={index === 0}
            >
              <summary>
                <div className="agent-planet-summary-main">
                  <span>Planet {index + 1}</span>
                  <strong>{group.planetID}</strong>
                </div>
                <span className="agent-count">
                  {group.evaluations.length} link
                  {group.evaluations.length === 1 ? "" : "s"}
                </span>
              </summary>

              {group.evaluations.length === 0 ? (
                <p className="agent-empty agent-planet-diagnostics-empty">
                  No link diagnostics recorded for this planet.
                </p>
              ) : (
                <div className="agent-planet-link-list">
                  {group.evaluations.map((evaluation, evaluationIndex) => (
                    <article
                      className="agent-planet-link-item"
                      key={`${group.planetID}-${evaluation.link_id}-${evaluationIndex}`}
                    >
                      <div className="agent-planet-link-main">
                        <div>
                          <strong>{evaluation.link_id}</strong>
                          {evaluation.reasons &&
                            evaluation.reasons.length > 0 && (
                              <small>{evaluation.reasons.join(" / ")}</small>
                            )}
                        </div>
                        <span className="agent-status-pill">
                          {evaluation.action ?? "SCORED"}
                        </span>
                      </div>

                      <dl className="agent-planet-link-metrics">
                        <div>
                          <dt>Physical</dt>
                          <dd>{latency(evaluation.physical_latency_ms)}</dd>
                        </div>
                        <div>
                          <dt>Congestion</dt>
                          <dd>
                            {latency(
                              evaluation.predicted_congestion_penalty_ms,
                            )}
                          </dd>
                        </div>
                        <div>
                          <dt>Trust</dt>
                          <dd>{score(evaluation.trust_score)}</dd>
                        </div>
                        <div>
                          <dt>Targeting</dt>
                          <dd>{score(evaluation.targeting_risk_score)}</dd>
                        </div>
                        <div>
                          <dt>Uncertainty</dt>
                          <dd>{score(evaluation.uncertainty_score)}</dd>
                        </div>
                        <div>
                          <dt>True Cost</dt>
                          <dd>{latency(evaluation.combined_cost)}</dd>
                        </div>
                      </dl>
                    </article>
                  ))}
                </div>
              )}
            </details>
          ))}
        </div>
      )}
    </section>
  );
}
