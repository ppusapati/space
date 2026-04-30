"""Inference handler — couples a job to the right `eo_ml` pipeline."""
from __future__ import annotations

import io
import json
from dataclasses import asdict, dataclass

import numpy as np
from eo_ml import Classifier, Detector, Segmenter
from ml_serving import ModelRegistry, ModelNotFound

from .bus import Job


@dataclass(frozen=True)
class HandlerError(Exception):
    """Raised when an inference job cannot be processed."""

    reason: str

    def __str__(self) -> str:  # pragma: no cover - trivial repr
        return self.reason


@dataclass
class Handler:
    """Routes a :class:`Job` to the matching pipeline.

    The registry must contain models registered under the same names
    used in the inbound jobs.
    """

    registry: ModelRegistry
    detectors: dict[str, Detector]
    segmenters: dict[str, Segmenter]
    classifiers: dict[str, Classifier]

    def handle(self, job: Job, image_npy: bytes) -> bytes:
        """Process one job and return the result as JSON bytes."""
        image = np.load(io.BytesIO(image_npy), allow_pickle=False)
        if job.kind == "detect":
            det = self.detectors.get(job.model_name)
            if det is None:
                raise HandlerError(f"no detector registered for `{job.model_name}`")
            results = [asdict(d) for d in det.detect(image)]
            payload = {"detections": results}
        elif job.kind == "segment":
            seg = self.segmenters.get(job.model_name)
            if seg is None:
                raise HandlerError(f"no segmenter registered for `{job.model_name}`")
            mask = seg.segment(image)
            payload = {"shape": list(mask.shape), "classes": mask.astype(int).tolist()}
        elif job.kind == "classify":
            cls = self.classifiers.get(job.model_name)
            if cls is None:
                raise HandlerError(f"no classifier registered for `{job.model_name}`")
            top, probs = cls.classify(image)
            payload = {"top_class": int(top), "probs": probs.tolist()}
        else:
            raise HandlerError(f"unknown job kind `{job.kind}`")
        # Surface the registry's existence implicitly — fail loudly when
        # the model name does not appear in either the registry or the
        # specialised dict, so operators get an early signal.
        try:
            self.registry.meta(job.model_name)
        except ModelNotFound as exc:
            raise HandlerError(f"model `{job.model_name}` not in registry") from exc
        return json.dumps(payload).encode("utf-8")
