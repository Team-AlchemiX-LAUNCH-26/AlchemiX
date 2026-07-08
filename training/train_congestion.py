"""
Train congestion penalty regressor from link_traffic_history.csv.

Key changes from v1:
- Saturation classifier REMOVED: only 3 saturated examples → hard rule used instead.
- Rolling features use shift(1) to prevent leakage.
- Missing load_units imputed using formula or training-partition median.
- observed_latency_ms NaN rows excluded from regression target only.
- Adds missing-value indicator columns.
- Full metadata exported with model JSON.
- Compares median baseline, polynomial ridge, random forest, extra trees,
  gradient boosting, and HistGradientBoosting. The Go runtime export remains
  HistGradientBoosting-compatible.
"""

import json
import os
import sys

import joblib
import numpy as np
import pandas as pd
from sklearn.dummy import DummyRegressor
from sklearn.ensemble import (
    ExtraTreesRegressor,
    GradientBoostingRegressor,
    HistGradientBoostingRegressor,
    RandomForestRegressor,
)
from sklearn.impute import SimpleImputer
from sklearn.linear_model import Ridge
from sklearn.metrics import mean_absolute_error, mean_squared_error
from sklearn.pipeline import make_pipeline
from sklearn.preprocessing import PolynomialFeatures, StandardScaler

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
JOBLIB_PATH = os.path.join(
    os.path.dirname(__file__), "..", "internal", "models", "congestion_model.joblib"
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


def scalar_baseline(value) -> float:
    return float(np.asarray(value).reshape(-1)[0])


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

    # Transform each partition using training-learned parameters, then build
    # features inside that partition only. No feature engineering is performed
    # on the full dataset before train/val/test separation.
    train = build_congestion_features(preprocessor.transform(raw_split.train))
    val = build_congestion_features(preprocessor.transform(raw_split.val))
    test = build_congestion_features(preprocessor.transform(raw_split.test))

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

    # Model comparison across several suitable regression families. The Go
    # runtime currently consumes HGB tree JSON, so we export the best HGB model
    # while still recording the broader validation comparison.
    best_validation_model = None
    best_validation_mae = float("inf")
    best_exportable_model = None
    best_exportable_mae = float("inf")
    model_comparison = []

    candidates = [
        ("Median baseline", DummyRegressor(strategy="median"), False),
        ("Polynomial Ridge", make_pipeline(
            SimpleImputer(strategy="median"),
            PolynomialFeatures(degree=2, include_bias=False),
            StandardScaler(),
            Ridge(alpha=10.0),
        ), False),
        ("Random Forest", make_pipeline(
            SimpleImputer(strategy="median"),
            RandomForestRegressor(
                n_estimators=200, max_depth=10, min_samples_leaf=5,
                random_state=RANDOM_STATE, n_jobs=1,
            ),
        ), False),
        ("Extra Trees", make_pipeline(
            SimpleImputer(strategy="median"),
            ExtraTreesRegressor(
                n_estimators=200, max_depth=10, min_samples_leaf=5,
                random_state=RANDOM_STATE, n_jobs=1,
            ),
        ), False),
        ("Gradient Boosting", make_pipeline(
            SimpleImputer(strategy="median"),
            GradientBoostingRegressor(
                learning_rate=0.05, max_depth=3, n_estimators=200,
                random_state=RANDOM_STATE,
            ),
        ), False),
        ("HistGradientBoosting", HistGradientBoostingRegressor(
            learning_rate=0.10, max_depth=5, max_iter=200,
            min_samples_leaf=10, random_state=RANDOM_STATE), True),
    ]

    for name, model, exportable in candidates:
        model.fit(X_train, y_train)
        preds = model.predict(X_val).clip(min=0.0)
        mae = mean_absolute_error(y_val, preds)
        rmse = np.sqrt(mean_squared_error(y_val, preds))
        p95 = float(np.percentile(np.abs(y_val.values - preds), 95))
        print(f"  {name}: MAE={mae:.2f} ms, RMSE={rmse:.2f} ms")
        model_comparison.append({
            "model": name,
            "val_mae_ms": round(float(mae), 4),
            "val_rmse_ms": round(float(rmse), 4),
            "val_p95_abs_error_ms": round(p95, 4),
            "go_exportable": exportable,
        })
        if mae < best_validation_mae:
            best_validation_mae = mae
            best_validation_model = (name, model)
        if exportable and mae < best_exportable_mae:
            best_exportable_mae = mae
            best_exportable_model = (name, model)

    best_validation_name, best_validation_estimator = best_validation_model
    selected_name, reg = best_exportable_model
    print(f"\n  Best validation model: {best_validation_name}")
    print(f"  Exported Go-compatible model: {selected_name}")

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
        "best_validation_model": best_validation_name,
        "model_comparison": model_comparison,
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
            "base_prediction": scalar_baseline(reg._baseline_prediction),
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
    joblib.dump(
        {
            "task": "congestion",
            "model_name": best_validation_name,
            "estimator": best_validation_estimator,
            "feature_names": CONGESTION_FEATURE_COLS,
            "metrics": metrics,
        },
        JOBLIB_PATH,
    )

    print(f"\nExported to {OUTPUT_PATH}")
    print(f"Best Python model exported to {JOBLIB_PATH}")
    print("=" * 60)


if __name__ == "__main__":
    main()
