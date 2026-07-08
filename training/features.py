"""
Leakage-safe feature engineering for the AlchemiX ML pipeline.

Golden rule: any feature at tick t uses only information from ticks < t.
This is enforced by calling shift(1) before any rolling or expanding window.

Do NOT call these functions before sorting by [link_id, tick].
"""
from __future__ import annotations

import numpy as np
import pandas as pd


def _lagged_rolling_mean(series: pd.Series, window: int) -> pd.Series:
    """Mean of [t-window, t-1] — never includes tick t."""
    return series.shift(1).rolling(window, min_periods=1).mean()


def _lagged_rolling_std(series: pd.Series, window: int) -> pd.Series:
    """Std of [t-window, t-1]."""
    return series.shift(1).rolling(window, min_periods=2).std().fillna(0.0)


def _lagged_rolling_sum(series: pd.Series, window: int) -> pd.Series:
    """Sum of [t-window, t-1]."""
    return series.shift(1).rolling(window, min_periods=1).sum()


def _lagged_expanding_mean(series: pd.Series) -> pd.Series:
    """Expanding mean over all rows before tick t."""
    return series.shift(1).expanding(min_periods=1).mean()


# ---------------------------------------------------------------------------
# Congestion features
# ---------------------------------------------------------------------------

def build_congestion_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Build leakage-safe congestion features.
    Input df must already have: load_ratio, load_units, capacity_units,
    physical_latency_ms, link_id_encoded, load_units_missing.
    Sorted by [link_id, tick].
    """
    df = df.copy()

    # Previous tick's load ratio (already 1-tick lagged by definition).
    df["previous_load_ratio"] = df.groupby("link_id")["load_ratio"].shift(1)
    # For tick 0 of each link, fill with current load_ratio (no history available).
    df["previous_load_ratio"] = df["previous_load_ratio"].fillna(df["load_ratio"])

    df["load_ratio_change"] = df["load_ratio"] - df["previous_load_ratio"]

    # Rolling mean of load_ratio using only past ticks.
    df["rolling_mean_load_ratio"] = (
        df.groupby("link_id")["load_ratio"]
        .transform(lambda x: _lagged_rolling_mean(x, window=5))
    )

    # Rolling std of load_ratio.
    df["rolling_std_load_ratio"] = (
        df.groupby("link_id")["load_ratio"]
        .transform(lambda x: _lagged_rolling_std(x, window=5))
    )

    # Rate of load increase: diff of the lagged rolling mean.
    df["rate_of_load_increase"] = (
        df.groupby("link_id")["rolling_mean_load_ratio"]
        .transform(lambda x: x.diff().fillna(0))
    )

    df["distance_to_saturation"] = (0.90 - df["load_ratio"]).clip(lower=0.0)

    return df


CONGESTION_FEATURE_COLS = [
    "load_ratio",
    "load_units",
    "capacity_units",
    "link_id_encoded",
    "previous_load_ratio",
    "load_ratio_change",
    "rolling_mean_load_ratio",
    "rolling_std_load_ratio",
    "rate_of_load_increase",
    "distance_to_saturation",
    "load_units_missing",
]


# ---------------------------------------------------------------------------
# Trust features
# ---------------------------------------------------------------------------

def build_trust_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Build leakage-safe trust features.
    Input df must already have: self_reported_latency_ms, physical_latency_ms,
    link_id_encoded, self_reported_latency_missing, under_report_ratio.
    Sorted by [link_id, tick].
    """
    df = df.copy()

    phys = df["physical_latency_ms"].replace(0, np.nan)

    df["self_to_physical_ratio"] = df["self_reported_latency_ms"] / phys
    df["deviation_from_baseline"] = df["self_reported_latency_ms"] - df["physical_latency_ms"]

    # Lagged self-reported (previous tick).
    df["prev_self_reported"] = df.groupby("link_id")["self_reported_latency_ms"].shift(1)
    df["prev_self_reported"] = df["prev_self_reported"].fillna(df["self_reported_latency_ms"])

    df["self_reported_change"] = df["self_reported_latency_ms"] - df["prev_self_reported"]

    # Rolling mean of self_reported_latency_ms using only past ticks.
    df["rolling_self_reported"] = (
        df.groupby("link_id")["self_reported_latency_ms"]
        .transform(lambda x: _lagged_rolling_mean(x, window=5))
    )

    # Historical under-reporting bias: expanding mean of past under_report_ratio.
    # Critically uses shift(1) — no leakage of current tick's ratio.
    df["historical_bias"] = (
        df.groupby("link_id")["under_report_ratio"]
        .transform(lambda x: _lagged_expanding_mean(x).fillna(0))
    )

    df["below_physical_min"] = (
        df["self_reported_latency_ms"] < df["physical_latency_ms"] * 0.95
    ).astype(float)

    return df


TRUST_FEATURE_COLS = [
    "self_reported_latency_ms",
    "self_reported_latency_missing",
    "physical_latency_ms",
    "self_to_physical_ratio",
    "deviation_from_baseline",
    "prev_self_reported",
    "self_reported_change",
    "rolling_self_reported",
    "historical_bias",
    "below_physical_min",
    "link_id_encoded",
]


# ---------------------------------------------------------------------------
# Targeting features
# ---------------------------------------------------------------------------

def build_targeting_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Build leakage-safe targeting features.
    Input df must already have: traffic_share, jammed, link_id_encoded,
    traffic_share_missing.
    Sorted by [link_id, tick].
    """
    df = df.copy()

    # Rank of traffic_share within each tick (higher share → rank 1).
    # Uses the current tick's traffic_share — this is the observable input
    # at decision time, not a future value.
    df["traffic_share_rank"] = df.groupby("tick")["traffic_share"].rank(
        ascending=False, method="min"
    )

    # Previous tick's traffic share.
    df["prev_traffic_share"] = df.groupby("link_id")["traffic_share"].shift(1)
    df["prev_traffic_share"] = df["prev_traffic_share"].fillna(df["traffic_share"])
    df["traffic_share_change"] = df["traffic_share"] - df["prev_traffic_share"]

    # Rolling mean of traffic_share using only past ticks.
    df["rolling_traffic_share"] = (
        df.groupby("link_id")["traffic_share"]
        .transform(lambda x: _lagged_rolling_mean(x, window=5))
    )

    # Historical jam rate: expanding mean over past ticks only.
    df["historical_jam_rate"] = (
        df.groupby("link_id")["jammed"]
        .transform(lambda x: _lagged_expanding_mean(x).fillna(0))
    )

    # Rolling jam frequency (last 10 ticks before current).
    df["rolling_jam_count"] = (
        df.groupby("link_id")["jammed"]
        .transform(lambda x: _lagged_rolling_sum(x, window=10))
    )

    # Consecutive high-usage: how many of the last 10 ticks was traffic_share above median.
    # Median computed from training-set context; here we use a fixed 1/n_links approximation.
    # Safer: use df["traffic_share"] lagged to avoid current-tick leakage.
    n_links = df["link_id"].nunique()
    median_share = 1.0 / max(n_links, 1)
    df["_above_median"] = (df["traffic_share"] > median_share).astype(float)
    df["consecutive_high_usage"] = (
        df.groupby("link_id")["_above_median"]
        .transform(lambda x: _lagged_rolling_sum(x, window=10))
    )
    df.drop(columns=["_above_median"], inplace=True)

    return df


TARGETING_FEATURE_COLS = [
    "traffic_share",
    "traffic_share_missing",
    "traffic_share_rank",
    "prev_traffic_share",
    "traffic_share_change",
    "rolling_traffic_share",
    "historical_jam_rate",
    "rolling_jam_count",
    "consecutive_high_usage",
    "link_id_encoded",
]
