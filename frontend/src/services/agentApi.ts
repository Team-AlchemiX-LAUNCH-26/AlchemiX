import {
  EvaluateTransmission,
  GetAgentState,
  GetDecisionAudit,
  GetPacketTimeline,
  ParseTransmissionRequest,
  ResetAgentHistory,
} from "../../wailsjs/go/main/AgentApp";

import type {
  AgentDecisionReport,
  AgentState,
  AuditRecord,
  PacketTimelineEntry,
  ParsedTransmissionRequest,
} from "../types/agent";

/**
 * This is the only frontend file that depends directly on the Wails-generated
 * package path. If Wails generates a different path, update the import above.
 */
export const agentApi = {
  async parse(raw: string): Promise<ParsedTransmissionRequest> {
    return (await ParseTransmissionRequest(raw)) as ParsedTransmissionRequest;
  },

  async evaluate(raw: string): Promise<AgentDecisionReport> {
    return (await EvaluateTransmission(raw)) as AgentDecisionReport;
  },

  async state(): Promise<AgentState> {
    return (await GetAgentState()) as AgentState;
  },

  async audit(): Promise<AuditRecord[]> {
    return (await GetDecisionAudit()) as AuditRecord[];
  },

  async timeline(): Promise<PacketTimelineEntry[]> {
    return (await GetPacketTimeline()) as PacketTimelineEntry[];
  },

  async reset(): Promise<void> {
    await ResetAgentHistory();
  },
};
