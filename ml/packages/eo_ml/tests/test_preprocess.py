"""Pre-processing tests."""
from __future__ import annotations

import numpy as np
import pytest

from eo_ml import (
    BandStats,
    normalise,
    tile_image,
    to_channels_first,
    to_channels_last,
    untile_segmentation,
)


def test_channels_first_round_trip():
    img = np.arange(24).reshape(2, 3, 4)
    assert to_channels_last(to_channels_first(img.transpose(1, 2, 0))).shape == (3, 4, 2)


def test_normalise_zero_mean_unit_std():
    img = np.array([[[1.0, 2.0], [3.0, 4.0]]], dtype=np.float32)  # (1, 2, 2)
    out = normalise(img, BandStats(mean=(2.5,), std=(1.118,)))
    np.testing.assert_allclose(out.mean(), 0.0, atol=1e-3)


def test_normalise_rejects_zero_std():
    img = np.zeros((1, 2, 2), dtype=np.float32)
    with pytest.raises(ValueError):
        normalise(img, BandStats(mean=(0.0,), std=(0.0,)))


def test_tile_image_covers_full_image():
    c, h, w = 3, 100, 80
    img = np.random.default_rng(0).standard_normal((c, h, w)).astype(np.float32)
    tiles = list(tile_image(img, tile_size=32, overlap=8))
    seen = np.zeros((h, w), dtype=bool)
    for t in tiles:
        seen[t.row : t.row + t.rows, t.col : t.col + t.cols] = True
    assert seen.all()


def test_untile_segmentation_round_trip_no_overlap():
    c, h, w = 3, 64, 64
    rng = np.random.default_rng(1)
    truth_logits = rng.standard_normal((4, h, w)).astype(np.float32)
    chw = rng.standard_normal((c, h, w)).astype(np.float32)
    tiles_with_probs = []
    for t in tile_image(chw, tile_size=32, overlap=0):
        sub = truth_logits[:, t.row : t.row + t.rows, t.col : t.col + t.cols]
        # Pad sub to (4, 32, 32) to mimic real-tile output.
        padded = np.zeros((4, 32, 32), dtype=np.float32)
        padded[:, : t.rows, : t.cols] = sub
        tiles_with_probs.append((t, padded))
    merged = untile_segmentation(tiles_with_probs, h, w, num_classes=4)
    np.testing.assert_allclose(merged, truth_logits, atol=1e-6)


def test_tile_image_rejects_invalid_overlap():
    img = np.zeros((1, 16, 16), dtype=np.float32)
    with pytest.raises(ValueError):
        list(tile_image(img, tile_size=8, overlap=8))
