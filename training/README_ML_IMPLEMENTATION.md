# AlchemiX Phase 2 Machine Learning Implementation

This repository keeps the ML implementation in `training/` and exports Go-compatible JSON models to `internal/models/`. It intentionally does not move files into a separate `analytics/` tree because the Go agent already loads the current model artifacts.

## Delivered ML Signals

- `congestion`: predicts extra latency pressure and uses a hard saturation rule.
- `trust`: estimates latency under-reporting and converts it to a trust score in Go.
- `targeting`: predicts Chimera jamming risk from traffic-share history.

The Go agent combines these outputs in True Cost:

```text
physical_latency_ms
+ predicted_congestion_penalty_ms
+ trust penalty
+ targeting-risk penalty
+ uncertainty penalty
+ switching penalty
```

## Data Layout

Required source files are in `datasets/`:

- `universe-config.json`
- `link_traffic_history.csv`
- `link_telemetry.csv`
- `link_incident_history.csv`
- `DATASETS.md`

Run validation before training:

```powershell
cd training
python validate_data.py
```

## Environment

From the repository root:

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
python -m pip install --upgrade pip
python -m pip install -r training\requirements.txt
```

In VS Code, select `.venv\Scripts\python.exe` as the Python interpreter.

## Training Workflow

Run from `training/`:

```powershell
python validate_data.py
python train_congestion.py
python train_trust.py
python train_targeting.py
python evaluate_all.py
python -m pytest tests
```

Outputs:

- `internal/models/congestion_model.json`
- `internal/models/trust_model.json`
- `internal/models/targeting_model.json`
- `internal/models/congestion_model.joblib`
- `internal/models/trust_model.joblib`
- `internal/models/targeting_model.joblib`
- `internal/models/model_manifest.json`
- `training/model_validation.json`
- `training/model_validation.md`
- `training/reports/model_selection_summary.csv`
- `training/reports/final_test_metrics.json`

## Model Families Compared

Each trainer compares multiple meaningful algorithm families on the validation split:

- Congestion regression: median baseline, polynomial ridge, random forest, extra trees, gradient boosting, HistGradientBoosting.
- Trust regression: median baseline, ridge, random forest, gradient boosting, HistGradientBoosting.
- Targeting classification: prior baseline, logistic regression, random forest, gradient boosting, HistGradientBoosting.

`best_validation_model` records the strongest validation model. `selected_model` records the exported runtime model. These can differ because the current Go backend reads the HistGradientBoosting JSON tree format.

## FastAPI Prediction Service

The Python service loads the `.joblib` best-validation models, so it can serve non-HGB winners such as Gradient Boosting or Logistic Regression.

Run from the repository root:

```powershell
.\.venv\Scripts\Activate.ps1
python -B -m uvicorn training.service.app:app --host 127.0.0.1 --port 8100
```

Endpoints:

- `GET /health`
- `GET /model-info`
- `POST /predict/congestion`
- `POST /predict/trust`
- `POST /predict/targeting`
- `POST /predict/all`

## Safety Rules

- Time-based train/validation/test split is mandatory.
- Data is split before feature engineering; each partition builds lagged and rolling features independently.
- Rolling and historical features use `shift(1)` before aggregation.
- Missing latency is never filled with zero.
- Saturated links are rejected by hard rules, not resurrected by ML.
- Model artifacts include feature order, dataset hash, metrics, split metadata, and preprocessing metadata.

## Jupyter Notebooks

Use notebooks for the staged ML workflow and inspection. Keep the authoritative model logic in the Python scripts above so exported artifacts stay reproducible.

Recommended notebook flow:

```powershell
cd training
python -m ipykernel install --user --name alchemix-ml --display-name "AlchemiX ML"
jupyter notebook notebooks\01_EDA.ipynb
```

Notebook sequence:

- `notebooks/01_EDA.ipynb`
- `notebooks/02_Data_Preprocessing.ipynb`
- `notebooks/03_Feature_Engineering.ipynb`
- `notebooks/04_Model_Training.ipynb`
- `notebooks/05_Model_Evaluation.ipynb`
- `notebooks/06_Model_Export.ipynb`

Notebook outputs should be saved to `training/reports/` when they are useful for demos. Avoid committing large generated notebook checkpoints or ad hoc model artifacts.
