"""Inference handler for GI prediction jobs."""
from __future__ import annotations

import io
import json
from dataclasses import dataclass

import numpy as np
from gi_ml import (
    FloodRiskInputs,
    drought_spi,
    flood_risk,
    forecast_ndvi_trend,
    project_landcover,
)

from .bus import Job


@dataclass(frozen=True)
class HandlerError(Exception):
    """Raised when a GI prediction job cannot be processed."""

    reason: str

    def __str__(self) -> str:  # pragma: no cover
        return self.reason


@dataclass
class Handler:
    """Routes a GI :class:`Job` to the matching gi_ml entry point.

    The job's ``kind`` selects the predictor; the payload is loaded from
    a JSON document at ``input_uri`` whose schema depends on the kind.
    """

    def handle(self, job: Job, input_bytes: bytes) -> bytes:
        try:
            payload = json.loads(input_bytes)
        except json.JSONDecodeError as exc:
            raise HandlerError(f"input is not valid JSON: {exc}") from exc

        if job.kind == "flood_risk":
            arrays = payload.get("inputs", {})
            try:
                inputs = FloodRiskInputs(
                    elevation_m=np.asarray(arrays["elevation_m"], dtype=np.float32),
                    distance_to_river_m=np.asarray(arrays["distance_to_river_m"], dtype=np.float32),
                    rainfall_mm=np.asarray(arrays["rainfall_mm"], dtype=np.float32),
                    soil_saturation=np.asarray(arrays["soil_saturation"], dtype=np.float32),
                )
            except KeyError as exc:
                raise HandlerError(f"missing flood_risk input `{exc.args[0]}`") from exc
            risk = flood_risk(inputs)
            return _ok_payload({"shape": list(risk.shape), "risk": risk.tolist()})

        if job.kind == "drought_spi":
            try:
                series = np.asarray(payload["monthly_precip_mm"], dtype=np.float32)
                window = int(payload["window_months"])
            except KeyError as exc:
                raise HandlerError(f"missing drought_spi input `{exc.args[0]}`") from exc
            spi = drought_spi(series, window)
            return _ok_payload({"window_months": window, "spi": spi.tolist()})

        if job.kind == "forecast_ndvi":
            try:
                series = np.asarray(payload["ndvi"], dtype=np.float32)
                horizon = int(payload["horizon"])
            except KeyError as exc:
                raise HandlerError(f"missing forecast_ndvi input `{exc.args[0]}`") from exc
            forecast = forecast_ndvi_trend(series, horizon)
            return _ok_payload(
                {
                    "mean": forecast.mean.tolist(),
                    "lower": forecast.lower.tolist(),
                    "upper": forecast.upper.tolist(),
                    "slope": forecast.slope,
                    "intercept": forecast.intercept,
                }
            )

        if job.kind == "urban_markov":
            try:
                lc = np.asarray(payload["landcover"], dtype=np.int32)
                t = np.asarray(payload["transition_matrix"], dtype=np.float64)
                steps = int(payload["steps"])
            except KeyError as exc:
                raise HandlerError(f"missing urban_markov input `{exc.args[0]}`") from exc
            projected = project_landcover(lc, t, steps)
            return _ok_payload(
                {"shape": list(projected.shape), "landcover": projected.tolist()}
            )

        raise HandlerError(f"unknown job kind `{job.kind}`")


def _ok_payload(body: dict) -> bytes:
    return json.dumps(body).encode("utf-8")


def _save_npy(arr: np.ndarray) -> bytes:  # pragma: no cover - convenience
    buf = io.BytesIO()
    np.save(buf, arr, allow_pickle=False)
    return buf.getvalue()
