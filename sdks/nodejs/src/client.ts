import { CustosConfig, EventPayload } from './types';

const SDK_VERSION = '0.1.0';
const DEFAULT_BATCH_SIZE = 10;
const DEFAULT_FLUSH_INTERVAL_MS = 5000;
const DEFAULT_MAX_RETRIES = 3;
const DEFAULT_RETRY_BASE_DELAY_MS = 1000;

export class CustosClient {
  private readonly config: Required<CustosConfig>;
  private queue: EventPayload[] = [];
  private timer: ReturnType<typeof setInterval> | null = null;

  constructor(config: CustosConfig) {
    this.config = {
      environment: 'production',
      batchSize: DEFAULT_BATCH_SIZE,
      flushIntervalMs: DEFAULT_FLUSH_INTERVAL_MS,
      maxRetries: DEFAULT_MAX_RETRIES,
      retryBaseDelayMs: DEFAULT_RETRY_BASE_DELAY_MS,
      ...config,
    };
    this.startFlushTimer();
  }

  enqueue(payload: Omit<EventPayload, 'sdk_version' | 'timestamp'>): void {
    try {
      this.queue.push({
        ...payload,
        sdk_version: SDK_VERSION,
        timestamp: new Date().toISOString(),
      });
      if (this.queue.length >= this.config.batchSize) {
        this.flush();
      }
    } catch {
      // Never throw into the host application
    }
  }

  flush(): void {
    if (this.queue.length === 0) return;
    const batch = this.queue.splice(0, this.config.batchSize);
    // Fire and forget — errors are swallowed intentionally
    this.sendWithRetry(batch, 0).catch(() => undefined);
  }

  shutdown(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
    this.flush();
  }

  private startFlushTimer(): void {
    this.timer = setInterval(() => this.flush(), this.config.flushIntervalMs);
    // Allow the process to exit even if this timer is running
    if (this.timer.unref) this.timer.unref();
  }

  private async sendWithRetry(batch: EventPayload[], attempt: number): Promise<void> {
    for (const event of batch) {
      try {
        await this.sendOne(event, attempt);
      } catch {
        // Individual failures are swallowed — never crash the host
      }
    }
  }

  private async sendOne(event: EventPayload, attempt: number): Promise<void> {
    const url = `${this.config.host}/api/v1/ingest`;
    try {
      const res = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Custos-Key': this.config.apiKey,
        },
        body: JSON.stringify(event),
      });

      if (res.status === 202) return;

      // Retry on server errors
      if (res.status >= 500 && attempt < this.config.maxRetries) {
        await this.delay(this.backoff(attempt));
        return this.sendOne(event, attempt + 1);
      }
    } catch {
      // Network error — retry
      if (attempt < this.config.maxRetries) {
        await this.delay(this.backoff(attempt));
        return this.sendOne(event, attempt + 1);
      }
    }
  }

  private backoff(attempt: number): number {
    return Math.min(this.config.retryBaseDelayMs * Math.pow(2, attempt), 30000);
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
