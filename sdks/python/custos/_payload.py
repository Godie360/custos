from __future__ import annotations

import datetime
from dataclasses import dataclass, field
from typing import List

SDK_VERSION = "0.1.0"


@dataclass
class EventPayload:
    service: str
    environment: str
    error_type: str
    message: str
    stack_trace: List[str]
    timestamp: str = field(default_factory=lambda: datetime.datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ"))
    sdk_version: str = SDK_VERSION

    def to_dict(self) -> dict:
        return {
            "service": self.service,
            "environment": self.environment,
            "error_type": self.error_type,
            "message": self.message,
            "stack_trace": self.stack_trace,
            "timestamp": self.timestamp,
            "sdk_version": self.sdk_version,
        }
