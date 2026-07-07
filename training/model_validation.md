# AlchemiX Model Validation Report

**Generated:** 2026-07-07T15:54:09.956807+00:00

## congestion_model.json (`congestion_regressor`)
- **Version:** 2.0.0
- **Dataset Hash:** `61c62f57ec7d183c`
- **Feature Count:** 11

### Metrics
- **val_mae_ms**: 17521.1512
- **val_rmse_ms**: 98609.3473
- **val_medae_ms**: 4309.7821
- **val_n_samples**: 858
- **train_n_samples**: 4023
- **saturated_train**: 2
- **saturated_val**: 1
- **saturated_test**: 0
- **error_by_load_band**:
  - `0.00-0.50`: {'n': 725, 'mae_ms': 6985.984308838837}
  - `0.50-0.75`: {'n': 123, 'mae_ms': 36618.544846756144}
  - `0.75-0.90`: {'n': 10, 'mae_ms': 546422.8081290955}
- **selected_model**: HGBR(l_rate=0.10,depth=5)

## trust_model.json (`trust_regressor`)
- **Version:** 2.0.0
- **Dataset Hash:** `f52fb59c227ab7fb`
- **Feature Count:** 11

### Metrics
- **val_mae**: 0.04316
- **val_rmse**: 0.086433
- **spoof_precision**: 0.6516
- **spoof_recall**: 0.8632
- **fp_rate_honest**: 0.069
- **spoof_threshold**: 0.15
- **trust_scale**: 0.5
- **selected_model**: HGBR(lr=0.05,d=4)

## targeting_model.json (`targeting_classifier`)
- **Version:** 2.0.0
- **Dataset Hash:** `26bca83b65dace37`
- **Feature Count:** 10

### Metrics
- **val_roc_auc**: 0.5511
- **val_pr_auc**: -0.0418
- **val_brier**: 0.0722
- **val_precision**: 0.0
- **val_recall**: 0.0
- **calibration_threshold**: 0.5
- **platt_a**: -1.358259
- **platt_b**: 2.842998
- **calibration_bins**: [{'bin': '0.00-0.20', 'n': 900, 'mean_predicted': 0.0789, 'mean_true': 0.0789}]
- **class_distribution**:
  - `train_jammed`: 343
  - `train_not_jammed`: 3845
  - `val_jammed`: 71
  - `val_not_jammed`: 829
- **selected_model**: HGBC(lr=0.10,d=5)

