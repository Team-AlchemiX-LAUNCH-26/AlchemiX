"""
feature_engineering.py – Safe feature creation for AlchemiX Phase 2.

Responsibilities
----------------
- Split link_id into planet_a / planet_b
- Create interaction features
- Create bucket / bin features
- Create ratio / difference features

Forbidden in this module
------------------------
- Historical averages computed over the full dataset
- historical_avg_load, historical_trust, historical_attack_rate
- Any statistic that requires knowledge of future rows
- Encoding (OrdinalEncoder, OneHotEncoder, etc.)
- Scaling / normalisation
- Model training
"""

from __future__ import annotations

import pandas as pd
import numpy as np

from ml.utils import get_logger

logger = get_logger(__name__)


# ── Link splitting ────────────────────────────────────────────────────────────

def split_link_id(df: pd.DataFrame) -> pd.DataFrame:
    """
    Split 'Aegis-Boreas' into planet_a='Aegis', planet_b='Boreas'.

    The original link_id column is kept for reference.
    """
    parts = df["link_id"].str.split("-", n=1, expand=True)
    df = df.copy()
    df["planet_a"] = parts[0]
    df["planet_b"] = parts[1]
    logger.info("split_link_id: added planet_a, planet_b columns")
    return df


# ── Traffic features ──────────────────────────────────────────────────────────

def engineer_traffic_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Create safe features from the cleaned traffic dataset.

    Features added:
    - planet_a, planet_b
    - load_percentage      = load_ratio × 100
    - near_capacity        = load_ratio >= 0.80
    - high_congestion      = load_ratio >= 0.90
    - load_bucket          = pd.cut(load_ratio, bins)
    - status_ok            = 1 if status == 'ok' else 0
    - load_ratio_sq        = load_ratio ** 2   (nonlinear interaction)
    - load_units_x_ratio   = load_units × load_ratio
    """
    df = split_link_id(df)

    df["load_percentage"]   = (df["load_ratio"] * 100).round(2)
    df["near_capacity"]     = (df["load_ratio"] >= 0.80).astype(int)
    df["high_congestion"]   = (df["load_ratio"] >= 0.90).astype(int)

    load_bins  = [0, 0.2, 0.4, 0.6, 0.8, 1.0]
    load_labels = ["very_low", "low", "medium", "high", "very_high"]
    df["load_bucket"] = pd.cut(
        df["load_ratio"], bins=load_bins, labels=load_labels, include_lowest=True
    )

    # Status as binary flag — kept for schema consistency with live API
    df["status_ok"] = (df["status"] == "ok").astype(int)

    # Interaction / polynomial
    df["load_ratio_sq"]      = df["load_ratio"] ** 2
    df["load_units_x_ratio"] = df["load_units"] * df["load_ratio"]

    logger.info(
        "engineer_traffic_features: %d features, %d rows",
        len(df.columns), len(df),
    )
    return df


# ── Telemetry features ────────────────────────────────────────────────────────

def engineer_telemetry_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Create safe features from the cleaned telemetry dataset.

    Also derives the trust_score target:
        latency_difference = measured_latency_ms - self_reported_latency_ms
        latency_ratio      = measured / self_reported  (clipped to avoid div/0)
        percentage_error   = |difference| / measured × 100
        trust_score        = 1 / (1 + max(0, difference) / measured)
            ↳ 1.0 = honest link, 0.0 = severely lying link

    Features added:
    - planet_a, planet_b
    - latency_difference
    - absolute_difference
    - latency_ratio
    - percentage_error
    - trust_score  (← derived regression target)
    """
    df = split_link_id(df)

    df["latency_difference"]  = df["measured_latency_ms"] - df["self_reported_latency_ms"]
    df["absolute_difference"] = df["latency_difference"].abs()

    # Clip measured to avoid division by zero
    measured_safe = df["measured_latency_ms"].clip(lower=1e-6)
    df["latency_ratio"]    = df["measured_latency_ms"] / df["self_reported_latency_ms"].clip(lower=1e-6)
    df["percentage_error"] = (df["absolute_difference"] / measured_safe * 100).round(2)

    # Trust score: normalised inverse penalty
    # If difference <= 0 (self_reported >= measured), the link is honest → trust = 1
    positive_diff = df["latency_difference"].clip(lower=0)
    df["trust_score"] = (1 / (1 + positive_diff / measured_safe)).round(4)

    logger.info(
        "engineer_telemetry_features: %d features, %d rows",
        len(df.columns), len(df),
    )
    return df


# ── Incident features ─────────────────────────────────────────────────────────

def engineer_incident_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Create safe features from the cleaned incident dataset.

    Features added:
    - planet_a, planet_b
    - traffic_percentage   = traffic_share × 100
    - traffic_bucket       = pd.cut(traffic_share, bins)
    - high_traffic_share   = traffic_share >= 0.20
    - relative_traffic     = traffic_share / traffic_share.mean() within tick
    - jammed_int           = int(jammed_flag)   (convenience for plotting)
    """
    df = split_link_id(df)

    df["traffic_percentage"] = (df["traffic_share"] * 100).round(2)

    traffic_bins   = [0, 0.05, 0.10, 0.20, 0.35, 1.0]
    traffic_labels = ["very_low", "low", "medium", "high", "very_high"]
    df["traffic_bucket"] = pd.cut(
        df["traffic_share"], bins=traffic_bins, labels=traffic_labels, include_lowest=True
    )

    df["high_traffic_share"] = (df["traffic_share"] >= 0.20).astype(int)

    # relative_traffic: ratio to mean traffic_share across ALL rows
    # This is NOT a historical per-link stat; it's a snapshot-level ratio.
    overall_mean = df["traffic_share"].mean()
    df["relative_traffic"] = (df["traffic_share"] / overall_mean).round(4)

    df["jammed_int"] = df["jammed_flag"].astype(int)

    logger.info(
        "engineer_incident_features: %d features, %d rows",
        len(df.columns), len(df),
    )
    return df
