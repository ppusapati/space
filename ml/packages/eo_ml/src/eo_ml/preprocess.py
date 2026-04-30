"""Pre-processing helpers for satellite imagery."""
from __future__ import annotations

from dataclasses import dataclass
from typing import Iterator

import numpy as np


@dataclass(frozen=True)
class BandStats:
    """Per-band normalisation statistics."""

    mean: tuple[float, ...]
    std: tuple[float, ...]


def to_channels_first(image: np.ndarray) -> np.ndarray:
    """Convert ``(H, W, C)`` → ``(C, H, W)``."""
    if image.ndim != 3:
        raise ValueError(f"expected 3-D image, got shape {image.shape}")
    return np.transpose(image, (2, 0, 1))


def to_channels_last(image: np.ndarray) -> np.ndarray:
    """Convert ``(C, H, W)`` → ``(H, W, C)``."""
    if image.ndim != 3:
        raise ValueError(f"expected 3-D image, got shape {image.shape}")
    return np.transpose(image, (1, 2, 0))


def normalise(image: np.ndarray, stats: BandStats) -> np.ndarray:
    """Apply per-band ``(x − μ) / σ`` to a channels-first array.

    Args:
        image: ``(C, H, W)`` float array.
        stats: per-band mean / std (length must equal C).
    """
    if image.ndim != 3:
        raise ValueError(f"expected (C, H, W) array, got {image.shape}")
    c = image.shape[0]
    if len(stats.mean) != c or len(stats.std) != c:
        raise ValueError(
            f"stats length ({len(stats.mean)}, {len(stats.std)}) does not match band count {c}"
        )
    mean = np.asarray(stats.mean, dtype=image.dtype).reshape(c, 1, 1)
    std = np.asarray(stats.std, dtype=image.dtype).reshape(c, 1, 1)
    if np.any(std == 0):
        raise ValueError("std contains zero — invalid normalisation")
    return (image - mean) / std


@dataclass(frozen=True)
class Tile:
    """One tile produced by :func:`tile_image`."""

    row: int
    """Top-left pixel row in the source image."""

    col: int
    """Top-left pixel column in the source image."""

    rows: int
    """Tile height (``= tile_size`` for interior tiles, smaller at the bottom edge)."""

    cols: int
    """Tile width (``= tile_size`` for interior tiles, smaller at the right edge)."""

    pixels: np.ndarray
    """``(C, tile_size, tile_size)`` array; for edge tiles the unfilled pixels are zero-padded."""


def tile_image(
    image: np.ndarray, tile_size: int, overlap: int
) -> Iterator[Tile]:
    """Iterate over fixed-size, optionally overlapping tiles.

    The image is assumed channels-first ``(C, H, W)``. Edge tiles smaller
    than ``tile_size`` are zero-padded; the ``rows`` and ``cols`` fields
    record how much of the tile contains real data so callers can crop on
    the way back.
    """
    if image.ndim != 3:
        raise ValueError(f"expected (C, H, W) array, got {image.shape}")
    if tile_size <= 0:
        raise ValueError("tile_size must be positive")
    if not 0 <= overlap < tile_size:
        raise ValueError(f"overlap must be in [0, {tile_size}), got {overlap}")

    c, h, w = image.shape
    stride = tile_size - overlap
    row = 0
    while row < h:
        col = 0
        while col < w:
            tile = np.zeros((c, tile_size, tile_size), dtype=image.dtype)
            rows = min(tile_size, h - row)
            cols = min(tile_size, w - col)
            tile[:, :rows, :cols] = image[:, row : row + rows, col : col + cols]
            yield Tile(row=row, col=col, rows=rows, cols=cols, pixels=tile)
            if col + tile_size >= w:
                break
            col += stride
        if row + tile_size >= h:
            break
        row += stride


def untile_segmentation(
    tiles: list[tuple[Tile, np.ndarray]],
    height: int,
    width: int,
    num_classes: int,
) -> np.ndarray:
    """Reassemble per-tile class-probability outputs into a full image.

    Each tile output should have shape ``(num_classes, tile_h, tile_w)``.
    For overlapping tiles the per-class probability is averaged across
    contributors before the final argmax.
    """
    if num_classes <= 0:
        raise ValueError("num_classes must be positive")
    accum = np.zeros((num_classes, height, width), dtype=np.float32)
    counts = np.zeros((height, width), dtype=np.float32)
    for tile, probs in tiles:
        if probs.shape[0] != num_classes:
            raise ValueError(
                f"tile output has {probs.shape[0]} classes, expected {num_classes}"
            )
        h_used = tile.rows
        w_used = tile.cols
        accum[:, tile.row : tile.row + h_used, tile.col : tile.col + w_used] += probs[
            :, :h_used, :w_used
        ]
        counts[tile.row : tile.row + h_used, tile.col : tile.col + w_used] += 1.0
    counts = np.maximum(counts, 1e-6)
    return accum / counts
