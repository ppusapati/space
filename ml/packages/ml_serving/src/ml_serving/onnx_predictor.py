"""ONNX Runtime-backed :class:`Predictor`."""
from __future__ import annotations

import threading
from pathlib import Path
from typing import Mapping, Sequence

import numpy as np
import onnxruntime as ort

from .predictor import PredictionError, Predictor


class OnnxPredictor(Predictor):
    """Thread-safe predictor wrapping an ``onnxruntime.InferenceSession``.

    ``providers`` defaults to ``["CPUExecutionProvider"]``; pass
    ``["CUDAExecutionProvider", "CPUExecutionProvider"]`` for GPU
    acceleration when the corresponding wheel is installed.
    """

    def __init__(
        self,
        model_path: str | Path,
        providers: Sequence[str] | None = None,
        intra_op_num_threads: int | None = None,
    ) -> None:
        path = Path(model_path)
        if not path.is_file():
            raise FileNotFoundError(f"ONNX model not found: {path}")
        options = ort.SessionOptions()
        if intra_op_num_threads is not None:
            options.intra_op_num_threads = int(intra_op_num_threads)
        self._session = ort.InferenceSession(
            str(path),
            sess_options=options,
            providers=list(providers) if providers else ["CPUExecutionProvider"],
        )
        self._input_names = tuple(i.name for i in self._session.get_inputs())
        self._output_names = tuple(o.name for o in self._session.get_outputs())
        self._lock = threading.Lock()

    @property
    def input_names(self) -> tuple[str, ...]:
        return self._input_names

    @property
    def output_names(self) -> tuple[str, ...]:
        return self._output_names

    def predict(self, inputs: Mapping[str, np.ndarray]) -> dict[str, np.ndarray]:
        for name in self._input_names:
            if name not in inputs:
                raise PredictionError(f"missing input `{name}`")
        feed = {name: np.asarray(inputs[name]) for name in self._input_names}
        try:
            with self._lock:
                outputs = self._session.run(list(self._output_names), feed)
        except (ort.OrtRuntimeException, RuntimeError) as exc:
            raise PredictionError(str(exc)) from exc
        return dict(zip(self._output_names, outputs))
