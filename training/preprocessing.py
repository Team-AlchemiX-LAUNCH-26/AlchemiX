"""
Shared preprocessing for the AlchemiX ML pipeline.

Rules:
- All imputers are fit on the training partition only.
- Missing-value indicator columns are added before imputation.
- No blanket zero-filling.
- Canonical link_id ordering is loaded once from universe-config.json.
"""
from __future__ import annotations

import hashlib
import json
import os
from typing import Any

import numpy as np
import pandas as pd

UNIVERSE_PATH = os.path.join(
    os.path.dirname(__file__), "..", "datasets", "universe-config.json"
)


# ---------------------------------------------------------------------------
# Universe helpers
# ---------------------------------------------------------------------------

def load_universe_config(path: str = UNIVERSE_PATH) -> dict:
    with open(path, "r") as f:
        return json.load(f)


def load_physical_baselines(cfg: dict) -> dict[str, float]:
    """Compute speed-of-light latency per link from universe config."""
    nodes = {n["id"]: n for n in cfg["nodes"]}
    sol   = cfg["universe_metadata"]["speed_of_light_kms"]
    scale = cfg["universe_metadata"]["coordinate_scale_unit_km"]
    baselines: dict[str, float] = {}
    for link in cfg["interplanetary_links"]:
        a = nodes[link["planet_a"]]
        b = nodes[link["planet_b"]]
        dx = (a["x"] - b["x"]) * scale
        dy = (a["y"] - b["y"]) * scale
        dist_km = np.sqrt(dx**2 + dy**2)
        latency_ms = (dist_km / sol) * 1000.0
        baselines[link["link_id"]] = round(latency_ms, 6)
    return baselines


def load_capacities(cfg: dict) -> dict[str, float]:
    return {
        link["link_id"]: float(link["capacity_units"])
        for link in cfg["interplanetary_links"]
    }


def canonical_link_id_map(cfg: dict) -> dict[str, int]:
    """Return a stable sorted mapping of link_id → integer index."""
    link_ids = sorted(link["link_id"] for link in cfg["interplanetary_links"])
    return {lid: i for i, lid in enumerate(link_ids)}


# ---------------------------------------------------------------------------
# Dataset validation
# ---------------------------------------------------------------------------

REQUIRED_TRAFFIC_COLS = ["link_id", "tick", "load_units", "load_ratio", "status", "observed_latency_ms"]
REQUIRED_TELEMETRY_COLS = ["link_id", "tick", "self_reported_latency_ms", "measured_latency_ms"]
REQUIRED_INCIDENT_COLS  = ["link_id", "tick", "traffic_share", "jammed_flag"]


def validate_schema(df: pd.DataFrame, required_cols: list[str], name: str) -> None:
    missing = [c for c in required_cols if c not in df.columns]
    if missing:
        raise ValueError(f"[{name}] Missing required columns: {missing}")

    if df["tick"].isnull().any():
        raise ValueError(f"[{name}] Column 'tick' contains NaN values")
    if df["link_id"].isnull().any():
        raise ValueError(f"[{name}] Column 'link_id' contains NaN values")

    # Check numeric types for known numeric columns.
    for col in ["tick", "load_ratio", "observed_latency_ms",
                "self_reported_latency_ms", "measured_latency_ms",
                "traffic_share", "load_units"]:
        if col in df.columns and df[col].dtype == object:
            raise ValueError(f"[{name}] Column '{col}' is object dtype — expected numeric")

    # Range checks.
    if "load_ratio" in df.columns:
        bad = df["load_ratio"].dropna()
        if (bad < 0).any() or (bad > 1).any():
            raise ValueError(f"[{name}] load_ratio out of [0, 1]")

    if "traffic_share" in df.columns:
        bad = df["traffic_share"].dropna()
        if (bad < 0).any() or (bad > 1).any():
            raise ValueError(f"[{name}] traffic_share out of [0, 1]")


# ---------------------------------------------------------------------------
# Traffic dataset preprocessing
# ---------------------------------------------------------------------------

class TrafficPreprocessor:
    """
    Fits on training data, then transforms any partition.

    Missing-value policy:
    - load_units missing: impute as load_ratio * capacity_units (physical formula).
      If load_ratio and capacity both present, use formula; else use training-partition median.
    - observed_latency_ms missing: mark with indicator; exclude from regression targets.
      Do NOT replace target with zero.
    """

    def __init__(self) -> None:
        self._load_units_median: float | None = None
        self._baselines: dict[str, float] = {}
        self._capacities: dict[str, float] = {}
        self._link_id_map: dict[str, int] = {}

    def fit(
        self,
        train_df: pd.DataFrame,
        baselines: dict[str, float],
        capacities: dict[str, float],
        link_id_map: dict[str, int],
    ) -> "TrafficPreprocessor":
        self._baselines  = baselines
        self._capacities = capacities
        self._link_id_map = link_id_map

        # Compute training-partition median for load_units as fallback.
        valid = train_df["load_units"].dropna()
        self._load_units_median = float(valid.median()) if len(valid) > 0 else 0.0
        return self

    def transform(self, df: pd.DataFrame) -> pd.DataFrame:
        df = df.copy()

        # Attach baselines and capacities.
        df["physical_latency_ms"] = df["link_id"].map(self._baselines)
        df["capacity_units"] = df["link_id"].map(self._capacities)
        df["link_id_encoded"] = df["link_id"].map(self._link_id_map).fillna(-1).astype(int)

        # --- load_units imputation ---
        df["load_units_missing"] = df["load_units"].isnull().astype(float)

        # Prefer formula: load_units = load_ratio * capacity_units.
        formula_possible = df["load_units"].isnull() & df["capacity_units"].notna()
        df.loc[formula_possible, "load_units"] = (
            df.loc[formula_possible, "load_ratio"] *
            df.loc[formula_possible, "capacity_units"]
        )

        # Still-missing: use training median.
        df["load_units"] = df["load_units"].fillna(self._load_units_median)

        # --- observed_latency_ms: indicator only, do NOT fill target ---
        df["observed_latency_missing"] = df["observed_latency_ms"].isnull().astype(float)
        # (observed_latency_ms intentionally left NaN for rows missing it —
        #  regression target will filter these out)

        # --- is_saturated (hard rule, not model-predicted) ---
        df["is_saturated"] = (
            (df["status"] == "saturated") | (df["load_ratio"] >= 0.90)
        ).astype(int)

        return df

    def metadata(self) -> dict[str, Any]:
        return {
            "load_units_training_median": self._load_units_median,
            "missing_indicator_features": ["load_units_missing", "observed_latency_missing"],
        }


# ---------------------------------------------------------------------------
# Telemetry dataset preprocessing
# ---------------------------------------------------------------------------

class TelemetryPreprocessor:
    """
    Missing-value policy for self_reported_latency_ms:
    - Add self_reported_latency_missing = 1 indicator.
    - Impute with training-partition median (learned here, applied anywhere).
    - Never treat missing as zero.
    """

    def __init__(self) -> None:
        self._self_reported_median: float | None = None
        self._baselines: dict[str, float] = {}
        self._link_id_map: dict[str, int] = {}

    def fit(
        self,
        train_df: pd.DataFrame,
        baselines: dict[str, float],
        link_id_map: dict[str, int],
    ) -> "TelemetryPreprocessor":
        self._baselines  = baselines
        self._link_id_map = link_id_map

        valid = train_df["self_reported_latency_ms"].dropna()
        self._self_reported_median = float(valid.median()) if len(valid) > 0 else 0.0
        return self

    def transform(self, df: pd.DataFrame) -> pd.DataFrame:
        df = df.copy()

        df["physical_latency_ms"] = df["link_id"].map(self._baselines)
        df["link_id_encoded"] = df["link_id"].map(self._link_id_map).fillna(-1).astype(int)

        # --- self_reported_latency_ms ---
        df["self_reported_latency_missing"] = df["self_reported_latency_ms"].isnull().astype(float)
        df["self_reported_latency_ms"] = df["self_reported_latency_ms"].fillna(
            self._self_reported_median
        )

        # Under-reporting ratio (training target for trust model).
        # Only defined when measured_latency_ms > 0.
        measured = df["measured_latency_ms"].replace(0, np.nan)
        df["under_report_ratio"] = (
            np.maximum(0, df["measured_latency_ms"] - df["self_reported_latency_ms"])
            / measured
        ).clip(0, 1)

        return df

    def metadata(self) -> dict[str, Any]:
        return {
            "self_reported_latency_ms_training_median": self._self_reported_median,
            "missing_indicator_features": ["self_reported_latency_missing"],
        }


# ---------------------------------------------------------------------------
# Incident dataset preprocessing
# ---------------------------------------------------------------------------

class IncidentPreprocessor:
    """
    Missing-value policy for traffic_share:
    - When one value is missing per tick, reconstruct as 1 - sum(others),
      provided the result is in [0, 1].
    - Otherwise add indicator and use training-partition median.
    """

    def __init__(self) -> None:
        self._traffic_share_median: float | None = None
        self._link_id_map: dict[str, int] = {}

    def fit(
        self,
        train_df: pd.DataFrame,
        link_id_map: dict[str, int],
    ) -> "IncidentPreprocessor":
        self._link_id_map = link_id_map
        valid = train_df["traffic_share"].dropna()
        self._traffic_share_median = float(valid.median()) if len(valid) > 0 else 0.0
        return self

    def transform(self, df: pd.DataFrame) -> pd.DataFrame:
        df = df.copy()
        df["link_id_encoded"] = df["link_id"].map(self._link_id_map).fillna(-1).astype(int)

        # --- traffic_share reconstruction ---
        df["traffic_share_missing"] = df["traffic_share"].isnull().astype(float)

        # Try per-tick reconstruction when exactly one value is missing.
        missing_counts = df.groupby("tick")["traffic_share"].transform(
            lambda x: x.isnull().sum()
        )
        tick_sums = df.groupby("tick")["traffic_share"].transform("sum")

        reconstructible = (missing_counts == 1) & df["traffic_share"].isnull()
        reconstructed_value = 1.0 - tick_sums  # sum of non-missing shares
        in_range = reconstructed_value.between(0, 1)
        can_reconstruct = reconstructible & in_range

        df.loc[can_reconstruct, "traffic_share"] = reconstructed_value[can_reconstruct]

        # Remaining missing: use training-partition median.
        df["traffic_share"] = df["traffic_share"].fillna(self._traffic_share_median)

        df["jammed"] = df["jammed_flag"].astype(int)

        return df

    def metadata(self) -> dict[str, Any]:
        return {
            "traffic_share_training_median": self._traffic_share_median,
            "missing_indicator_features": ["traffic_share_missing"],
        }


# ---------------------------------------------------------------------------
# Dataset hash
# ---------------------------------------------------------------------------

def csv_sha256(path: str) -> str:
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            h.update(chunk)
    return h.hexdigest()[:16]
