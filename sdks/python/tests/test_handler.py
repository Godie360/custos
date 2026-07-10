"""Handler integration tests — logging.Handler contract and exception isolation."""

import logging
import sys
from unittest.mock import MagicMock, patch

import pytest

from custos import CustosHandler


def _make_handler(**overrides) -> CustosHandler:
    defaults = dict(
        dsn="http://localhost:19999",
        api_key="custos_test",
        service="test-svc",
        environment="test",
        max_retries=0,
    )
    defaults.update(overrides)
    # Patch the queue so no real thread or HTTP call is made.
    with patch("custos._handler.EventQueue") as MockQueue:
        MockQueue.return_value.start = MagicMock()
        h = CustosHandler(**defaults)
        h._queue = MockQueue.return_value
    return h


class TestHandlerContract:
    def test_level_is_error(self) -> None:
        h = _make_handler()
        assert h.level == logging.ERROR

    def test_emit_calls_enqueue(self) -> None:
        h = _make_handler()
        record = logging.LogRecord(
            name="test", level=logging.ERROR, pathname="x.py",
            lineno=1, msg="oops", args=(), exc_info=None,
        )
        h.emit(record)
        h._queue.enqueue.assert_called_once_with(record)

    def test_close_calls_stop(self) -> None:
        h = _make_handler()
        h.close()
        h._queue.stop.assert_called_once()

    def test_below_error_not_forwarded(self) -> None:
        h = _make_handler()
        logger = logging.getLogger("custos.test.below")
        logger.addHandler(h)
        logger.warning("this should not be forwarded")
        h._queue.enqueue.assert_not_called()
        logger.removeHandler(h)


class TestExceptionIsolation:
    def test_emit_never_raises_even_if_enqueue_raises(self) -> None:
        h = _make_handler()
        h._queue.enqueue.side_effect = RuntimeError("queue exploded")
        record = logging.LogRecord(
            name="test", level=logging.ERROR, pathname="x.py",
            lineno=1, msg="error", args=(), exc_info=None,
        )
        # Must not raise — CustosHandler must never crash the host app.
        h.emit(record)

    def test_close_never_raises_even_if_stop_raises(self) -> None:
        h = _make_handler()
        h._queue.stop.side_effect = RuntimeError("stop exploded")
        h.close()  # must not propagate


class TestLoggingIntegration:
    def test_attaches_to_logger_and_captures_exception(self) -> None:
        """End-to-end: logger.exception() triggers enqueue with exc_info."""
        enqueued = []

        class CapturingQueue:
            def enqueue(self, record):
                enqueued.append(record)
            def start(self): pass
            def stop(self, **kw): pass

        with patch("custos._handler.EventQueue", return_value=CapturingQueue()):
            h = CustosHandler(
                dsn="http://localhost:19999",
                api_key="key",
                service="svc",
                environment="test",
            )

        logger = logging.getLogger("custos.test.integration")
        logger.addHandler(h)
        try:
            raise TypeError("bad type")
        except TypeError:
            logger.exception("caught an error")
        logger.removeHandler(h)

        assert len(enqueued) == 1
        assert enqueued[0].exc_info is not None
        assert enqueued[0].exc_info[0] is TypeError
