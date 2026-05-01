"""End-to-end tests for the EO ML worker using in-memory bus + filesystem
storage + fake predictors."""
from __future__ import annotations

import io
import json
import threading

import numpy as np
import pytest
from eo_ml import BandStats, Classifier, Detector, Segmenter
from ml_serving import ModelRegistry

from eo_ml_worker import (
    Handler,
    InMemoryBus,
    Job,
    JobStatus,
    LocalStorage,
)
from eo_ml_worker.main import _process, run_loop


class _FakeDetectorPredictor:
    @property
    def input_names(self):
        return ("images",)

    @property
    def output_names(self):
        return ("dets",)

    def predict(self, inputs):
        det = np.array([[0.0, 0.0, 10.0, 10.0, 0.95, 0]], dtype=np.float32)
        return {"dets": det[np.newaxis]}


def _build_handler():
    reg = ModelRegistry()
    p = _FakeDetectorPredictor()
    reg.register("yolov8-port", p)
    detector = Detector(predictor=p, band_stats=BandStats(mean=(0, 0, 0), std=(1, 1, 1)))
    return Handler(
        registry=reg,
        detectors={"yolov8-port": detector},
        segmenters={},
        classifiers={},
    )


def _save_npy(arr: np.ndarray) -> bytes:
    buf = io.BytesIO()
    np.save(buf, arr, allow_pickle=False)
    return buf.getvalue()


def test_process_detect_writes_output(tmp_path):
    store = LocalStorage(tmp_path)
    img = np.zeros((3, 64, 64), dtype=np.float32)
    store.put_bytes("local://input.npy", _save_npy(img))
    job = Job(
        job_id="j1",
        tenant_id="t1",
        model_name="yolov8-port",
        kind="detect",
        input_uri="local://input.npy",
        output_uri="local://result.json",
    )
    handler = _build_handler()
    result = _process(job, store, handler)
    assert result.status == JobStatus.OK
    out_bytes = store.get_bytes("local://result.json")
    payload = json.loads(out_bytes)
    assert "detections" in payload
    assert len(payload["detections"]) == 1
    assert payload["detections"][0]["score"] == pytest.approx(0.95, rel=1e-5)


def test_process_unknown_kind_marks_failed(tmp_path):
    store = LocalStorage(tmp_path)
    img = np.zeros((3, 16, 16), dtype=np.float32)
    store.put_bytes("local://x.npy", _save_npy(img))
    handler = _build_handler()
    job = Job(
        job_id="j2",
        tenant_id="t1",
        model_name="yolov8-port",
        kind="bogus",
        input_uri="local://x.npy",
        output_uri="local://out.json",
    )
    result = _process(job, store, handler)
    assert result.status == JobStatus.FAILED
    assert "unknown job kind" in (result.error or "")


def test_process_unknown_model_marks_failed(tmp_path):
    store = LocalStorage(tmp_path)
    img = np.zeros((3, 16, 16), dtype=np.float32)
    store.put_bytes("local://x.npy", _save_npy(img))
    handler = _build_handler()
    job = Job(
        job_id="j3",
        tenant_id="t1",
        model_name="nonexistent",
        kind="detect",
        input_uri="local://x.npy",
        output_uri="local://out.json",
    )
    result = _process(job, store, handler)
    assert result.status == JobStatus.FAILED


def test_run_loop_consumes_until_close(tmp_path):
    store = LocalStorage(tmp_path)
    img = np.zeros((3, 32, 32), dtype=np.float32)
    store.put_bytes("local://x.npy", _save_npy(img))
    handler = _build_handler()
    bus = InMemoryBus()
    bus.push(Job(
        job_id="a",
        tenant_id="t",
        model_name="yolov8-port",
        kind="detect",
        input_uri="local://x.npy",
        output_uri="local://out-a.json",
    ))
    bus.push(Job(
        job_id="b",
        tenant_id="t",
        model_name="yolov8-port",
        kind="detect",
        input_uri="local://x.npy",
        output_uri="local://out-b.json",
    ))

    thread = threading.Thread(target=run_loop, args=(bus, store, handler))
    thread.start()
    # Tear the bus down — consume() returns when None is enqueued.
    bus.close()
    thread.join(timeout=5)
    assert not thread.is_alive()
    statuses = {r.job_id: r.status for r in bus.results}
    assert statuses == {"a": JobStatus.OK, "b": JobStatus.OK}


def test_local_storage_rejects_path_traversal(tmp_path):
    store = LocalStorage(tmp_path)
    with pytest.raises(ValueError):
        store.get_bytes("local://../../etc/passwd")
