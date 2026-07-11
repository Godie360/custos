import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { api } from '#/lib/api'
import type { FilterRule, Project } from '#/lib/types'

export const Route = createFileRoute('/filters')({ component: Filters })

const FIELDS = ['error_type', 'message', 'service', 'environment'] as const
const OPERATORS = ['equals', 'contains', 'starts_with'] as const

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
    <div className="p-6 max-w-2xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-white">Filter Rules</h1>
        <p className="text-xs text-gray-500 max-w-xs text-right">
          Matching events are silently dropped before storage and AI analysis.
        </p>
      </div>

      {/* Project picker */}
      {projects.length > 1 && (
        <select
          value={selectedProject}
          onChange={e => setSelectedProject(e.target.value)}
          className="mb-5 px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 focus:outline-none focus:border-indigo-500"
        >
          {(projects as Project[]).map(p => (
            <option key={p.id} value={p.id}>{p.name}</option>
          ))}
        </select>
      )}

      {/* New rule form */}
      <form
        onSubmit={e => {
          e.preventDefault()
          if (value.trim() && projectId) createRule.mutate()
        }}
        className="flex gap-2 mb-6 flex-wrap"
      >
        <select
          value={field}
          onChange={e => setField(e.target.value as FilterRule['field'])}
          className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 focus:outline-none focus:border-indigo-500"
        >
          {FIELDS.map(f => <option key={f} value={f}>{f}</option>)}
        </select>
        <select
          value={operator}
          onChange={e => setOperator(e.target.value as FilterRule['operator'])}
          className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 focus:outline-none focus:border-indigo-500"
        >
          {OPERATORS.map(o => <option key={o} value={o}>{o}</option>)}
        </select>
        <input
          value={value}
          onChange={e => setValue(e.target.value)}
          placeholder="Value…"
          className="flex-1 min-w-32 px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-indigo-500"
        />
        <button
          type="submit"
          disabled={createRule.isPending || !value.trim() || !projectId}
          className="flex items-center gap-1 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-sm rounded transition-colors disabled:opacity-50"
        >
          <Plus size={13} /> Add rule
        </button>
      </form>

      {!projectId && (
        <p className="text-gray-500 text-sm">Create a project first to add filter rules.</p>
      )}

      {isLoading && <p className="text-gray-500 text-sm">Loading…</p>}

      {!isLoading && rules.length === 0 && projectId && (
        <p className="text-gray-600 text-sm">No filter rules yet.</p>
      )}

      <div className="space-y-2">
        {(rules as FilterRule[]).map(rule => (
          <div
            key={rule.id}
            className="flex items-center justify-between bg-gray-900 border border-gray-800 rounded-lg px-4 py-3"
          >
            <div className="flex items-center gap-2 text-sm">
              <span className="px-2 py-0.5 bg-gray-800 rounded text-gray-300 font-mono text-xs">{rule.field}</span>
              <span className="text-gray-600">{rule.operator}</span>
              <span className="text-gray-200 font-mono">"{rule.value}"</span>
            </div>
            <button
              onClick={() => deleteRule.mutate(rule.id)}
              disabled={deleteRule.isPending}
              className="text-red-600 hover:text-red-400 disabled:opacity-40 ml-3"
            >
              <Trash2 size={14} />
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
