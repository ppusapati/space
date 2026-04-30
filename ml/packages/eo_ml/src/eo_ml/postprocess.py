"""Post-processing helpers."""
from __future__ import annotations

from dataclasses import dataclass

import numpy as np


@dataclass(frozen=True)
class Detection:
    """Single detected object."""

    x1: float
    """Left edge in pixels."""

    y1: float
    """Top edge in pixels."""

    x2: float
    """Right edge in pixels."""

    y2: float
    """Bottom edge in pixels."""

    score: float
    """Confidence score in [0, 1]."""

    class_id: int
    """Predicted class index."""


def _iou(a: tuple[float, float, float, float], b: tuple[float, float, float, float]) -> float:
    ax1, ay1, ax2, ay2 = a
    bx1, by1, bx2, by2 = b
    ix1 = max(ax1, bx1)
    iy1 = max(ay1, by1)
    ix2 = min(ax2, bx2)
    iy2 = min(ay2, by2)
    iw = max(0.0, ix2 - ix1)
    ih = max(0.0, iy2 - iy1)
    inter = iw * ih
    a_area = max(0.0, ax2 - ax1) * max(0.0, ay2 - ay1)
    b_area = max(0.0, bx2 - bx1) * max(0.0, by2 - by1)
    union = a_area + b_area - inter
    return inter / union if union > 0 else 0.0


def nms(
    boxes: np.ndarray,
    scores: np.ndarray,
    class_ids: np.ndarray,
    iou_threshold: float,
    score_threshold: float = 0.0,
) -> list[Detection]:
    """Class-aware Non-Maximum Suppression.

    Args:
        boxes: ``(N, 4)`` ``float`` array of ``[x1, y1, x2, y2]`` coordinates.
        scores: ``(N,)`` confidence scores.
        class_ids: ``(N,)`` integer class predictions.
        iou_threshold: IoU above which the lower-scored detection is suppressed.
        score_threshold: drop detections with score below this value.

    Returns the surviving detections in descending score order.
    """
    if boxes.ndim != 2 or boxes.shape[1] != 4:
        raise ValueError(f"boxes must be (N, 4); got {boxes.shape}")
    if scores.shape != (boxes.shape[0],):
        raise ValueError("scores shape mismatch")
    if class_ids.shape != (boxes.shape[0],):
        raise ValueError("class_ids shape mismatch")
    if not 0.0 <= iou_threshold <= 1.0:
        raise ValueError("iou_threshold must be in [0, 1]")
    keep_mask = scores >= score_threshold
    if not keep_mask.any():
        return []
    idxs = np.argsort(-scores[keep_mask])
    sub_boxes = boxes[keep_mask][idxs]
    sub_scores = scores[keep_mask][idxs]
    sub_classes = class_ids[keep_mask][idxs]

    surviving: list[Detection] = []
    suppressed = np.zeros(len(sub_boxes), dtype=bool)
    for i in range(len(sub_boxes)):
        if suppressed[i]:
            continue
        bi = tuple(float(v) for v in sub_boxes[i])
        surviving.append(
            Detection(
                x1=bi[0],
                y1=bi[1],
                x2=bi[2],
                y2=bi[3],
                score=float(sub_scores[i]),
                class_id=int(sub_classes[i]),
            )
        )
        for j in range(i + 1, len(sub_boxes)):
            if suppressed[j] or sub_classes[j] != sub_classes[i]:
                continue
            bj = tuple(float(v) for v in sub_boxes[j])
            if _iou(bi, bj) > iou_threshold:
                suppressed[j] = True
    return surviving


def argmax_segmentation(probs: np.ndarray) -> np.ndarray:
    """Convert a ``(C, H, W)`` probability cube to an ``(H, W)`` class-id map."""
    if probs.ndim != 3:
        raise ValueError(f"probs must be (C, H, W); got {probs.shape}")
    return np.argmax(probs, axis=0).astype(np.int32)
