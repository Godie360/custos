"""Queue tests — async batching, retry on 503, no exception propagation."""

import http.server
import json
import logging
import threading
import time
from typing import List
from unittest.mock import MagicMock, patch

import pytest

from custos._payload import EventPayload
from custos._queue import EventQueue
from custos._redactor import Redactor


def _make_queue(**overrides) -> EventQueue:
    defaults = dict(
        dsn="http://localhost:19999",
        api_key="custos_test",
        service="test-svc",
        environment="test",
        batch_size=10,
        flush_interval=0.1,
        max_retries=0,
        timeout=2.0,
        redactor=Redactor(),
    )
    defaults.update(overrides)
    return EventQueue(**defaults)


def _error_record(msg: str = "boom") -> logging.LogRecord:
    try:
        raise ValueError(msg)
    except ValueError:
        import sys
        record = logging.LogRecord(
            name="test",
            level=logging.ERROR,
            pathname="test.py",
            lineno=1,
            msg=msg,
            args=(),
            exc_info=sys.exc_info(),
        )
    return record


class TestBuildPayload:
    def test_extracts_exception_type(self) -> None:
        q = _make_queue()
        record = _error_record("something went wrong")
        payload = q._build_payload(record)
        assert payload.error_type == "ValueError"

    def test_extracts_message(self) -> None:
        q = _make_queue()
        record = _error_record("something went wrong")
        payload = q._build_payload(record)
        assert payload.message == "something went wrong"

    def test_stack_trace_is_list_of_strings(self) -> None:
        q = _make_queue()
        payload = q._build_payload(_error_record())
        assert isinstance(payload.stack_trace, list)
        assert all(isinstance(s, str) for s in payload.stack_trace)

    def test_redaction_applied_to_message(self) -> None:
        q = _make_queue()
        record = logging.LogRecord(
            name="test",
            level=logging.ERROR,
            pathname="test.py",
            lineno=1,
            msg="failed: password=s3cr3t",
            args=(),
            exc_info=None,
        )
        payload = q._build_payload(record)
        assert "s3cr3t" not in payload.message

    def test_service_and_environment_set(self) -> None:
        q = _make_queue(service="auth-svc", environment="staging")
        payload = q._build_payload(_error_record())
        assert payload.service == "auth-svc"
        assert payload.environment == "staging"

    def test_timestamp_format(self) -> None:
        import re

        q = _make_queue()
        payload = q._build_payload(_error_record())
        assert re.match(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z", payload.timestamp)


class TestQueueFull:
    def test_enqueue_when_full_does_not_raise(self) -> None:
        q = _make_queue(batch_size=1, flush_interval=60)
        # Fill the queue without starting the worker
        for _ in range(1001):
            try:
                q._q.put_nowait(object())
            except Exception:
                pass
        # enqueue must not raise
        q.enqueue(_error_record())


class TestRetry:
    def test_retries_on_503_then_gives_up(self) -> None:
        """Verify exponential backoff fires on 5xx and gives up after max_retries."""
        call_count = 0

        def fake_urlopen(req, timeout=None):
            nonlocal call_count
            call_count += 1
            import urllib.error
            raise urllib.error.HTTPError(
                url=req.full_url, code=503, msg="Service Unavailable", hdrs=None, fp=None
            )

        q = _make_queue(max_retries=2, flush_interval=0.05)

        with patch("custos._queue.urllib.request.urlopen", side_effect=fake_urlopen):
            with patch("custos._queue.time.sleep"):  # skip real backoff delays
                q._send_with_retry(
                    EventPayload(
                        service="svc",
                        environment="test",
                        error_type="Err",
                        message="msg",
                        stack_trace=[],
                        timestamp="2026-01-01T00:00:00Z",
                    )
                )

        # 1 initial attempt + 2 retries
        assert call_count == 3

    def test_no_retry_on_400(self) -> None:
        """4xx errors are not retried — bad payload is never going to succeed."""
        call_count = 0

        def fake_urlopen(req, timeout=None):
            nonlocal call_count
            call_count += 1
            import urllib.error
            raise urllib.error.HTTPError(
                url=req.full_url, code=400, msg="Bad Request", hdrs=None, fp=None
            )

        q = _make_queue(max_retries=5, flush_interval=0.05)

        with patch("custos._queue.urllib.request.urlopen", side_effect=fake_urlopen):
            q._send_with_retry(
                EventPayload(
                    service="svc",
                    environment="test",
                    error_type="Err",
                    message="msg",
                    stack_trace=[],
                    timestamp="2026-01-01T00:00:00Z",
                )
            )

        assert call_count == 1  # no retry


class TestWorkerThread:
    def test_flush_on_stop(self) -> None:
        """Events queued before stop() are flushed before the thread exits."""
        sent: List[dict] = []

        class FakeResp:
            status = 202
            def __enter__(self): return self
            def __exit__(self, *a): pass

        def fake_urlopen(req, timeout=None):
            sent.append(json.loads(req.data))
            return FakeResp()

        with patch("custos._queue.urllib.request.urlopen", side_effect=fake_urlopen):
            q = _make_queue(max_retries=0, flush_interval=60)
            q.start()
            q.enqueue(_error_record("first"))
            q.enqueue(_error_record("second"))
            q.stop(drain_timeout=3.0)

        assert len(sent) == 2
        messages = {p["message"] for p in sent}
        assert "first" in messages
        assert "second" in messages
