import { CustosClient } from '../src/client';

const mockFetch = jest.fn();
global.fetch = mockFetch;

const config = {
  apiKey: 'test-key',
  host: 'http://localhost:8080',
  service: 'test-service',
  environment: 'test',
  flushIntervalMs: 60000, // disable auto-flush in tests
  batchSize: 5,
  maxRetries: 2,
};

beforeEach(() => {
  mockFetch.mockReset();
});

describe('CustosClient', () => {
  it('sends event to ingest endpoint with correct headers', async () => {
    mockFetch.mockResolvedValue({ status: 202 });

    const client = new CustosClient(config);
    client.enqueue({
      service: 'svc',
      environment: 'prod',
      error_type: 'TypeError',
      message: 'something went wrong',
      stack_trace: ['at foo (bar.js:1)'],
    });
    client.flush();
    client.shutdown();

    await new Promise(r => setTimeout(r, 50));

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/ingest',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'X-Custos-Key': 'test-key',
          'Content-Type': 'application/json',
        }),
      })
    );
  });

  it('retries on 503 up to maxRetries times', async () => {
    mockFetch.mockResolvedValue({ status: 503 });

    // retryBaseDelayMs: 0 makes retries instant so the test doesn't need to sleep
    const client = new CustosClient({ ...config, maxRetries: 2, retryBaseDelayMs: 0 });
    client.enqueue({
      service: 'svc',
      environment: 'prod',
      error_type: 'Error',
      message: 'fail',
      stack_trace: [],
    });
    client.flush();

    await new Promise(r => setTimeout(r, 50));
    client.shutdown();

    // Initial attempt + 2 retries = 3 calls
    expect(mockFetch).toHaveBeenCalledTimes(3);
  });

  it('never throws when fetch rejects', async () => {
    mockFetch.mockRejectedValue(new Error('network error'));

    const client = new CustosClient({ ...config, maxRetries: 0 });
    expect(() => {
      client.enqueue({
        service: 'svc',
        environment: 'prod',
        error_type: 'Error',
        message: 'fail',
        stack_trace: [],
      });
      client.flush();
    }).not.toThrow();

    client.shutdown();
    await new Promise(r => setTimeout(r, 50));
  });

  it('auto-flushes when batch size is reached', async () => {
    mockFetch.mockResolvedValue({ status: 202 });

    const client = new CustosClient({ ...config, batchSize: 2 });
    const event = {
      service: 'svc', environment: 'prod',
      error_type: 'Error', message: 'err', stack_trace: [],
    };
    client.enqueue(event);
    client.enqueue(event); // triggers auto-flush at batchSize=2

    await new Promise(r => setTimeout(r, 50));
    client.shutdown();

    expect(mockFetch).toHaveBeenCalled();
  });
});
