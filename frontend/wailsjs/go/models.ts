export namespace domain {
	
	export class VoidTransitBreakdown {
	    from_id: string;
	    to_id: string;
	    sending_tower: number;
	    receiving_tower: number;
	    void_distance_km: number;
	    source_atmosphere_seconds: number;
	    vacuum_seconds: number;
	    destination_atmosphere_seconds: number;
	    total_seconds: number;
	
	    static createFrom(source: any = {}) {
	        return new VoidTransitBreakdown(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from_id = source["from_id"];
	        this.to_id = source["to_id"];
	        this.sending_tower = source["sending_tower"];
	        this.receiving_tower = source["receiving_tower"];
	        this.void_distance_km = source["void_distance_km"];
	        this.source_atmosphere_seconds = source["source_atmosphere_seconds"];
	        this.vacuum_seconds = source["vacuum_seconds"];
	        this.destination_atmosphere_seconds = source["destination_atmosphere_seconds"];
	        this.total_seconds = source["total_seconds"];
	    }
	}
	export class PlanetTransitBreakdown {
	    planet_id: string;
	    entry_tower: number;
	    exit_tower: number;
	    ring_path: number[];
	    ring_direction: string;
	    segments: number;
	    distinct_towers: number;
	    fiber_seconds: number;
	    tower_seconds: number;
	    total_seconds: number;
	
	    static createFrom(source: any = {}) {
	        return new PlanetTransitBreakdown(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.planet_id = source["planet_id"];
	        this.entry_tower = source["entry_tower"];
	        this.exit_tower = source["exit_tower"];
	        this.ring_path = source["ring_path"];
	        this.ring_direction = source["ring_direction"];
	        this.segments = source["segments"];
	        this.distinct_towers = source["distinct_towers"];
	        this.fiber_seconds = source["fiber_seconds"];
	        this.tower_seconds = source["tower_seconds"];
	        this.total_seconds = source["total_seconds"];
	    }
	}
	export class HopLogEntry {
	    sequence: number;
	    planet_id: string;
	    previous_planet_id?: string;
	    next_planet_id?: string;
	    entry_tower: number;
	    exit_tower: number;
	    ring_path: number[];
	    ring_direction: string;
	    segments: number;
	    distinct_towers: number;
	    local_codex: number;
	    next_hop_codex?: number;
	    decoded_payload: string;
	    encoded_payload?: string[];
	    binary_stream?: string;
	    planet_transit: PlanetTransitBreakdown;
	    void_transit?: VoidTransitBreakdown;
	    step_latency_seconds: number;
	    cumulative_latency_seconds: number;
	
	    static createFrom(source: any = {}) {
	        return new HopLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sequence = source["sequence"];
	        this.planet_id = source["planet_id"];
	        this.previous_planet_id = source["previous_planet_id"];
	        this.next_planet_id = source["next_planet_id"];
	        this.entry_tower = source["entry_tower"];
	        this.exit_tower = source["exit_tower"];
	        this.ring_path = source["ring_path"];
	        this.ring_direction = source["ring_direction"];
	        this.segments = source["segments"];
	        this.distinct_towers = source["distinct_towers"];
	        this.local_codex = source["local_codex"];
	        this.next_hop_codex = source["next_hop_codex"];
	        this.decoded_payload = source["decoded_payload"];
	        this.encoded_payload = source["encoded_payload"];
	        this.binary_stream = source["binary_stream"];
	        this.planet_transit = this.convertValues(source["planet_transit"], PlanetTransitBreakdown);
	        this.void_transit = this.convertValues(source["void_transit"], VoidTransitBreakdown);
	        this.step_latency_seconds = source["step_latency_seconds"];
	        this.cumulative_latency_seconds = source["cumulative_latency_seconds"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LatencyBreakdown {
	    fiber_seconds: number;
	    tower_seconds: number;
	    atmosphere_seconds: number;
	    void_seconds: number;
	    total_seconds: number;
	
	    static createFrom(source: any = {}) {
	        return new LatencyBreakdown(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fiber_seconds = source["fiber_seconds"];
	        this.tower_seconds = source["tower_seconds"];
	        this.atmosphere_seconds = source["atmosphere_seconds"];
	        this.void_seconds = source["void_seconds"];
	        this.total_seconds = source["total_seconds"];
	    }
	}
	export class Link {
	    a: string;
	    b: string;
	    void_distance_km: number;
	    a_tower: number;
	    b_tower: number;
	
	    static createFrom(source: any = {}) {
	        return new Link(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.a = source["a"];
	        this.b = source["b"];
	        this.void_distance_km = source["void_distance_km"];
	        this.a_tower = source["a_tower"];
	        this.b_tower = source["b_tower"];
	    }
	}
	export class PlanetStatus {
	    id: string;
	    endpoint: string;
	    healthy: boolean;
	    manually_disabled: boolean;
	    available: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PlanetStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.endpoint = source["endpoint"];
	        this.healthy = source["healthy"];
	        this.manually_disabled = source["manually_disabled"];
	        this.available = source["available"];
	    }
	}
	export class Planet {
	    id: string;
	    codex: number;
	    x: number;
	    y: number;
	    radius_km: number;
	    active_towers: number;
	    atmosphere_thickness_km: number;
	    refraction_index: number;
	
	    static createFrom(source: any = {}) {
	        return new Planet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.codex = source["codex"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.radius_km = source["radius_km"];
	        this.active_towers = source["active_towers"];
	        this.atmosphere_thickness_km = source["atmosphere_thickness_km"];
	        this.refraction_index = source["refraction_index"];
	    }
	}
	export class UniverseMetadata {
	    system_name: string;
	    speed_of_light_kms: number;
	    max_void_hop_distance_km: number;
	    coordinate_scale_unit_km: number;
	    tower_processing_delay_ms: number;
	    fiber_speed_fraction: number;
	
	    static createFrom(source: any = {}) {
	        return new UniverseMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.system_name = source["system_name"];
	        this.speed_of_light_kms = source["speed_of_light_kms"];
	        this.max_void_hop_distance_km = source["max_void_hop_distance_km"];
	        this.coordinate_scale_unit_km = source["coordinate_scale_unit_km"];
	        this.tower_processing_delay_ms = source["tower_processing_delay_ms"];
	        this.fiber_speed_fraction = source["fiber_speed_fraction"];
	    }
	}
	export class NetworkSnapshot {
	    metadata: UniverseMetadata;
	    planets: Planet[];
	    planet_status: PlanetStatus[];
	    links: Link[];
	    disabled_links: string[];
	    disabled_towers: Record<string, Array<number>>;
	
	    static createFrom(source: any = {}) {
	        return new NetworkSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metadata = this.convertValues(source["metadata"], UniverseMetadata);
	        this.planets = this.convertValues(source["planets"], Planet);
	        this.planet_status = this.convertValues(source["planet_status"], PlanetStatus);
	        this.links = this.convertValues(source["links"], Link);
	        this.disabled_links = source["disabled_links"];
	        this.disabled_towers = source["disabled_towers"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Packet {
	    id: string;
	    origin_id: string;
	    destination_id: string;
	    current_id: string;
	    payload: string;
	    encoded_payload?: string[];
	    binary_stream?: string;
	    hop_log: HopLogEntry[];
	    route: string[];
	    route_index: number;
	    status: string;
	    decoded_payload?: string;
	    cumulative_latency_seconds: number;
	
	    static createFrom(source: any = {}) {
	        return new Packet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.origin_id = source["origin_id"];
	        this.destination_id = source["destination_id"];
	        this.current_id = source["current_id"];
	        this.payload = source["payload"];
	        this.encoded_payload = source["encoded_payload"];
	        this.binary_stream = source["binary_stream"];
	        this.hop_log = this.convertValues(source["hop_log"], HopLogEntry);
	        this.route = source["route"];
	        this.route_index = source["route_index"];
	        this.status = source["status"];
	        this.decoded_payload = source["decoded_payload"];
	        this.cumulative_latency_seconds = source["cumulative_latency_seconds"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class Route {
	    path: string[];
	    planet_transits: PlanetTransitBreakdown[];
	    void_transits: VoidTransitBreakdown[];
	    latency: LatencyBreakdown;
	
	    static createFrom(source: any = {}) {
	        return new Route(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.planet_transits = this.convertValues(source["planet_transits"], PlanetTransitBreakdown);
	        this.void_transits = this.convertValues(source["void_transits"], VoidTransitBreakdown);
	        this.latency = this.convertValues(source["latency"], LatencyBreakdown);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TransmissionResult {
	    status: string;
	    message?: string;
	    packet: Packet;
	    route: Route;
	    original_payload: string;
	    decoded_payload: string;
	    total_latency_seconds: number;
	    latency: LatencyBreakdown;
	
	    static createFrom(source: any = {}) {
	        return new TransmissionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.message = source["message"];
	        this.packet = this.convertValues(source["packet"], Packet);
	        this.route = this.convertValues(source["route"], Route);
	        this.original_payload = source["original_payload"];
	        this.decoded_payload = source["decoded_payload"];
	        this.total_latency_seconds = source["total_latency_seconds"];
	        this.latency = this.convertValues(source["latency"], LatencyBreakdown);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace main {
	
	export class AgentStateDTO {
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
	
	    static createFrom(source: any = {}) {
	        return new AgentStateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.current_tick = source["current_tick"];
	        this.current_planet = source["current_planet"];
	        this.destination_planet = source["destination_planet"];
	        this.active_message_id = source["active_message_id"];
	        this.last_confirmed_planet = source["last_confirmed_planet"];
	        this.current_path = source["current_path"];
	        this.queued_packets = source["queued_packets"];
	        this.quarantined_links = source["quarantined_links"];
	        this.last_error = source["last_error"];
	    }
	}
	export class AuditRecordDTO {
	    audit_id: string;
	    tick: number;
	    current_planet: string;
	    link_id: string;
	    action: string;
	    physical_latency_ms: number;
	    predicted_congestion_penalty_ms: number;
	    trust_score: number;
	    targeting_risk_score: number;
	    uncertainty_score: number;
	    combined_cost: number;
	    reasons: string[];
	    alternative_path?: string[];
	    timestamp?: string;
	
	    static createFrom(source: any = {}) {
	        return new AuditRecordDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.audit_id = source["audit_id"];
	        this.tick = source["tick"];
	        this.current_planet = source["current_planet"];
	        this.link_id = source["link_id"];
	        this.action = source["action"];
	        this.physical_latency_ms = source["physical_latency_ms"];
	        this.predicted_congestion_penalty_ms = source["predicted_congestion_penalty_ms"];
	        this.trust_score = source["trust_score"];
	        this.targeting_risk_score = source["targeting_risk_score"];
	        this.uncertainty_score = source["uncertainty_score"];
	        this.combined_cost = source["combined_cost"];
	        this.reasons = source["reasons"];
	        this.alternative_path = source["alternative_path"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class LinkEvaluationDTO {
	    link_id: string;
	    physical_latency_ms?: number;
	    predicted_congestion_penalty_ms: number;
	    trust_score: number;
	    targeting_risk_score: number;
	    uncertainty_score?: number;
	    combined_cost: number;
	    action?: string;
	    reasons?: string[];
	
	    static createFrom(source: any = {}) {
	        return new LinkEvaluationDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.link_id = source["link_id"];
	        this.physical_latency_ms = source["physical_latency_ms"];
	        this.predicted_congestion_penalty_ms = source["predicted_congestion_penalty_ms"];
	        this.trust_score = source["trust_score"];
	        this.targeting_risk_score = source["targeting_risk_score"];
	        this.uncertainty_score = source["uncertainty_score"];
	        this.combined_cost = source["combined_cost"];
	        this.action = source["action"];
	        this.reasons = source["reasons"];
	    }
	}
	export class DecisionReportDTO {
	    origin_id: string;
	    destination_id: string;
	    parsed_payload?: string;
	    baseline_path?: string[];
	    chosen_path: string[];
	    link_evaluations: LinkEvaluationDTO[];
	    final_latency_estimate_ms: number;
	    explanation: string;
	    status?: string;
	    confidence?: number;
	
	    static createFrom(source: any = {}) {
	        return new DecisionReportDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.origin_id = source["origin_id"];
	        this.destination_id = source["destination_id"];
	        this.parsed_payload = source["parsed_payload"];
	        this.baseline_path = source["baseline_path"];
	        this.chosen_path = source["chosen_path"];
	        this.link_evaluations = this.convertValues(source["link_evaluations"], LinkEvaluationDTO);
	        this.final_latency_estimate_ms = source["final_latency_estimate_ms"];
	        this.explanation = source["explanation"];
	        this.status = source["status"];
	        this.confidence = source["confidence"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class PacketTimelineEntryDTO {
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
	
	    static createFrom(source: any = {}) {
	        return new PacketTimelineEntryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message_id = source["message_id"];
	        this.packet_id = source["packet_id"];
	        this.sequence_number = source["sequence_number"];
	        this.total_packets = source["total_packets"];
	        this.state = source["state"];
	        this.planet_id = source["planet_id"];
	        this.link_id = source["link_id"];
	        this.route_version = source["route_version"];
	        this.retry_count = source["retry_count"];
	        this.tick = source["tick"];
	        this.detail = source["detail"];
	    }
	}
	export class ParsedTransmissionRequestDTO {
	    origin_id: string;
	    destination_id: string;
	    payload: string;
	    confidence: number;
	    ambiguities: string[];
	
	    static createFrom(source: any = {}) {
	        return new ParsedTransmissionRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.origin_id = source["origin_id"];
	        this.destination_id = source["destination_id"];
	        this.payload = source["payload"];
	        this.confidence = source["confidence"];
	        this.ambiguities = source["ambiguities"];
	    }
	}

}

export namespace protocol {
	
	export class TransmissionRequest {
	    origin_id: string;
	    destination_id: string;
	    payload: string;
	
	    static createFrom(source: any = {}) {
	        return new TransmissionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.origin_id = source["origin_id"];
	        this.destination_id = source["destination_id"];
	        this.payload = source["payload"];
	    }
	}
	export class UniverseResponse {
	    status: string;
	    snapshot: domain.NetworkSnapshot;
	
	    static createFrom(source: any = {}) {
	        return new UniverseResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.snapshot = this.convertValues(source["snapshot"], domain.NetworkSnapshot);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

