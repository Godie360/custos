# custos-sdk

Python SDK for [Custos](https://github.com/Godie360/custos) — AI-powered log intelligence.

## Installation

```bash
pip install custos-sdk
```

## Quickstart

```python
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
```

That's it. Any `logging.error(...)` or `logging.exception(...)` call is automatically captured, redacted, and forwarded to Custos in the background.

## How it works

- **Drop-in**: subclasses `logging.Handler` — works with Django, Flask, FastAPI, and any other framework that uses stdlib logging.
- **Non-blocking**: events are queued and sent from a background daemon thread. Your application never waits on the network.
- **Redaction**: API keys, passwords, tokens, credit card numbers, emails, and IP addresses are stripped before any data leaves your host.
- **Retry**: transient network errors and 5xx responses trigger exponential backoff (default: 5 retries, cap 60 s).
- **Safe**: the handler catches all its own exceptions internally and never propagates them to your application.

## Configuration

| Parameter | Default | Description |
|---|---|---|
| `dsn` | required | Base URL of your Custos server |
| `api_key` | required | `X-Custos-Key` for your project |
| `service` | required | Name of the service (e.g. `payments-api`) |
| `environment` | `"production"` | Deployment environment |
| `batch_size` | `20` | Max events per flush |
| `flush_interval` | `5.0` | Seconds between flushes |
| `max_retries` | `5` | Max send attempts per event |
| `timeout` | `10.0` | HTTP request timeout in seconds |

## Framework examples

### FastAPI

```python
from fastapi import FastAPI
import logging
from custos import CustosHandler

logging.getLogger().addHandler(
    CustosHandler(dsn="...", api_key="...", service="my-api")
)
app = FastAPI()
```

### Django

In `settings.py`:

```python
LOGGING = {
    "version": 1,
    "handlers": {
        "custos": {
            "()": "custos.CustosHandler",
            "dsn": "https://custos.example.com",
            "api_key": "custos_...",
            "service": "django-app",
            "environment": "production",
        }
    },
    "root": {"handlers": ["custos"], "level": "ERROR"},
}
```

## Development

```bash
cd sdks/python
pip install -e ".[dev]"
pytest
```

## License

MIT
