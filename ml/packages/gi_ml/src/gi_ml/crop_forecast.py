"""Crop NDVI trend forecasting via least-squares linear regression."""
from __future__ import annotations

from dataclasses import dataclass

import numpy as np


@dataclass(frozen=True)
class CropForecast:
    """Forecast result.

    `mean[k]` is the central NDVI prediction at horizon `k+1` (one
    forecast step), and `lower[k]` / `upper[k]` form a 95 % confidence
    band derived from the residual standard error.
    """

    mean: np.ndarray
    lower: np.ndarray
    upper: np.ndarray
    slope: float
    intercept: float


def forecast_ndvi_trend(
    ndvi: np.ndarray, horizon: int, confidence: float = 0.95
) -> CropForecast:
    """Fit ``y = a · t + b`` to a 1-D NDVI time series and project
    ``horizon`` steps ahead with a Gaussian confidence band.

    Args:
        ndvi: ``(T,)`` array of historical NDVI values.
        horizon: number of future steps to forecast.
        confidence: confidence level (default 95 %).

    Raises:
        ValueError: for non-1-D input, fewer than 2 samples, or invalid
        confidence.
    """
    arr = np.asarray(ndvi, dtype=np.float32)
    if arr.ndim != 1:
        raise ValueError("ndvi must be 1-D")
    if arr.size < 2:
        raise ValueError("need at least 2 samples to fit a trend")
    if not 0.0 < confidence < 1.0:
        raise ValueError("confidence must be in (0, 1)")
    if horizon <= 0:
        raise ValueError("horizon must be positive")
    t = np.arange(arr.size, dtype=np.float64)
    # Least-squares fit y = a t + b
    a, b = np.polyfit(t, arr, deg=1)
    residuals = arr - (a * t + b)
    sigma = float(np.std(residuals, ddof=1)) if arr.size > 2 else 0.0
    # Approximate Z-score for the requested confidence (two-sided).
    # Use the inverse-normal via numpy's built-in `np.sqrt(2) * erfinv(p)`
    # to avoid SciPy.
    p = 0.5 + confidence / 2.0
    z = float(np.sqrt(2.0) * _erfinv(2.0 * p - 1.0))
    future_t = np.arange(arr.size, arr.size + horizon, dtype=np.float64)
    mean = (a * future_t + b).astype(np.float32)
    half_width = np.full_like(mean, z * sigma) if sigma > 0 else np.zeros_like(mean)
    return CropForecast(
        mean=mean,
        lower=mean - half_width,
        upper=mean + half_width,
        slope=float(a),
        intercept=float(b),
    )


def _erfinv(x: float) -> float:
    """Inverse error function, Winitzki approximation (max abs error 1.3·10⁻⁴)."""
    if not -1.0 < x < 1.0:
        raise ValueError("erfinv argument must be in (-1, 1)")
    a = 0.147
    ln1mx2 = float(np.log(1.0 - x * x))
    term = 2.0 / (np.pi * a) + ln1mx2 / 2.0
    return float(np.sign(x) * np.sqrt(np.sqrt(term * term - ln1mx2 / a) - term))
