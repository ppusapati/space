"""Thread-safe model registry."""
from __future__ import annotations

import threading
from dataclasses import dataclass, field
from typing import Iterable

from .predictor import Predictor


class ModelNotFound(KeyError):
    """Raised when a model name is not registered."""


@dataclass(frozen=True)
class ModelMeta:
    """Lightweight metadata attached to a registered model."""

    name: str
    """Stable, URL-safe identifier."""

    version: str
    """Semantic version (e.g. ``1.2.3``) or git commit hash."""

    input_names: tuple[str, ...]
    """Ordered tuple of input names expected by the underlying predictor."""

    output_names: tuple[str, ...]
    """Ordered tuple of output names produced by the underlying predictor."""

    description: str = ""
    """Human-readable description."""

    labels: dict[str, str] = field(default_factory=dict)
    """Free-form labels (e.g. mission, classification)."""


@dataclass
class _Entry:
    meta: ModelMeta
    predictor: Predictor


class ModelRegistry:
    """Mapping of ``name -> Predictor`` with safe concurrent access."""

    def __init__(self) -> None:
        self._lock = threading.RLock()
        self._entries: dict[str, _Entry] = {}

    def register(
        self,
        name: str,
        predictor: Predictor,
        *,
        version: str = "0.0.0",
        description: str = "",
        labels: dict[str, str] | None = None,
    ) -> ModelMeta:
        if not name:
            raise ValueError("model name must not be empty")
        meta = ModelMeta(
            name=name,
            version=version,
            input_names=tuple(predictor.input_names),
            output_names=tuple(predictor.output_names),
            description=description,
            labels=dict(labels or {}),
        )
        with self._lock:
            self._entries[name] = _Entry(meta=meta, predictor=predictor)
        return meta

    def unregister(self, name: str) -> None:
        with self._lock:
            self._entries.pop(name, None)

    def get(self, name: str) -> Predictor:
        with self._lock:
            entry = self._entries.get(name)
        if entry is None:
            raise ModelNotFound(name)
        return entry.predictor

    def meta(self, name: str) -> ModelMeta:
        with self._lock:
            entry = self._entries.get(name)
        if entry is None:
            raise ModelNotFound(name)
        return entry.meta

    def list(self) -> list[ModelMeta]:
        with self._lock:
            return [e.meta for e in self._entries.values()]

    def names(self) -> Iterable[str]:
        with self._lock:
            return list(self._entries.keys())

    def __len__(self) -> int:
        with self._lock:
            return len(self._entries)

    def __contains__(self, name: object) -> bool:
        with self._lock:
            return isinstance(name, str) and name in self._entries
