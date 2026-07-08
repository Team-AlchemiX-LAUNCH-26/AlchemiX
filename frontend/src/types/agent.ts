export type AgentAction =
  | "CONTINUE"
  | "REROUTE"
  | "QUEUE"
  | "PROBE"
  | "QUARANTINE"
  | "DELIVERED"
  | "FAILED_SAFE"
  | string;

export interface ParsedTransmissionRequest {
  origin_id: string;
  destination_id: string;
  payload: string;
  confidence: number;
  ambiguities: string[];
}

export interface LinkEvaluation {
  link_id: string;
  physical_latency_ms?: number;
  predicted_congestion_penalty_ms: number;
  trust_score: number;
  targeting_risk_score: number;
  uncertainty_score?: number;
  combined_cost: number;
  action?: AgentAction;
  reasons?: string[];
}

export interface HopDecision {
  tick: number;
  current_planet: string;
  next_planet: string;
  link_id: string;
  action: AgentAction;
  evaluation: LinkEvaluation;
  reasons: string[];
  alternative_path?: string[];
}

export interface AgentDecisionReport {
  origin_id: string;
  destination_id: string;
  parsed_payload?: string;
  baseline_path?: string[];
  chosen_path: string[];
  link_evaluations: LinkEvaluation[];
  final_latency_estimate_ms: number;
  explanation: string;
  status?: string;
  confidence?: number;
}

export interface AgentState {
  status: string;
  current_tick: number;
  current_planet?: string;
  destination_planet?: string;
  active_message_id?: string;
  last_confirmed_planet?: string;
  current_path?: string[];
  queued_packets: number;
  quarantined_links?: string[];
  last_error?: string;
}

export interface AuditRecord {
  audit_id: string;
  tick: number;
  current_planet: string;
  link_id: string;
  action: AgentAction;
  physical_latency_ms: number;
  predicted_congestion_penalty_ms: number;
  trust_score: number;
  targeting_risk_score: number;
  uncertainty_score: number;
  combined_cost: number;
  reasons: string[];
  alternative_path?: string[];
  timestamp?: string;
}

export interface PacketTimelineEntry {
  message_id: string;
  packet_id: string;
  sequence_number: number;
  total_packets: number;
  state: string;
  planet_id?: string;
  link_id?: string;
  route_version: number;
  retry_count: number;
  tick: number;
  detail?: string;
}
