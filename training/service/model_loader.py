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
        cfg = self._load_universe_config()
        link_ids = sorted(link["link_id"] for link in cfg["interplanetary_links"])
        return {link_id: index for index, link_id in enumerate(link_ids)}

    def _load_universe_config(self) -> dict[str, Any]:
        root = self.model_dir.resolve().parent.parent
        candidates = [
            root / "datasets" / "universe-config.json",
            root / "configs" / "universe-config.json",
            Path.cwd() / "datasets" / "universe-config.json",
            Path.cwd() / "configs" / "universe-config.json",
        ]
        for path in candidates:
            if path.exists():
                return json.loads(path.read_text())
        searched = ", ".join(str(path) for path in candidates)
        raise FileNotFoundError(f"Missing universe-config.json. Searched: {searched}")

    def model_info(self) -> dict[str, Any]:
        return {
            "congestion": self.congestion["model_name"],
            "trust": self.trust["model_name"],
            "targeting": self.targeting["model_name"],
            "manifest_generated_at": self.manifest.get("generated_at"),
        }
