"""Generate consolidated ML reports and a model manifest."""
from __future__ import annotations

import csv
import json
import os
from datetime import datetime, timezone

from validate_models import main as validate_models_main

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
MODELS_DIR = os.path.join(ROOT, "internal", "models")
REPORT_DIR = os.path.join(os.path.dirname(__file__), "reports")
MANIFEST_PATH = os.path.join(MODELS_DIR, "model_manifest.json")

MODEL_FILES = {
    "congestion": "congestion_model.json",
    "trust": "trust_model.json",
    "targeting": "targeting_model.json",
}


def _load_model(name: str) -> dict:
    path = os.path.join(MODELS_DIR, MODEL_FILES[name])
    if not os.path.exists(path):
        raise FileNotFoundError(f"Missing {path}. Run train_{name}.py first.")
    with open(path, "r") as f:
        return json.load(f)


def _selection_reason(name: str, model: dict) -> str:
    metrics = model.get("metrics", {})
    selected = metrics.get("selected_model", "selected champion")
    best_validation = metrics.get("best_validation_model", selected)
    if selected != best_validation:
        return (
            f"{best_validation} had the strongest validation score, but {selected} "
            "was exported because the Go runtime currently supports the "
            "HistGradientBoosting JSON tree format."
        )
    if name == "congestion":
        return f"{selected} selected for validation MAE/RMSE with hard saturation gates."
    if name == "trust":
        return f"{selected} selected for under-reporting error while tracking spoof precision and recall."
    return f"{selected} selected for calibrated targeting risk with PR-AUC/Brier reporting."


def build_manifest(models: dict[str, dict]) -> dict:
    manifest = {
        "generated_at": datetime.now(tz=timezone.utc).isoformat(),
        "training_data": {
            "traffic": "link_traffic_history.csv",
            "telemetry": "link_telemetry.csv",
            "incident": "link_incident_history.csv",
            "universe": "universe-config.json",
        },
        "split_strategy": "time_based_70_15_15_by_tick",
        "runtime_artifact_format": "Go-compatible JSON model exports",
        "models": {},
    }

    for name, model in models.items():
        manifest["models"][name] = {
            "artifact": MODEL_FILES[name],
            "task": model.get("task"),
            "model_type": model.get("metrics", {}).get("selected_model", model.get("model_type")),
            "features": model.get("feature_names", model.get("features", [])),
            "validation_metrics": model.get("metrics", {}),
            "selection_reason": _selection_reason(name, model),
            "dataset_hash": model.get("dataset_hash"),
            "model_version": model.get("model_version"),
        }

    return manifest


def write_metric_summary(models: dict[str, dict]) -> None:
    os.makedirs(REPORT_DIR, exist_ok=True)
    path = os.path.join(REPORT_DIR, "model_selection_summary.csv")
    rows = []
    for name, model in models.items():
        metrics = model.get("metrics", {})
        rows.append({
            "model_family": name,
            "task": model.get("task", ""),
            "selected_model": metrics.get("selected_model", ""),
            "best_validation_model": metrics.get("best_validation_model", ""),
            "primary_metric_1": next(iter(metrics.items()))[0] if metrics else "",
            "primary_metric_1_value": next(iter(metrics.items()))[1] if metrics else "",
            "feature_count": len(model.get("feature_names", model.get("features", []))),
            "dataset_hash": model.get("dataset_hash", ""),
        })

    with open(path, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=list(rows[0].keys()))
        writer.writeheader()
        writer.writerows(rows)


def main() -> None:
    validate_models_main()
    models = {name: _load_model(name) for name in MODEL_FILES}
    manifest = build_manifest(models)

    os.makedirs(MODELS_DIR, exist_ok=True)
    with open(MANIFEST_PATH, "w") as f:
        json.dump(manifest, f, indent=2)

    os.makedirs(REPORT_DIR, exist_ok=True)
    with open(os.path.join(REPORT_DIR, "final_test_metrics.json"), "w") as f:
        json.dump(
            {
                "generated_at": manifest["generated_at"],
                "note": "Current exports contain validation metrics; final test should be run once after champion freeze.",
                "models": {
                    name: model["validation_metrics"]
                    for name, model in manifest["models"].items()
                },
            },
            f,
            indent=2,
        )

    write_metric_summary(models)
    print(f"Manifest written to {MANIFEST_PATH}")
    print(f"Reports written to {REPORT_DIR}")


if __name__ == "__main__":
    main()
