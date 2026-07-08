"""
evaluation.py – Model evaluation metrics and charts for AlchemiX Phase 2.

Responsibilities
----------------
- Regression: MAE, RMSE, R², MAPE
- Classification: Accuracy, Precision, Recall, F1, ROC-AUC, Confusion Matrix
- Feature importance
- Residual analysis
- Learning curves
- Calibration curve
- Professional matplotlib charts
"""

from __future__ import annotations

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.gridspec as gridspec
from sklearn.metrics import (
    mean_absolute_error,
    mean_squared_error,
    r2_score,
    accuracy_score,
    precision_score,
    recall_score,
    f1_score,
    roc_auc_score,
    confusion_matrix,
    RocCurveDisplay,
    PrecisionRecallDisplay,
)
from sklearn.calibration import CalibrationDisplay
from sklearn.inspection import permutation_importance
from sklearn.model_selection import learning_curve, KFold, StratifiedKFold

from ml.utils import get_logger, RANDOM_SEED

logger = get_logger(__name__)

# ── Style ─────────────────────────────────────────────────────────────────────
plt.rcParams.update({
    "figure.facecolor": "#0d1117",
    "axes.facecolor": "#161b22",
    "axes.edgecolor": "#30363d",
    "axes.labelcolor": "#c9d1d9",
    "xtick.color": "#c9d1d9",
    "ytick.color": "#c9d1d9",
    "text.color": "#c9d1d9",
    "grid.color": "#21262d",
    "grid.linewidth": 0.8,
    "figure.titlesize": 14,
    "axes.titlesize": 12,
    "axes.labelsize": 10,
    "font.family": "monospace",
})
ACCENT  = "#58a6ff"
SUCCESS = "#3fb950"
DANGER  = "#f85149"
WARN    = "#e3b341"


# ── Regression metrics ────────────────────────────────────────────────────────

def regression_metrics(y_true: np.ndarray, y_pred: np.ndarray) -> dict[str, float]:
    """Return MAE, RMSE, R², MAPE as a dictionary."""
    mae  = mean_absolute_error(y_true, y_pred)
    rmse = np.sqrt(mean_squared_error(y_true, y_pred))
    r2   = r2_score(y_true, y_pred)
    # MAPE – avoid division by zero
    mask = y_true != 0
    mape = np.mean(np.abs((y_true[mask] - y_pred[mask]) / y_true[mask])) * 100 if mask.any() else float("nan")
    metrics = {"MAE": mae, "RMSE": rmse, "R2": r2, "MAPE_%": mape}
    for k, v in metrics.items():
        logger.info("  %-10s = %.4f", k, v)
    return metrics


# ── Classification metrics ────────────────────────────────────────────────────

def classification_metrics(y_true: np.ndarray, y_pred: np.ndarray, y_prob: np.ndarray | None = None) -> dict[str, float]:
    """Return Accuracy, Precision, Recall, F1, ROC-AUC."""
    metrics = {
        "Accuracy":  accuracy_score(y_true, y_pred),
        "Precision": precision_score(y_true, y_pred, zero_division=0),
        "Recall":    recall_score(y_true, y_pred, zero_division=0),
        "F1":        f1_score(y_true, y_pred, zero_division=0),
    }
    if y_prob is not None:
        metrics["ROC_AUC"] = roc_auc_score(y_true, y_prob)
    for k, v in metrics.items():
        logger.info("  %-12s = %.4f", k, v)
    return metrics


# ── Residual plot ─────────────────────────────────────────────────────────────

def plot_residuals(y_true: np.ndarray, y_pred: np.ndarray, title: str = "Residual Analysis") -> plt.Figure:
    residuals = y_true - y_pred
    fig, axes = plt.subplots(1, 2, figsize=(14, 5))
    fig.suptitle(title, fontsize=14, color="#c9d1d9")

    axes[0].scatter(y_pred, residuals, alpha=0.4, color=ACCENT, s=10)
    axes[0].axhline(0, color=DANGER, linewidth=1.5, linestyle="--")
    axes[0].set_xlabel("Predicted")
    axes[0].set_ylabel("Residual")
    axes[0].set_title("Residuals vs Predicted")
    axes[0].grid(True)

    axes[1].hist(residuals, bins=50, color=ACCENT, edgecolor="#21262d", alpha=0.85)
    axes[1].set_xlabel("Residual")
    axes[1].set_ylabel("Count")
    axes[1].set_title("Residual Distribution")
    axes[1].grid(True)

    plt.tight_layout()
    return fig


# ── Confusion matrix ──────────────────────────────────────────────────────────

def plot_confusion_matrix(y_true: np.ndarray, y_pred: np.ndarray, title: str = "Confusion Matrix") -> plt.Figure:
    cm = confusion_matrix(y_true, y_pred)
    fig, ax = plt.subplots(figsize=(5, 4))
    fig.patch.set_facecolor("#0d1117")
    ax.set_facecolor("#161b22")
    im = ax.imshow(cm, cmap="Blues")
    ax.set_xticks([0, 1]); ax.set_yticks([0, 1])
    ax.set_xticklabels(["False", "True"]); ax.set_yticklabels(["False", "True"])
    ax.set_xlabel("Predicted"); ax.set_ylabel("Actual")
    ax.set_title(title, color="#c9d1d9")
    for i in range(2):
        for j in range(2):
            ax.text(j, i, str(cm[i, j]), ha="center", va="center",
                    color="white" if cm[i, j] > cm.max() / 2 else "#0d1117", fontsize=14)
    plt.colorbar(im, ax=ax)
    plt.tight_layout()
    return fig


# ── Feature importance ────────────────────────────────────────────────────────

def plot_feature_importance(
    model: object,
    feature_names: list[str],
    top_n: int = 20,
    title: str = "Feature Importance",
) -> plt.Figure:
    """Plot feature importance from tree-based models."""
    try:
        importances = model.feature_importances_
    except AttributeError:
        logger.warning("Model has no feature_importances_ attribute – skipping.")
        return plt.figure()

    idx = np.argsort(importances)[-top_n:]
    fig, ax = plt.subplots(figsize=(10, max(4, top_n * 0.35)))
    ax.barh(
        [feature_names[i] for i in idx],
        importances[idx],
        color=ACCENT, edgecolor="#21262d",
    )
    ax.set_xlabel("Importance")
    ax.set_title(title, color="#c9d1d9")
    ax.grid(axis="x")
    plt.tight_layout()
    return fig


# ── Learning curve ────────────────────────────────────────────────────────────

def plot_learning_curve(
    model: object,
    X: np.ndarray,
    y: np.ndarray,
    task: str = "regression",
    title: str = "Learning Curve",
) -> plt.Figure:
    scoring = "neg_root_mean_squared_error" if task == "regression" else "roc_auc"
    cv = (
        KFold(n_splits=5, shuffle=True, random_state=RANDOM_SEED)
        if task == "regression"
        else StratifiedKFold(n_splits=5, shuffle=True, random_state=RANDOM_SEED)
    )
    train_sizes, train_scores, val_scores = learning_curve(
        model, X, y,
        cv=cv,
        scoring=scoring,
        train_sizes=np.linspace(0.1, 1.0, 10),
        n_jobs=-1,
    )
    fig, ax = plt.subplots(figsize=(10, 5))
    ax.plot(train_sizes, -train_scores.mean(axis=1), label="Train", color=SUCCESS)
    ax.fill_between(train_sizes,
                    -train_scores.mean(axis=1) - train_scores.std(axis=1),
                    -train_scores.mean(axis=1) + train_scores.std(axis=1),
                    alpha=0.2, color=SUCCESS)
    ax.plot(train_sizes, -val_scores.mean(axis=1), label="Validation", color=ACCENT)
    ax.fill_between(train_sizes,
                    -val_scores.mean(axis=1) - val_scores.std(axis=1),
                    -val_scores.mean(axis=1) + val_scores.std(axis=1),
                    alpha=0.2, color=ACCENT)
    ax.set_xlabel("Training Samples")
    ax.set_ylabel(scoring.replace("neg_", ""))
    ax.set_title(title, color="#c9d1d9")
    ax.legend()
    ax.grid(True)
    plt.tight_layout()
    return fig


# ── Prediction vs Actual ──────────────────────────────────────────────────────

def plot_pred_vs_actual(y_true: np.ndarray, y_pred: np.ndarray, title: str = "Predicted vs Actual") -> plt.Figure:
    fig, ax = plt.subplots(figsize=(7, 7))
    ax.scatter(y_true, y_pred, alpha=0.3, s=8, color=ACCENT)
    lo, hi = min(y_true.min(), y_pred.min()), max(y_true.max(), y_pred.max())
    ax.plot([lo, hi], [lo, hi], color=DANGER, linewidth=1.5, linestyle="--", label="Perfect fit")
    ax.set_xlabel("Actual")
    ax.set_ylabel("Predicted")
    ax.set_title(title, color="#c9d1d9")
    ax.legend()
    ax.grid(True)
    plt.tight_layout()
    return fig


# ── Model comparison table ────────────────────────────────────────────────────

def print_comparison_table(results_df: pd.DataFrame) -> None:
    """Pretty-print a model comparison DataFrame."""
    print("\n" + "=" * 60)
    print("  MODEL COMPARISON")
    print("=" * 60)
    print(results_df.to_string(index=False))
    print("=" * 60 + "\n")
