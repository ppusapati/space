"""End-to-end tests for the GI ML worker."""
from __future__ import annotations

import json
import threading

import pytest

from gi_ml_worker import (
    Handler,
    InMemoryBus,
    Job,
    JobStatus,
    LocalStorage,
)
from gi_ml_worker.main import _process, run_loop


def _job(kind: str, input_uri: str, output_uri: str, jid: str = "j1") -> Job:
    return Job(
        job_id=jid,
        tenant_id="t1",
        model_name=kind,
        kind=kind,
        input_uri=input_uri,
        output_uri=output_uri,
    )


def test_flood_risk_pipeline(tmp_path):
    store = LocalStorage(tmp_path)
    payload = {
        "inputs": {
            "elevation_m": [[5.0]],
            "distance_to_river_m": [[20.0]],
            "rainfall_mm": [[150.0]],
            "soil_saturation": [[0.95]],
        }
    }
    store.put_bytes("local://input.json", json.dumps(payload).encode("utf-8"))
    handler = Handler()
    result = _process(_job("flood_risk", "local://input.json", "local://out.json"), store, handler)
    assert result.status == JobStatus.OK
    out = json.loads(store.get_bytes("local://out.json"))
    assert out["shape"] == [1, 1]
    assert out["risk"][0][0] > 0.7


def test_drought_spi_pipeline(tmp_path):
    store = LocalStorage(tmp_path)
    payload = {"monthly_precip_mm": [50.0] * 24, "window_months": 3}
    store.put_bytes("local://in.json", json.dumps(payload).encode("utf-8"))
    handler = Handler()
    result = _process(_job("drought_spi", "local://in.json", "local://out.json"), store, handler)
    assert result.status == JobStatus.OK
    out = json.loads(store.get_bytes("local://out.json"))
    assert len(out["spi"]) == 22  # 24 - 3 + 1


def test_forecast_ndvi_pipeline(tmp_path):
    store = LocalStorage(tmp_path)
    payload = {"ndvi": [0.5 + i * 0.01 for i in range(24)], "horizon": 6}
    store.put_bytes("local://in.json", json.dumps(payload).encode("utf-8"))
    handler = Handler()
    result = _process(_job("forecast_ndvi", "local://in.json", "local://out.json"), store, handler)
    assert result.status == JobStatus.OK
    out = json.loads(store.get_bytes("local://out.json"))
    assert len(out["mean"]) == 6
    assert abs(out["slope"] - 0.01) < 1e-3


def test_urban_markov_pipeline(tmp_path):
    store = LocalStorage(tmp_path)
    payload = {
        "landcover": [[0, 1], [1, 0]],
        "transition_matrix": [[0.05, 0.95], [0.0, 1.0]],
        "steps": 5,
    }
    store.put_bytes("local://in.json", json.dumps(payload).encode("utf-8"))
    handler = Handler()
    result = _process(_job("urban_markov", "local://in.json", "local://out.json"), store, handler)
    assert result.status == JobStatus.OK
    out = json.loads(store.get_bytes("local://out.json"))
    assert out["landcover"] == [[1, 1], [1, 1]]


def test_unknown_kind_marks_failed(tmp_path):
    store = LocalStorage(tmp_path)
    store.put_bytes("local://in.json", json.dumps({}).encode("utf-8"))
    handler = Handler()
    result = _process(_job("unknown", "local://in.json", "local://out.json"), store, handler)
    assert result.status == JobStatus.FAILED


def test_run_loop_processes_until_close(tmp_path):
    store = LocalStorage(tmp_path)
    payload = {"monthly_precip_mm": [50.0] * 12, "window_months": 3}
    store.put_bytes("local://in.json", json.dumps(payload).encode("utf-8"))
    handler = Handler()
    bus = InMemoryBus()
    bus.push(_job("drought_spi", "local://in.json", "local://a.json", jid="a"))
    bus.push(_job("drought_spi", "local://in.json", "local://b.json", jid="b"))
    thread = threading.Thread(target=run_loop, args=(bus, store, handler))
    thread.start()
    bus.close()
    thread.join(timeout=5)
    assert not thread.is_alive()
    assert {r.job_id: r.status for r in bus.results} == {"a": JobStatus.OK, "b": JobStatus.OK}
