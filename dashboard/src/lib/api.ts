import type { AnalyticsSummary, FilterRule, Issue, ListIssuesParams, Project } from './types'

const BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText)
    throw new Error(`${res.status}: ${text}`)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export const api = {
  health: () => request<{ status: string }>('/health'),

  issues: {
    list: (params: ListIssuesParams = {}) => {
      const q = new URLSearchParams()
      if (params.service) q.set('service', params.service)
      if (params.environment) q.set('environment', params.environment)
      if (params.severity) q.set('severity', params.severity)
      q.set('limit', String(params.limit ?? 50))
      if (params.offset) q.set('offset', String(params.offset))
      return request<Issue[]>(`/api/v1/issues?${q}`)
    },
    get: (id: string) => request<Issue>(`/api/v1/issues/${id}`),
    patch: (id: string, body: Partial<Pick<Issue, 'status' | 'severity'>>) =>
      request<Issue>(`/api/v1/issues/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
  },

  analytics: {
    summary: () => request<AnalyticsSummary>('/api/v1/analytics/summary'),
  },

  projects: {
    list: () => request<Project[]>('/api/v1/projects'),
    create: (name: string) =>
      request<Project>('/api/v1/projects', { method: 'POST', body: JSON.stringify({ name }) }),
    createKey: (projectId: string, label: string) =>
      request<{ key: string }>(`/api/v1/projects/${projectId}/keys`, {
        method: 'POST',
        body: JSON.stringify({ label }),
      }),
    revokeKey: (projectId: string, keyId: string) =>
      request<void>(`/api/v1/projects/${projectId}/keys/${keyId}`, { method: 'DELETE' }),
  },

  filters: {
    list: (projectId: string) =>
      request<FilterRule[]>(`/api/v1/projects/${projectId}/filters`),
    create: (projectId: string, rule: Pick<FilterRule, 'field' | 'operator' | 'value'>) =>
      request<FilterRule>(`/api/v1/projects/${projectId}/filters`, {
        method: 'POST',
        body: JSON.stringify(rule),
      }),
    delete: (projectId: string, filterId: string) =>
      request<void>(`/api/v1/projects/${projectId}/filters/${filterId}`, { method: 'DELETE' }),
  },
}
