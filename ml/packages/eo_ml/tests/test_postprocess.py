"""Post-processing tests."""
from __future__ import annotations

import numpy as np
import pytest

from eo_ml import argmax_segmentation, nms


def test_nms_keeps_high_scoring_box_and_suppresses_overlap():
    boxes = np.array(
        [
            [0.0, 0.0, 10.0, 10.0],
            [1.0, 1.0, 11.0, 11.0],  # overlaps box 0
            [50.0, 50.0, 60.0, 60.0],
        ]
    )
    scores = np.array([0.95, 0.80, 0.90])
    classes = np.array([0, 0, 0])
    keep = nms(boxes, scores, classes, iou_threshold=0.5)
    assert len(keep) == 2
    assert keep[0].score == 0.95
    assert keep[1].score == 0.90


def test_nms_class_aware_keeps_overlapping_different_classes():
    boxes = np.array(
        [
            [0.0, 0.0, 10.0, 10.0],
            [1.0, 1.0, 11.0, 11.0],
        ]
    )
    scores = np.array([0.9, 0.85])
    classes = np.array([0, 1])
    keep = nms(boxes, scores, classes, iou_threshold=0.5)
    assert len(keep) == 2  # different classes are not suppressed against each other


def test_nms_score_threshold_filters():
    boxes = np.array([[0.0, 0.0, 10.0, 10.0], [50.0, 50.0, 60.0, 60.0]])
    scores = np.array([0.9, 0.1])
    classes = np.array([0, 0])
    keep = nms(boxes, scores, classes, iou_threshold=0.5, score_threshold=0.5)
    assert len(keep) == 1
    assert keep[0].score == 0.9


def test_nms_invalid_arguments():
    with pytest.raises(ValueError):
        nms(
            boxes=np.zeros((3, 3)),
            scores=np.zeros(3),
            class_ids=np.zeros(3),
            iou_threshold=0.5,
        )


def test_argmax_segmentation_picks_highest_class():
    probs = np.array(
        [
            [[0.1, 0.7], [0.3, 0.2]],
            [[0.5, 0.2], [0.4, 0.5]],
            [[0.4, 0.1], [0.3, 0.3]],
        ],
        dtype=np.float32,
    )
    out = argmax_segmentation(probs)
    np.testing.assert_array_equal(out, [[1, 0], [1, 1]])
