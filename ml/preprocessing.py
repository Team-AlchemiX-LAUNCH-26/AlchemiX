"""
preprocessing.py – Data cleaning functions for AlchemiX Phase 2.

Responsibilities
----------------
- Duplicate removal
- KNN imputation for non-target columns
- Missing target removal
- Schema validation
- Save cleaned CSV files

Forbidden in this module
------------------------
- Feature engineering
- Encoding
- Scaling
- Model training
- Computing historical statistics
"""

from __future__ import annotations

import pandas as pd
import numpy as np
from sklearn.impute import KNNImputer

from ml.utils import get_logger, RANDOM_SEED

logger = get_logger(__name__)

# ── Configuration ─────────────────────────────────────────────────────────────
KNN_NEIGHBORS: int = 5
SATURATION_STATUS: str = "saturated"

EXPECTED_TRAFFIC_COLS  = {"link_id", "tick", "load_units", "load_ratio", "status", "observed_latency_ms"}
EXPECTED_TELEMETRY_COLS = {"link_id", "tick", "self_reported_latency_ms", "measured_latency_ms"}
EXPECTED_INCIDENT_COLS  = {"link_id", "tick", "traffic_share", "jammed_flag"}


# ── Schema validation ─────────────────────────────────────────────────────────
def validate_schema(df: pd.DataFrame, expected_cols: set[str], name: str) -> None:
    """Raise ValueError if any expected column is missing."""
    missing = expected_cols - set(df.columns)
    if missing:
        raise ValueError(f"[{name}] Missing columns: {missing}")
    logger.info("[%s] Schema OK – %d rows, %d columns", name, len(df), len(df.columns))


# ── Duplicate removal ─────────────────────────────────────────────────────────
def remove_duplicates(df: pd.DataFrame, subset: list[str], name: str) -> pd.DataFrame:
    """
    Remove duplicate rows based on key columns.

    NOTE: `tick` must be included in the subset to avoid removing legitimate
    observations that share the same link_id but occur at different times.
    """
    before = len(df)
    df = df.drop_duplicates(subset=subset, keep="first")
    removed = before - len(df)
    logger.info("[%s] Duplicates removed: %d (rows remaining: %d)", name, removed, len(df))
    return df.reset_index(drop=True)


# ── KNN Imputation ────────────────────────────────────────────────────────────
def knn_impute_numeric(
    df: pd.DataFrame,
    numeric_cols: list[str],
    n_neighbors: int = KNN_NEIGHBORS,
    name: str = "",
) -> pd.DataFrame:
    """
    Apply KNN imputation to the given numeric columns.

    Categorical and non-numeric columns are excluded from imputation.
    Only rows with missing values in `numeric_cols` are affected.
    """
    missing_before = df[numeric_cols].isna().sum().sum()
    if missing_before == 0:
        logger.info("[%s] No missing values in numeric columns – skipping KNN.", name)
        return df

    imputer = KNNImputer(n_neighbors=n_neighbors)
    df_copy = df.copy()
    df_copy[numeric_cols] = imputer.fit_transform(df_copy[numeric_cols])
    missing_after = df_copy[numeric_cols].isna().sum().sum()
    logger.info(
        "[%s] KNN imputation: %d → %d missing values in %s",
        name, missing_before, missing_after, numeric_cols,
    )
    return df_copy


# ── Missing target removal ────────────────────────────────────────────────────
def drop_missing_target(df: pd.DataFrame, target_col: str, name: str) -> pd.DataFrame:
    """
    Drop rows where the target column is null.

    For the traffic dataset, `observed_latency_ms` is null when the link is
    saturated — those rows are intentionally removed.
    """
    before = len(df)
    df = df.dropna(subset=[target_col]).reset_index(drop=True)
    removed = before - len(df)
    logger.info(
        "[%s] Rows removed due to missing target '%s': %d (rows remaining: %d)",
        name, target_col, removed, len(df),
    )
    return df


# ── Missing value report ──────────────────────────────────────────────────────
def missing_value_report(df: pd.DataFrame, name: str) -> pd.DataFrame:
    """Return and log a summary of missing values per column."""
    report = (
        df.isna().sum()
          .rename("missing_count")
          .to_frame()
    )
    report["missing_pct"] = (report["missing_count"] / len(df) * 100).round(2)
    report = report[report["missing_count"] > 0]
    if report.empty:
        logger.info("[%s] No missing values remaining.", name)
    else:
        logger.info("[%s] Missing values:\n%s", name, report.to_string())
    return report


# ── Dataset-specific cleaning pipelines ───────────────────────────────────────

def clean_traffic(df: pd.DataFrame) -> pd.DataFrame:
    """
    Clean the link traffic history dataset.

    Steps:
    1. Validate schema
    2. Remove duplicates on (link_id, tick)
    3. KNN impute load_units and load_ratio  (NOT observed_latency_ms)
    4. Remove rows where observed_latency_ms is null
    5. Keep status column as-is (schema consistency)
    """
    name = "traffic"
    validate_schema(df, EXPECTED_TRAFFIC_COLS, name)
    df = remove_duplicates(df, subset=["link_id", "tick"], name=name)

    # Impute only the input features – never the target
    impute_cols = ["load_units", "load_ratio"]
    df = knn_impute_numeric(df, impute_cols, name=name)

    # Remove rows with missing target
    df = drop_missing_target(df, "observed_latency_ms", name=name)

    missing_value_report(df, name)
    logger.info("[%s] Cleaning complete – %d rows", name, len(df))
    return df


def clean_telemetry(df: pd.DataFrame) -> pd.DataFrame:
    """
    Clean the link telemetry dataset.

    Steps:
    1. Validate schema
    2. Remove duplicates on (link_id, tick)
    3. KNN impute self_reported_latency_ms
    4. Remove rows where measured_latency_ms is null (ground truth)
    """
    name = "telemetry"
    validate_schema(df, EXPECTED_TELEMETRY_COLS, name)
    df = remove_duplicates(df, subset=["link_id", "tick"], name=name)

    impute_cols = ["self_reported_latency_ms"]
    df = knn_impute_numeric(df, impute_cols, name=name)

    df = drop_missing_target(df, "measured_latency_ms", name=name)

    missing_value_report(df, name)
    logger.info("[%s] Cleaning complete – %d rows", name, len(df))
    return df


def clean_incident(df: pd.DataFrame) -> pd.DataFrame:
    """
    Clean the link incident history dataset.

    Steps:
    1. Validate schema
    2. Remove duplicates on (link_id, tick)
    3. KNN impute traffic_share
    4. Keep jammed_flag rows (boolean, should have no nulls)
    5. Ensure jammed_flag is boolean
    """
    name = "incident"
    validate_schema(df, EXPECTED_INCIDENT_COLS, name)
    df = remove_duplicates(df, subset=["link_id", "tick"], name=name)

    impute_cols = ["traffic_share"]
    df = knn_impute_numeric(df, impute_cols, name=name)

    # Ensure target is boolean
    df["jammed_flag"] = df["jammed_flag"].astype(bool)
    drop_missing_target(df, "jammed_flag", name=name)

    missing_value_report(df, name)
    logger.info("[%s] Cleaning complete – %d rows", name, len(df))
    return df
