"""Unit tests for the FastAPI server scaffolding."""
from __future__ import annotations

import numpy as np
from fastapi.testclient import TestClient

from ml_serving import ModelRegistry, decode_array_field, encode_array_field
from ml_serving.server import build_app


class _DoublePredictor:
    @property
    def input_names(self) -> tuple[str, ...]:
        return ("x",)

    @property
    def output_names(self) -> tuple[str, ...]:
        return ("y",)

    def predict(self, inputs):
        return {"y": 2.0 * np.asarray(inputs["x"])}


def _client():
    registry = ModelRegistry()
    registry.register("double", _DoublePredictor(), version="1.0.0", description="2x")
    return TestClient(build_app(registry))


def test_healthz():
    resp = _client().get("/healthz")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok"}


def test_list_models():
    resp = _client().get("/v1/models")
    assert resp.status_code == 200
    body = resp.json()
    assert len(body) == 1
    assert body[0]["name"] == "double"
    assert body[0]["input_names"] == ["x"]


def test_model_metadata():
    resp = _client().get("/v1/models/double")
    assert resp.status_code == 200
    assert resp.json()["version"] == "1.0.0"


def test_model_metadata_404():
    resp = _client().get("/v1/models/missing")
    assert resp.status_code == 404


def test_inference_round_trip():
    client = _client()
    payload = {"x": encode_array_field(np.array([1.0, 2.0, 3.0], dtype=np.float32))}
    resp = client.post("/v1/models/double/infer", json={"inputs": payload})
    assert resp.status_code == 200, resp.text
    out = decode_array_field(resp.json()["outputs"]["y"])
    np.testing.assert_allclose(out, [2.0, 4.0, 6.0])
