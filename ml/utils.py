"""
utils.py – Shared helpers used across all AlchemiX ML modules.

Responsibilities
----------------
- Path resolution (no hardcoded absolute paths)
- Logging setup
- Reproducibility seed
- Dataset discovery
"""

from __future__ import annotations

import logging
import os
from pathlib import Path

# ── Reproducibility ──────────────────────────────────────────────────────────
RANDOM_SEED: int = 42

# ── Project-relative paths ───────────────────────────────────────────────────
# The root of the repository is two directories above this file:
# e.g.  e:/AlchemiX/ml/utils.py  →  root = e:/AlchemiX
PROJECT_ROOT: Path = Path(__file__).resolve().parents[1]

RAW_DIR: Path       = PROJECT_ROOT / "datasets" / "raw"
CLEANED_DIR: Path   = PROJECT_ROOT / "datasets" / "cleaned"
ENGINEERED_DIR: Path = PROJECT_ROOT / "datasets" / "engineered"
MODELS_DIR: Path    = PROJECT_ROOT / "models"
NOTEBOOKS_DIR: Path = PROJECT_ROOT / "notebooks"

# Ensure output directories exist at import time
for _d in (CLEANED_DIR, ENGINEERED_DIR, MODELS_DIR):
    _d.mkdir(parents=True, exist_ok=True)


# ── Dataset file names ───────────────────────────────────────────────────────
TRAFFIC_RAW_FILE: Path       = RAW_DIR / "link_traffic_history.csv"
TELEMETRY_RAW_FILE: Path     = RAW_DIR / "link_telemetry.csv"
INCIDENT_RAW_FILE: Path      = RAW_DIR / "link_incident_history.csv"

TRAFFIC_CLEAN_FILE: Path     = CLEANED_DIR / "traffic_clean.csv"
TELEMETRY_CLEAN_FILE: Path   = CLEANED_DIR / "telemetry_clean.csv"
INCIDENT_CLEAN_FILE: Path    = CLEANED_DIR / "incident_clean.csv"

# Train / Test splits (written by Notebook 02, read by Notebook 03)
TRAFFIC_TRAIN_CLEAN: Path    = CLEANED_DIR / "traffic_train.csv"
TRAFFIC_TEST_CLEAN: Path     = CLEANED_DIR / "traffic_test.csv"
TELEMETRY_TRAIN_CLEAN: Path  = CLEANED_DIR / "telemetry_train.csv"
TELEMETRY_TEST_CLEAN: Path   = CLEANED_DIR / "telemetry_test.csv"
INCIDENT_TRAIN_CLEAN: Path   = CLEANED_DIR / "incident_train.csv"
INCIDENT_TEST_CLEAN: Path    = CLEANED_DIR / "incident_test.csv"

TRAFFIC_ENG_FILE: Path       = ENGINEERED_DIR / "traffic_engineered.csv"
TELEMETRY_ENG_FILE: Path     = ENGINEERED_DIR / "telemetry_engineered.csv"
INCIDENT_ENG_FILE: Path      = ENGINEERED_DIR / "incident_engineered.csv"

# Engineered train / test splits (written by Notebook 03, read by Notebook 04)
TRAFFIC_TRAIN_ENG: Path      = ENGINEERED_DIR / "traffic_train_eng.csv"
TRAFFIC_TEST_ENG: Path       = ENGINEERED_DIR / "traffic_test_eng.csv"
TELEMETRY_TRAIN_ENG: Path    = ENGINEERED_DIR / "telemetry_train_eng.csv"
TELEMETRY_TEST_ENG: Path     = ENGINEERED_DIR / "telemetry_test_eng.csv"
INCIDENT_TRAIN_ENG: Path     = ENGINEERED_DIR / "incident_train_eng.csv"
INCIDENT_TEST_ENG: Path      = ENGINEERED_DIR / "incident_test_eng.csv"


# ── Model file names ─────────────────────────────────────────────────────────
CONGESTION_MODEL_FILE: Path  = MODELS_DIR / "congestion.joblib"
TRUST_MODEL_FILE: Path       = MODELS_DIR / "trust.joblib"
TARGETING_MODEL_FILE: Path   = MODELS_DIR / "targeting.joblib"


# ── Known planet names and link IDs ─────────────────────────────────────────
PLANETS: list[str] = [
    "Aegis", "Boreas", "Dawn", "Elysium", "Fenix", "Caelum"
]

LINKS: list[str] = [
    "Aegis-Boreas", "Aegis-Dawn", "Aegis-Elysium",
    "Boreas-Dawn", "Boreas-Elysium", "Boreas-Fenix",
    "Caelum-Dawn", "Caelum-Elysium", "Caelum-Fenix",
    "Dawn-Elysium", "Dawn-Fenix", "Elysium-Fenix",
]


# ── Logging ──────────────────────────────────────────────────────────────────
def get_logger(name: str) -> logging.Logger:
    """Return a consistently formatted logger."""
    logger = logging.getLogger(name)
    if not logger.handlers:
        handler = logging.StreamHandler()
        fmt = logging.Formatter(
            "%(asctime)s  %(levelname)-8s  %(name)s  %(message)s",
            datefmt="%H:%M:%S",
        )
        handler.setFormatter(fmt)
        logger.addHandler(handler)
        logger.setLevel(logging.INFO)
    return logger


# ── Convenience ──────────────────────────────────────────────────────────────
def env(key: str, default: str = "") -> str:
    """Read an environment variable with a safe default."""
    return os.environ.get(key, default)
