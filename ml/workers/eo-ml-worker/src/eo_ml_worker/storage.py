"""Object-store abstraction with local-filesystem and S3 backends."""
from __future__ import annotations

from pathlib import Path
from typing import Protocol


class ObjectStore(Protocol):
    """Pluggable byte-blob store keyed by URI."""

    def get_bytes(self, uri: str) -> bytes:
        """Fetch a blob as raw bytes."""

    def put_bytes(self, uri: str, data: bytes) -> None:
        """Write a blob."""


class LocalStorage:
    """Filesystem-backed store. URIs are interpreted relative to ``root``
    after stripping the ``local://`` scheme."""

    def __init__(self, root: Path | str) -> None:
        self.root = Path(root)
        self.root.mkdir(parents=True, exist_ok=True)

    def _path(self, uri: str) -> Path:
        if uri.startswith("local://"):
            uri = uri[len("local://") :]
        path = self.root / uri.lstrip("/")
        if not path.resolve().is_relative_to(self.root.resolve()):
            raise ValueError(f"path {path} escapes storage root")
        return path

    def get_bytes(self, uri: str) -> bytes:
        return self._path(uri).read_bytes()

    def put_bytes(self, uri: str, data: bytes) -> None:
        path = self._path(uri)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)
