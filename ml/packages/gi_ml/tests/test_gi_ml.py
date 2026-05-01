"""Unit tests for gi_ml."""
from __future__ import annotations

import numpy as np
import pytest

from gi_ml import (
    CropForecast,
    FloodRiskInputs,
    drought_spi,
    flood_risk,
    forecast_ndvi_trend,
    project_landcover,
)


# ---- Flood risk -------------------------------------------------------

def test_flood_risk_low_for_high_dry_terrain():
    inputs = FloodRiskInputs(
        elevation_m=np.array([[200.0]]),
        distance_to_river_m=np.array([[5_000.0]]),
        rainfall_mm=np.array([[5.0]]),
        soil_saturation=np.array([[0.1]]),
    )
    r = flood_risk(inputs)
    assert r.shape == (1, 1)
    assert r[0, 0] < 0.2


def test_flood_risk_high_for_riverside_saturated_terrain():
    inputs = FloodRiskInputs(
        elevation_m=np.array([[5.0]]),
        distance_to_river_m=np.array([[20.0]]),
        rainfall_mm=np.array([[150.0]]),
        soil_saturation=np.array([[0.95]]),
    )
    r = flood_risk(inputs)
    assert r[0, 0] > 0.7


def test_flood_risk_rejects_shape_mismatch():
    with pytest.raises(ValueError):
        flood_risk(
            FloodRiskInputs(
                elevation_m=np.zeros((2, 2)),
                distance_to_river_m=np.zeros((2, 3)),
                rainfall_mm=np.zeros((2, 2)),
                soil_saturation=np.zeros((2, 2)),
            )
        )


# ---- Drought SPI ------------------------------------------------------

def test_drought_spi_returns_zero_for_constant_series():
    arr = np.full(24, 50.0, dtype=np.float32)
    out = drought_spi(arr, window_months=3)
    assert out.shape == (22,)
    np.testing.assert_allclose(out, 0.0, atol=1e-6)


def test_drought_spi_distinguishes_wet_and_dry():
    rng = np.random.default_rng(0)
    base = rng.normal(loc=80, scale=10, size=240)
    base[:60] -= 40  # very dry first 60 months
    base[180:] += 40  # very wet last 60 months
    out = drought_spi(base.astype(np.float32), window_months=12)
    assert out[:30].mean() < 0
    assert out[-30:].mean() > 0


# ---- NDVI forecast ----------------------------------------------------

def test_forecast_ndvi_trend_recovers_known_slope():
    t = np.arange(24)
    ndvi = (0.01 * t + 0.50 + np.random.default_rng(0).normal(0, 0.005, 24)).astype(np.float32)
    f = forecast_ndvi_trend(ndvi, horizon=6, confidence=0.95)
    assert isinstance(f, CropForecast)
    assert f.mean.shape == (6,)
    assert abs(f.slope - 0.01) < 1e-3
    # Confidence band should bracket the central forecast.
    assert (f.lower <= f.mean).all()
    assert (f.mean <= f.upper).all()


def test_forecast_ndvi_rejects_short_series():
    with pytest.raises(ValueError):
        forecast_ndvi_trend(np.array([0.5]), horizon=1)


# ---- Urban-growth Markov ---------------------------------------------

def test_project_landcover_keeps_state_under_identity_matrix():
    lc = np.array([[0, 1], [2, 0]], dtype=np.int32)
    t = np.eye(3)
    out = project_landcover(lc, t, steps=10)
    np.testing.assert_array_equal(out, lc)


def test_project_landcover_evolves_with_strong_transition():
    lc = np.zeros((4, 4), dtype=np.int32)
    # Class 0 evolves into class 1 with probability 0.95 per step.
    t = np.array([[0.05, 0.95], [0.0, 1.0]])
    out = project_landcover(lc, t, steps=5)
    # After 5 steps probability of state 1 dominates → all pixels become 1.
    assert (out == 1).all()


def test_project_landcover_rejects_non_stochastic_matrix():
    lc = np.zeros((2, 2), dtype=np.int32)
    bad_t = np.array([[0.5, 0.4], [0.0, 1.0]])  # row 0 sums to 0.9
    with pytest.raises(ValueError):
        project_landcover(lc, bad_t, steps=1)
