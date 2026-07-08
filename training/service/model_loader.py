from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import joblib


class PythonModelStore:
    def __init__(self, model_dir: Path) -> None:
        self.model_dir = model_dir
        self.congestion = self._load("congestion_model.joblib")
        self.trust = self._load("trust_model.joblib")
        self.targeting = self._load("targeting_model.joblib")
        self.manifest = self._load_manifest()

    def _load(self, filename: str) -> dict[str, Any]:
        path = self.model_dir / filename
        if not path.exists():
            raise FileNotFoundError(
                f"Missing {path}. Run the training scripts to export joblib models."
            )
        return joblib.load(path)

    def _load_manifest(self) -> dict[str, Any]:
        path = self.model_dir / "model_manifest.json"
        if not path.exists():
            return {}
        return json.loads(path.read_text())

    def link_id_map(self) -> dict[str, int]:
        path = self.model_dir / "congestion_model.json"
        if not path.exists():
            return {}
        model = json.loads(path.read_text())
        return model.get("link_id_map", {})

    def model_info(self) -> dict[str, Any]:
        return {
            "congestion": self.congestion["model_name"],
            "trust": self.trust["model_name"],
            "targeting": self.targeting["model_name"],
            "manifest_generated_at": self.manifest.get("generated_at"),
        }
