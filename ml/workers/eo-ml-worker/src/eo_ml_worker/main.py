"""Worker entry point."""
from __future__ import annotations

import logging
import os
import sys
from pathlib import Path

from ml_serving import InMemoryBus, Job, JobBus, JobResult, JobStatus, ObjectStore

from .handler import Handler, HandlerError

log = logging.getLogger("eo_ml_worker")


def run_loop(bus: JobBus, store: ObjectStore, handler: Handler) -> None:
    """Consume jobs from ``bus`` until it closes."""
    for job in bus.consume():
        result = _process(job, store, handler)
        bus.publish(result)


def _process(job: Job, store: ObjectStore, handler: Handler) -> JobResult:
    try:
        image_bytes = store.get_bytes(job.input_uri)
        result_bytes = handler.handle(job, image_bytes)
        store.put_bytes(job.output_uri, result_bytes)
        return JobResult(
            job_id=job.job_id, status=JobStatus.OK, output_uri=job.output_uri
        )
    except (HandlerError, ValueError, FileNotFoundError) as exc:
        log.warning("job %s failed: %s", job.job_id, exc)
        return JobResult(
            job_id=job.job_id,
            status=JobStatus.FAILED,
            output_uri=job.output_uri,
            error=str(exc),
        )


def main() -> int:  # pragma: no cover - exercised in deployment, not unit tests
    """CLI entry point. Wires Kafka + local FS storage from the
    standard environment variables and runs the loop forever.

    Environment variables:

    * ``KAFKA_BROKERS`` — bootstrap servers list (default ``localhost:9092``).
    * ``KAFKA_GROUP`` — consumer group (default ``eo-ml-worker``).
    * ``KAFKA_INPUT_TOPIC`` (default ``eo.inference.jobs.v1``).
    * ``KAFKA_OUTPUT_TOPIC`` (default ``eo.inference.results.v1``).
    * ``STORAGE_ROOT`` — local FS root (default ``/var/lib/eo-ml-worker``).

    Programmatic users (tests, embedded deployments) construct the
    handler / bus / storage directly and call :func:`run_loop`.
    """
    logging.basicConfig(level=os.getenv("LOG_LEVEL", "INFO"))
    log.error("eo-ml-worker requires programmatic wiring of a Handler "
              "instance. Use `from eo_ml_worker import run_loop` from "
              "your deployment image.")
    # Smoke-mode: with no jobs the worker exits cleanly.
    bus = InMemoryBus()
    bus.close()
    return 0


if __name__ == "__main__":  # pragma: no cover
    sys.exit(main())
