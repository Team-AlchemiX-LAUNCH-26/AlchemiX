from __future__ import annotations

from collections import defaultdict
from statistics import mean, pstdev

import pandas as pd

from features import CONGESTION_FEATURE_COLS, TARGETING_FEATURE_COLS, TRUST_FEATURE_COLS
from service.schemas import LinkObservation


class FeatureHistory:
    def __init__(self) -> None:
        self._by_link: dict[str, list[LinkObservation]] = defaultdict(list)

    def add(self, obs: LinkObservation) -> None:
        self._by_link[obs.link_id].append(obs)

    def values(self, link_id: str, attr: str, window: int | None = None) -> list[float]:
        rows = self._by_link.get(link_id, [])
        if window is not None:
            rows = rows[-window:]
        values: list[float] = []
        for row in rows:
            val = getattr(row, attr)
            if val is not None:
                values.append(float(val))
        return values


class LiveFeatureBuilder:
    def __init__(self, link_id_map: dict[str, int]) -> None:
        self.history = FeatureHistory()
        self.link_id_map = link_id_map

    def add_observation(self, obs: LinkObservation) -> None:
        self.history.add(obs)

    def congestion(self, obs: LinkObservation) -> pd.DataFrame:
        previous = self.history.values(obs.link_id, "load_ratio", 1)
        prev_load = previous[-1] if previous else obs.load_ratio
        rolling_values = self.history.values(obs.link_id, "load_ratio", 5)
        rolling_mean = mean(rolling_values) if rolling_values else obs.load_ratio
        rolling_std = pstdev(rolling_values) if len(rolling_values) > 1 else 0.0
        older_values = self.history.values(obs.link_id, "load_ratio", 6)[:-1]
        older_mean = mean(older_values) if older_values else rolling_mean

        row = {
            "load_ratio": obs.load_ratio,
            "load_units": obs.current_load,
            "capacity_units": obs.capacity_units,
            "link_id_encoded": float(self.link_id_map.get(obs.link_id, -1)),
            "previous_load_ratio": prev_load,
            "load_ratio_change": obs.load_ratio - prev_load,
            "rolling_mean_load_ratio": rolling_mean,
            "rolling_std_load_ratio": rolling_std,
            "rate_of_load_increase": rolling_mean - older_mean,
            "distance_to_saturation": max(0.0, 0.90 - obs.load_ratio),
            "load_units_missing": 0.0,
        }
        return pd.DataFrame([row], columns=CONGESTION_FEATURE_COLS)

    def trust(self, obs: LinkObservation) -> pd.DataFrame:
        physical = obs.physical_latency_ms or obs.physical_baseline_latency_ms or 1.0
        self_latency = obs.self_reported_latency_ms
        missing = 1.0 if self_latency is None else 0.0
        self_value = physical if self_latency is None else self_latency

        previous = self.history.values(obs.link_id, "self_reported_latency_ms", 1)
        prev_self = previous[-1] if previous else self_value
        rolling_values = self.history.values(obs.link_id, "self_reported_latency_ms", 5)
        rolling_self = mean(rolling_values) if rolling_values else self_value

        row = {
            "self_reported_latency_ms": self_value,
            "self_reported_latency_missing": missing,
            "physical_latency_ms": physical,
            "self_to_physical_ratio": self_value / physical if physical else 0.0,
            "deviation_from_baseline": self_value - physical,
            "prev_self_reported": prev_self,
            "self_reported_change": self_value - prev_self,
            "rolling_self_reported": rolling_self,
            "historical_bias": 0.0,
            "below_physical_min": 1.0 if self_value < physical * 0.95 else 0.0,
            "link_id_encoded": float(self.link_id_map.get(obs.link_id, -1)),
        }
        return pd.DataFrame([row], columns=TRUST_FEATURE_COLS)

    def targeting(self, obs: LinkObservation) -> pd.DataFrame:
        previous = self.history.values(obs.link_id, "traffic_share", 1)
        prev_share = previous[-1] if previous else obs.traffic_share
        rolling_values = self.history.values(obs.link_id, "traffic_share", 5)
        rolling_share = mean(rolling_values) if rolling_values else obs.traffic_share

        n_links = max(1, len(self.link_id_map))
        median_share = 1.0 / n_links
        high_usage_values = [
            1.0 if value > median_share else 0.0
            for value in self.history.values(obs.link_id, "traffic_share", 10)
        ]

        row = {
            "traffic_share": obs.traffic_share,
            "traffic_share_missing": 0.0,
            "traffic_share_rank": 1.0,
            "prev_traffic_share": prev_share,
            "traffic_share_change": obs.traffic_share - prev_share,
            "rolling_traffic_share": rolling_share,
            "historical_jam_rate": 0.0,
            "rolling_jam_count": 0.0,
            "consecutive_high_usage": sum(high_usage_values),
            "link_id_encoded": float(self.link_id_map.get(obs.link_id, -1)),
        }
        return pd.DataFrame([row], columns=TARGETING_FEATURE_COLS)
