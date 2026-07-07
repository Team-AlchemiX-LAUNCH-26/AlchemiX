"""
Train telemetry trust regressor from link_telemetry.csv.

Key changes from v1:
- self_reported_latency_ms NaN: adds indicator column, imputes with training-partition median.
- historical_bias feature uses shift(1) before expanding mean — no leakage.
- rolling_self_reported uses shift(1) — no leakage.
- prev_self_reported uses shift(1) — no leakage of imputed value.
- Full metadata exported.
- HistGradientBoostingRegressor for robustness to NaN inputs.
"""

import json
import os
import sys

import numpy as np
import pandas as pd
from sklearn.ensemble import HistGradientBoostingRegressor
from sklearn.metrics import mean_absolute_error, mean_squared_error

sys.path.insert(0, os.path.dirname(__file__))

from preprocessing import (
    REQUIRED_TELEMETRY_COLS,
    TelemetryPreprocessor,
    canonical_link_id_map,
    csv_sha256,
    load_physical_baselines,
    load_universe_config,
    validate_schema,
)
from split import assert_no_overlap, tick_split
from features import TRUST_FEATURE_COLS, build_trust_features
from model_metadata import RANDOM_STATE, build_metadata, split_info_from_result

DATASET_PATH = os.path.join(
    os.path.dirname(__file__), "..", "datasets", "link_telemetry.csv"
)
UNIVERSE_PATH = os.path.join(
    os.path.dirname(__file__), "..", "datasets", "universe-config.json"
)
OUTPUT_PATH = os.path.join(
    os.path.dirname(__file__), "..", "internal", "models", "trust_model.json"
)

# Trust score conversion: trust = clamp(1 - ratio / TRUST_SCALE, 0, 1)
# A ratio of TRUST_SCALE maps to trust_score = 0.
# Documented here and exported in model metadata.
TRUST_SCALE = 0.50
SPOOF_THRESHOLD = 0.15  # under_report_ratio above this is considered spoofed.


def tree_to_dict_hgbr(estimator, feature_names: list[str]) -> list[list[dict]]:
    """Export a HistGradientBoostingRegressor as flat node arrays."""
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
                })
            trees.append(tree_nodes)
    return trees


def main() -> None:
    print("=" * 60)
    print("TRUST MODEL TRAINING  (v2 — leakage-safe)")
    print("=" * 60)

    dataset_hash = csv_sha256(DATASET_PATH)
    df_raw = pd.read_csv(DATASET_PATH)
    print(f"Loaded {len(df_raw)} rows | hash={dataset_hash}")
    validate_schema(df_raw, REQUIRED_TELEMETRY_COLS, "link_telemetry")

    cfg = load_universe_config(UNIVERSE_PATH)
    baselines = load_physical_baselines(cfg)
    link_id_map = canonical_link_id_map(cfg)

    df_raw = df_raw.sort_values(["link_id", "tick"]).reset_index(drop=True)

    # --- Split FIRST, fit preprocessor on training partition ---
    raw_split = tick_split(df_raw)
    assert_no_overlap(raw_split)
    print(f"\nTick-based split:")
    print(raw_split.describe())

    preprocessor = TelemetryPreprocessor()
    preprocessor.fit(raw_split.train, baselines, link_id_map)

    # Transform full dataset using training-learned imputation values.
    full_df = preprocessor.transform(df_raw)

    # Missing value stats.
    n_missing_self = int(df_raw["self_reported_latency_ms"].isnull().sum())
    print(f"\n  Missing self_reported_latency_ms: {n_missing_self} "
          f"-> imputed with training median "
          f"({preprocessor._self_reported_median:.2f} ms) + indicator column added")

    # Build leakage-safe features on full df (already sorted).
    full_df = build_trust_features(full_df)

    # Drop rows where under_report_ratio target is NaN (measured_latency_ms == 0).
    full_df = full_df.dropna(subset=["under_report_ratio"])
    print(f"  Rows after dropping invalid targets: {len(full_df)}")

    # Re-split after feature engineering.
    train = full_df[full_df["tick"] <= raw_split.train_tick_range[1]]
    val   = full_df[
        (full_df["tick"] > raw_split.train_tick_range[1]) &
        (full_df["tick"] <= raw_split.val_tick_range[1])
    ]
    test  = full_df[full_df["tick"] > raw_split.val_tick_range[1]]

    print(f"\nAfter feature engineering:")
    print(f"  Train: {len(train)}, Val: {len(val)}, Test: {len(test)}")

    # Verify features present.
    for col in TRUST_FEATURE_COLS:
        if col not in train.columns:
            raise RuntimeError(f"Missing feature column: {col}")

    X_train = train[TRUST_FEATURE_COLS].astype(float)
    y_train = train["under_report_ratio"]
    X_val   = val[TRUST_FEATURE_COLS].astype(float)
    y_val   = val["under_report_ratio"]

    # Model selection.
    best_model = None
    best_mae = float("inf")

    candidates = [
        ("HGBR(lr=0.05,d=4)", HistGradientBoostingRegressor(
            learning_rate=0.05, max_depth=4, max_iter=200,
            min_samples_leaf=10, random_state=RANDOM_STATE)),
        ("HGBR(lr=0.10,d=5)", HistGradientBoostingRegressor(
            learning_rate=0.10, max_depth=5, max_iter=200,
            min_samples_leaf=10, random_state=RANDOM_STATE)),
        ("HGBR(lr=0.10,d=3)", HistGradientBoostingRegressor(
            learning_rate=0.10, max_depth=3, max_iter=150,
            min_samples_leaf=15, random_state=RANDOM_STATE)),
    ]

    print("\n--- Trust Penalty Regressor ---")
    for name, model in candidates:
        model.fit(X_train, y_train)
        preds = np.clip(model.predict(X_val), 0.0, 1.0)
        mae = mean_absolute_error(y_val, preds)
        rmse = np.sqrt(mean_squared_error(y_val, preds))
        print(f"  {name}: MAE={mae:.4f}, RMSE={rmse:.4f}")
        if mae < best_mae:
            best_mae = mae
            best_model = (name, model)

    selected_name, reg = best_model
    print(f"\n  Selected: {selected_name}")

    y_pred_val = np.clip(reg.predict(X_val), 0.0, 1.0)
    val_mae  = mean_absolute_error(y_val, y_pred_val)
    val_rmse = np.sqrt(mean_squared_error(y_val, y_pred_val))

    # Spoof detection metrics.
    y_val_spoof  = (y_val > SPOOF_THRESHOLD).astype(int)
    y_pred_spoof = (y_pred_val > SPOOF_THRESHOLD).astype(int)

    from sklearn.metrics import precision_score, recall_score
    spoof_prec = precision_score(y_val_spoof, y_pred_spoof, zero_division=0)
    spoof_rec  = recall_score(y_val_spoof, y_pred_spoof, zero_division=0)

    # False-positive rate on honest links.
    honest_mask = y_val_spoof == 0
    fp_rate = (
        float((y_pred_spoof[honest_mask] == 1).sum()) / int(honest_mask.sum())
        if honest_mask.sum() > 0 else 0.0
    )

    print(f"  Val MAE:            {val_mae:.4f}")
    print(f"  Val RMSE:           {val_rmse:.4f}")
    print(f"  Spoof Precision:    {spoof_prec:.4f} (threshold={SPOOF_THRESHOLD})")
    print(f"  Spoof Recall:       {spoof_rec:.4f}")
    print(f"  FP rate (honest):   {fp_rate:.4f}")

    # --- Export ---
    os.makedirs(os.path.dirname(OUTPUT_PATH), exist_ok=True)

    split_meta = split_info_from_result(raw_split, dataset_hash)
    prep_meta  = preprocessor.metadata()
    metrics    = {
        "val_mae":          round(val_mae, 6),
        "val_rmse":         round(val_rmse, 6),
        "spoof_precision":  round(spoof_prec, 4),
        "spoof_recall":     round(spoof_rec, 4),
        "fp_rate_honest":   round(fp_rate, 4),
        "spoof_threshold":  SPOOF_THRESHOLD,
        "trust_scale":      TRUST_SCALE,
        "selected_model":   selected_name,
    }

    estimator_trees = tree_to_dict_hgbr(reg, TRUST_FEATURE_COLS)

    meta = build_metadata(
        task="trust_regressor",
        feature_names=TRUST_FEATURE_COLS,
        split_info=split_meta,
        preprocessing_meta=prep_meta,
        metrics=metrics,
        extra={
            "model_type": "trust",
            "features": TRUST_FEATURE_COLS,
            "learning_rate": reg.learning_rate,
            "n_estimators": len(reg._predictors),
            "base_prediction": float(reg._baseline_prediction),
            "trust_scale": TRUST_SCALE,
            "spoof_threshold": SPOOF_THRESHOLD,
            "link_id_map": link_id_map,
            "physical_baselines_ms": baselines,
            "hgbr_trees": estimator_trees,
        },
    )

    with open(OUTPUT_PATH, "w") as f:
        json.dump(meta, f, indent=2)

    print(f"\nExported to {OUTPUT_PATH}")
    print("=" * 60)


if __name__ == "__main__":
    main()
