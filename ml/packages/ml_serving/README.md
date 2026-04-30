# ml_serving

Inference plumbing shared by `eo_ml` and `gi_ml`. Provides a model registry, a
`Predictor` protocol, micro-batching, an ONNX Runtime predictor, and a
FastAPI server scaffolding.

## Quick start

```python
from ml_serving import ModelRegistry, OnnxPredictor

registry = ModelRegistry()
registry.register("yolov8-port", OnnxPredictor("/models/yolov8-port.onnx"))

scores = registry.get("yolov8-port").predict({"images": batch_array})
```

## Server

`ml_serving.server.build_app(registry)` returns a FastAPI `app` with two
endpoints:

* `GET /healthz` — liveness probe.
* `POST /v1/models/{name}/infer` — JSON body containing input arrays
  (base64-encoded `npy` bytes); returns each output array similarly encoded.
