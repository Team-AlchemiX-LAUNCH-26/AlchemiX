"""
feature_engineering.py – Safe feature creation for AlchemiX Phase 2.

Responsibilities
----------------
- Split link_id into planet_a / planet_b
- Create interaction features
- Create bucket / bin features
- Create ratio / difference features
- Derive trust_score target from telemetry

Forbidden in this module
------------------------
- Historical averages computed over the full dataset (those go in training.py post-split)
- Any statistic that requires knowledge of other rows (relative_traffic removed — leakage risk)
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

    The original link_id column is kept for reference and for join operations
    in the historical feature step (training.py).
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
    Create safe row-wise features from the cleaned traffic dataset.

    Features added:
    - planet_a, planet_b          (from link_id split)
    - near_capacity               = load_ratio >= 0.80 (int flag)
    - load_bucket                 = pd.cut(load_ratio, 5 bins)
    - load_ratio_sq               = load_ratio ** 2  (nonlinear signal)
    - load_units_x_ratio          = load_units × load_ratio  (interaction)

    Intentionally NOT created here (moved post-split to training.py):
    - hist_mean_load_ratio, hist_std_load_ratio, hist_mean_load_units

    Intentionally removed (zero-variance / leakage in this dataset):
    - load_percentage  → exact linear duplicate of load_ratio (×100)
    - high_congestion  → threshold 0.90 has only 3 rows (0.05% of data)
    - status_ok        → 5997/6000 rows are 'ok'; only 3 are 'saturated'
                         which are also removed as missing-target rows
    - relative_traffic → would require cross-row mean → data leakage risk

    Verified ranges (on raw data):
    - load_ratio: 0.003–0.944
    - near_capacity = 1:  16 rows (0.3%)
    - load_bucket 'very_high' (>0.8): 16 rows
    """
    df = split_link_id(df)

    # Flag: link is near its capacity limit
    df["near_capacity"] = (df["load_ratio"] >= 0.80).astype(int)

    # Bucket the load into 5 qualitative bands
    load_bins   = [0.0, 0.2, 0.4, 0.6, 0.8, 1.0]
    load_labels = ["very_low", "low", "medium", "high", "very_high"]
    df["load_bucket"] = pd.cut(
        df["load_ratio"],
        bins=load_bins,
        labels=load_labels,
        include_lowest=True,
    )

    # Polynomial and interaction terms
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
    Create safe row-wise features from the cleaned telemetry dataset.
    Also derives the trust_score regression target.

    trust_score formula
    -------------------
    latency_difference = measured_latency_ms - self_reported_latency_ms
    positive_diff      = max(0, latency_difference)  (only penalise under-reporters)
    trust_score        = 1 / (1 + positive_diff / measured_latency_ms)
        → 1.0 = perfectly honest link (self_reported >= measured)
        → ~0.5 = link under-reports by ~100% of true latency
        → approaches 0 = pathologically dishonest link

    Verified on cleaned data:
    - trust_score range: 0.53–1.0
    - mean trust_score: 0.95 (most links are honest most of the time)
    - 2695 / 6000 rows have diff < 0 → trust = 1.0

    Features added:
    - planet_a, planet_b
    - latency_difference          = measured − self_reported
    - absolute_difference         = |latency_difference|
    - latency_ratio               = measured / self_reported
    - percentage_error            = |diff| / measured × 100
    - trust_score                 ← derived regression TARGET
    """
    df = split_link_id(df)

    df["latency_difference"]  = df["measured_latency_ms"] - df["self_reported_latency_ms"]
    df["absolute_difference"] = df["latency_difference"].abs()

    # Clip measured to avoid division by zero (should never occur after cleaning)
    measured_safe = df["measured_latency_ms"].clip(lower=1e-6)
    self_safe     = df["self_reported_latency_ms"].clip(lower=1e-6)

    df["latency_ratio"]    = df["measured_latency_ms"] / self_safe
    df["percentage_error"] = (df["absolute_difference"] / measured_safe * 100).round(2)

    # Trust score — penalise only positive difference (link under-reporting)
    positive_diff = df["latency_difference"].clip(lower=0)
    df["trust_score"] = (1.0 / (1.0 + positive_diff / measured_safe)).round(4)

    logger.info(
        "engineer_telemetry_features: %d features, %d rows",
        len(df.columns), len(df),
    )
    return df


# ── Incident features ─────────────────────────────────────────────────────────

def engineer_incident_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Create safe row-wise features from the cleaned incident dataset.

    Features added:
    - planet_a, planet_b
    - traffic_percentage   = traffic_share × 100  (human-readable scale)
    - traffic_bucket       = pd.cut(traffic_share, 5 bins)
    - high_traffic_share   = traffic_share >= 0.20  (int flag)
    - jammed_int           = int(jammed_flag)  (convenience for plots/metrics)

    Intentionally NOT created here (data leakage risk):
    - relative_traffic: would require dividing by full-dataset mean before split.
      Moved to training.py post-split as a historical feature instead.

    Verified ranges (on raw data):
    - traffic_share: 0.00001–0.504
    - high_traffic_share = 1:  501 rows (8.4%)
    - jammed_flag True:  493/6000 (8.2% positive class)
    - traffic_bucket distribution:
        very_low  (<0.05):  2487 rows
        low    (0.05–0.10): 1697 rows
        medium (0.10–0.20): 1315 rows
        high   (0.20–0.35):  458 rows
        very_high (>0.35):    43 rows
    """
    df = split_link_id(df)

    # Percentage scale (for readability in plots and reports)
    df["traffic_percentage"] = (df["traffic_share"] * 100).round(2)

    # Categorical bucket
    traffic_bins   = [0.0, 0.05, 0.10, 0.20, 0.35, 1.0]
    traffic_labels = ["very_low", "low", "medium", "high", "very_high"]
    df["traffic_bucket"] = pd.cut(
        df["traffic_share"],
        bins=traffic_bins,
        labels=traffic_labels,
        include_lowest=True,
    )

    # Binary flag: high-traffic links are more attractive Chimera targets
    df["high_traffic_share"] = (df["traffic_share"] >= 0.20).astype(int)

    # Integer target for sklearn metrics and matplotlib
    df["jammed_int"] = df["jammed_flag"].astype(int)

    logger.info(
        "engineer_incident_features: %d features, %d rows",
        len(df.columns), len(df),
    )
    return df
