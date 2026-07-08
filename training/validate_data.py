"""Validate all ML source datasets before training."""
from __future__ import annotations

import os

import pandas as pd

from preprocessing import (
    REQUIRED_INCIDENT_COLS,
    REQUIRED_TELEMETRY_COLS,
    REQUIRED_TRAFFIC_COLS,
    canonical_link_id_map,
    load_universe_config,
    validate_schema,
)

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
DATASET_DIR = os.path.join(ROOT, "datasets")
UNIVERSE_PATH = os.path.join(DATASET_DIR, "universe-config.json")

DATASETS = {
    "link_traffic_history.csv": REQUIRED_TRAFFIC_COLS,
    "link_telemetry.csv": REQUIRED_TELEMETRY_COLS,
    "link_incident_history.csv": REQUIRED_INCIDENT_COLS,
}


def _read_csv_checked(filename: str, required_cols: list[str]) -> pd.DataFrame:
    path = os.path.join(DATASET_DIR, filename)
    if not os.path.exists(path):
        raise FileNotFoundError(f"Missing required dataset: {path}")

    df = pd.read_csv(path)
    validate_schema(df, required_cols, filename)
    return df


def _validate_link_ids(df: pd.DataFrame, filename: str, valid_link_ids: set[str]) -> None:
    unknown = sorted(set(df["link_id"].dropna()) - valid_link_ids)
    if unknown:
        raise ValueError(f"[{filename}] Unknown link_id values: {unknown}")


def _validate_values(filename: str, df: pd.DataFrame) -> None:
    if not pd.api.types.is_numeric_dtype(df["tick"]):
        raise ValueError(f"[{filename}] tick must be numeric")

    if "status" in df.columns:
        allowed = {"ok", "saturated"}
        bad_status = sorted(set(df["status"].dropna()) - allowed)
        if bad_status:
            raise ValueError(f"[{filename}] status contains invalid values: {bad_status}")

    for col in ["observed_latency_ms", "self_reported_latency_ms", "measured_latency_ms"]:
        if col in df.columns:
            values = df[col].dropna()
            if (values < 0).any():
                raise ValueError(f"[{filename}] {col} contains negative values")

    if "jammed_flag" in df.columns:
        normalized = df["jammed_flag"].dropna().astype(str).str.lower()
        valid = {"true", "false", "0", "1"}
        bad = sorted(set(normalized) - valid)
        if bad:
            raise ValueError(f"[{filename}] jammed_flag is not boolean-like: {bad}")


def main() -> None:
    if not os.path.exists(UNIVERSE_PATH):
        raise FileNotFoundError(f"Missing universe config: {UNIVERSE_PATH}")

    cfg = load_universe_config(UNIVERSE_PATH)
    valid_link_ids = set(canonical_link_id_map(cfg))
    print("OK universe-config.json loaded")

    for filename, required_cols in DATASETS.items():
        df = _read_csv_checked(filename, required_cols)
        _validate_link_ids(df, filename, valid_link_ids)
        _validate_values(filename, df)
        print(f"OK {filename} valid ({len(df)} rows)")

    print("OK all link IDs match universe-config.json")


if __name__ == "__main__":
    main()
