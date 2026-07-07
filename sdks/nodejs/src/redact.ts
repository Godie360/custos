const REDACTED = '[REDACTED]';

const PATTERNS: RegExp[] = [
  // Authorization headers and bearer tokens
  /Bearer\s+[A-Za-z0-9\-._~+/]+=*/gi,
  // Generic API keys (key=value style)
  /([a-z_-]*(api[_-]?key|apikey|secret|password|passwd|token|auth)[a-z_-]*\s*[=:]\s*)(['"]?)[\w\-./+]{6,}(\3)/gi,
  // Connection strings
  /(?:postgres|mysql|mongodb|redis|amqp):\/\/[^\s"']+/gi,
  // AWS keys
  /AKIA[0-9A-Z]{16}/g,
  // Private keys
  /-----BEGIN\s+(?:RSA\s+)?PRIVATE KEY-----[\s\S]*?-----END\s+(?:RSA\s+)?PRIVATE KEY-----/g,
  // JWT tokens
  /eyJ[A-Za-z0-9\-_]+\.eyJ[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_.+/]*/g,
];

export function redact(input: string): string {
  let output = input;
  for (const pattern of PATTERNS) {
    output = output.replace(pattern, (match, ...groups) => {
      // For key=value patterns, preserve the key part
      if (groups.length >= 2 && typeof groups[0] === 'string') {
        return groups[0] + REDACTED;
      }
      return REDACTED;
    });
  }
  return output;
}

export function redactStack(frames: string[]): string[] {
  return frames.map(redact);
}
