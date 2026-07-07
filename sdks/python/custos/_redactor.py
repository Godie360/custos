from __future__ import annotations

import re
from typing import List, Tuple

# Each entry: (compiled pattern, replacement). Applied in order.
_PATTERNS: List[Tuple[re.Pattern, str]] = [
    # PEM private keys — must run before generic key/secret patterns
    (
        re.compile(
            r"-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----.*?-----END (?:RSA |EC |OPENSSH )?PRIVATE KEY-----",
            re.DOTALL,
        ),
        "<private-key-redacted>",
    ),
    # Full Authorization header line (covers Bearer, Basic, Digest, etc.)
    (re.compile(r"(?i)(Authorization)\s*[:=]\s*.+"), r"\1=<redacted>"),
    # Bearer token outside of Authorization header context
    (re.compile(r"(?i)Bearer\s+[A-Za-z0-9\-._~+/]+=*"), "Bearer <redacted>"),
    # Generic key / token / secret / password assignments
    (
        re.compile(
            r"(?i)(api[_\-]?key|apikey|api[_\-]?token|token|secret|password|passwd|pwd|private[_\-]?key)\s*[:=]\s*\S+"
        ),
        r"\1=<redacted>",
    ),
    # AWS credentials
    (
        re.compile(r"(?i)(aws_access_key_id|aws_secret_access_key|aws_session_token)\s*[:=]\s*\S+"),
        r"\1=<redacted>",
    ),
    # Credit card numbers (major schemes, unformatted)
    (
        re.compile(
            r"\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b"
        ),
        "<card-redacted>",
    ),
    # Email addresses
    (re.compile(r"\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b"), "<email-redacted>"),
    # IPv4 addresses
    (re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b"), "<ip-redacted>"),
]


class Redactor:
    """Applies regex-based secret redaction to strings before they leave the host."""

    def redact(self, text: str) -> str:
        for pattern, replacement in _PATTERNS:
            text = pattern.sub(replacement, text)
        return text

    def redact_lines(self, lines: List[str]) -> List[str]:
        return [self.redact(line) for line in lines]
