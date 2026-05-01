"""Job bus abstraction with in-memory and Kafka backends."""
from __future__ import annotations

import json
import queue
from collections.abc import Iterable, Iterator
from dataclasses import asdict, dataclass
from enum import Enum
from typing import Protocol


class JobStatus(str, Enum):
    """Outcome status of a job."""

    OK = "ok"
    FAILED = "failed"


@dataclass(frozen=True)
class Job:
    """One inference request."""

    job_id: str
    tenant_id: str
    model_name: str
    kind: str  # "detect" | "segment" | "classify"
    input_uri: str
    output_uri: str


@dataclass(frozen=True)
class JobResult:
    """One inference outcome."""

    job_id: str
    status: JobStatus
    output_uri: str
    error: str | None = None

    def to_json(self) -> str:
        d = asdict(self)
        d["status"] = self.status.value
        return json.dumps(d)


class JobBus(Protocol):
    """Bidirectional job queue."""

    def consume(self) -> Iterator[Job]:
        """Iterate forever, yielding jobs as they arrive."""

    def publish(self, result: JobResult) -> None:
        """Publish a completed result."""

    def close(self) -> None:
        """Release resources."""


class InMemoryBus:
    """Thread-safe in-memory bus for unit tests / integration tests."""

    def __init__(self) -> None:
        self._jobs: queue.Queue[Job | None] = queue.Queue()
        self.results: list[JobResult] = []
        self._closed = False

    def push(self, job: Job) -> None:
        self._jobs.put(job)

    def consume(self) -> Iterator[Job]:
        # Drain everything queued up to and including the close-sentinel
        # (``None``) so jobs published before close() are all delivered
        # regardless of thread scheduling.
        while True:
            item = self._jobs.get()
            if item is None:
                return
            yield item

    def publish(self, result: JobResult) -> None:
        self.results.append(result)

    def close(self) -> None:
        self._closed = True
        self._jobs.put(None)


class KafkaBus:
    """Kafka-backed bus using ``confluent_kafka``.

    Constructed lazily so that the rest of the worker remains importable
    without the Kafka native library installed (useful for unit tests).
    """

    def __init__(
        self,
        brokers: str,
        consumer_group: str,
        input_topic: str,
        output_topic: str,
    ) -> None:
        try:
            from confluent_kafka import Consumer, Producer
        except ImportError as exc:  # pragma: no cover - exercised only on prod
            raise RuntimeError(
                "confluent_kafka is not installed; install it to use KafkaBus"
            ) from exc
        self._consumer = Consumer(
            {
                "bootstrap.servers": brokers,
                "group.id": consumer_group,
                "enable.auto.commit": False,
                "auto.offset.reset": "earliest",
            }
        )
        self._consumer.subscribe([input_topic])
        self._producer = Producer({"bootstrap.servers": brokers})
        self._output_topic = output_topic
        self._closed = False

    def consume(self) -> Iterator[Job]:
        from confluent_kafka import KafkaError  # pragma: no cover

        while not self._closed:
            msg = self._consumer.poll(timeout=1.0)
            if msg is None:
                continue
            if msg.error():  # pragma: no cover - error paths exercised in integration only
                if msg.error().code() == KafkaError._PARTITION_EOF:
                    continue
                continue
            payload = json.loads(msg.value())
            yield Job(
                job_id=payload["job_id"],
                tenant_id=payload["tenant_id"],
                model_name=payload["model_name"],
                kind=payload["kind"],
                input_uri=payload["input_uri"],
                output_uri=payload["output_uri"],
            )
            self._consumer.commit(msg, asynchronous=False)

    def publish(self, result: JobResult) -> None:
        self._producer.produce(self._output_topic, value=result.to_json().encode("utf-8"))
        self._producer.poll(0)

    def close(self) -> None:
        self._closed = True
        self._producer.flush(5)
        self._consumer.close()


def jobs_from_messages(payloads: Iterable[bytes]) -> list[Job]:
    """Helper to build a list of Jobs from raw JSON bytes — used by tests."""
    out: list[Job] = []
    for raw in payloads:
        d = json.loads(raw)
        out.append(
            Job(
                job_id=d["job_id"],
                tenant_id=d["tenant_id"],
                model_name=d["model_name"],
                kind=d["kind"],
                input_uri=d["input_uri"],
                output_uri=d["output_uri"],
            )
        )
    return out
