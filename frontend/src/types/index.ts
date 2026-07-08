export interface Metadata { system_name:string; speed_of_light_kms:number; max_void_hop_distance_km:number; coordinate_scale_unit_km:number; tower_processing_delay_ms:number; fiber_speed_fraction:number }
export interface Planet { id:string; codex:number; x:number; y:number; radius_km:number; active_towers:number; atmosphere_thickness_km:number; refraction_index:number }
export interface PlanetStatus { id:string; endpoint:string; healthy:boolean; manually_disabled:boolean; available:boolean }
export interface Link { a:string; b:string; void_distance_km:number; a_tower:number; b_tower:number }
export interface Snapshot {
    metadata: Metadata;
    planets: Planet[];
    planet_status: PlanetStatus[];
    links: Link[];
    disabled_links: string[];
    disabled_towers: Record<string, number[]>;
  }
export interface UniverseResponse { status:string; snapshot:Snapshot }
export interface Transit {
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
  }
export interface VoidTransit { from_id:string; to_id:string; sending_tower:number; receiving_tower:number; void_distance_km:number; source_atmosphere_seconds:number; vacuum_seconds:number; destination_atmosphere_seconds:number; total_seconds:number }
export interface Latency { fiber_seconds:number; tower_seconds:number; atmosphere_seconds:number; void_seconds:number; total_seconds:number }
export interface Route { path:string[]; planet_transits:Transit[]; void_transits:VoidTransit[]; latency:Latency }
export interface HopLog {
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
  
    step_latency_seconds: number;
    cumulative_latency_seconds: number;
  }
export interface TransmissionResult { status:string; message?:string; route:Route; original_payload:string; decoded_payload:string; total_latency_seconds:number; latency:Latency; packet:{id:string;hop_log:HopLog[]} }
export interface NetworkEvent { type:string; timestamp:string; packet_id?:string; planet_id?:string; message?:string; data?:unknown }
