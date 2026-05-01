"""Flood and drought risk indices."""
from __future__ import annotations

from dataclasses import dataclass

import numpy as np


@dataclass(frozen=True)
class FloodRiskInputs:
    """Per-pixel features for the flood-risk index."""

    elevation_m: np.ndarray
    """``(H, W)`` elevation above mean sea level (m)."""

    distance_to_river_m: np.ndarray
    """``(H, W)`` distance from the nearest river / drainage line (m)."""

    rainfall_mm: np.ndarray
    """``(H, W)`` recent rainfall accumulation (mm)."""

    soil_saturation: np.ndarray
    """``(H, W)`` 0..1 fraction of soil moisture saturation."""


def flood_risk(inputs: FloodRiskInputs) -> np.ndarray:
    """Composite flood-risk index in `[0, 1]`.

    The score combines four normalised terms:

    * ``f_elev = exp(−elevation / 50)`` — low-lying pixels score higher.
    * ``f_dist = exp(−distance / 500)`` — nearer to a river scores higher.
    * ``f_rain = clip(rainfall / 100, 0, 1)`` — saturating at 100 mm.
    * ``f_soil = clip(soil_saturation, 0, 1)``.

    The four terms are averaged with weights `(0.25, 0.25, 0.30, 0.20)`
    that empirically reflect their relative importance in regional
    flood-risk literature.
    """
    elev = np.asarray(inputs.elevation_m, dtype=np.float32)
    dist = np.asarray(inputs.distance_to_river_m, dtype=np.float32)
    rain = np.asarray(inputs.rainfall_mm, dtype=np.float32)
    soil = np.asarray(inputs.soil_saturation, dtype=np.float32)
    if not (elev.shape == dist.shape == rain.shape == soil.shape):
        raise ValueError("all input rasters must share the same shape")
    f_elev = np.exp(-np.maximum(elev, 0.0) / 50.0)
    f_dist = np.exp(-np.maximum(dist, 0.0) / 500.0)
    f_rain = np.clip(rain / 100.0, 0.0, 1.0)
    f_soil = np.clip(soil, 0.0, 1.0)
    return 0.25 * f_elev + 0.25 * f_dist + 0.30 * f_rain + 0.20 * f_soil


def drought_spi(monthly_precip_mm: np.ndarray, window_months: int) -> np.ndarray:
    """Standardised Precipitation Index over a rolling window.

    Args:
        monthly_precip_mm: ``(T,)`` monthly precipitation (mm).
        window_months: rolling-window length (e.g. 3 for SPI-3, 12 for SPI-12).

    Returns an array of length ``len(monthly_precip_mm) − window_months + 1``
    of Z-scores (signed standard deviations from the long-run mean).
    """
    arr = np.asarray(monthly_precip_mm, dtype=np.float32)
    if arr.ndim != 1:
        raise ValueError("monthly_precip_mm must be 1-D")
    if window_months <= 0 or window_months > len(arr):
        raise ValueError(
            f"window_months must be in [1, {len(arr)}], got {window_months}"
        )
    cumulative = np.convolve(arr, np.ones(window_months, dtype=np.float32), mode="valid")
    mean = float(cumulative.mean())
    std = float(cumulative.std())
    if std == 0.0:
        return np.zeros_like(cumulative)
    return (cumulative - mean) / std
