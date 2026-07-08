# AlchemiX Model Validation Report

**Generated:** 2026-07-08T18:00:50.178229+00:00

## congestion_model.json (`congestion_regressor`)
- **Version:** 2.0.0
- **Dataset Hash:** `3073b5903981d6bf`
- **Feature Count:** 11

### Metrics
- **val_mae_ms**: 17787.3489
- **val_rmse_ms**: 105668.3497
- **val_medae_ms**: 4306.9693
- **val_n_samples**: 858
- **train_n_samples**: 4023
- **saturated_train**: 2
- **saturated_val**: 1
- **saturated_test**: 0
- **error_by_load_band**:
  - `0.00-0.50`: {'n': 725, 'mae_ms': 7060.672531919238}
  - `0.50-0.75`: {'n': 123, 'mae_ms': 34213.391115047496}
  - `0.75-0.90`: {'n': 10, 'mae_ms': 593431.0630320776}
- **selected_model**: HistGradientBoosting
- **best_validation_model**: Gradient Boosting
- **model_comparison**: [{'model': 'Median baseline', 'val_mae_ms': 74515.7584, 'val_rmse_ms': 153245.3413, 'val_p95_abs_error_ms': 328216.7347, 'go_exportable': False}, {'model': 'Polynomial Ridge', 'val_mae_ms': 33922.9355, 'val_rmse_ms': 54190.3077, 'val_p95_abs_error_ms': 117766.5305, 'go_exportable': False}, {'model': 'Random Forest', 'val_mae_ms': 16823.9037, 'val_rmse_ms': 57369.5917, 'val_p95_abs_error_ms': 49807.1562, 'go_exportable': False}, {'model': 'Extra Trees', 'val_mae_ms': 17245.1721, 'val_rmse_ms': 51579.5839, 'val_p95_abs_error_ms': 53760.6439, 'go_exportable': False}, {'model': 'Gradient Boosting', 'val_mae_ms': 16345.3029, 'val_rmse_ms': 62559.6227, 'val_p95_abs_error_ms': 46615.443, 'go_exportable': False}, {'model': 'HistGradientBoosting', 'val_mae_ms': 17787.3489, 'val_rmse_ms': 105668.3497, 'val_p95_abs_error_ms': 44300.006, 'go_exportable': True}]

## trust_model.json (`trust_regressor`)
- **Version:** 2.0.0
- **Dataset Hash:** `70220057ce46587b`
- **Feature Count:** 11

### Metrics
- **val_mae**: 0.042811
- **val_rmse**: 0.087755
- **spoof_precision**: 0.7227
- **spoof_recall**: 0.735
- **fp_rate_honest**: 0.0421
- **spoof_threshold**: 0.15
- **trust_scale**: 0.5
- **selected_model**: HistGradientBoosting
- **best_validation_model**: HistGradientBoosting
- **model_comparison**: [{'model': 'Median baseline', 'val_mae': 0.06387, 'val_rmse': 0.150687, 'go_exportable': False}, {'model': 'Ridge', 'val_mae': 0.047108, 'val_rmse': 0.092669, 'go_exportable': False}, {'model': 'Random Forest', 'val_mae': 0.043321, 'val_rmse': 0.087843, 'go_exportable': False}, {'model': 'Gradient Boosting', 'val_mae': 0.044017, 'val_rmse': 0.089778, 'go_exportable': False}, {'model': 'HistGradientBoosting', 'val_mae': 0.042811, 'val_rmse': 0.087755, 'go_exportable': True}]

## targeting_model.json (`targeting_classifier`)
- **Version:** 2.0.0
- **Dataset Hash:** `db5261b71a46e8b9`
- **Feature Count:** 10

### Metrics
- **val_roc_auc**: 0.5287
- **val_pr_auc**: 0.1212
- **val_brier**: 0.0715
- **val_precision**: 0.0
- **val_recall**: 0.0
- **calibration_threshold**: 0.5
- **platt_a**: -0.821862
- **platt_b**: 2.74107
- **calibration_bins**: [{'bin': '0.00-0.20', 'n': 888, 'mean_predicted': 0.0777, 'mean_true': 0.0777}]
- **class_distribution**:
  - `train_jammed`: 343
  - `train_not_jammed`: 3845
  - `val_jammed`: 69
  - `val_not_jammed`: 819
- **selected_model**: HistGradientBoosting
- **best_validation_model**: Logistic Regression
- **model_comparison**: [{'model': 'Prior baseline', 'val_roc_auc': 0.5, 'val_pr_auc': 0.0777, 'val_brier': 0.0717, 'val_log_loss': 0.2732, 'go_exportable': False}, {'model': 'Logistic Regression', 'val_roc_auc': 0.6203, 'val_pr_auc': 0.1407, 'val_brier': 0.2338, 'val_log_loss': 0.6605, 'go_exportable': False}, {'model': 'Random Forest', 'val_roc_auc': 0.5718, 'val_pr_auc': 0.1218, 'val_brier': 0.1605, 'val_log_loss': 0.5027, 'go_exportable': False}, {'model': 'Gradient Boosting', 'val_roc_auc': 0.6013, 'val_pr_auc': 0.1247, 'val_brier': 0.0742, 'val_log_loss': 0.2772, 'go_exportable': False}, {'model': 'HistGradientBoosting', 'val_roc_auc': 0.5287, 'val_pr_auc': 0.1212, 'val_brier': 0.1576, 'val_log_loss': 0.4908, 'go_exportable': True}]

