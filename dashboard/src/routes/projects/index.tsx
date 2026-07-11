import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { KeyRound, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { api } from '#/lib/api'
import type { Project } from '#/lib/types'

export const Route = createFileRoute('/projects/')({ component: Projects })

function Projects() {
  const qc = useQueryClient()
  const [newName, setNewName] = useState('')
  const [newKey, setNewKey] = useState<string | null>(null)
  const [keyLabels, setKeyLabels] = useState<Record<string, string>>({})

  const { data: projects = [], isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: api.projects.list,
  })

  const createProject = useMutation({
    mutationFn: (name: string) => api.projects.create(name),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['projects'] })
      setNewName('')
    },
  })

  const createKey = useMutation({
    mutationFn: ({ projectId, label }: { projectId: string; label: string }) =>
      api.projects.createKey(projectId, label),
    onSuccess: (data) => {
      setNewKey(data.key)
      void qc.invalidateQueries({ queryKey: ['projects'] })
    },
  })

  return (
    <div className="p-6 max-w-2xl">
      <h1 className="text-xl font-semibold text-white mb-6">Projects</h1>

      {/* New project */}
      <form
        onSubmit={e => {
          e.preventDefault()
          if (newName.trim()) createProject.mutate(newName.trim())
        }}
        className="flex gap-2 mb-7"
      >
        <input
          value={newName}
          onChange={e => setNewName(e.target.value)}
          placeholder="New project name…"
          className="flex-1 px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-indigo-500"
        />
        <button
          type="submit"
          disabled={createProject.isPending || !newName.trim()}
          className="flex items-center gap-1 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-sm rounded transition-colors disabled:opacity-50"
        >
          <Plus size={13} /> Create
        </button>
      </form>

      {newKey && (
        <div className="mb-5 p-3 bg-green-950 border border-green-800 rounded-lg">
          <p className="text-xs text-green-400 mb-1 font-medium">API key created — copy it now, it won't be shown again</p>
          <code className="text-sm text-green-200 font-mono break-all">{newKey}</code>
          <button
            onClick={() => setNewKey(null)}
            className="mt-2 text-xs text-green-600 hover:text-green-400"
          >
            Dismiss
          </button>
        </div>
      )}

      {isLoading && <p className="text-gray-500 text-sm">Loading…</p>}

      <div className="space-y-4">
        {projects.map((p: Project) => (
          <ProjectCard
            key={p.id}
            project={p}
            keyLabel={keyLabels[p.id] ?? ''}
            onKeyLabelChange={v => setKeyLabels(prev => ({ ...prev, [p.id]: v }))}
            onCreateKey={label => createKey.mutate({ projectId: p.id, label })}
            creatingKey={createKey.isPending}
          />
        ))}
      </div>
    </div>
  )
}

function ProjectCard({
  project,
  keyLabel,
  onKeyLabelChange,
  onCreateKey,
  creatingKey,
}: {
  project: Project
  keyLabel: string
  onKeyLabelChange: (v: string) => void
  onCreateKey: (label: string) => void
  creatingKey: boolean
}) {
  const qc = useQueryClient()

  const { data: keys = [] } = useQuery({
    queryKey: ['apikeys', project.id],
    queryFn: () =>
      // The projects list endpoint doesn't return keys; we fetch them via a
      // dedicated endpoint when available. For now surface what the project
      // object might carry, falling back to an empty array.
      Promise.resolve((project as unknown as { api_keys?: unknown[] }).api_keys ?? []),
  })

  const revokeKey = useMutation({
    mutationFn: (keyId: string) => api.projects.revokeKey(project.id, keyId),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['apikeys', project.id] }),
  })

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-3">
        <div>
          <p className="text-sm font-medium text-white">{project.name}</p>
          <p className="text-xs text-gray-500 font-mono">{project.slug}</p>
        </div>
        <p className="text-xs text-gray-600">{new Date(project.created_at).toLocaleDateString()}</p>
      </div>

      {keys.length > 0 && (
        <div className="mb-3 space-y-1">
          {(keys as Array<{ id: string; label: string; key_prefix: string }>).map(k => (
            <div key={k.id} className="flex items-center justify-between px-3 py-1.5 bg-gray-800 rounded text-xs">
              <span className="text-gray-300 font-mono">{k.key_prefix}… <span className="text-gray-500">{k.label}</span></span>
              <button
                onClick={() => revokeKey.mutate(k.id)}
                disabled={revokeKey.isPending}
                className="text-red-500 hover:text-red-400 disabled:opacity-40"
              >
                <Trash2 size={11} />
              </button>
            </div>
          ))}
        </div>
      )}

      <form
        onSubmit={e => {
          e.preventDefault()
          if (keyLabel.trim()) onCreateKey(keyLabel.trim())
        }}
        className="flex gap-2"
      >
        <input
          value={keyLabel}
          onChange={e => onKeyLabelChange(e.target.value)}
          placeholder="Key label…"
          className="flex-1 px-2 py-1 bg-gray-800 border border-gray-700 rounded text-xs text-gray-200 placeholder-gray-600 focus:outline-none focus:border-indigo-500"
        />
        <button
          type="submit"
          disabled={creatingKey || !keyLabel.trim()}
          className="flex items-center gap-1 px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 text-xs rounded transition-colors disabled:opacity-50"
        >
          <KeyRound size={11} /> New key
        </button>
      </form>
    </div>
  )
}
