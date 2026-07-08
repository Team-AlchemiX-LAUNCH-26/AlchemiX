import numpy as np
import pandas as pd
from preprocessing import TrafficPreprocessor, TelemetryPreprocessor

def test_traffic_preprocessor_imputation():
    train_df = pd.DataFrame({
        "link_id": ["A-B", "B-C"],
        "tick": [0, 0],
        "load_units": [100.0, np.nan],
        "load_ratio": [0.5, 0.8],
        "status": ["ok", "ok"],
        "observed_latency_ms": [50.0, 60.0]
    })
    
    baselines = {"A-B": 10.0, "B-C": 15.0}
    capacities = {"A-B": 200.0, "B-C": 300.0}
    link_map = {"A-B": 0, "B-C": 1}
    
    prep = TrafficPreprocessor().fit(train_df, baselines, capacities, link_map)
    
    # The training median of valid load_units is 100.0
    assert prep._load_units_median == 100.0
    
    test_df = pd.DataFrame({
        "link_id": ["B-C", "A-B"],
        "tick": [1, 1],
        "load_units": [np.nan, np.nan],
        "load_ratio": [0.2, 0.4],
        "status": ["ok", "saturated"],
        "observed_latency_ms": [np.nan, 40.0]
    })
    
    out = prep.transform(test_df)
    
    # Row 0: B-C. load_units is missing, capacity is 300. 300 * 0.2 = 60.0
    assert out.iloc[0]["load_units"] == 60.0
    assert out.iloc[0]["load_units_missing"] == 1.0
    assert out.iloc[0]["observed_latency_missing"] == 1.0
    
    # Row 1: A-B. is_saturated should be 1 because status is 'saturated'
    assert out.iloc[1]["is_saturated"] == 1
    # load_units is missing, capacity is 200. 200 * 0.4 = 80.0
    assert out.iloc[1]["load_units"] == 80.0

def test_telemetry_preprocessor_imputation():
    train_df = pd.DataFrame({
        "link_id": ["A-B", "B-C", "C-D"],
        "tick": [0, 0, 0],
        "self_reported_latency_ms": [10.0, 20.0, np.nan],
        "measured_latency_ms": [12.0, 22.0, 30.0]
    })
    
    baselines = {"A-B": 5.0, "B-C": 5.0, "C-D": 5.0}
    link_map = {"A-B": 0, "B-C": 1, "C-D": 2}
    
    prep = TelemetryPreprocessor().fit(train_df, baselines, link_map)
    assert prep._self_reported_median == 15.0 # median of [10.0, 20.0]
    
    test_df = pd.DataFrame({
        "link_id": ["C-D"],
        "tick": [1],
        "self_reported_latency_ms": [np.nan],
        "measured_latency_ms": [20.0]
    })
    
    out = prep.transform(test_df)
    assert out.iloc[0]["self_reported_latency_ms"] == 15.0
    assert out.iloc[0]["self_reported_latency_missing"] == 1.0
