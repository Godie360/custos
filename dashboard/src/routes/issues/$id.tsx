import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft, Bot, CheckCircle, Clock, EyeOff, Hash, Layers, Lightbulb, Terminal } from 'lucide-react'
import { api } from '#/lib/api'
import { SeverityBadge, StatusBadge } from '#/components/Badge'
import { timeAgo } from '#/lib/utils'

export const Route = createFileRoute('/issues/$id')({ component: IssueDetail })

function InfoBlock({ label, value, mono = false }: { label: string; value: string | number; mono?: boolean }) {
  return (
    <div style={{ background: 'rgba(0,0,0,0.2)', borderRadius: 8, padding: '12px 16px' }}>
      <p style={{ color: 'var(--text-3)', fontSize: 11, marginBottom: 5, textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</p>
      <p style={{ color: 'var(--text-1)', fontSize: 13.5, fontFamily: mono ? 'monospace' : undefined }}>{value}</p>
    </div>
  )
}

function AISection({ icon: Icon, title, color, children }: {
  icon: React.ElementType; title: string; color: string; children: React.ReactNode
}) {
  return (
    <div style={{ border: `1px solid ${color}25`, borderRadius: 10, overflow: 'hidden' }}>
      <div style={{ background: `${color}08`, borderBottom: `1px solid ${color}20`, padding: '10px 16px', display: 'flex', alignItems: 'center', gap: 8 }}>
        <Icon size={14} style={{ color }} />
        <span style={{ color, fontSize: 12.5, fontWeight: 600 }}>{title}</span>
      </div>
      <div style={{ padding: '14px 16px', fontSize: 13.5, lineHeight: 1.65, color: 'var(--text-1)' }}>
        {children}
      </div>
    </div>
  )
}

function IssueDetail() {
  const { id } = Route.useParams()
  const qc = useQueryClient()

  const { data: issue, isLoading, error } = useQuery({
    queryKey: ['issue', id],
    queryFn: () => api.issues.get(id),
  })

  const patch = useMutation({
    mutationFn: (status: 'resolved' | 'ignored' | 'open') =>
      api.issues.patch(id, { status }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['issue', id] })
      void qc.invalidateQueries({ queryKey: ['issues'] })
      void qc.invalidateQueries({ queryKey: ['analytics'] })
    },
  })

  if (isLoading) {
    return (
      <div className="space-y-4 max-w-3xl">
        <div className="skeleton" style={{ height: 24, width: 200 }} />
        <div className="skeleton" style={{ height: 100, borderRadius: 10 }} />
        <div className="skeleton" style={{ height: 160, borderRadius: 10 }} />
      </div>
    )
  }

  if (error || !issue) {
    return (
      <div className="card" style={{ padding: 32, textAlign: 'center' }}>
        <p style={{ color: 'var(--text-3)' }}>Issue not found or failed to load.</p>
        <Link to="/issues" className="btn-ghost" style={{ marginTop: 16, display: 'inline-block' }}>
          ← Back to issues
        </Link>
      </div>
    )
  }

  const hasAI = !!(issue.ai_explanation || issue.likely_cause || (issue.suggested_checks?.length))

  return (
    <div className="max-w-3xl space-y-5">
      {/* Back */}
      <Link
        to="/issues"
        style={{ display: 'inline-flex', alignItems: 'center', gap: 6, color: 'var(--text-2)', fontSize: 13 }}
        className="hover:text-[#2DD4BF] transition-colors"
      >
        <ArrowLeft size={13} /> Back to issues
      </Link>

      {/* Header card */}
      <div className="card" style={{ padding: '20px 24px' }}>
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="flex items-center gap-2 mb-3 flex-wrap">
              <SeverityBadge severity={issue.severity} />
              <StatusBadge status={issue.status} />
              <span style={{ color: 'var(--text-3)', fontSize: 11 }}>
                ID: <span style={{ fontFamily: 'monospace' }}>{issue.id.slice(0, 8)}</span>
              </span>
            </div>
            <h1 style={{ color: 'var(--text-1)', fontSize: 20, fontWeight: 700, letterSpacing: '-0.3px', marginBottom: 8 }}>
              {issue.error_type}
            </h1>
            <p style={{ color: 'var(--text-2)', fontSize: 14, lineHeight: 1.5 }}>{issue.message}</p>
          </div>
          <div className="flex gap-2 shrink-0">
            {issue.status !== 'open' && (
              <button onClick={() => patch.mutate('open')} disabled={patch.isPending} className="btn-ghost flex items-center gap-1.5" style={{ fontSize: 12 }}>
                Reopen
              </button>
            )}
            {issue.status !== 'resolved' && (
              <button onClick={() => patch.mutate('resolved')} disabled={patch.isPending} className="btn-ghost flex items-center gap-1.5" style={{ fontSize: 12, color: '#4ADE80', borderColor: '#4ADE8030' }}>
                <CheckCircle size={13} /> Resolve
              </button>
            )}
            {issue.status !== 'ignored' && (
              <button onClick={() => patch.mutate('ignored')} disabled={patch.isPending} className="btn-ghost flex items-center gap-1.5" style={{ fontSize: 12 }}>
                <EyeOff size={13} /> Ignore
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Metadata grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
        <InfoBlock label="Service" value={issue.service} mono />
        <InfoBlock label="Environment" value={issue.environment} />
        <InfoBlock label="Occurrences" value={issue.occurrence_count.toLocaleString()} />
        <InfoBlock label="First seen" value={new Date(issue.first_seen).toLocaleString()} />
        <InfoBlock label="Last seen" value={new Date(issue.last_seen).toLocaleString()} />
        <InfoBlock label="Time since first" value={timeAgo(issue.first_seen)} />
      </div>

      {/* AI Analysis */}
      {hasAI ? (
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <Bot size={15} style={{ color: '#2DD4BF' }} />
            <span style={{ color: 'var(--text-1)', fontSize: 14, fontWeight: 600 }}>AI Analysis</span>
          </div>

          {issue.ai_explanation && (
            <AISection icon={Layers} title="Explanation" color="#2DD4BF">
              {issue.ai_explanation}
            </AISection>
          )}

          {issue.likely_cause && (
            <AISection icon={Lightbulb} title="Likely Cause" color="#FCD34D">
              {issue.likely_cause}
            </AISection>
          )}

          {issue.suggested_checks && issue.suggested_checks.length > 0 && (
            <AISection icon={CheckCircle} title="Suggested Checks" color="#4ADE80">
              <ul className="space-y-2">
                {issue.suggested_checks.map((c, i) => (
                  <li key={i} style={{ display: 'flex', gap: 10, color: 'var(--text-1)' }}>
                    <span style={{ color: '#4ADE80', fontWeight: 700, shrink: 0 }}>→</span>
                    {c}
                  </li>
                ))}
              </ul>
            </AISection>
          )}
        </div>
      ) : (
        <div className="card" style={{ padding: '20px 24px', display: 'flex', alignItems: 'center', gap: 12 }}>
          <Clock size={16} style={{ color: 'var(--text-3)' }} />
          <p style={{ color: 'var(--text-3)', fontSize: 13 }}>
            AI analysis is pending or the AI provider is not configured. Results appear here after analysis completes.
          </p>
        </div>
      )}

      {/* Stack trace */}
      {issue.stack_trace && issue.stack_trace.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Terminal size={14} style={{ color: 'var(--text-2)' }} />
            <span style={{ color: 'var(--text-1)', fontSize: 14, fontWeight: 600 }}>Stack Trace</span>
            <span style={{ color: 'var(--text-3)', fontSize: 11 }}>({issue.stack_trace.length} frames)</span>
          </div>
          <pre
            style={{
              background: '#050708',
              border: '1px solid var(--border)',
              borderRadius: 10,
              padding: '16px 20px',
              fontSize: 12,
              color: '#7A9DB8',
              overflowX: 'auto',
              lineHeight: 1.7,
              fontFamily: 'ui-monospace, monospace',
            }}
          >
            {issue.stack_trace.map((f, i) => (
              <div key={i} style={{ color: i === 0 ? '#2DD4BF' : undefined }}>
                {f}
              </div>
            ))}
          </pre>
        </div>
      )}

      {/* Fingerprint */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: 'var(--text-3)', fontSize: 11 }}>
        <Hash size={11} />
        <span>Fingerprint: <span style={{ fontFamily: 'monospace' }}>{issue.fingerprint}</span></span>
      </div>
    </div>
  )
}
