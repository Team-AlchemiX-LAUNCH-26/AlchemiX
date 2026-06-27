import type { HopLog } from "../types";

export function HopLogTable({ logs }: { logs: HopLog[] }) {
  return (
    <div className="hop-scroll">
      <table>
        <thead>
          <tr>
            <th>#</th>
            <th>Planet</th>
            <th>Towers</th>
            <th>Ring path</th>
            <th>Direction</th>
            <th>Codex</th>
            <th>Segments</th>
            <th>Step</th>
            <th>Cumulative</th>
          </tr>
        </thead>

        <tbody>
          {logs.map((hop) => (
            <tr key={hop.sequence}>
              <td>{hop.sequence}</td>
              <td>{hop.planet_id}</td>
              <td>
                {hop.entry_tower} → {hop.exit_tower}
              </td>
              <td>
                {hop.ring_path?.length
                  ? hop.ring_path.join(" → ")
                  : "-"}
              </td>
              <td>{hop.ring_direction ?? "-"}</td>
              <td>
                B{hop.local_codex}
                {hop.next_hop_codex
                  ? ` → B${hop.next_hop_codex}`
                  : ""}
              </td>
              <td>{hop.segments}</td>
              <td>{hop.step_latency_seconds.toFixed(5)}s</td>
              <td>{hop.cumulative_latency_seconds.toFixed(5)}s</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}