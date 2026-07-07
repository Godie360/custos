"""Redaction tests — verify secrets are stripped before leaving the host."""

import pytest

from custos._redactor import Redactor


@pytest.fixture()
def r() -> Redactor:
    return Redactor()


class TestApiKeysAndTokens:
    def test_api_key_equals(self, r: Redactor) -> None:
        assert "sk-abc123" not in r.redact("api_key=sk-abc123")

    def test_api_key_colon(self, r: Redactor) -> None:
        assert "mysecret" not in r.redact("apikey: mysecret")

    def test_token_assignment(self, r: Redactor) -> None:
        assert "tok_xyz" not in r.redact("token=tok_xyz")

    def test_bearer_header(self, r: Redactor) -> None:
        result = r.redact("Authorization: Bearer eyJhbGci.payload.sig")
        assert "eyJhbGci" not in result
        assert "Authorization=<redacted>" in result

    def test_authorization_basic(self, r: Redactor) -> None:
        result = r.redact("Authorization=Basic dXNlcjpwYXNz")
        assert "dXNlcjpwYXNz" not in result


class TestPasswords:
    def test_password_equals(self, r: Redactor) -> None:
        assert "hunter2" not in r.redact("password=hunter2")

    def test_passwd_colon(self, r: Redactor) -> None:
        assert "s3cr3t" not in r.redact("passwd: s3cr3t")

    def test_secret_key(self, r: Redactor) -> None:
        assert "super_secret" not in r.redact("secret=super_secret")


class TestAwsCredentials:
    def test_aws_access_key_id(self, r: Redactor) -> None:
        assert "AKIAIOSFODNN7EXAMPLE" not in r.redact(
            "aws_access_key_id=AKIAIOSFODNN7EXAMPLE"
        )

    def test_aws_secret_key(self, r: Redactor) -> None:
        assert "wJalrXUtnFEMI" not in r.redact(
            "aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
        )


class TestPrivateKeys:
    def test_rsa_private_key(self, r: Redactor) -> None:
        pem = (
            "-----BEGIN RSA PRIVATE KEY-----\n"
            "MIIEowIBAAKCAQEA0Z3VS5JJcds3xHn/ygWep4PAtE\n"
            "-----END RSA PRIVATE KEY-----"
        )
        result = r.redact(pem)
        assert "MIIEowIBAAKCAQEA" not in result
        assert "<private-key-redacted>" in result


class TestCreditCards:
    @pytest.mark.parametrize(
        "number",
        [
            "4111111111111111",  # Visa
            "5500005555555559",  # MC
            "371449635398431",   # Amex
            "6011111111111117",  # Discover
        ],
    )
    def test_card_number_redacted(self, r: Redactor, number: str) -> None:
        assert number not in r.redact(f"card: {number}")

    def test_non_card_number_preserved(self, r: Redactor) -> None:
        # Short numbers must not be redacted
        result = r.redact("error code: 12345")
        assert "12345" in result


class TestEmail:
    def test_email_redacted(self, r: Redactor) -> None:
        assert "user@example.com" not in r.redact("contact: user@example.com")

    def test_non_email_preserved(self, r: Redactor) -> None:
        result = r.redact("status ok")
        assert "status ok" == result


class TestIpAddresses:
    def test_ipv4_redacted(self, r: Redactor) -> None:
        assert "192.168.1.100" not in r.redact("connecting to 192.168.1.100")


class TestRedactLines:
    def test_redact_list_applies_to_each_line(self, r: Redactor) -> None:
        lines = [
            "  File app.py, line 10",
            "    password=hunter2",
            "    raise ValueError",
        ]
        result = r.redact_lines(lines)
        assert "hunter2" not in result[1]
        assert "File app.py" in result[0]


class TestSafeInput:
    """Safe strings must pass through unchanged."""

    @pytest.mark.parametrize(
        "text",
        [
            "ValueError: list index out of range",
            "KeyError: 'user_id'",
            "File '/app/service.py', line 42, in handle_request",
            "",
        ],
    )
    def test_clean_text_unchanged(self, r: Redactor, text: str) -> None:
        assert r.redact(text) == text
