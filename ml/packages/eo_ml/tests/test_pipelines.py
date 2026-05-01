"""Pipeline tests using fake predictors (no model weights required)."""
from __future__ import annotations

import numpy as np

from eo_ml import BandStats, Classifier, Detector, Segmenter


class _FakeDetectorPredictor:
    @property
    def input_names(self) -> tuple[str, ...]:
        return ("images",)

    @property
    def output_names(self) -> tuple[str, ...]:
        return ("dets",)

    def predict(self, inputs):
        # Return two boxes per image: one high-confidence, one duplicate
        # nearby to verify NMS suppression.
        det = np.array(
            [[5.0, 5.0, 25.0, 25.0, 0.9, 0],
             [6.0, 6.0, 26.0, 26.0, 0.85, 0]],
            dtype=np.float32,
        )
        return {"dets": det[np.newaxis]}


def test_detector_runs_nms_on_predictor_output():
    p = _FakeDetectorPredictor()
    d = Detector(
        predictor=p,
        band_stats=BandStats(mean=(0.0, 0.0, 0.0), std=(1.0, 1.0, 1.0)),
        iou_threshold=0.5,
        score_threshold=0.0,
    )
    img = np.zeros((3, 64, 64), dtype=np.float32)
    out = d.detect(img)
    assert len(out) == 1  # NMS dropped the near-duplicate
    assert abs(out[0].score - 0.9) < 1e-5


class _FakeSegPredictor:
    """Returns a tile output whose argmax is everywhere class 1."""

    @property
    def input_names(self) -> tuple[str, ...]:
        return ("images",)

    @property
    def output_names(self) -> tuple[str, ...]:
        return ("logits",)

    def predict(self, inputs):
        x = inputs["images"]
        b, _c, h, w = x.shape
        # 3 classes; class 1 wins everywhere.
        out = np.zeros((b, 3, h, w), dtype=np.float32)
        out[:, 1, :, :] = 10.0
        return {"logits": out}


def test_segmenter_argmax_uniform_class():
    p = _FakeSegPredictor()
    s = Segmenter(
        predictor=p,
        band_stats=BandStats(mean=(0.0, 0.0, 0.0), std=(1.0, 1.0, 1.0)),
        num_classes=3,
        tile_size=32,
        overlap=8,
    )
    img = np.zeros((3, 64, 64), dtype=np.float32)
    out = s.segment(img)
    assert out.shape == (64, 64)
    assert (out == 1).all()


class _FakeClsPredictor:
    @property
    def input_names(self) -> tuple[str, ...]:
        return ("images",)

    @property
    def output_names(self) -> tuple[str, ...]:
        return ("probs",)

    def predict(self, inputs):
        # 5 classes; class 3 wins.
        return {"probs": np.array([[0.05, 0.05, 0.10, 0.70, 0.10]], dtype=np.float32)}


def test_classifier_returns_top_class():
    p = _FakeClsPredictor()
    c = Classifier(
        predictor=p,
        band_stats=BandStats(mean=(0.0, 0.0, 0.0), std=(1.0, 1.0, 1.0)),
    )
    img = np.zeros((3, 32, 32), dtype=np.float32)
    top, vec = c.classify(img)
    assert top == 3
    assert vec.shape == (5,)
