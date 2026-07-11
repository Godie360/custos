import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { CheckCircle, EyeOff, RefreshCw, SlidersHorizontal } from 'lucide-react'
import { api } from '#/lib/api'
import { SeverityBadge, StatusBadge } from '#/components/Badge'
import { timeAgo } from '#/lib/utils'
import type { Issue } from '#/lib/types'

export const Route = createFileRoute('/issues/')({ component: IssuesFeed })

const SEVERITIES = ['', 'low', 'medium', 'high', 'critical']
const STATUSES = ['', 'open', 'resolved', 'ignored']

function IssueRow({ issue, onPatch }: { issue: Issue; onPatch: (id: string, status: string) => void }) {
  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: '2fr 1fr 1fr 90px 80px 70px 100px',
        alignItems: 'center',
        padding: '12px 16px',
        borderBottom: '1px solid var(--border)',
        gap: 12,
      }}
      className="hover:bg-[#0F1915] transition-colors group"
    >
      {/* Error info */}
      <div className="min-w-0">
        <Link
          to="/issues/$id"
          params={{ id: issue.id }}
          style={{ color: 'var(--text-1)', fontSize: 13.5, fontWeight: 500 }}
          className="hover:text-[#2DD4BF] transition-colors block truncate"
        >
          {issue.error_type}
        </Link>
        <p style={{ color: 'var(--text-3)', fontSize: 11.5, marginTop: 2 }} className="truncate">
          {issue.message}
        </p>
      </div>

      {/* Service */}
      <div className="min-w-0">
        <span
          style={{
            background: 'rgba(45,212,191,0.07)', color: '#2DD4BF',
            fontSize: 11.5, padding: '3px 8px', borderRadius: 5, fontFamily: 'monospace',
          }}
          className="truncate block"
        >
          {issue.service}
        </span>
      </div>

      {/* Environment */}
      <span style={{ color: 'var(--text-2)', fontSize: 12 }}>{issue.environment}</span>

      {/* Severity */}
      <SeverityBadge severity={issue.severity} />

      {/* Status */}
      <StatusBadge status={issue.status} />

      {/* Count */}
      <div style={{ textAlign: 'right' }}>
        <span style={{ color: 'var(--text-1)', fontSize: 13, fontWeight: 600 }}>
          {issue.occurrence_count.toLocaleString()}
        </span>
        <span style={{ color: 'var(--text-3)', fontSize: 10, display: 'block' }}>occurrences</span>
      </div>

      {/* Actions */}
      <div className="flex gap-2 justify-end opacity-0 group-hover:opacity-100 transition-opacity">
        {issue.status !== 'resolved' && (
          <button
            onClick={() => onPatch(issue.id, 'resolved')}
            style={{ color: '#4ADE80' }}
            className="hover:opacity-70 transition-opacity"
            title="Resolve"
          >
            <CheckCircle size={14} />
          </button>
        )}
        {issue.status !== 'ignored' && (
          <button
            onClick={() => onPatch(issue.id, 'ignored')}
            style={{ color: 'var(--text-3)' }}
            className="hover:opacity-70 transition-opacity"
            title="Ignore"
          >
            <EyeOff size={14} />
          </button>
        )}
      </div>
    </div>
  )
}

function IssuesFeed() {
  const qc = useQueryClient()
  const [service, setService] = useState('')
  const [environment, setEnvironment] = useState('')
  const [severity, setSeverity] = useState('')
  const [status, setStatus] = useState('')
  const [showFilters, setShowFilters] = useState(false)

  const { data: issues = [], isLoading, isFetching, refetch } = useQuery({
    queryKey: ['issues', { service, environment, severity, status }],
    queryFn: () =>
      api.issues.list({
        service: service || undefined,
        environment: environment || undefined,
        severity: severity || undefined,
        limit: 100,
      }),
    refetchInterval: 15_000,
    select: data =>
      status ? data.filter(i => i.status === status) : data,
  })

  const patch = useMutation({
    mutationFn: ({ id, s }: { id: string; s: string }) =>
      api.issues.patch(id, { status: s as Issue['status'] }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['issues'] })
      void qc.invalidateQueries({ queryKey: ['analytics'] })
    },
  })

  const open = issues.filter(i => i.status === 'open').length
  const resolved = issues.filter(i => i.status === 'resolved').length
  const critical = issues.filter(i => i.severity === 'critical').length

  return (
    <div>
      {/* Page header */}
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 style={{ color: 'var(--text-1)', fontSize: 20, fontWeight: 700, marginBottom: 4 }}>Issues</h1>
          <div className="flex gap-4">
            {[
              { label: 'Open', val: open, color: '#2DD4BF' },
              { label: 'Resolved', val: resolved, color: '#4ADE80' },
              { label: 'Critical', val: critical, color: '#F87171' },
            ].map(({ label, val, color }) => (
              <span key={label} style={{ fontSize: 12, color: 'var(--text-2)' }}>
                <span style={{ color, fontWeight: 700 }}>{val}</span> {label}
              </span>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowFilters(v => !v)}
            className="btn-ghost flex items-center gap-1.5"
            style={{ color: showFilters ? '#2DD4BF' : undefined }}
          >
            <SlidersHorizontal size={13} /> Filters
          </button>
          <button
            onClick={() => refetch()}
            className="btn-ghost flex items-center gap-1.5"
            disabled={isFetching}
          >
            <RefreshCw size={13} className={isFetching ? 'animate-spin' : ''} />
            Refresh
          </button>
        </div>
      </div>

      {/* Filters bar */}
      {showFilters && (
        <div className="card flex gap-3 flex-wrap mb-4" style={{ padding: '14px 16px' }}>
          <input
            value={service}
            onChange={e => setService(e.target.value)}
            placeholder="Service…"
            style={{ padding: '6px 12px', fontSize: 12.5, width: 160 }}
          />
          <input
            value={environment}
            onChange={e => setEnvironment(e.target.value)}
            placeholder="Environment…"
            style={{ padding: '6px 12px', fontSize: 12.5, width: 140 }}
          />
          <select
            value={severity}
            onChange={e => setSeverity(e.target.value)}
            style={{ padding: '6px 12px', fontSize: 12.5 }}
          >
            {SEVERITIES.map(s => <option key={s} value={s}>{s || 'All severities'}</option>)}
          </select>
          <select
            value={status}
            onChange={e => setStatus(e.target.value)}
            style={{ padding: '6px 12px', fontSize: 12.5 }}
          >
            {STATUSES.map(s => <option key={s} value={s}>{s || 'All statuses'}</option>)}
          </select>
          {(service || environment || severity || status) && (
            <button
              className="btn-ghost"
              style={{ fontSize: 12 }}
              onClick={() => { setService(''); setEnvironment(''); setSeverity(''); setStatus('') }}
            >
              Clear
            </button>
          )}
        </div>
      )}

      {/* Table */}
      <div className="card overflow-hidden">
        {/* Header */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '2fr 1fr 1fr 90px 80px 70px 100px',
            padding: '10px 16px',
            borderBottom: '1px solid var(--border)',
            gap: 12,
            background: 'rgba(0,0,0,0.2)',
          }}
        >
          {['Error', 'Service', 'Environment', 'Severity', 'Status', 'Count', ''].map(h => (
            <span key={h} style={{ color: 'var(--text-3)', fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
              {h}
            </span>
          ))}
        </div>

        {isLoading ? (
          <div className="space-y-0">
            {[...Array(8)].map((_, i) => (
              <div key={i} style={{ padding: '14px 16px', borderBottom: '1px solid var(--border)' }}>
                <div className="skeleton" style={{ height: 14, width: '60%', marginBottom: 6 }} />
                <div className="skeleton" style={{ height: 10, width: '40%' }} />
              </div>
            ))}
          </div>
        ) : issues.length === 0 ? (
          <div style={{ padding: '60px 0', textAlign: 'center' }}>
            <p style={{ color: 'var(--text-3)', fontSize: 14 }}>No issues match your filters.</p>
          </div>
        ) : (
          issues.map(issue => (
            <IssueRow
              key={issue.id}
              issue={issue}
              onPatch={(id, s) => patch.mutate({ id, s })}
            />
          ))
        )}

        {!isLoading && issues.length > 0 && (
          <div style={{ padding: '10px 16px', borderTop: '1px solid var(--border)' }}>
            <span style={{ color: 'var(--text-3)', fontSize: 12 }}>
              {issues.length} issue{issues.length !== 1 ? 's' : ''} · refreshes every 15s
            </span>
          </div>
        )}
      </div>
    </div>
  )
}
