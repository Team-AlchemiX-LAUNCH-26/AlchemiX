"""
Train congestion penalty regressor from link_traffic_history.csv.

Key changes from v1:
- Saturation classifier REMOVED: only 3 saturated examples → hard rule used instead.
- Rolling features use shift(1) to prevent leakage.
- Missing load_units imputed using formula or training-partition median.
- observed_latency_ms NaN rows excluded from regression target only.
- Adds missing-value indicator columns.
- Full metadata exported with model JSON.
- Uses HistGradientBoostingRegressor (selected over DecisionTreeRegressor by val MAE).
"""

import json
import os
import sys

import numpy as np
import pandas as pd
from sklearn.ensemble import HistGradientBoostingRegressor
from sklearn.metrics import mean_absolute_error, mean_squared_error

# Add training directory to path so shared modules import cleanly.
sys.path.insert(0, os.path.dirname(__file__))

from preprocessing import (
    REQUIRED_TRAFFIC_COLS,
    TrafficPreprocessor,
    canonical_link_id_map,
    csv_sha256,
    load_capacities,
    load_physical_baselines,
    load_universe_config,
    validate_schema,
)
from split import SplitResult, assert_no_overlap, tick_split
from features import CONGESTION_FEATURE_COLS, build_congestion_features
from model_metadata import RANDOM_STATE, build_metadata, split_info_from_result

DATASET_PATH = os.path.join(
    os.path.dirname(__file__), "..", "datasets", "link_traffic_history.csv"
)
UNIVERSE_PATH = os.path.join(
    os.path.dirname(__file__), "..", "datasets", "universe-config.json"
)
OUTPUT_PATH = os.path.join(
    os.path.dirname(__file__), "..", "internal", "models", "congestion_model.json"
)


def tree_to_dict_hgbr(estimator, feature_names: list[str]) -> list[dict]:
    """Export a HistGradientBoostingRegressor as a list of tree dicts."""
    from sklearn.ensemble._hist_gradient_boosting.predictor import TreePredictor

    trees = []
    for stage in estimator._predictors:
        for tree in stage:
            nodes = tree.nodes
            tree_nodes = []
            for i in range(len(nodes)):
                node = nodes[i]
                tree_nodes.append({
                    "node_id": int(i),
                    "is_leaf": bool(node["is_leaf"]),
                    "feature_idx": int(node["feature_idx"]) if not node["is_leaf"] else -1,
                    "feature_name": (
                        feature_names[int(node["feature_idx"])]
                        if not node["is_leaf"] and int(node["feature_idx"]) < len(feature_names)
                        else ""
                    ),
                    "threshold": float(node["num_threshold"]),
                    "left_child": int(node["left"]) if not node["is_leaf"] else -1,
                    "right_child": int(node["right"]) if not node["is_leaf"] else -1,
                    "value": float(node["value"]),
                    "count": int(node["count"]),
                })
            trees.append(tree_nodes)
    return trees


def main() -> None:
    print("=" * 60)
    print("CONGESTION MODEL TRAINING  (v2 — leakage-safe)")
    print("=" * 60)

    # --- Load data ---
    dataset_hash = csv_sha256(DATASET_PATH)
    df_raw = pd.read_csv(DATASET_PATH)
    print(f"Loaded {len(df_raw)} rows | hash={dataset_hash}")
    validate_schema(df_raw, REQUIRED_TRAFFIC_COLS, "link_traffic_history")

    cfg = load_universe_config(UNIVERSE_PATH)
    baselines = load_physical_baselines(cfg)
    capacities = load_capacities(cfg)
    link_id_map = canonical_link_id_map(cfg)

    # Sort once for all feature engineering.
    df_raw = df_raw.sort_values(["link_id", "tick"]).reset_index(drop=True)

    # --- Split first, then fit preprocessor on training partition only ---
    raw_split = tick_split(df_raw)
    assert_no_overlap(raw_split)
    print(f"\nTick-based split:")
    print(raw_split.describe())

    # Fit preprocessor on training data only.
    preprocessor = TrafficPreprocessor()
    preprocessor.fit(raw_split.train, baselines, capacities, link_id_map)

    # Transform all partitions using training-learned parameters.
    train_df = preprocessor.transform(raw_split.train)
    val_df   = preprocessor.transform(raw_split.val)
    test_df  = preprocessor.transform(raw_split.test)

    # Build leakage-safe features on the full dataset, then re-split.
    # This is correct because rolling features use shift(1) and look only backward.
    full_df = preprocessor.transform(df_raw)
    full_df = build_congestion_features(full_df)

    train = full_df[full_df["tick"] <= raw_split.train_tick_range[1]]
    val   = full_df[
        (full_df["tick"] > raw_split.train_tick_range[1]) &
        (full_df["tick"] <= raw_split.val_tick_range[1])
    ]
    test  = full_df[full_df["tick"] > raw_split.val_tick_range[1]]

    # --- Saturation statistics (informational only — using hard rule) ---
    n_sat_train = (train["is_saturated"] == 1).sum()
    n_sat_val   = (val["is_saturated"] == 1).sum()
    n_sat_test  = (test["is_saturated"] == 1).sum()
    print(f"\nSaturation (hard rule load_ratio>=0.90 or status=saturated):")
    print(f"  Train: {n_sat_train} / {len(train)}")
    print(f"  Val:   {n_sat_val} / {len(val)}")
    print(f"  Test:  {n_sat_test} / {len(test)}")
    print("  NOTE: Saturation handled by hard rule only — no classifier trained.")

    # --- Congestion penalty regressor ---
    print("\n--- Congestion Penalty Regressor ---")

    # Target: max(0, observed - baseline). Exclude saturated and missing latency rows.
    def make_regression_target(df: pd.DataFrame) -> pd.Series:
        usable = (df["is_saturated"] == 0) & (df["observed_latency_ms"].notna())
        penalty = (df["observed_latency_ms"] - df["physical_latency_ms"]).clip(lower=0.0)
        return penalty.where(usable)

    train["congestion_penalty_ms"] = make_regression_target(train)
    val["congestion_penalty_ms"]   = make_regression_target(val)

    train_reg = train[train["congestion_penalty_ms"].notna()]
    val_reg   = val[val["congestion_penalty_ms"].notna()]

    print(f"  Training rows with valid penalty target: {len(train_reg)}")
    print(f"  Validation rows with valid penalty target: {len(val_reg)}")

    missing_load_units_count = int(df_raw["load_units"].isnull().sum())
    missing_latency_count    = int(df_raw["observed_latency_ms"].isnull().sum())
    print(f"  Missing load_units (dataset): {missing_load_units_count}")
    print(f"  Missing observed_latency_ms (dataset): {missing_latency_count} (excluded from target)")

    # Verify feature columns are all present.
    for col in CONGESTION_FEATURE_COLS:
        if col not in train_reg.columns:
            raise RuntimeError(f"Missing feature column: {col}")

    X_train = train_reg[CONGESTION_FEATURE_COLS].astype(float)
    y_train = train_reg["congestion_penalty_ms"]
    X_val   = val_reg[CONGESTION_FEATURE_COLS].astype(float)
    y_val   = val_reg["congestion_penalty_ms"]

    # Model selection: HistGradientBoostingRegressor (handles NaN natively).
    best_model = None
    best_mae = float("inf")

    candidates = [
        ("HGBR(l_rate=0.05,depth=4)", HistGradientBoostingRegressor(
            learning_rate=0.05, max_depth=4, max_iter=200,
            min_samples_leaf=10, random_state=RANDOM_STATE)),
        ("HGBR(l_rate=0.10,depth=5)", HistGradientBoostingRegressor(
            learning_rate=0.10, max_depth=5, max_iter=200,
            min_samples_leaf=10, random_state=RANDOM_STATE)),
        ("HGBR(l_rate=0.10,depth=3)", HistGradientBoostingRegressor(
            learning_rate=0.10, max_depth=3, max_iter=150,
            min_samples_leaf=15, random_state=RANDOM_STATE)),
    ]

    for name, model in candidates:
        model.fit(X_train, y_train)
        preds = model.predict(X_val).clip(min=0.0)
        mae = mean_absolute_error(y_val, preds)
        rmse = np.sqrt(mean_squared_error(y_val, preds))
        print(f"  {name}: MAE={mae:.2f} ms, RMSE={rmse:.2f} ms")
        if mae < best_mae:
            best_mae = mae
            best_model = (name, model)

    selected_name, reg = best_model
    print(f"\n  Selected: {selected_name}")

    # Final val metrics.
    y_pred_val = reg.predict(X_val).clip(min=0.0)
    val_mae  = mean_absolute_error(y_val, y_pred_val)
    val_rmse = np.sqrt(mean_squared_error(y_val, y_pred_val))
    val_medae = float(np.median(np.abs(y_val.values - y_pred_val)))
    print(f"  Val MAE:   {val_mae:.2f} ms")
    print(f"  Val RMSE:  {val_rmse:.2f} ms")
    print(f"  Val MedAE: {val_medae:.2f} ms")

    # Error by load band.
    val_reg_copy = val_reg.copy()
    val_reg_copy["_pred"] = y_pred_val
    val_reg_copy["_err"]  = np.abs(val_reg_copy["congestion_penalty_ms"] - val_reg_copy["_pred"])
    bands = [(0.0, 0.5), (0.5, 0.75), (0.75, 0.90)]
    band_metrics = {}
    for lo, hi in bands:
        mask = val_reg_copy["load_ratio"].between(lo, hi, inclusive="left")
        n = mask.sum()
        band_mae = float(val_reg_copy.loc[mask, "_err"].mean()) if n > 0 else None
        band_metrics[f"{lo:.2f}-{hi:.2f}"] = {"n": int(n), "mae_ms": band_mae}
        if n > 0:
            print(f"    load_ratio [{lo:.2f},{hi:.2f}): n={n}, MAE={band_mae:.2f} ms")

    # --- Export ---
    os.makedirs(os.path.dirname(OUTPUT_PATH), exist_ok=True)

    split_meta = split_info_from_result(raw_split, dataset_hash)
    prep_meta  = preprocessor.metadata()
    metrics    = {
        "val_mae_ms":   round(val_mae, 4),
        "val_rmse_ms":  round(val_rmse, 4),
        "val_medae_ms": round(val_medae, 4),
        "val_n_samples": int(len(y_val)),
        "train_n_samples": int(len(y_train)),
        "saturated_train": int(n_sat_train),
        "saturated_val": int(n_sat_val),
        "saturated_test": int(n_sat_test),
        "error_by_load_band": band_metrics,
        "selected_model": selected_name,
    }

    # Export HGBR as flat node arrays (flat format for Go).
    estimator_trees = tree_to_dict_hgbr(reg, CONGESTION_FEATURE_COLS)

    meta = build_metadata(
        task="congestion_regressor",
        feature_names=CONGESTION_FEATURE_COLS,
        split_info=split_meta,
        preprocessing_meta=prep_meta,
        metrics=metrics,
        extra={
            "model_type": "congestion",
            "features": CONGESTION_FEATURE_COLS,
            "learning_rate": reg.learning_rate,
            "n_estimators": len(reg._predictors),
            "base_prediction": float(reg._baseline_prediction),
            "saturation_threshold": 0.90,
            "saturation_rule": "hard: load_ratio >= 0.90 OR status == 'saturated'",
            "link_id_map": link_id_map,
            "physical_baselines_ms": baselines,
            "capacities": capacities,
            "hgbr_trees": estimator_trees,
        },
    )

    with open(OUTPUT_PATH, "w") as f:
        json.dump(meta, f, indent=2)

    print(f"\nExported to {OUTPUT_PATH}")
    print("=" * 60)


if __name__ == "__main__":
    main()
