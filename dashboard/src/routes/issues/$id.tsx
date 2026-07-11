import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft, CheckCircle, EyeOff } from 'lucide-react'
import { api } from '#/lib/api'

export const Route = createFileRoute('/issues/$id')({ component: IssueDetail })

function IssueDetail() {
  const { id } = Route.useParams()
  const qc = useQueryClient()

  const { data: issue, isLoading, error } = useQuery({
    queryKey: ['issue', id],
    queryFn: () => api.issues.get(id),
  })

  const patch = useMutation({
    mutationFn: (status: 'resolved' | 'ignored') =>
      api.issues.patch(id, { status }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['issue', id] })
      void qc.invalidateQueries({ queryKey: ['issues'] })
    },
  })

  if (isLoading) return <div className="p-6 text-gray-500 text-sm">Loading…</div>
  if (error || !issue) return <div className="p-6 text-red-400 text-sm">Issue not found.</div>

  const severityColor: Record<string, string> = {
    critical: 'text-red-400',
    high: 'text-orange-400',
    medium: 'text-yellow-400',
    low: 'text-gray-400',
    error: 'text-orange-400',
  }

  return (
    <div className="p-6 max-w-3xl">
      <Link to="/issues" className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-300 mb-5">
        <ArrowLeft size={12} /> Back to issues
      </Link>

      <div className="flex items-start justify-between gap-4 mb-6">
        <div>
          <p className={`text-xs font-medium uppercase tracking-wide mb-1 ${severityColor[issue.severity] ?? 'text-gray-400'}`}>
            {issue.severity}
          </p>
          <h1 className="text-lg font-semibold text-white">{issue.error_type}</h1>
          <p className="text-sm text-gray-400 mt-1">{issue.message}</p>
        </div>
        <div className="flex gap-2 shrink-0">
          {issue.status !== 'resolved' && (
            <button
              onClick={() => patch.mutate('resolved')}
              disabled={patch.isPending}
              className="flex items-center gap-1 px-3 py-1.5 bg-green-900 hover:bg-green-800 text-green-300 text-xs rounded transition-colors disabled:opacity-50"
            >
              <CheckCircle size={12} /> Resolve
            </button>
          )}
          {issue.status !== 'ignored' && (
            <button
              onClick={() => patch.mutate('ignored')}
              disabled={patch.isPending}
              className="flex items-center gap-1 px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-gray-400 text-xs rounded transition-colors disabled:opacity-50"
            >
              <EyeOff size={12} /> Ignore
            </button>
          )}
        </div>
      </div>

      {/* Metadata */}
      <div className="grid grid-cols-2 gap-3 mb-6">
        {[
          { label: 'Service', value: issue.service },
          { label: 'Environment', value: issue.environment },
          { label: 'Occurrences', value: issue.occurrence_count.toLocaleString() },
          { label: 'Status', value: issue.status },
          { label: 'First seen', value: new Date(issue.first_seen).toLocaleString() },
          { label: 'Last seen', value: new Date(issue.last_seen).toLocaleString() },
        ].map(({ label, value }) => (
          <div key={label} className="bg-gray-900 border border-gray-800 rounded-lg px-4 py-3">
            <p className="text-xs text-gray-500 mb-1">{label}</p>
            <p className="text-sm text-gray-200 font-mono">{value}</p>
          </div>
        ))}
      </div>

      {/* AI analysis */}
      {issue.ai_explanation && (
        <section className="mb-6">
          <h2 className="text-sm font-semibold text-indigo-400 mb-2">AI Explanation</h2>
          <div className="bg-gray-900 border border-indigo-900 rounded-lg p-4 text-sm text-gray-300 leading-relaxed">
            {issue.ai_explanation}
          </div>
        </section>
      )}

      {issue.likely_cause && (
        <section className="mb-6">
          <h2 className="text-sm font-semibold text-yellow-500 mb-2">Likely Cause</h2>
          <div className="bg-gray-900 border border-yellow-900 rounded-lg p-4 text-sm text-gray-300">
            {issue.likely_cause}
          </div>
        </section>
      )}

      {issue.suggested_checks && issue.suggested_checks.length > 0 && (
        <section className="mb-6">
          <h2 className="text-sm font-semibold text-green-500 mb-2">Suggested Checks</h2>
          <ul className="space-y-1.5">
            {issue.suggested_checks.map((c, i) => (
              <li key={i} className="flex gap-2 text-sm text-gray-300">
                <span className="text-green-600 shrink-0">→</span>
                {c}
              </li>
            ))}
          </ul>
        </section>
      )}

      {/* Stack trace */}
      {issue.stack_trace && issue.stack_trace.length > 0 && (
        <section>
          <h2 className="text-sm font-semibold text-gray-400 mb-2">Stack Trace</h2>
          <pre className="bg-gray-900 border border-gray-800 rounded-lg p-4 text-xs text-gray-400 overflow-x-auto font-mono leading-5">
            {issue.stack_trace.join('\n')}
          </pre>
        </section>
      )}
    </div>
  )
}
