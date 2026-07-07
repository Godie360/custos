import { redact, redactStack } from '../src/redact';

describe('redact', () => {
  it('strips bearer tokens', () => {
    const input = 'Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.payload.signature';
    expect(redact(input)).not.toContain('eyJ');
    expect(redact(input)).toContain('[REDACTED]');
  });

  it('strips password in connection strings', () => {
    const input = 'postgres://user:supersecret@localhost:5432/db';
    const result = redact(input);
    expect(result).not.toContain('supersecret');
    expect(result).toContain('[REDACTED]');
  });

  it('strips API keys in key=value pairs', () => {
    const input = 'api_key=sk-abc123xyz789';
    const result = redact(input);
    expect(result).not.toContain('sk-abc123xyz789');
    expect(result).toContain('[REDACTED]');
  });

  it('strips JWT tokens', () => {
    const jwt = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0In0.hash';
    expect(redact(jwt)).not.toContain('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9');
  });

  it('leaves normal text untouched', () => {
    const input = 'TypeError: Cannot read property id of undefined';
    expect(redact(input)).toBe(input);
  });

  it('never throws on any input', () => {
    expect(() => redact('')).not.toThrow();
    expect(() => redact('normal log line')).not.toThrow();
    expect(() => redact('Bearer token123')).not.toThrow();
  });
});

describe('redactStack', () => {
  it('redacts sensitive values in stack frames', () => {
    const frames = [
      'at processPayment (payment.js:42)',
      'at callWithApiKey api_key=secret123 (auth.js:10)',
    ];
    const result = redactStack(frames);
    expect(result[0]).toBe('at processPayment (payment.js:42)');
    expect(result[1]).not.toContain('secret123');
  });
});
