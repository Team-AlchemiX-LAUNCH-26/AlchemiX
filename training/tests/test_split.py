import pandas as pd
import pytest
from split import tick_split, assert_no_overlap

def test_tick_split_basic():
    df = pd.DataFrame({
        "tick": [0, 0, 1, 1, 2, 2, 3, 3, 4, 4],
        "val": range(10)
    })
    
    # 5 ticks total (0..4). Range=5.
    # 70% train = 3.5 -> 3 ticks (0, 1, 2)
    # 15% val   = 0.75 -> 1 tick (3)
    # 15% test  = remainder -> 1 tick (4)
    split = tick_split(df, train_frac=0.6, val_frac=0.2)
    
    assert list(split.train["tick"].unique()) == [0, 1, 2]
    assert list(split.val["tick"].unique()) == [3]
    assert list(split.test["tick"].unique()) == [4]
    
    assert_no_overlap(split)

def test_tick_split_no_overlap():
    df = pd.DataFrame({
        "tick": list(range(100)) * 2,
        "val": range(200)
    })
    split = tick_split(df)
    assert_no_overlap(split)
    
    # Verify all ticks are assigned
    total_ticks = len(split.train["tick"].unique()) + len(split.val["tick"].unique()) + len(split.test["tick"].unique())
    assert total_ticks == 100

def test_tick_split_overlap_raises():
    df = pd.DataFrame({"tick": [0, 1, 2]})
    split = tick_split(df)
    
    # Force overlap
    split.val.loc[0, "tick"] = split.train["tick"].iloc[0]
    with pytest.raises(AssertionError):
        assert_no_overlap(split)
