"""High-level inference pipelines."""
from __future__ import annotations

from dataclasses import dataclass
from typing import Iterable

import numpy as np
from ml_serving import Predictor

from .postprocess import Detection, argmax_segmentation, nms
from .preprocess import (
    BandStats,
    Tile,
    normalise,
    tile_image,
    to_channels_first,
    untile_segmentation,
)


@dataclass
class Detector:
    """ONNX-based detector wrapper.

    The underlying model is expected to consume ``(B, C, H, W)`` float
    tensors and produce, per sample, an ``(N_i, 6)`` ``[x1, y1, x2, y2,
    score, class]`` output.
    """

    predictor: Predictor
    band_stats: BandStats
    iou_threshold: float = 0.5
    score_threshold: float = 0.25

    def detect(self, image: np.ndarray) -> list[Detection]:
        """Run detection on one ``(H, W, C)`` or ``(C, H, W)`` image."""
        chw = image if image.shape[0] <= 16 and image.ndim == 3 else to_channels_first(image)
        chw = normalise(chw.astype(np.float32), self.band_stats)
        batched = chw[np.newaxis]
        out = self.predictor.predict({self.predictor.input_names[0]: batched})
        raw = next(iter(out.values()))
        if raw.ndim == 3 and raw.shape[0] == 1:
            raw = raw[0]
        if raw.size == 0:
            return []
        if raw.shape[1] < 6:
            raise ValueError(
                f"detector output has {raw.shape[1]} columns; expected ≥ 6 (x1,y1,x2,y2,score,class)"
            )
        boxes = raw[:, :4].astype(np.float32)
        scores = raw[:, 4].astype(np.float32)
        classes = raw[:, 5].astype(np.int32)
        return nms(
            boxes,
            scores,
            classes,
            iou_threshold=self.iou_threshold,
            score_threshold=self.score_threshold,
        )


@dataclass
class Segmenter:
    """ONNX-based semantic segmentation wrapper.

    The underlying model consumes ``(B, C, H, W)`` and produces
    ``(B, num_classes, H, W)`` per-pixel logits / softmax probabilities.
    """

    predictor: Predictor
    band_stats: BandStats
    num_classes: int
    tile_size: int = 512
    overlap: int = 64

    def segment(self, image: np.ndarray) -> np.ndarray:
        """Tile-and-stitch semantic segmentation. Returns an ``(H, W)``
        class-id map."""
        if image.ndim != 3:
            raise ValueError(f"expected 3-D image, got {image.shape}")
        chw = image if image.shape[0] <= 16 else to_channels_first(image)
        c, h, w = chw.shape
        chw = normalise(chw.astype(np.float32), self.band_stats)
        results: list[tuple[Tile, np.ndarray]] = []
        for tile in tile_image(chw, self.tile_size, self.overlap):
            inputs = {self.predictor.input_names[0]: tile.pixels[np.newaxis]}
            out = self.predictor.predict(inputs)
            probs = next(iter(out.values()))
            if probs.ndim == 4 and probs.shape[0] == 1:
                probs = probs[0]
            if probs.shape != (self.num_classes, self.tile_size, self.tile_size):
                raise ValueError(
                    f"segmenter output {probs.shape} does not match "
                    f"({self.num_classes}, {self.tile_size}, {self.tile_size})"
                )
            results.append((tile, probs.astype(np.float32)))
        merged_probs = untile_segmentation(results, h, w, self.num_classes)
        return argmax_segmentation(merged_probs)


@dataclass
class Classifier:
    """ONNX-based classification wrapper."""

    predictor: Predictor
    band_stats: BandStats

    def classify(self, image: np.ndarray) -> tuple[int, np.ndarray]:
        """Return ``(top_class, full_probability_vector)``."""
        chw = image if image.shape[0] <= 16 and image.ndim == 3 else to_channels_first(image)
        chw = normalise(chw.astype(np.float32), self.band_stats)
        out = self.predictor.predict({self.predictor.input_names[0]: chw[np.newaxis]})
        probs = next(iter(out.values()))
        if probs.ndim == 2 and probs.shape[0] == 1:
            probs = probs[0]
        if probs.ndim != 1:
            raise ValueError(f"classifier output shape {probs.shape} is not 1-D after squeeze")
        return int(np.argmax(probs)), probs.astype(np.float32)


def batched_detect(
    detector: Detector, images: Iterable[np.ndarray]
) -> list[list[Detection]]:
    """Run a Detector over many images sequentially."""
    return [detector.detect(img) for img in images]
