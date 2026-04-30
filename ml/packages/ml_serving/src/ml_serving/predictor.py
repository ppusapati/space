"""Predictor protocol — the only contract every model implementation
must satisfy."""
from __future__ import annotations

from typing import Mapping, Protocol, runtime_checkable

import numpy as np


class PredictionError(RuntimeError):
    """Raised when a predictor fails to produce a result."""


@runtime_checkable
class Predictor(Protocol):
    """A callable that consumes a dict of named ``ndarray`` inputs and
    returns a dict of named ``ndarray`` outputs.

    Implementations must be **thread-safe** for concurrent ``predict``
    calls; ``ml_serving`` uses a thread pool inside the FastAPI server.
    """

    @property
    def input_names(self) -> tuple[str, ...]:
        """Ordered tuple of expected input names."""

    @property
    def output_names(self) -> tuple[str, ...]:
        """Ordered tuple of output names produced by the model."""

    def predict(self, inputs: Mapping[str, np.ndarray]) -> dict[str, np.ndarray]:
        """Run inference and return the output map.

        Raises:
            PredictionError: if the runtime fails.
        """
