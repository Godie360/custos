import { useQuery } from '@tanstack/react-query'
import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { api } from '#/lib/api'
import type { Issue } from '#/lib/types'

export const Route = createFileRoute('/issues/')({ component: IssuesFeed })

const SEVERITIES = ['', 'low', 'medium', 'high', 'critical']

const severityBadge: Record<string, string> = {
  low: 'bg-gray-700 text-gray-300',
  medium: 'bg-yellow-900 text-yellow-300',
  high: 'bg-orange-900 text-orange-300',
  critical: 'bg-red-900 text-red-400',
  error: 'bg-orange-900 text-orange-300',
}

function SeverityBadge({ s }: { s: string }) {
  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${severityBadge[s] ?? 'bg-gray-700 text-gray-400'}`}>
      {s}
    </span>
  )
}

function timeAgo(iso: string) {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

function IssuesFeed() {
  const [service, setService] = useState('')
  const [environment, setEnvironment] = useState('')
  const [severity, setSeverity] = useState('')

  const { data: issues = [], isLoading, error } = useQuery({
    queryKey: ['issues', { service, environment, severity }],
    queryFn: () => api.issues.list({ service: service || undefined, environment: environment || undefined, severity: severity || undefined, limit: 50 }),
    refetchInterval: 15_000,
  })

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-white">Issues</h1>
        <span className="text-xs text-gray-500">auto-refreshes every 15s</span>
      </div>

      {/* Filters */}
      <div className="flex gap-3 mb-5">
        <input
          value={service}
          onChange={e => setService(e.target.value)}
          placeholder="Filter by service…"
          className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-indigo-500 w-44"
        />
        <input
          value={environment}
          onChange={e => setEnvironment(e.target.value)}
          placeholder="Environment…"
          className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-indigo-500 w-36"
        />
        <select
          value={severity}
          onChange={e => setSeverity(e.target.value)}
          className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 focus:outline-none focus:border-indigo-500"
        >
          {SEVERITIES.map(s => (
            <option key={s} value={s}>{s || 'All severities'}</option>
          ))}
        </select>
      </div>

      {isLoading && <p className="text-gray-500 text-sm">Loading…</p>}
      {error && <p className="text-red-400 text-sm">Failed to load issues.</p>}

      {!isLoading && !error && issues.length === 0 && (
        <p className="text-gray-500 text-sm">No issues match your filters.</p>
      )}

      <div className="space-y-2">
        {issues.map((issue: Issue) => (
          <Link
            key={issue.id}
            to="/issues/$id"
            params={{ id: issue.id }}
            className="block bg-gray-900 border border-gray-800 rounded-lg px-4 py-3 hover:border-gray-600 transition-colors"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <SeverityBadge s={issue.severity} />
                  <span className="text-xs text-gray-500 font-mono">{issue.service}</span>
                  <span className="text-xs text-gray-600">{issue.environment}</span>
                </div>
                <p className="text-sm font-medium text-gray-100 truncate">{issue.error_type}</p>
                <p className="text-xs text-gray-500 truncate mt-0.5">{issue.message}</p>
              </div>
              <div className="shrink-0 text-right">
                <p className="text-xs text-gray-400">{issue.occurrence_count}×</p>
                <p className="text-xs text-gray-600 mt-0.5">{timeAgo(issue.last_seen)}</p>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
