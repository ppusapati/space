"""Unit tests for the registry, batching, and codec helpers."""
from __future__ import annotations

import numpy as np
import pytest

from ml_serving import (
    ModelNotFound,
    ModelRegistry,
    Predictor,
    collate_batch,
    decode_array_field,
    encode_array_field,
    split_batch,
)
from ml_serving.batching import collate_batch as cb


class _IdentityPredictor:
    """Trivial predictor whose output equals its input."""

    @property
    def input_names(self) -> tuple[str, ...]:
        return ("x",)

    @property
    def output_names(self) -> tuple[str, ...]:
        return ("y",)

    def predict(self, inputs):
        return {"y": np.asarray(inputs["x"])}


def test_predictor_protocol_satisfied():
    p = _IdentityPredictor()
    assert isinstance(p, Predictor)


def test_registry_register_and_get():
    r = ModelRegistry()
    p = _IdentityPredictor()
    meta = r.register("identity", p, version="1.0.0", description="echo")
    assert "identity" in r
    assert r.get("identity") is p
    assert r.meta("identity") == meta
    assert meta.input_names == ("x",)
    assert meta.output_names == ("y",)
    assert len(r.list()) == 1


def test_registry_unregister_and_missing():
    r = ModelRegistry()
    r.register("identity", _IdentityPredictor())
    r.unregister("identity")
    with pytest.raises(ModelNotFound):
        r.get("identity")


def test_registry_rejects_empty_name():
    r = ModelRegistry()
    with pytest.raises(ValueError):
        r.register("", _IdentityPredictor())


def test_codec_round_trip():
    arr = np.arange(24, dtype=np.float32).reshape(2, 3, 4)
    payload = encode_array_field(arr)
    decoded = decode_array_field(payload)
    np.testing.assert_array_equal(decoded, arr)
    assert decoded.dtype == arr.dtype


def test_collate_and_split_batch_round_trip():
    items = [
        {"x": np.array([1.0, 2.0]), "y": np.array([10.0])},
        {"x": np.array([3.0, 4.0]), "y": np.array([20.0])},
        {"x": np.array([5.0, 6.0]), "y": np.array([30.0])},
    ]
    batched = collate_batch(items)
    assert batched["x"].shape == (3, 2)
    assert batched["y"].shape == (3, 1)
    rebuilt = split_batch(batched, len(items))
    assert len(rebuilt) == len(items)
    for original, recovered in zip(items, rebuilt):
        np.testing.assert_array_equal(original["x"], recovered["x"])
        np.testing.assert_array_equal(original["y"], recovered["y"])


def test_collate_empty_batch_rejected():
    with pytest.raises(ValueError):
        cb([])


def test_collate_shape_mismatch_rejected():
    items = [
        {"x": np.array([1.0, 2.0])},
        {"x": np.array([1.0, 2.0, 3.0])},
    ]
    with pytest.raises(ValueError):
        cb(items)
