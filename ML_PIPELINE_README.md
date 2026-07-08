# AlchemiX Machine Learning Pipeline

This document describes the current machine-learning pipeline used by the AlchemiX Phase 2 agent. The notebooks are a guided interface, but the source of truth is the Python code in `training/`.

## Pipeline Overview

```text
raw CSV datasets
-> dataset validation
-> time-based train/validation/test split
-> preprocessing fitted on train only
-> feature engineering per split
-> model comparison
-> champion selection
-> model export
-> validation reports
-> FastAPI prediction service
-> Go/backend/frontend consumption
```

## Source Data

Input files:

```text
datasets/universe-config.json
datasets/link_traffic_history.csv
datasets/link_telemetry.csv
datasets/link_incident_history.csv
```

Dataset roles:

| Dataset | Purpose | Target |
|---|---|---|
| `link_traffic_history.csv` | Congestion model | Congestion penalty from observed latency |
| `link_telemetry.csv` | Trust model | Under-reporting ratio / trust score |
| `link_incident_history.csv` | Targeting model | Jammed flag / targeting risk |
| `universe-config.json` | Topology metadata | Link IDs, capacity, physical latency baseline |

Validation script:

```powershell
cd training
python validate_data.py
```

Validation checks include required files, required columns, numeric tick values, link IDs, valid ranges, boolean jam flags, and non-negative latency values.

## Split Strategy

File:

```text
training/split.py
```

The split is time-based by `tick`:

```text
Train:      earliest 70%
Validation: next 15%
Test:       latest 15%
```

Important rule:

```text
Split first, then preprocess and feature-engineer each split independently.
```

The pipeline does not feature-engineer the full dataset before splitting.

## Preprocessing

File:

```text
training/preprocessing.py
```

Preprocessors are fitted on the training partition only, then applied to train, validation, and test.

### Shared Preprocessing Rules

- Missing values get indicator columns.
- Missing latency is never filled with zero.
- Link IDs are encoded using a stable map from `universe-config.json`.
- Physical latency baselines and capacities are loaded from `universe-config.json`.
- Dataset hashes and preprocessing metadata are exported with model artifacts.

### Traffic Preprocessing

Class:

```python
TrafficPreprocessor
```

Steps:

- Add `physical_latency_ms`.
- Add `capacity_units`.
- Add `link_id_encoded`.
- Add `load_units_missing`.
- Impute missing `load_units` using:

```text
load_ratio * capacity_units
```

- Fall back to training median if needed.
- Add `observed_latency_missing`.
- Keep `observed_latency_ms` as missing for target filtering.
- Add hard saturation flag:

```text
status == saturated OR load_ratio >= 0.90
```

### Telemetry Preprocessing

Class:

```python
TelemetryPreprocessor
```

Steps:

- Add `physical_latency_ms`.
- Add `link_id_encoded`.
- Add `self_reported_latency_missing`.
- Impute missing `self_reported_latency_ms` with training median.
- Compute training target:

```text
under_report_ratio =
max(0, measured_latency_ms - self_reported_latency_ms) / measured_latency_ms
```

### Incident Preprocessing

Class:

```python
IncidentPreprocessor
```

Steps:

- Add `link_id_encoded`.
- Add `traffic_share_missing`.
- If exactly one `traffic_share` is missing in a tick, reconstruct it:

```text
1 - sum(other traffic shares in the tick)
```

- Fall back to training median if needed.
- Convert `jammed_flag` into integer target `jammed`.

## Feature Engineering

File:

```text
training/features.py
```

Feature engineering is leakage-safe. Rolling and historical features use `shift(1)` before aggregation, so a row at tick `t` only uses information from ticks before `t`.

### Congestion Features

Function:

```python
build_congestion_features
```

Features:

```text
load_ratio
load_units
capacity_units
link_id_encoded
previous_load_ratio
load_ratio_change
rolling_mean_load_ratio
rolling_std_load_ratio
rate_of_load_increase
distance_to_saturation
load_units_missing
```

Target:

```text
congestion_penalty_ms = max(0, observed_latency_ms - physical_latency_ms)
```

Rows with saturated links or missing observed latency are excluded from the regression target.

### Trust Features

Function:

```python
build_trust_features
```

Features:

```text
self_reported_latency_ms
self_reported_latency_missing
physical_latency_ms
self_to_physical_ratio
deviation_from_baseline
prev_self_reported
self_reported_change
rolling_self_reported
historical_bias
below_physical_min
link_id_encoded
```

Target:

```text
under_report_ratio
```

The service converts predicted under-report ratio into:

```text
trust_score = clamp(1 - under_report_ratio / trust_scale)
```

### Targeting Features

Function:

```python
build_targeting_features
```

Features:

```text
traffic_share
traffic_share_missing
traffic_share_rank
prev_traffic_share
traffic_share_change
rolling_traffic_share
historical_jam_rate
rolling_jam_count
consecutive_high_usage
link_id_encoded
```

Target:

```text
jammed
```

## Model Training Stages

Training scripts:

```text
training/train_congestion.py
training/train_trust.py
training/train_targeting.py
```

Each script:

1. Loads raw CSV data.
2. Validates schema.
3. Sorts by `link_id` and `tick`.
4. Splits by tick.
5. Fits preprocessor on train only.
6. Transforms each split separately.
7. Builds features for each split separately.
8. Trains candidate models.
9. Compares on validation split.
10. Records `best_validation_model`.
11. Exports Go-compatible JSON model.
12. Exports Python `.joblib` best-validation model.

## Models Compared

### Congestion Regression

Script:

```text
training/train_congestion.py
```

Compared models:

```text
Median baseline
Polynomial Ridge Regression
Random Forest Regressor
Extra Trees Regressor
Gradient Boosting Regressor
HistGradientBoosting Regressor
```

Metrics:

```text
MAE
RMSE
P95 absolute error
Median absolute error
Error by load-ratio band
```

Current best validation model:

```text
Gradient Boosting
```

Current Go-compatible exported model:

```text
HistGradientBoosting
```

### Trust Regression

Script:

```text
training/train_trust.py
```

Compared models:

```text
Median baseline
Ridge Regression
Random Forest Regressor
Gradient Boosting Regressor
HistGradientBoosting Regressor
```

Metrics:

```text
MAE
RMSE
Spoof precision
Spoof recall
False-positive rate on honest links
```

Current best validation model:

```text
HistGradientBoosting
```

Current Go-compatible exported model:

```text
HistGradientBoosting
```

### Targeting Classification

Script:

```text
training/train_targeting.py
```

Compared models:

```text
Prior baseline
Logistic Regression
Random Forest Classifier
Gradient Boosting Classifier
HistGradientBoosting Classifier
```

Metrics:

```text
ROC-AUC
PR-AUC
Brier score
Log loss
Precision
Recall
Calibration bins
```

Current best validation model:

```text
Logistic Regression
```

Current Go-compatible exported model:

```text
HistGradientBoosting
```

## Exported Artifacts

Model outputs:

```text
internal/models/congestion_model.json
internal/models/trust_model.json
internal/models/targeting_model.json
internal/models/congestion_model.joblib
internal/models/trust_model.joblib
internal/models/targeting_model.joblib
internal/models/model_manifest.json
```

Reports:

```text
training/model_validation.json
training/model_validation.md
training/reports/model_selection_summary.csv
training/reports/final_test_metrics.json
```

JSON artifacts are retained for compatibility with the older Go in-process HistGradientBoosting evaluator and model inspection. They are not required by the current Go/Wails runtime.

`.joblib` artifacts are used by the Python FastAPI service and can serve non-HistGradientBoosting winners.

## FastAPI Prediction Service

Files:

```text
training/service/app.py
training/service/model_loader.py
training/service/feature_builder.py
training/service/schemas.py
```

Run:

```powershell
python -B -m uvicorn training.service.app:app --host 127.0.0.1 --port 8100
```

Endpoints:

```text
GET  /health
GET  /model-info
POST /predict/congestion
POST /predict/trust
POST /predict/targeting
POST /predict/all
```

The service currently loads:

```text
congestion_model.joblib -> Gradient Boosting
trust_model.joblib      -> HistGradientBoosting
targeting_model.joblib  -> Logistic Regression
```

## Go Runtime Integration

The Go backend now calls the Python FastAPI prediction service:

```text
POST http://127.0.0.1:8100/predict/all
```

That service loads the `.joblib` artifacts and returns congestion, trust, and targeting predictions in one response. This lets the Go/Wails runtime use the best validation models, including non-HistGradientBoosting winners.

The default service URL is `http://127.0.0.1:8100`. Override it before launching Go/Wails if needed:

```powershell
$env:ML_SERVICE_URL = "http://127.0.0.1:8101"
wails dev -skipbindings -tags native_webview2loader
```

The JSON artifacts may remain in `internal/models/` for compatibility with the older in-process evaluator and for metadata inspection, but the active runtime path is FastAPI plus `.joblib`. The FastAPI service derives live `link_id_map` values from `datasets/universe-config.json`, not from the JSON model files.

## Notebooks

Notebook workflow:

```text
training/notebooks/01_EDA.ipynb
training/notebooks/02_Data_Preprocessing.ipynb
training/notebooks/03_Feature_Engineering.ipynb
training/notebooks/04_Model_Training.ipynb
training/notebooks/05_Model_Evaluation.ipynb
training/notebooks/06_Model_Export.ipynb
```

The notebooks are for inspection and guided execution. The production logic remains in the Python scripts.

## Main Commands

Install dependencies:

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
python -m pip install -r training\requirements.txt
```

Validate:

```powershell
cd training
python validate_data.py
python -m pytest tests
cd ..
```

Train:

```powershell
python -B training\train_congestion.py
python -B training\train_trust.py
python -B training\train_targeting.py
python -B training\evaluate_all.py
```

Serve:

```powershell
python -B -m uvicorn training.service.app:app --host 127.0.0.1 --port 8100
```
