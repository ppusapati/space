"""Wire-format helpers for transporting numpy arrays in JSON.

Inference requests / responses are JSON-friendly when arrays are
serialised as base64-encoded ``.npy`` payloads. ``decode_array_field``
and ``encode_array_field`` provide that conversion.
"""
from __future__ import annotations

import base64
import io

import numpy as np


def encode_array_field(arr: np.ndarray) -> str:
    """Encode an array as a base64-encoded ``.npy`` byte string."""
    buf = io.BytesIO()
    np.save(buf, arr, allow_pickle=False)
    return base64.b64encode(buf.getvalue()).decode("ascii")


def decode_array_field(payload: str) -> np.ndarray:
    """Decode a base64-encoded ``.npy`` byte string back to an array."""
    raw = base64.b64decode(payload.encode("ascii"))
    buf = io.BytesIO(raw)
    return np.load(buf, allow_pickle=False)
