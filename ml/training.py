"""
training.py – Model training pipeline for AlchemiX Phase 2.

Responsibilities
----------------
- Train/Test split (MUST happen BEFORE historical feature engineering)
- Historical feature engineering (on training split ONLY, then mapped to test)
- Encoding (OrdinalEncoder for categorical columns)
- Scaling  (StandardScaler for numeric columns)
- Benchmark multiple algorithms (RF, XGBoost, LightGBM, CatBoost, DT)
- Cross-validation
- Hyperparameter search (RandomizedSearchCV)
- Select and return best model

Forbidden
---------
- Computing historical statistics before split
- Using target column as an input feature
- Encoding before split
"""

from __future__ import annotations

import time
from typing import Any

import numpy as np
import pandas as pd
from sklearn.ensemble import RandomForestRegressor, RandomForestClassifier
from sklearn.tree import DecisionTreeRegressor, DecisionTreeClassifier
from sklearn.preprocessing import OrdinalEncoder, StandardScaler
from sklearn.model_selection import (
    train_test_split,
    cross_val_score,
    RandomizedSearchCV,
    StratifiedKFold,
    KFold,
)
from sklearn.pipeline import Pipeline
from sklearn.compose import ColumnTransformer
import xgboost as xgb
import lightgbm as lgb

from ml.utils import get_logger, RANDOM_SEED

logger = get_logger(__name__)

TEST_SIZE: float = 0.20
CV_FOLDS: int = 5
N_ITER_SEARCH: int = 30

# ── Train / Test Split ────────────────────────────────────────────────────────

def split_data(
    df: pd.DataFrame,
    target_col: str,
    test_size: float = TEST_SIZE,
    stratify_col: str | None = None,
) -> tuple[pd.DataFrame, pd.DataFrame]:
    """
    Split engineered data into train and test sets.

    Returns (train_df, test_df) as DataFrames to allow historical
    feature engineering to be applied AFTER splitting.

    The split is done on row indices; tick ordering is respected.
    For time-series style data (same link, sequential ticks) we sort
    by tick and use the last `test_size` fraction as test.
    """
    df = df.sort_values(["link_id", "tick"]).reset_index(drop=True)
    if stratify_col:
        train_df, test_df = train_test_split(
            df,
            test_size=test_size,
            random_state=RANDOM_SEED,
            stratify=df[stratify_col],
        )
    else:
        train_df, test_df = train_test_split(
            df,
            test_size=test_size,
            random_state=RANDOM_SEED,
            shuffle=True,
        )
    logger.info(
        "split_data: train=%d  test=%d  target='%s'",
        len(train_df), len(test_df), target_col,
    )
    return train_df.reset_index(drop=True), test_df.reset_index(drop=True)


# ── Historical Feature Engineering (post-split only) ─────────────────────────

def add_historical_features_traffic(
    train_df: pd.DataFrame,
    test_df: pd.DataFrame,
) -> tuple[pd.DataFrame, pd.DataFrame]:
    """
    Compute per-link historical statistics from the TRAINING SET ONLY.
    Map the computed values to the test set via link_id lookup.

    Features added:
    - hist_mean_load_ratio   per link
    - hist_mean_load_units   per link
    - hist_std_load_ratio    per link
    """
    stats = (
        train_df
        .groupby("link_id")[["load_ratio", "load_units"]]
        .agg(
            hist_mean_load_ratio=("load_ratio", "mean"),
            hist_std_load_ratio=("load_ratio", "std"),
            hist_mean_load_units=("load_units", "mean"),
        )
        .reset_index()
    )

    train_df = train_df.merge(stats, on="link_id", how="left")
    test_df  = test_df.merge(stats, on="link_id", how="left")
    # Fill any unseen test links with training global mean
    global_defaults = {
        "hist_mean_load_ratio": train_df["load_ratio"].mean(),
        "hist_std_load_ratio":  train_df["load_ratio"].std(),
        "hist_mean_load_units": train_df["load_units"].mean(),
    }
    train_df = train_df.fillna(global_defaults)
    test_df  = test_df.fillna(global_defaults)
    logger.info("add_historical_features_traffic: added 3 historical columns")
    return train_df, test_df


def add_historical_features_telemetry(
    train_df: pd.DataFrame,
    test_df: pd.DataFrame,
) -> tuple[pd.DataFrame, pd.DataFrame]:
    """
    Add per-link historical trust statistics computed from training only.
    """
    stats = (
        train_df
        .groupby("link_id")[["trust_score", "latency_difference"]]
        .agg(
            hist_mean_trust=("trust_score", "mean"),
            hist_std_trust=("trust_score", "std"),
            hist_mean_latency_diff=("latency_difference", "mean"),
        )
        .reset_index()
    )
    train_df = train_df.merge(stats, on="link_id", how="left")
    test_df  = test_df.merge(stats, on="link_id", how="left")
    defaults = {
        "hist_mean_trust":         train_df["trust_score"].mean(),
        "hist_std_trust":          train_df["trust_score"].std(),
        "hist_mean_latency_diff":  train_df["latency_difference"].mean(),
    }
    train_df = train_df.fillna(defaults)
    test_df  = test_df.fillna(defaults)
    logger.info("add_historical_features_telemetry: added 3 historical columns")
    return train_df, test_df


def add_historical_features_incident(
    train_df: pd.DataFrame,
    test_df: pd.DataFrame,
) -> tuple[pd.DataFrame, pd.DataFrame]:
    """
    Add per-link historical attack rate computed from training only.
    """
    stats = (
        train_df
        .groupby("link_id")["jammed_flag"]
        .agg(hist_attack_rate="mean")
        .reset_index()
    )
    train_df = train_df.merge(stats, on="link_id", how="left")
    test_df  = test_df.merge(stats, on="link_id", how="left")
    defaults = {"hist_attack_rate": train_df["jammed_flag"].mean()}
    train_df = train_df.fillna(defaults)
    test_df  = test_df.fillna(defaults)
    logger.info("add_historical_features_incident: added 1 historical column")
    return train_df, test_df


# ── Preprocessor builder ──────────────────────────────────────────────────────

def build_preprocessor(
    numeric_features: list[str],
    categorical_features: list[str],
) -> ColumnTransformer:
    """Build a ColumnTransformer with StandardScaler + OrdinalEncoder."""
    transformers: list = []
    if numeric_features:
        transformers.append(("num", StandardScaler(), numeric_features))
    if categorical_features:
        transformers.append(
            ("cat", OrdinalEncoder(handle_unknown="use_encoded_value", unknown_value=-1), categorical_features)
        )
    return ColumnTransformer(transformers=transformers, remainder="drop")


# ── Candidate models ──────────────────────────────────────────────────────────

def regression_candidates(seed: int = RANDOM_SEED) -> dict[str, Any]:
    return {
        "RandomForest": RandomForestRegressor(random_state=seed, n_jobs=-1),
        "XGBoost":      xgb.XGBRegressor(random_state=seed, n_jobs=-1, verbosity=0),
        "LightGBM":     lgb.LGBMRegressor(random_state=seed, n_jobs=-1, verbose=-1),
        "DecisionTree": DecisionTreeRegressor(random_state=seed),
    }


def classification_candidates(seed: int = RANDOM_SEED) -> dict[str, Any]:
    return {
        "RandomForest": RandomForestClassifier(random_state=seed, n_jobs=-1),
        "XGBoost":      xgb.XGBClassifier(random_state=seed, n_jobs=-1, verbosity=0),
        "LightGBM":     lgb.LGBMClassifier(random_state=seed, n_jobs=-1, verbose=-1),
        "DecisionTree": DecisionTreeClassifier(random_state=seed),
    }


# ── Benchmarking ──────────────────────────────────────────────────────────────

def benchmark_models(
    candidates: dict[str, Any],
    X_train: np.ndarray,
    y_train: np.ndarray,
    task: str = "regression",
    cv_folds: int = CV_FOLDS,
) -> pd.DataFrame:
    """
    Run cross-validation on each candidate and return a comparison DataFrame.

    Parameters
    ----------
    task : 'regression' | 'classification'
    """
    scoring   = "neg_root_mean_squared_error" if task == "regression" else "roc_auc"
    cv_method = KFold(n_splits=cv_folds, shuffle=True, random_state=RANDOM_SEED) \
                if task == "regression" \
                else StratifiedKFold(n_splits=cv_folds, shuffle=True, random_state=RANDOM_SEED)

    records = []
    for name, model in candidates.items():
        t0 = time.time()
        scores = cross_val_score(model, X_train, y_train, cv=cv_method, scoring=scoring, n_jobs=-1)
        elapsed = time.time() - t0
        record = {
            "model": name,
            "cv_mean": scores.mean(),
            "cv_std":  scores.std(),
            "train_time_s": round(elapsed, 2),
        }
        records.append(record)
        logger.info(
            "  %-15s  cv_mean=%.4f  cv_std=%.4f  time=%.1fs",
            name, scores.mean(), scores.std(), elapsed,
        )

    results = pd.DataFrame(records).sort_values("cv_mean", ascending=(task == "regression"))
    return results.reset_index(drop=True)


# ── Hyperparameter search ─────────────────────────────────────────────────────

REGRESSION_PARAM_GRIDS: dict[str, dict] = {
    "RandomForest": {
        "n_estimators": [100, 200, 300],
        "max_depth": [None, 10, 20, 30],
        "min_samples_split": [2, 5, 10],
        "min_samples_leaf": [1, 2, 4],
        "max_features": ["sqrt", "log2", 0.5],
    },
    "XGBoost": {
        "n_estimators": [100, 200, 300],
        "max_depth": [3, 5, 7, 9],
        "learning_rate": [0.01, 0.05, 0.1, 0.2],
        "subsample": [0.7, 0.8, 0.9, 1.0],
        "colsample_bytree": [0.7, 0.8, 1.0],
    },
    "LightGBM": {
        "n_estimators": [100, 200, 300],
        "num_leaves": [31, 63, 127],
        "learning_rate": [0.01, 0.05, 0.1],
        "subsample": [0.7, 0.8, 1.0],
    },
    "DecisionTree": {
        "max_depth": [None, 5, 10, 20],
        "min_samples_split": [2, 5, 10],
        "min_samples_leaf": [1, 2, 4],
    },
}

CLASSIFICATION_PARAM_GRIDS: dict[str, dict] = {
    "RandomForest": REGRESSION_PARAM_GRIDS["RandomForest"],
    "XGBoost": {
        **REGRESSION_PARAM_GRIDS["XGBoost"],
        "scale_pos_weight": [1, 2, 3, 5],
    },
    "LightGBM": {
        **REGRESSION_PARAM_GRIDS["LightGBM"],
        "is_unbalance": [True, False],
    },
    "DecisionTree": REGRESSION_PARAM_GRIDS["DecisionTree"],
}


def tune_model(
    model: Any,
    model_name: str,
    X_train: np.ndarray,
    y_train: np.ndarray,
    task: str = "regression",
    n_iter: int = N_ITER_SEARCH,
) -> Any:
    """
    Run RandomizedSearchCV on the best model and return the tuned estimator.
    """
    grids = REGRESSION_PARAM_GRIDS if task == "regression" else CLASSIFICATION_PARAM_GRIDS
    param_grid = grids.get(model_name, {})
    if not param_grid:
        logger.warning("No param grid for %s – returning default model.", model_name)
        model.fit(X_train, y_train)
        return model

    scoring = "neg_root_mean_squared_error" if task == "regression" else "roc_auc"
    cv = KFold(n_splits=5, shuffle=True, random_state=RANDOM_SEED) \
         if task == "regression" \
         else StratifiedKFold(n_splits=5, shuffle=True, random_state=RANDOM_SEED)

    search = RandomizedSearchCV(
        estimator=model,
        param_distributions=param_grid,
        n_iter=n_iter,
        scoring=scoring,
        cv=cv,
        random_state=RANDOM_SEED,
        n_jobs=-1,
        verbose=0,
    )
    search.fit(X_train, y_train)
    logger.info(
        "tune_model [%s]: best score=%.4f  params=%s",
        model_name, search.best_score_, search.best_params_,
    )
    return search.best_estimator_
