import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { ChevronDown, Filter, Plus, ShieldOff, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { api } from '#/lib/api'
import type { FilterRule, Project } from '#/lib/types'

export const Route = createFileRoute('/filters')({ component: Filters })

const FIELDS = ['error_type', 'message', 'service', 'environment'] as const
const OPERATORS = ['equals', 'contains', 'starts_with'] as const

const OPERATOR_LABEL: Record<string, string> = {
  equals: '=',
  contains: '∋',
  starts_with: '^',
}

function Filters() {
  const qc = useQueryClient()
  const [selectedProject, setSelectedProject] = useState('')
  const [field, setField] = useState<FilterRule['field']>('message')
  const [operator, setOperator] = useState<FilterRule['operator']>('contains')
  const [value, setValue] = useState('')

  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: api.projects.list,
  })

  const projectId = selectedProject || (projects[0] as Project | undefined)?.id || ''
  const projectName = (projects as Project[]).find(p => p.id === projectId)?.name ?? ''

  const { data: rules = [], isLoading } = useQuery({
    queryKey: ['filters', projectId],
    queryFn: () => api.filters.list(projectId),
    enabled: !!projectId,
  })

  const createRule = useMutation({
    mutationFn: () => api.filters.create(projectId, { field, operator, value }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['filters', projectId] })
      setValue('')
    },
  })

  const deleteRule = useMutation({
    mutationFn: (filterId: string) => api.filters.delete(projectId, filterId),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['filters', projectId] }),
  })

  return (
    <div>
      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 style={{ color: 'var(--text-1)', fontSize: 20, fontWeight: 700, marginBottom: 4 }}>Filter Rules</h1>
          <p style={{ color: 'var(--text-2)', fontSize: 13, maxWidth: 520 }}>
            Events matching a rule are silently dropped at ingestion — no storage, no AI analysis, no notification.
            Useful for suppressing noisy health-check pings or expected errors.
          </p>
        </div>
        <div
          style={{
            display: 'flex', alignItems: 'center', gap: 6,
            background: 'rgba(248,113,113,0.08)', border: '1px solid rgba(248,113,113,0.2)',
            padding: '6px 12px', borderRadius: 8,
          }}
        >
          <ShieldOff size={13} style={{ color: '#F87171' }} />
          <span style={{ fontSize: 12, color: '#F87171' }}>
            {(rules as FilterRule[]).length} active rule{rules.length !== 1 ? 's' : ''}
          </span>
        </div>
      </div>

      {/* Project selector */}
      {projects.length > 0 && (
        <div className="card mb-5" style={{ padding: '14px 18px' }}>
          <div className="flex items-center gap-3">
            <Filter size={13} style={{ color: 'var(--text-3)' }} />
            <span style={{ color: 'var(--text-2)', fontSize: 12.5, fontWeight: 500 }}>Project</span>
            <div className="relative">
              <select
                value={selectedProject || projectId}
                onChange={e => setSelectedProject(e.target.value)}
                style={{ padding: '5px 28px 5px 12px', fontSize: 13, appearance: 'none', paddingRight: 28 }}
              >
                {(projects as Project[]).map(p => (
                  <option key={p.id} value={p.id}>{p.name}</option>
                ))}
              </select>
              <ChevronDown
                size={12}
                style={{ position: 'absolute', right: 8, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-3)', pointerEvents: 'none' }}
              />
            </div>
          </div>
        </div>
      )}

      {/* Add rule */}
      {projectId && (
        <div className="card mb-5" style={{ padding: '16px 18px' }}>
          <p style={{ color: 'var(--text-2)', fontSize: 12.5, fontWeight: 500, marginBottom: 12 }}>
            Add rule for <span style={{ color: '#2DD4BF' }}>{projectName}</span>
          </p>
          <form
            onSubmit={e => { e.preventDefault(); if (value.trim()) createRule.mutate() }}
            className="flex gap-3 items-end flex-wrap"
          >
            <div>
              <p style={{ color: 'var(--text-3)', fontSize: 10.5, marginBottom: 5, textTransform: 'uppercase', letterSpacing: '0.07em' }}>Field</p>
              <select
                value={field}
                onChange={e => setField(e.target.value as FilterRule['field'])}
                style={{ padding: '7px 12px', fontSize: 13 }}
              >
                {FIELDS.map(f => <option key={f} value={f}>{f}</option>)}
              </select>
            </div>
            <div>
              <p style={{ color: 'var(--text-3)', fontSize: 10.5, marginBottom: 5, textTransform: 'uppercase', letterSpacing: '0.07em' }}>Operator</p>
              <select
                value={operator}
                onChange={e => setOperator(e.target.value as FilterRule['operator'])}
                style={{ padding: '7px 12px', fontSize: 13 }}
              >
                {OPERATORS.map(o => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
            <div style={{ flex: 1, minWidth: 180 }}>
              <p style={{ color: 'var(--text-3)', fontSize: 10.5, marginBottom: 5, textTransform: 'uppercase', letterSpacing: '0.07em' }}>Value</p>
              <input
                value={value}
                onChange={e => setValue(e.target.value)}
                placeholder="e.g. health check…"
                style={{ width: '100%', padding: '7px 12px', fontSize: 13 }}
              />
            </div>
            <button
              type="submit"
              disabled={createRule.isPending || !value.trim()}
              className="btn-primary flex items-center gap-1.5"
            >
              <Plus size={13} /> Add Rule
            </button>
          </form>

          {/* Preview */}
          {value.trim() && (
            <div style={{ marginTop: 12, padding: '8px 12px', background: 'rgba(0,0,0,0.2)', borderRadius: 6, fontSize: 12, color: 'var(--text-2)' }}>
              Will drop events where <code style={{ color: '#2DD4BF', background: 'rgba(45,212,191,0.08)', padding: '1px 4px', borderRadius: 3 }}>{field}</code>{' '}
              {operator === 'equals' ? 'equals' : operator === 'contains' ? 'contains' : 'starts with'}{' '}
              <code style={{ color: '#FCD34D', background: 'rgba(252,211,77,0.08)', padding: '1px 4px', borderRadius: 3 }}>"{value}"</code>
            </div>
          )}
        </div>
      )}

      {/* Rules list */}
      {!projectId ? (
        <div className="card" style={{ padding: 48, textAlign: 'center' }}>
          <p style={{ color: 'var(--text-3)', fontSize: 14 }}>Create a project first to add filter rules.</p>
        </div>
      ) : isLoading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="card skeleton" style={{ height: 52 }} />
          ))}
        </div>
      ) : (rules as FilterRule[]).length === 0 ? (
        <div className="card" style={{ padding: '40px', textAlign: 'center' }}>
          <Filter size={28} style={{ color: 'var(--text-3)', margin: '0 auto 10px' }} />
          <p style={{ color: 'var(--text-3)', fontSize: 13 }}>No filter rules yet. Add one above to start suppressing noisy events.</p>
        </div>
      ) : (
        <div className="card overflow-hidden">
          <div
            style={{
              display: 'grid', gridTemplateColumns: '120px 110px 1fr 36px',
              padding: '9px 16px', borderBottom: '1px solid var(--border)',
              background: 'rgba(0,0,0,0.2)', gap: 12,
            }}
          >
            {['Field', 'Operator', 'Value', ''].map(h => (
              <span key={h} style={{ color: 'var(--text-3)', fontSize: 10.5, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.07em' }}>
                {h}
              </span>
            ))}
          </div>

          {(rules as FilterRule[]).map((rule, i) => (
            <div
              key={rule.id}
              style={{
                display: 'grid', gridTemplateColumns: '120px 110px 1fr 36px',
                padding: '13px 16px',
                borderBottom: i < rules.length - 1 ? '1px solid var(--border)' : undefined,
                alignItems: 'center', gap: 12,
              }}
              className="hover:bg-[#0D1510] transition-colors group"
            >
              <span
                style={{
                  fontFamily: 'monospace', fontSize: 12, color: '#2DD4BF',
                  background: 'rgba(45,212,191,0.08)', padding: '3px 8px', borderRadius: 5, display: 'inline-block',
                }}
              >
                {rule.field}
              </span>
              <div className="flex items-center gap-2">
                <span
                  style={{
                    fontSize: 15, color: '#FCD34D', fontWeight: 700,
                    background: 'rgba(252,211,77,0.08)', padding: '2px 8px', borderRadius: 5, fontFamily: 'monospace',
                  }}
                >
                  {OPERATOR_LABEL[rule.operator]}
                </span>
                <span style={{ color: 'var(--text-3)', fontSize: 11 }}>{rule.operator}</span>
              </div>
              <span style={{ color: 'var(--text-1)', fontSize: 13, fontFamily: 'monospace' }}>"{rule.value}"</span>
              <button
                onClick={() => deleteRule.mutate(rule.id)}
                disabled={deleteRule.isPending}
                style={{ color: 'var(--text-3)', opacity: 0 }}
                className="group-hover:!opacity-100 hover:!text-red-400 transition-all disabled:opacity-40"
              >
                <Trash2 size={14} />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
