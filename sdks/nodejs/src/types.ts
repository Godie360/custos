export interface CustosConfig {
  apiKey: string;
  host: string;
  service: string;
  environment?: string;
  /** Max events to hold before flushing. Default: 10 */
  batchSize?: number;
  /** Flush interval in ms. Default: 5000 */
  flushIntervalMs?: number;
  /** Max retry attempts on failed send. Default: 3 */
  maxRetries?: number;
  /** Base delay in ms for exponential backoff. Default: 1000 */
  retryBaseDelayMs?: number;
}

export interface EventPayload {
  service: string;
  environment: string;
  error_type: string;
  message: string;
  stack_trace: string[];
  timestamp: string;
  sdk_version: string;
}
