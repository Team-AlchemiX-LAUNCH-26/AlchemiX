import type { PacketTimelineEntry } from "../types/agent";

interface PacketTimelineProps {
  entries: PacketTimelineEntry[];
}

export function PacketTimeline({ entries }: PacketTimelineProps) {
  const ordered = [...entries].sort(
    (left, right) =>
      left.tick - right.tick ||
      left.sequence_number - right.sequence_number,
  );

  return (
    <section className="agent-card" aria-labelledby="packet-timeline-title">
      <div className="agent-card-header">
        <div>
          <p className="agent-eyebrow">Store and forward</p>
          <h2 id="packet-timeline-title">Packet timeline</h2>
        </div>
        <span className="agent-count">{entries.length} events</span>
      </div>

      {ordered.length === 0 ? (
        <p className="agent-empty">No packet activity yet.</p>
      ) : (
        <ol className="agent-timeline">
          {ordered.map((entry, index) => (
            <li
              key={`${entry.packet_id}-${entry.state}-${entry.tick}-${index}`}
            >
              <span className="agent-timeline-node" aria-hidden="true" />
              <div>
                <div className="agent-timeline-heading">
                  <span className="agent-status-pill">{entry.state}</span>
                  <strong>
                    Packet {entry.sequence_number}/{entry.total_packets}
                  </strong>
                  <span>Tick {entry.tick}</span>
                </div>
                <p>
                  {entry.planet_id && <>Planet: {entry.planet_id}. </>}
                  {entry.link_id && <>Link: {entry.link_id}. </>}
                  Route v{entry.route_version}; retries {entry.retry_count}.
                </p>
                {entry.detail && <small>{entry.detail}</small>}
              </div>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}
