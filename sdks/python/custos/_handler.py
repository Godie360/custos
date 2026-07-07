from __future__ import annotations

import logging

from ._queue import EventQueue
from ._redactor import Redactor


class CustosHandler(logging.Handler):
    """A logging.Handler that forwards ERROR-and-above records to Custos.

    Drop-in for any Python application that uses stdlib logging:

        import logging
        from custos import CustosHandler

        logging.getLogger().addHandler(
            CustosHandler(
                dsn="https://custos.example.com",
                api_key="custos_...",
                service="payments-api",
                environment="production",
            )
        )

    The handler never raises an exception into the calling application.
    Events are batched and sent on a background daemon thread.
    """

    def __init__(
        self,
        dsn: str,
        *,
        api_key: str,
        service: str,
        environment: str = "production",
        batch_size: int = 20,
        flush_interval: float = 5.0,
        max_retries: int = 5,
        timeout: float = 10.0,
    ) -> None:
        super().__init__(level=logging.ERROR)
        redactor = Redactor()
        self._queue = EventQueue(
            dsn=dsn,
            api_key=api_key,
            service=service,
            environment=environment,
            batch_size=batch_size,
            flush_interval=flush_interval,
            max_retries=max_retries,
            timeout=timeout,
            redactor=redactor,
        )
        self._queue.start()

    def emit(self, record: logging.LogRecord) -> None:
        try:
            self._queue.enqueue(record)
        except Exception:
            self.handleError(record)

    def close(self) -> None:
        try:
            self._queue.stop()
        except Exception:
            pass
        finally:
            super().close()
