from __future__ import annotations

import os
import sys
from pathlib import Path

import numpy as np
from fastapi import FastAPI

TRAINING_DIR = Path(__file__).resolve().parents[1]
ROOT = TRAINING_DIR.parent
sys.path.insert(0, str(TRAINING_DIR))

from service.feature_builder import LiveFeatureBuilder
from service.model_loader import PythonModelStore
from service.schemas import (
    CombinedResponse,
    CongestionResponse,
    LinkObservation,
    TargetingResponse,
    TrustResponse,
)

MODEL_DIR = Path(os.environ.get("MODEL_DIR", ROOT / "internal" / "models"))

store = PythonModelStore(MODEL_DIR)
features = LiveFeatureBuilder(store.link_id_map())
app = FastAPI(title="AlchemiX ML Prediction Service")


def _clip(value: float, lower: float = 0.0, upper: float = 1.0) -> float:
    return float(min(upper, max(lower, value)))


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/model-info")
def model_info() -> dict[str, object]:
    return store.model_info()


@app.post("/predict/congestion", response_model=CongestionResponse)
def predict_congestion(obs: LinkObservation) -> CongestionResponse:
    x = features.congestion(obs)
    estimator = store.congestion["estimator"]
    penalty = float(np.clip(estimator.predict(x)[0], 0.0, None))
    sat_prob = 1.0 if obs.status == "saturated" or obs.load_ratio >= 0.90 else 0.0
    if sat_prob == 0.0 and obs.load_ratio > 0.80:
        sat_prob = (obs.load_ratio - 0.80) / 0.10
    return CongestionResponse(
        link_id=obs.link_id,
        predicted_congestion_penalty_ms=penalty,
        saturation_probability=_clip(sat_prob),
        confidence=0.95 if sat_prob == 1.0 else _clip(1.0 - sat_prob * 0.5),
        model_name=store.congestion["model_name"],
    )


@app.post("/predict/trust", response_model=TrustResponse)
def predict_trust(obs: LinkObservation) -> TrustResponse:
    x = features.trust(obs)
    estimator = store.trust["estimator"]
    ratio = _clip(float(estimator.predict(x)[0]))
    scale = float(store.trust.get("trust_scale", 0.50)) or 0.50
    trust_score = _clip(1.0 - ratio / scale)
    reason = "Reported latency is plausible against expected latency."
    if ratio > float(store.trust.get("spoof_threshold", 0.15)):
        reason = "High probability of latency under-reporting detected."
    return TrustResponse(
        link_id=obs.link_id,
        trust_score=trust_score,
        confidence=0.40 if obs.self_reported_latency_ms is None else 0.90,
        reason=reason,
        model_name=store.trust["model_name"],
    )


@app.post("/predict/targeting", response_model=TargetingResponse)
def predict_targeting(obs: LinkObservation) -> TargetingResponse:
    x = features.targeting(obs)
    estimator = store.targeting["estimator"]
    if hasattr(estimator, "predict_proba"):
        score = float(estimator.predict_proba(x)[0, 1])
    else:
        score = _clip(float(estimator.predict(x)[0]))
    reason = "Traffic-share pattern is within normal targeting range."
    if score > 0.50:
        reason = "High structural targeting risk based on traffic patterns."
    return TargetingResponse(
        link_id=obs.link_id,
        targeting_risk_score=_clip(score),
        confidence=0.85,
        reason=reason,
        model_name=store.targeting["model_name"],
    )


@app.post("/predict/all", response_model=CombinedResponse)
def predict_all(obs: LinkObservation) -> CombinedResponse:
    congestion = predict_congestion(obs)
    trust = predict_trust(obs)
    targeting = predict_targeting(obs)
    features.add_observation(obs)
    return CombinedResponse(
        link_id=obs.link_id,
        congestion=congestion,
        trust=trust,
        targeting=targeting,
        model_versions={
            "congestion": congestion.model_name,
            "trust": trust.model_name,
            "targeting": targeting.model_name,
        },
    )
