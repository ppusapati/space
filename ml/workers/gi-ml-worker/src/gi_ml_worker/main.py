"""GI ML worker entry point."""
from __future__ import annotations

import logging
import os
import sys

from ml_serving import InMemoryBus, Job, JobBus, JobResult, JobStatus, ObjectStore

from .handler import Handler, HandlerError

log = logging.getLogger("gi_ml_worker")


def run_loop(bus: JobBus, store: ObjectStore, handler: Handler) -> None:
    """Consume jobs from ``bus`` until it closes."""
    for job in bus.consume():
        result = _process(job, store, handler)
        bus.publish(result)


def _process(job: Job, store: ObjectStore, handler: Handler) -> JobResult:
    try:
        input_bytes = store.get_bytes(job.input_uri)
        result_bytes = handler.handle(job, input_bytes)
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
    """CLI entry point. See ``eo-ml-worker`` for the equivalent
    deployment-time wiring contract."""
    logging.basicConfig(level=os.getenv("LOG_LEVEL", "INFO"))
    log.error("gi-ml-worker requires programmatic wiring of a Handler "
              "instance via `from gi_ml_worker import run_loop`.")
    bus = InMemoryBus()
    bus.close()
    return 0


if __name__ == "__main__":  # pragma: no cover
    sys.exit(main())
