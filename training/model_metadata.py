"""
Shared model metadata builder for the AlchemiX ML pipeline.

Every exported model artifact includes:
- model_version
- task type
- feature_names (in exact inference order)
- random_seed
- tick ranges for all three splits
- dataset hash
- creation timestamp
- Python version
- scikit-learn version
- preprocessing metadata
- metrics
"""
from __future__ import annotations

import platform
import sys
from datetime import datetime, timezone
from typing import Any

import sklearn


MODEL_VERSION = "2.0.0"
RANDOM_STATE = 42


def build_metadata(
    task: str,
    feature_names: list[str],
    split_info: dict[str, Any],
    preprocessing_meta: dict[str, Any],
    metrics: dict[str, Any],
    extra: dict[str, Any] | None = None,
) -> dict[str, Any]:
    """
    Build a complete model metadata dict suitable for JSON export.

    Parameters
    ----------
    task : one of "congestion_regressor", "trust_regressor", "targeting_classifier"
    feature_names : ordered list matching the training feature matrix columns
    split_info : dict with keys train_tick_range, val_tick_range, test_tick_range,
                 train_rows, val_rows, test_rows, dataset_hash
    preprocessing_meta : dict from a preprocessor's .metadata() method
    metrics : dict of metric_name → value (all float)
    extra : additional model-specific keys (e.g. link_id_map, physical_baselines_ms)
    """
    meta: dict[str, Any] = {
        "model_version": MODEL_VERSION,
        "task": task,
        "feature_names": feature_names,
        "feature_count": len(feature_names),
        "random_state": RANDOM_STATE,
        "training_tick_range": split_info["train_tick_range"],
        "validation_tick_range": split_info["val_tick_range"],
        "test_tick_range": split_info["test_tick_range"],
        "training_rows": split_info.get("train_rows", None),
        "validation_rows": split_info.get("val_rows", None),
        "test_rows": split_info.get("test_rows", None),
        "dataset_hash": split_info.get("dataset_hash", None),
        "created_at": datetime.now(tz=timezone.utc).isoformat(),
        "python_version": sys.version.split(" ")[0],
        "sklearn_version": sklearn.__version__,
        "platform": platform.system(),
        "preprocessing": preprocessing_meta,
        "metrics": metrics,
    }
    if extra:
        meta.update(extra)
    return meta


def split_info_from_result(split, dataset_hash: str) -> dict[str, Any]:
    """Convert a SplitResult into the dict that build_metadata expects."""
    return {
        "train_tick_range": list(split.train_tick_range),
        "val_tick_range":   list(split.val_tick_range),
        "test_tick_range":  list(split.test_tick_range),
        "train_rows": len(split.train),
        "val_rows":   len(split.val),
        "test_rows":  len(split.test),
        "dataset_hash": dataset_hash,
    }
