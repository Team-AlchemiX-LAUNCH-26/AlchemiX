import numpy as np
import pandas as pd
from features import build_congestion_features, build_trust_features, build_targeting_features

def test_congestion_leakage():
    df = pd.DataFrame({
        "link_id": ["A", "A", "A", "B"],
        "tick": [0, 1, 2, 0],
        "load_ratio": [0.1, 0.5, 0.9, 0.2],
        "load_units": [10, 50, 90, 20],
        "capacity_units": [100, 100, 100, 100],
        "physical_latency_ms": [10.0, 10.0, 10.0, 20.0],
        "link_id_encoded": [0, 0, 0, 1],
        "load_units_missing": [0, 0, 0, 0]
    })
    
    out = build_congestion_features(df)
    
    # Tick 2 for link A (index 2)
    # rolling_mean_load_ratio should use ticks 0 and 1 -> (0.1 + 0.5) / 2 = 0.3
    # It must NOT include tick 2's load_ratio of 0.9.
    assert out.iloc[2]["rolling_mean_load_ratio"] == 0.3
    
    # Tick 1 for link A (index 1)
    # previous_load_ratio should be tick 0's value: 0.1
    assert out.iloc[1]["previous_load_ratio"] == 0.1
    
    # Tick 0 for link A (index 0)
    # previous_load_ratio should fallback to current: 0.1
    assert out.iloc[0]["previous_load_ratio"] == 0.1

def test_trust_leakage():
    df = pd.DataFrame({
        "link_id": ["A", "A", "A"],
        "tick": [0, 1, 2],
        "self_reported_latency_ms": [10.0, 15.0, 20.0],
        "physical_latency_ms": [10.0, 10.0, 10.0],
        "self_reported_latency_missing": [0, 0, 0],
        "under_report_ratio": [0.0, 0.1, 0.5],
        "link_id_encoded": [0, 0, 0]
    })
    
    out = build_trust_features(df)
    
    # historical_bias is an expanding mean.
    # At tick 2, it should be the mean of under_report_ratio from tick 0 and 1.
    # (0.0 + 0.1) / 2 = 0.05
    # It must NOT include tick 2's ratio of 0.5.
    assert out.iloc[2]["historical_bias"] == 0.05
    
    # rolling_self_reported at tick 2 should use tick 0 and 1
    # (10.0 + 15.0) / 2 = 12.5
    assert out.iloc[2]["rolling_self_reported"] == 12.5

def test_targeting_leakage():
    df = pd.DataFrame({
        "link_id": ["A", "A", "A"],
        "tick": [0, 1, 2],
        "traffic_share": [0.1, 0.6, 0.9],
        "traffic_share_missing": [0, 0, 0],
        "jammed": [0, 1, 1],
        "link_id_encoded": [0, 0, 0]
    })
    
    out = build_targeting_features(df)
    
    # consecutive_high_usage uses rolling sum of (traffic_share > median)
    # Median of traffic_share is 1.0 / 1 link = 1.0. None are > 1.0.
    # We should just test historical_jam_rate.
    # At tick 2, it uses tick 0 and 1: (0 + 1) / 2 = 0.5
    assert out.iloc[2]["historical_jam_rate"] == 0.5
    
    # rolling_traffic_share at tick 2 uses tick 0 and 1: (0.1 + 0.6) / 2 = 0.35
    np.testing.assert_almost_equal(out.iloc[2]["rolling_traffic_share"], 0.35)
