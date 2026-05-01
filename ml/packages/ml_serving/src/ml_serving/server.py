"""FastAPI server scaffolding for the model registry."""
from __future__ import annotations

from typing import Any, Mapping

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from .codec import decode_array_field, encode_array_field
from .predictor import PredictionError
from .registry import ModelMeta, ModelNotFound, ModelRegistry


class InferRequest(BaseModel):
    """Inference request body. Each input array is base64-encoded ``.npy``."""

    inputs: dict[str, str] = Field(
        ..., description="Map of input-name → base64(npy) payload."
    )


class InferResponse(BaseModel):
    """Inference response with base64-encoded ``.npy`` payloads."""

    outputs: dict[str, str]


class ModelInfo(BaseModel):
    """Model metadata."""

    name: str
    version: str
    input_names: list[str]
    output_names: list[str]
    description: str
    labels: dict[str, str]


def _meta_to_model(meta: ModelMeta) -> ModelInfo:
    return ModelInfo(
        name=meta.name,
        version=meta.version,
        input_names=list(meta.input_names),
        output_names=list(meta.output_names),
        description=meta.description,
        labels=dict(meta.labels),
    )


def build_app(registry: ModelRegistry) -> FastAPI:
    """Build a FastAPI app exposing the registered models.

    Endpoints:

    * ``GET /healthz`` — liveness.
    * ``GET /v1/models`` — list registered models.
    * ``GET /v1/models/{name}`` — single-model metadata.
    * ``POST /v1/models/{name}/infer`` — run inference.
    """
    app = FastAPI(title="P9E ML Serving", version="0.1.0")

    @app.get("/healthz")
    def healthz() -> dict[str, str]:
        return {"status": "ok"}

    @app.get("/v1/models", response_model=list[ModelInfo])
    def list_models() -> list[ModelInfo]:
        return [_meta_to_model(m) for m in registry.list()]

    @app.get("/v1/models/{name}", response_model=ModelInfo)
    def get_model(name: str) -> ModelInfo:
        try:
            return _meta_to_model(registry.meta(name))
        except ModelNotFound as exc:
            raise HTTPException(status_code=404, detail=f"model `{name}` not found") from exc

    @app.post("/v1/models/{name}/infer", response_model=InferResponse)
    def infer(name: str, body: InferRequest) -> InferResponse:
        try:
            predictor = registry.get(name)
        except ModelNotFound as exc:
            raise HTTPException(status_code=404, detail=f"model `{name}` not found") from exc
        feed: Mapping[str, Any] = {k: decode_array_field(v) for k, v in body.inputs.items()}
        try:
            outputs = predictor.predict(feed)
        except PredictionError as exc:
            raise HTTPException(status_code=500, detail=str(exc)) from exc
        return InferResponse(outputs={k: encode_array_field(v) for k, v in outputs.items()})

    return app
