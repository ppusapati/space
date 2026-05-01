"""Inference plumbing shared by ``eo_ml`` and ``gi_ml``.

The package deliberately decouples the prediction *interface* (the
:class:`Predictor` protocol) from any specific runtime so that the same
``ModelRegistry``, batching helpers, and FastAPI scaffolding work with
ONNX Runtime, PyTorch, TensorFlow, or a unit-test stub.

The :mod:`ml_serving.jobbus` and :mod:`ml_serving.storage` submodules
host the worker-side abstractions (Kafka / in-memory job bus,
filesystem-backed object store) shared by every ML worker daemon.
"""
from .predictor import Predictor, PredictionError
from .registry import ModelMeta, ModelRegistry, ModelNotFound
from .onnx_predictor import OnnxPredictor
from .batching import collate_batch, split_batch
from .codec import decode_array_field, encode_array_field
from .jobbus import InMemoryBus, Job, JobBus, JobResult, JobStatus, KafkaBus
from .storage import LocalStorage, ObjectStore

__all__ = [
    "Predictor",
    "PredictionError",
    "ModelMeta",
    "ModelRegistry",
    "ModelNotFound",
    "OnnxPredictor",
    "collate_batch",
    "split_batch",
    "decode_array_field",
    "encode_array_field",
    "InMemoryBus",
    "Job",
    "JobBus",
    "JobResult",
    "JobStatus",
    "KafkaBus",
    "LocalStorage",
    "ObjectStore",
]
