from __future__ import annotations

import datetime
import json
import logging
import queue
import random
import threading
import time
import traceback
import urllib.error
import urllib.request
from typing import List, Optional

from ._payload import EventPayload, SDK_VERSION
from ._redactor import Redactor

# Internal logger — silent by default; host app must opt in.
_log = logging.getLogger("custos.internal")
_log.addHandler(logging.NullHandler())

_SENTINEL = object()  # signals the worker thread to drain and exit
_MAX_QUEUE = 1_000    # drop events silently once the in-memory queue is full


class EventQueue:
    """Thread-safe queue that batches events and POSTs them to the Custos ingest endpoint."""

    def __init__(
        self,
        *,
        dsn: str,
        api_key: str,
        service: str,
        environment: str,
        batch_size: int,
        flush_interval: float,
        max_retries: int,
        timeout: float,
        redactor: Redactor,
    ) -> None:
        self._url = dsn.rstrip("/") + "/api/v1/ingest"
        self._api_key = api_key
        self._service = service
        self._environment = environment
        self._batch_size = batch_size
        self._flush_interval = flush_interval
        self._max_retries = max_retries
        self._timeout = timeout
        self._redactor = redactor
        self._q: queue.Queue = queue.Queue(maxsize=_MAX_QUEUE)
        self._thread: Optional[threading.Thread] = None

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    def start(self) -> None:
        self._thread = threading.Thread(
            target=self._run, name="custos-worker", daemon=True
        )
        self._thread.start()

    def stop(self, drain_timeout: float = 5.0) -> None:
        try:
            self._q.put_nowait(_SENTINEL)
        except queue.Full:
            pass
        if self._thread is not None:
            self._thread.join(timeout=drain_timeout)

    # ------------------------------------------------------------------
    # Enqueue
    # ------------------------------------------------------------------

    def enqueue(self, record: logging.LogRecord) -> None:
        try:
            payload = self._build_payload(record)
            self._q.put_nowait(payload)
        except queue.Full:
            _log.warning("event queue full — dropping event")

    # ------------------------------------------------------------------
    # Worker loop
    # ------------------------------------------------------------------

    def _run(self) -> None:
        batch: List[EventPayload] = []
        deadline = time.monotonic() + self._flush_interval

        while True:
            wait = max(0.0, deadline - time.monotonic())
            try:
                item = self._q.get(timeout=wait)
            except queue.Empty:
                # Flush whatever has accumulated during this window.
                if batch:
                    self._flush(batch)
                    batch = []
                deadline = time.monotonic() + self._flush_interval
                continue

            if item is _SENTINEL:
                if batch:
                    self._flush(batch)
                return

            batch.append(item)  # type: ignore[arg-type]
            if len(batch) >= self._batch_size:
                self._flush(batch)
                batch = []
                deadline = time.monotonic() + self._flush_interval

    # ------------------------------------------------------------------
    # Sending
    # ------------------------------------------------------------------

    def _flush(self, batch: List[EventPayload]) -> None:
        for payload in batch:
            self._send_with_retry(payload)

    def _send_with_retry(self, payload: EventPayload) -> None:
        body = json.dumps(payload.to_dict()).encode()

        for attempt in range(self._max_retries + 1):
            try:
                req = urllib.request.Request(
                    self._url,
                    data=body,
                    headers={
                        "Content-Type": "application/json",
                        "X-Custos-Key": self._api_key,
                    },
                    method="POST",
                )
                with urllib.request.urlopen(req, timeout=self._timeout) as resp:
                    if resp.status == 202:
                        return
                    # Unexpected 2xx — log and bail; server accepted it.
                    _log.warning("unexpected status %d from server", resp.status)
                    return

            except urllib.error.HTTPError as exc:
                if 400 <= exc.code < 500:
                    # Client error — not retryable (bad payload, invalid key, etc.)
                    _log.warning("HTTP %d — dropping event (not retryable)", exc.code)
                    return
                _log.warning(
                    "HTTP %d on attempt %d/%d", exc.code, attempt + 1, self._max_retries + 1
                )

            except Exception as exc:  # noqa: BLE001
                _log.warning(
                    "send error on attempt %d/%d: %s",
                    attempt + 1,
                    self._max_retries + 1,
                    exc,
                )

            if attempt < self._max_retries:
                # Exponential backoff with ±0–0.5 s jitter, capped at 60 s.
                backoff = min(60.0, (2**attempt) * 0.5) + random.uniform(0, 0.5)
                time.sleep(backoff)

        _log.warning("gave up sending event after %d attempts", self._max_retries + 1)

    # ------------------------------------------------------------------
    # Payload construction
    # ------------------------------------------------------------------

    def _build_payload(self, record: logging.LogRecord) -> EventPayload:
        exc_type = ""
        stack_lines: List[str] = []

        if record.exc_info and record.exc_info[0] is not None:
            exc_type = record.exc_info[0].__name__
            raw_frames = traceback.format_tb(record.exc_info[2])
            stack_lines = [
                line.strip()
                for block in raw_frames
                for line in block.splitlines()
                if line.strip()
            ]

        if not exc_type:
            exc_type = record.levelname

        message = self._redactor.redact(record.getMessage())
        stack_lines = self._redactor.redact_lines(stack_lines)

        ts = datetime.datetime.fromtimestamp(record.created, tz=datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

        return EventPayload(
            service=self._service,
            environment=self._environment,
            error_type=exc_type,
            message=message,
            stack_trace=stack_lines,
            timestamp=ts,
        )
