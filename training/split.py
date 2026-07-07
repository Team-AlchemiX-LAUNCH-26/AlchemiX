"""
Shared tick-based train/validation/test splitter.

Splitting philosophy:
- Never use random row splitting. 
- All rows from the same tick stay in the same partition.
- Uses actual min/max ticks, not assumed tick values.
- Approximately 70/15/15 split by tick range.
"""
from __future__ import annotations

import math
from dataclasses import dataclass

import pandas as pd


RANDOM_STATE = 42  # Kept for downstream reproducibility documentation.

@dataclass(frozen=True)
class SplitResult:
    train: pd.DataFrame
    val: pd.DataFrame
    test: pd.DataFrame
    train_tick_range: tuple[int, int]
    val_tick_range: tuple[int, int]
    test_tick_range: tuple[int, int]

    def describe(self) -> str:
        lines = [
            f"  Train: ticks {self.train_tick_range[0]}–{self.train_tick_range[1]}, "
            f"{len(self.train)} rows",
            f"  Val:   ticks {self.val_tick_range[0]}–{self.val_tick_range[1]}, "
            f"{len(self.val)} rows",
            f"  Test:  ticks {self.test_tick_range[0]}–{self.test_tick_range[1]}, "
            f"{len(self.test)} rows",
        ]
        return "\n".join(lines)


def tick_split(
    df: pd.DataFrame,
    tick_col: str = "tick",
    train_frac: float = 0.70,
    val_frac: float = 0.15,
) -> SplitResult:
    """
    Split df by tick, keeping all rows from the same tick together.

    Parameters
    ----------
    df : DataFrame with a tick column containing integer tick indices.
    tick_col : name of the tick column (default "tick").
    train_frac : fraction of ticks assigned to train (default 0.70).
    val_frac : fraction of ticks assigned to val (default 0.15).
    test gets the remainder.

    Returns
    -------
    SplitResult with train/val/test DataFrames and their tick ranges.
    """
    if tick_col not in df.columns:
        raise ValueError(f"Column '{tick_col}' not found in DataFrame")

    min_tick = int(df[tick_col].min())
    max_tick = int(df[tick_col].max())
    tick_range = max_tick - min_tick  # inclusive count is tick_range+1

    # Compute boundaries based on the actual range.
    train_end = min_tick + math.floor(tick_range * train_frac)
    val_end   = min_tick + math.floor(tick_range * (train_frac + val_frac))

    # Clamp so val_end < max_tick to guarantee a non-empty test set.
    if val_end >= max_tick:
        val_end = max_tick - 1
    if train_end >= val_end:
        train_end = val_end - 1

    train = df[df[tick_col] <= train_end].copy()
    val   = df[(df[tick_col] > train_end) & (df[tick_col] <= val_end)].copy()
    test  = df[df[tick_col] > val_end].copy()

    return SplitResult(
        train=train,
        val=val,
        test=test,
        train_tick_range=(min_tick, train_end),
        val_tick_range=(train_end + 1, val_end),
        test_tick_range=(val_end + 1, max_tick),
    )


def assert_no_overlap(split: SplitResult, tick_col: str = "tick") -> None:
    """Raise AssertionError if any tick appears in more than one partition."""
    train_ticks = set(split.train[tick_col])
    val_ticks   = set(split.val[tick_col])
    test_ticks  = set(split.test[tick_col])

    tv = train_ticks & val_ticks
    tt = train_ticks & test_ticks
    vt = val_ticks   & test_ticks

    assert not tv, f"Ticks in both train and val: {sorted(tv)[:5]}"
    assert not tt, f"Ticks in both train and test: {sorted(tt)[:5]}"
    assert not vt, f"Ticks in both val and test: {sorted(vt)[:5]}"

    assert max(train_ticks) < min(val_ticks),  \
        f"Train max tick {max(train_ticks)} >= val min tick {min(val_ticks)}"
    assert max(val_ticks)   < min(test_ticks), \
        f"Val max tick {max(val_ticks)} >= test min tick {min(test_ticks)}"
