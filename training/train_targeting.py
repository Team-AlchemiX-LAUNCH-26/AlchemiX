"""
Train targeting-risk classifier from link_incident_history.csv.

Key changes from v1:
- rolling_traffic_share uses shift(1) — no leakage.
- recent_jam_count uses shift(1) — no leakage.
- consecutive_high_usage uses shift(1) — no leakage.
- class_weight="balanced" for 8.2% minority class.
- Reports PR-AUC, ROC-AUC, Brier score (not raw accuracy).
- Platt sigmoid calibration on validation set.
- Full metadata exported including calibration parameters.
"""

import json
import os
import sys

import joblib
import numpy as np
import pandas as pd
from sklearn.dummy import DummyClassifier
from sklearn.ensemble import (
    GradientBoostingClassifier,
    HistGradientBoostingClassifier,
    RandomForestClassifier,
)
from sklearn.impute import SimpleImputer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import (
    average_precision_score,
    brier_score_loss,
    log_loss,
    precision_score,
    recall_score,
    roc_auc_score,
)
from sklearn.pipeline import make_pipeline
from sklearn.preprocessing import StandardScaler

sys.path.insert(0, os.path.dirname(__file__))

from preprocessing import (
    REQUIRED_INCIDENT_COLS,
    IncidentPreprocessor,
    canonical_link_id_map,
    csv_sha256,
    load_universe_config,
    validate_schema,
)
from split import assert_no_overlap, tick_split
from features import TARGETING_FEATURE_COLS, build_targeting_features
from model_metadata import RANDOM_STATE, build_metadata, split_info_from_result

DATASET_PATH = os.path.join(
    os.path.dirname(__file__), "..", "datasets", "link_incident_history.csv"
)
OUTPUT_PATH = os.path.join(
    os.path.dirname(__file__), "..", "internal", "models", "targeting_model.json"
)
JOBLIB_PATH = os.path.join(
    os.path.dirname(__file__), "..", "internal", "models", "targeting_model.joblib"
)
UNIVERSE_PATH = os.path.join(
    os.path.dirname(__file__), "..", "datasets", "universe-config.json"
)


def tree_to_dict_hgbc(estimator, feature_names: list[str]) -> list[list[dict]]:
    """Export a HistGradientBoostingClassifier as flat node arrays."""
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


def scalar_baseline(value) -> float:
    return float(np.asarray(value).reshape(-1)[0])


def pr_auc(y_true: pd.Series, y_proba: np.ndarray) -> float:
    return float(average_precision_score(y_true, y_proba))


def sigmoid_calibration_params(
    y_true: np.ndarray, y_score: np.ndarray
) -> tuple[float, float]:
    """
    Fit Platt scaling: P_calibrated = 1 / (1 + exp(a * score + b)).
    Returns (a, b) fit on y_true, y_score.
    """
    from scipy.optimize import minimize
    from scipy.special import expit

    def neg_log_likelihood(params):
        a, b = params
        p = expit(-(a * y_score + b))
        p = np.clip(p, 1e-12, 1 - 1e-12)
        return -np.mean(y_true * np.log(p) + (1 - y_true) * np.log(1 - p))

    res = minimize(neg_log_likelihood, x0=[1.0, 0.0], method="Nelder-Mead")
    return float(res.x[0]), float(res.x[1])


def apply_sigmoid_calibration(score: np.ndarray, a: float, b: float) -> np.ndarray:
    from scipy.special import expit
    return expit(-(a * score + b))


def main() -> None:
    print("=" * 60)
    print("TARGETING-RISK MODEL TRAINING  (v2 — leakage-safe)")
    print("=" * 60)

    dataset_hash = csv_sha256(DATASET_PATH)
    df_raw = pd.read_csv(DATASET_PATH)
    print(f"Loaded {len(df_raw)} rows | hash={dataset_hash}")
    validate_schema(df_raw, REQUIRED_INCIDENT_COLS, "link_incident_history")

    cfg = load_universe_config(UNIVERSE_PATH)
    link_id_map = canonical_link_id_map(cfg)

    df_raw = df_raw.sort_values(["link_id", "tick"]).reset_index(drop=True)

    # Class distribution.
    n_jammed     = int(df_raw["jammed_flag"].sum())
    n_not_jammed = int((~df_raw["jammed_flag"]).sum())
    print(f"\nClass distribution: jammed={n_jammed} ({n_jammed/len(df_raw)*100:.1f}%), "
          f"not_jammed={n_not_jammed}")
    print(f"  Missing traffic_share: {df_raw['traffic_share'].isnull().sum()}")

    raw_split = tick_split(df_raw)
    assert_no_overlap(raw_split)
    print(f"\nTick-based split:")
    print(raw_split.describe())

    preprocessor = IncidentPreprocessor()
    preprocessor.fit(raw_split.train, link_id_map)

    # Transform each partition using training-learned imputation values, then
    # build features inside that partition only. No feature engineering is
    # performed on the full dataset before train/val/test separation.
    train = build_targeting_features(preprocessor.transform(raw_split.train))
    val = build_targeting_features(preprocessor.transform(raw_split.val))
    test = build_targeting_features(preprocessor.transform(raw_split.test))

    train = train.dropna(subset=TARGETING_FEATURE_COLS + ["jammed"])
    val = val.dropna(subset=TARGETING_FEATURE_COLS + ["jammed"])
    test = test.dropna(subset=TARGETING_FEATURE_COLS + ["jammed"])
    print(f"  Rows after dropping invalid: {len(train) + len(val) + len(test)}")

    for col in TARGETING_FEATURE_COLS:
        if col not in train.columns:
            raise RuntimeError(f"Missing feature column: {col}")

    X_train = train[TARGETING_FEATURE_COLS].astype(float)
    y_train = train["jammed"]
    X_val   = val[TARGETING_FEATURE_COLS].astype(float)
    y_val   = val["jammed"]

    # Class distributions per partition.
    n_jam_train = int(y_train.sum())
    n_jam_val   = int(y_val.sum())
    print(f"\n  Train jammed: {n_jam_train}/{len(y_train)} "
          f"({n_jam_train/max(1,len(y_train))*100:.1f}%)")
    print(f"  Val jammed:   {n_jam_val}/{len(y_val)} "
          f"({n_jam_val/max(1,len(y_val))*100:.1f}%)")

    # --- Model selection ---
    print("\n--- Targeting Classifier (class_weight=balanced) ---")

    best_validation_model = None
    best_validation_pr_auc = -float("inf")
    best_exportable_model = None
    best_exportable_pr_auc = -float("inf")
    model_comparison = []

    candidates = [
        ("Prior baseline", DummyClassifier(strategy="prior"), False),
        ("Logistic Regression", make_pipeline(
            SimpleImputer(strategy="median"),
            StandardScaler(),
            LogisticRegression(
                class_weight="balanced", max_iter=1000,
                random_state=RANDOM_STATE,
            ),
        ), False),
        ("Random Forest", make_pipeline(
            SimpleImputer(strategy="median"),
            RandomForestClassifier(
                n_estimators=300, max_depth=10, min_samples_leaf=5,
                class_weight="balanced", random_state=RANDOM_STATE, n_jobs=1,
            ),
        ), False),
        ("Gradient Boosting", make_pipeline(
            SimpleImputer(strategy="median"),
            GradientBoostingClassifier(
                learning_rate=0.05, max_depth=3, n_estimators=200,
                random_state=RANDOM_STATE,
            ),
        ), False),
        ("HistGradientBoosting", HistGradientBoostingClassifier(
            learning_rate=0.05, max_depth=4, max_iter=200,
            min_samples_leaf=5, random_state=RANDOM_STATE,
            class_weight="balanced"), True),
    ]

    for name, model, exportable in candidates:
        model.fit(X_train, y_train)
        proba = model.predict_proba(X_val)[:, 1]
        roc   = roc_auc_score(y_val, proba)
        prauc = pr_auc(y_val, proba)
        brier = brier_score_loss(y_val, proba)
        ll = log_loss(y_val, np.clip(proba, 1e-12, 1 - 1e-12))
        print(f"  {name}: ROC-AUC={roc:.4f}, PR-AUC={prauc:.4f}, Brier={brier:.4f}")
        model_comparison.append({
            "model": name,
            "val_roc_auc": round(float(roc), 4),
            "val_pr_auc": round(float(prauc), 4),
            "val_brier": round(float(brier), 4),
            "val_log_loss": round(float(ll), 4),
            "go_exportable": exportable,
        })
        if prauc > best_validation_pr_auc:
            best_validation_pr_auc = prauc
            best_validation_model = (name, model)
        if exportable and prauc > best_exportable_pr_auc:
            best_exportable_pr_auc = prauc
            best_exportable_model = (name, model)

    best_validation_name, best_validation_estimator = best_validation_model
    selected_name, clf = best_exportable_model
    print(f"\n  Best validation model: {best_validation_name}")
    print(f"  Exported Go-compatible model: {selected_name}")

    # --- Platt sigmoid calibration on validation set ---
    y_val_proba_raw = clf.predict_proba(X_val)[:, 1]
    cal_a, cal_b = sigmoid_calibration_params(y_val.values.astype(float), y_val_proba_raw)
    y_val_calibrated = apply_sigmoid_calibration(y_val_proba_raw, cal_a, cal_b)

    # Calibration diagnostics.
    n_bins = 5
    bins = np.linspace(0, 1, n_bins + 1)
    calibration_bins = []
    for i in range(n_bins):
        lo, hi = bins[i], bins[i + 1]
        mask = (y_val_calibrated >= lo) & (y_val_calibrated < hi)
        if mask.sum() > 0:
            mean_pred  = float(y_val_calibrated[mask].mean())
            mean_true  = float(y_val.values[mask].mean())
            calibration_bins.append({
                "bin": f"{lo:.2f}-{hi:.2f}",
                "n": int(mask.sum()),
                "mean_predicted": round(mean_pred, 4),
                "mean_true": round(mean_true, 4),
            })

    roc_cal  = roc_auc_score(y_val, y_val_calibrated)
    brier_cal = brier_score_loss(y_val, y_val_calibrated)
    prauc_cal = pr_auc(y_val, y_val_calibrated)

    # Hard threshold metrics.
    threshold = 0.5
    y_pred_hard = (y_val_calibrated >= threshold).astype(int)
    prec  = precision_score(y_val, y_pred_hard, zero_division=0)
    rec   = recall_score(y_val, y_pred_hard, zero_division=0)

    print(f"\n  After Platt calibration (a={cal_a:.4f}, b={cal_b:.4f}):")
    print(f"  ROC-AUC:   {roc_cal:.4f}")
    print(f"  PR-AUC:    {prauc_cal:.4f}")
    print(f"  Brier:     {brier_cal:.4f}")
    print(f"  Precision: {prec:.4f}  Recall: {rec:.4f}  (threshold={threshold})")
    print(f"  Calibration bins:")
    for b in calibration_bins:
        print(f"    {b['bin']}: n={b['n']}, pred={b['mean_predicted']:.3f}, true={b['mean_true']:.3f}")

    # --- Export ---
    os.makedirs(os.path.dirname(OUTPUT_PATH), exist_ok=True)

    split_meta = split_info_from_result(raw_split, dataset_hash)
    prep_meta  = preprocessor.metadata()
    metrics    = {
        "val_roc_auc":       round(roc_cal, 4),
        "val_pr_auc":        round(prauc_cal, 4),
        "val_brier":         round(brier_cal, 4),
        "val_precision":     round(prec, 4),
        "val_recall":        round(rec, 4),
        "calibration_threshold": threshold,
        "platt_a":           round(cal_a, 6),
        "platt_b":           round(cal_b, 6),
        "calibration_bins":  calibration_bins,
        "class_distribution": {
            "train_jammed": n_jam_train,
            "train_not_jammed": len(y_train) - n_jam_train,
            "val_jammed": n_jam_val,
            "val_not_jammed": len(y_val) - n_jam_val,
        },
        "selected_model": selected_name,
        "best_validation_model": best_validation_name,
        "model_comparison": model_comparison,
    }

    estimator_trees = tree_to_dict_hgbc(clf, TARGETING_FEATURE_COLS)

    meta = build_metadata(
        task="targeting_classifier",
        feature_names=TARGETING_FEATURE_COLS,
        split_info=split_meta,
        preprocessing_meta=prep_meta,
        metrics=metrics,
        extra={
            "model_type": "targeting",
            "features": TARGETING_FEATURE_COLS,
            "learning_rate": clf.learning_rate,
            "n_estimators": len(clf._predictors),
            "base_prediction": scalar_baseline(clf._baseline_prediction),
            "platt_a": round(cal_a, 6),
            "platt_b": round(cal_b, 6),
            "link_id_map": link_id_map,
            "hgbr_trees": estimator_trees,
        },
    )

    with open(OUTPUT_PATH, "w") as f:
        json.dump(meta, f, indent=2)
    joblib.dump(
        {
            "task": "targeting",
            "model_name": best_validation_name,
            "estimator": best_validation_estimator,
            "feature_names": TARGETING_FEATURE_COLS,
            "metrics": metrics,
        },
        JOBLIB_PATH,
    )

    print(f"\nExported to {OUTPUT_PATH}")
    print(f"Best Python model exported to {JOBLIB_PATH}")
    print("=" * 60)


if __name__ == "__main__":
    main()
