"""Micro-batching helpers."""
from __future__ import annotations

from typing import Mapping, Sequence

import numpy as np


def collate_batch(items: Sequence[Mapping[str, np.ndarray]]) -> dict[str, np.ndarray]:
    """Stack a sequence of single-sample feed dicts into a batched feed.

    All items must share the same key set and per-key shape; the resulting
    arrays are stacked along a new leading batch axis.
    """
    if not items:
        raise ValueError("cannot collate an empty batch")
    keys = set(items[0].keys())
    for i, item in enumerate(items[1:], start=1):
        if set(item.keys()) != keys:
            raise ValueError(
                f"item {i} has key set {set(item.keys())}, expected {keys}"
            )
    batched: dict[str, np.ndarray] = {}
    for k in items[0].keys():
        arrays = [np.asarray(item[k]) for item in items]
        shape0 = arrays[0].shape
        for j, a in enumerate(arrays[1:], start=1):
            if a.shape != shape0:
                raise ValueError(
                    f"key `{k}`: item {j} has shape {a.shape}, expected {shape0}"
                )
        batched[k] = np.stack(arrays, axis=0)
    return batched


def split_batch(
    outputs: Mapping[str, np.ndarray],
    batch_size: int,
) -> list[dict[str, np.ndarray]]:
    """Split a batched output dict back into per-sample dicts."""
    if batch_size <= 0:
        raise ValueError("batch_size must be positive")
    for k, arr in outputs.items():
        if arr.shape[0] != batch_size:
            raise ValueError(
                f"output `{k}` has leading dim {arr.shape[0]}, expected {batch_size}"
            )
    out: list[dict[str, np.ndarray]] = []
    for i in range(batch_size):
        out.append({k: np.asarray(v[i]) for k, v in outputs.items()})
    return out
