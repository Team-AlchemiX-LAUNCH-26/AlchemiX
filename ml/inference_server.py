"""
inference_server.py – FastAPI inference server for AlchemiX Phase 2.

Exposes a single POST /predict endpoint that accepts InferenceRequest JSON
and returns all three model scores in one response.

Usage
-----
    python -m ml.inference_server          # default port 7070
    AI_SERVER_PORT=8888 python -m ml.inference_server

The Go AI Agent connects to this server via ai/inference.go.

Model loading
-------------
Models are loaded ONCE at startup from the models/ directory.
No retraining happens at inference time.
"""

from __future__ import annotations

import json
import os
import sys
import pathlib
import logging
from typing import Annotated

import joblib
import numpy as np
import pandas as pd
import uvicorn
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field

# ── Project root on sys.path ──────────────────────────────────────────────────
PROJECT_ROOT = pathlib.Path(__file__).resolve().parents[1]
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

from ml.utils import MODELS_DIR

# ── Logging ──────────────────────────────────────────────────────────────────
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s  %(levelname)-8s  %(name)s  %(message)s",
    datefmt="%H:%M:%S",
)
log = logging.getLogger("inference_server")

# ── Load models once at startup ───────────────────────────────────────────────
def _load(name: str):
    path = MODELS_DIR / name
    if not path.exists():
        raise FileNotFoundError(
            f"Model not found: {path}. "
            "Run notebooks 04 and 06 first to train and export the models."
        )
    return joblib.load(path)


log.info("Loading models from %s ...", MODELS_DIR)
congestion_model     = _load("congestion.joblib")
trust_model          = _load("trust.joblib")
targeting_model      = _load("targeting.joblib")
preprocessor_cong    = _load("preprocessor_congestion.joblib")
preprocessor_trust   = _load("preprocessor_trust.joblib")
preprocessor_target  = _load("preprocessor_targeting.joblib")
model_meta           = json.loads((MODELS_DIR / "model_metadata.json").read_text())
log.info("All models loaded ✓")

CONG_NUMERIC   = model_meta["congestion"]["numeric"]
CONG_CAT       = model_meta["congestion"]["categorical"]
TRUST_NUMERIC  = model_meta["trust"]["numeric"]
TRUST_CAT      = model_meta["trust"]["categorical"]
INC_NUMERIC    = model_meta["targeting"]["numeric"]
INC_CAT        = model_meta["targeting"]["categorical"]

# Trained historical statistics for fallback defaults
# (these are global means from training — safe to use at inference time).
HIST_DEFAULTS = {
    "hist_mean_load_ratio":    0.35,
    "hist_std_load_ratio":     0.20,
    "hist_mean_load_units":    60.0,
    "hist_mean_trust":         0.90,
    "hist_std_trust":          0.10,
    "hist_mean_latency_diff":  5_000.0,
    "hist_attack_rate":        0.10,
}


# ── Pydantic schemas ──────────────────────────────────────────────────────────

class InferenceRequest(BaseModel):
    link_id: str
    load_units: float = Field(default=50.0)
    load_ratio: float = Field(default=0.30, ge=0.0, le=1.0)
    status: str = Field(default="ok")
    self_reported_latency_ms: float = Field(default=100_000.0)
    measured_latency_ms: float = Field(default=100_000.0)
    traffic_share: float = Field(default=0.083, ge=0.0, le=1.0)


class InferenceResponse(BaseModel):
    link_id: str
    congestion_penalty_ms: float
    trust_score: float
    trust_penalty_ms: float
    jam_probability: float
    targeting_penalty_ms: float


# ── Feature builders ──────────────────────────────────────────────────────────

def _build_congestion_features(req: InferenceRequest) -> pd.DataFrame:
    planet_a, planet_b = req.link_id.split("-", 1)
    row = {
        "load_units":            req.load_units,
        "load_ratio":            req.load_ratio,
        "load_percentage":       req.load_ratio * 100,
        "load_ratio_sq":         req.load_ratio ** 2,
        "load_units_x_ratio":    req.load_units * req.load_ratio,
        "near_capacity":         int(req.load_ratio >= 0.80),
        "high_congestion":       int(req.load_ratio >= 0.90),
        "status_ok":             int(req.status == "ok"),
        "hist_mean_load_ratio":  HIST_DEFAULTS["hist_mean_load_ratio"],
        "hist_std_load_ratio":   HIST_DEFAULTS["hist_std_load_ratio"],
        "hist_mean_load_units":  HIST_DEFAULTS["hist_mean_load_units"],
        "planet_a": planet_a,
        "planet_b": planet_b,
    }
    return pd.DataFrame([row])[CONG_NUMERIC + CONG_CAT]


def _build_trust_features(req: InferenceRequest) -> pd.DataFrame:
    planet_a, planet_b = req.link_id.split("-", 1)
    diff    = req.measured_latency_ms - req.self_reported_latency_ms
    abs_diff = abs(diff)
    safe_meas = max(req.measured_latency_ms, 1e-6)
    safe_self = max(req.self_reported_latency_ms, 1e-6)
    row = {
        "self_reported_latency_ms": req.self_reported_latency_ms,
        "measured_latency_ms":      req.measured_latency_ms,
        "latency_difference":       diff,
        "absolute_difference":      abs_diff,
        "latency_ratio":            req.measured_latency_ms / safe_self,
        "percentage_error":         abs_diff / safe_meas * 100,
        "hist_mean_trust":          HIST_DEFAULTS["hist_mean_trust"],
        "hist_std_trust":           HIST_DEFAULTS["hist_std_trust"],
        "hist_mean_latency_diff":   HIST_DEFAULTS["hist_mean_latency_diff"],
        "planet_a": planet_a,
        "planet_b": planet_b,
    }
    return pd.DataFrame([row])[TRUST_NUMERIC + TRUST_CAT]


def _build_targeting_features(req: InferenceRequest) -> pd.DataFrame:
    planet_a, planet_b = req.link_id.split("-", 1)
    overall_mean = 1.0 / 12  # approx mean for 12 links
    row = {
        "traffic_share":       req.traffic_share,
        "traffic_percentage":  req.traffic_share * 100,
        "high_traffic_share":  int(req.traffic_share >= 0.20),
        "relative_traffic":    req.traffic_share / max(overall_mean, 1e-9),
        "hist_attack_rate":    HIST_DEFAULTS["hist_attack_rate"],
        "planet_a": planet_a,
        "planet_b": planet_b,
    }
    return pd.DataFrame([row])[INC_NUMERIC + INC_CAT]


# ── FastAPI app ───────────────────────────────────────────────────────────────

app = FastAPI(
    title="AlchemiX ML Inference Server",
    description="Serves the three Phase 2 ML models to the Go AI Agent.",
    version="1.0.0",
)


@app.get("/health")
def health():
    return {"status": "ok", "models": ["congestion", "trust", "targeting"]}


@app.post("/predict", response_model=InferenceResponse)
def predict(req: InferenceRequest) -> InferenceResponse:
    """
    Run all three models for a single link and return the combined scores.

    Called once per link in the route by the Go ai/inference.go client.
    """
    try:
        # ── Model 1: Congestion ───────────────────────────────────────────────
        X_cong = preprocessor_cong.transform(_build_congestion_features(req))
        congestion_penalty_ms = float(congestion_model.predict(X_cong)[0])

        # ── Model 2: Trust ────────────────────────────────────────────────────
        X_trust = preprocessor_trust.transform(_build_trust_features(req))
        trust_score = float(np.clip(trust_model.predict(X_trust)[0], 0.0, 1.0))
        trust_penalty_ms = (1.0 - trust_score) * 50_000

        # ── Model 3: Targeting ────────────────────────────────────────────────
        X_target = preprocessor_target.transform(_build_targeting_features(req))
        jam_probability = float(
            np.clip(targeting_model.predict_proba(X_target)[0][1], 0.0, 1.0)
        )
        targeting_penalty_ms = jam_probability * 200_000

    except Exception as exc:
        log.exception("Prediction failed for link %s", req.link_id)
        raise HTTPException(status_code=500, detail=str(exc))

    log.info(
        "link=%-20s  cong=+%7.0f ms  trust=%.3f  jam=%.3f",
        req.link_id, congestion_penalty_ms, trust_score, jam_probability,
    )

    return InferenceResponse(
        link_id=req.link_id,
        congestion_penalty_ms=round(congestion_penalty_ms, 2),
        trust_score=round(trust_score, 4),
        trust_penalty_ms=round(trust_penalty_ms, 2),
        jam_probability=round(jam_probability, 4),
        targeting_penalty_ms=round(targeting_penalty_ms, 2),
    )


# ── Entry point ───────────────────────────────────────────────────────────────
if __name__ == "__main__":
    port = int(os.environ.get("AI_SERVER_PORT", 7070))
    log.info("Starting inference server on port %d ...", port)
    uvicorn.run(app, host="0.0.0.0", port=port, log_level="warning")
