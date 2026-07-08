from __future__ import annotations

from pydantic import BaseModel, Field


class LinkObservation(BaseModel):
    tick: int
    link_id: str
    planet_a: str | None = None
    planet_b: str | None = None
    capacity_units: float = 0.0
    current_load: float = 0.0
    load_ratio: float = Field(ge=0.0)
    self_reported_latency_ms: float | None = None
    traffic_share: float = Field(default=0.0, ge=0.0)
    status: str = "ok"
    physical_latency_ms: float | None = None
    physical_baseline_latency_ms: float | None = None


class CongestionResponse(BaseModel):
    link_id: str
    predicted_congestion_penalty_ms: float
    saturation_probability: float
    confidence: float
    model_name: str


class TrustResponse(BaseModel):
    link_id: str
    trust_score: float
    confidence: float
    reason: str
    model_name: str


class TargetingResponse(BaseModel):
    link_id: str
    targeting_risk_score: float
    confidence: float
    reason: str
    model_name: str


class CombinedResponse(BaseModel):
    link_id: str
    congestion: CongestionResponse
    trust: TrustResponse
    targeting: TargetingResponse
    model_versions: dict[str, str]
