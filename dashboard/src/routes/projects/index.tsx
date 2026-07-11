import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { Copy, KeyRound, Plus, Trash2, X } from 'lucide-react'
import { useState } from 'react'
import { api } from '#/lib/api'
import type { Project } from '#/lib/types'

export const Route = createFileRoute('/projects/')({ component: Projects })

type CreatedKey = { projectId: string; key: string; label: string }

function Projects() {
  const qc = useQueryClient()
  const [newName, setNewName] = useState('')
  const [createdKeys, setCreatedKeys] = useState<CreatedKey[]>([])
  const [keyLabels, setKeyLabels] = useState<Record<string, string>>({})
  const [copied, setCopied] = useState<string | null>(null)

  const { data: projects = [], isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: api.projects.list,
  })

  const createProject = useMutation({
    mutationFn: (name: string) => api.projects.create(name),
    onSuccess: () => { void qc.invalidateQueries({ queryKey: ['projects'] }); setNewName('') },
  })

  const createKey = useMutation({
    mutationFn: ({ projectId, label }: { projectId: string; label: string }) =>
      api.projects.createKey(projectId, label),
    onSuccess: (data, { projectId, label }) => {
      setCreatedKeys(prev => [...prev, { projectId, key: data.key, label }])
      setKeyLabels(prev => ({ ...prev, [projectId]: '' }))
    },
  })

  const revokeKey = useMutation({
    mutationFn: ({ projectId, keyId }: { projectId: string; keyId: string }) =>
      api.projects.revokeKey(projectId, keyId),
    onSuccess: (_, { projectId, keyId }) => {
      setCreatedKeys(prev => prev.filter(k => !(k.projectId === projectId && k.key.startsWith(keyId))))
      void qc.invalidateQueries({ queryKey: ['projects'] })
    },
  })

  const copyKey = (key: string) => {
    void navigator.clipboard.writeText(key)
    setCopied(key)
    setTimeout(() => setCopied(null), 2000)
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 style={{ color: 'var(--text-1)', fontSize: 20, fontWeight: 700, marginBottom: 4 }}>Projects</h1>
          <p style={{ color: 'var(--text-2)', fontSize: 13 }}>Manage projects and API keys for SDK authentication.</p>
        </div>
      </div>

      {/* New project form */}
      <div className="card mb-6" style={{ padding: '16px 20px' }}>
        <p style={{ color: 'var(--text-2)', fontSize: 12.5, marginBottom: 10, fontWeight: 500 }}>New Project</p>
        <form
          onSubmit={e => { e.preventDefault(); if (newName.trim()) createProject.mutate(newName.trim()) }}
          className="flex gap-3"
        >
          <input
            value={newName}
            onChange={e => setNewName(e.target.value)}
            placeholder="Project name…"
            style={{ flex: 1, padding: '7px 12px', fontSize: 13 }}
          />
          <button
            type="submit"
            disabled={createProject.isPending || !newName.trim()}
            className="btn-primary flex items-center gap-1.5"
          >
            <Plus size={13} /> Create Project
          </button>
        </form>
      </div>

      {/* Projects list */}
      {isLoading ? (
        <div className="space-y-4">
          {[...Array(2)].map((_, i) => (
            <div key={i} className="card skeleton" style={{ height: 120 }} />
          ))}
        </div>
      ) : projects.length === 0 ? (
        <div className="card" style={{ padding: 48, textAlign: 'center' }}>
          <p style={{ color: 'var(--text-3)', fontSize: 14 }}>No projects yet. Create one above.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {(projects as Project[]).map(p => {
            const projKeys = createdKeys.filter(k => k.projectId === p.id)
            const label = keyLabels[p.id] ?? ''
            return (
              <div key={p.id} className="card" style={{ padding: '20px 24px' }}>
                {/* Project header */}
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 style={{ color: 'var(--text-1)', fontSize: 16, fontWeight: 600, marginBottom: 4 }}>
                      {p.name}
                    </h3>
                    <div className="flex items-center gap-3">
                      <span
                        style={{
                          fontFamily: 'monospace', fontSize: 12,
                          background: 'rgba(45,212,191,0.08)', color: '#2DD4BF',
                          padding: '2px 8px', borderRadius: 5,
                        }}
                      >
                        {p.slug}
                      </span>
                      <span style={{ color: 'var(--text-3)', fontSize: 11 }}>
                        Created {new Date(p.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                  <span
                    style={{
                      fontFamily: 'monospace', fontSize: 10, color: 'var(--text-3)',
                      background: 'rgba(0,0,0,0.3)', padding: '3px 8px', borderRadius: 4,
                    }}
                  >
                    {p.id.slice(0, 8)}
                  </span>
                </div>

                {/* Created keys (in-session) */}
                {projKeys.length > 0 && (
                  <div className="mb-4 space-y-2">
                    {projKeys.map(k => (
                      <div
                        key={k.key}
                        style={{
                          background: 'rgba(45,212,191,0.06)',
                          border: '1px solid rgba(45,212,191,0.2)',
                          borderRadius: 7,
                          padding: '10px 14px',
                          display: 'flex',
                          alignItems: 'center',
                          gap: 12,
                        }}
                      >
                        <div className="flex-1 min-w-0">
                          <p style={{ color: 'var(--text-3)', fontSize: 11, marginBottom: 3 }}>
                            {k.label} · copy now, won't be shown again
                          </p>
                          <p style={{ fontFamily: 'monospace', fontSize: 12, color: '#2DD4BF' }} className="truncate">
                            {k.key}
                          </p>
                        </div>
                        <button
                          onClick={() => copyKey(k.key)}
                          style={{ color: copied === k.key ? '#4ADE80' : '#2DD4BF', shrink: 0 }}
                          className="hover:opacity-70 transition-opacity"
                          title="Copy key"
                        >
                          {copied === k.key ? <span style={{ fontSize: 11 }}>Copied!</span> : <Copy size={14} />}
                        </button>
                        <button
                          onClick={() => setCreatedKeys(prev => prev.filter(x => x.key !== k.key))}
                          style={{ color: 'var(--text-3)' }}
                          className="hover:opacity-70 transition-opacity"
                        >
                          <X size={14} />
                        </button>
                      </div>
                    ))}
                  </div>
                )}

                {/* Add key form */}
                <div style={{ borderTop: '1px solid var(--border)', paddingTop: 14, display: 'flex', gap: 10 }}>
                  <div className="flex items-center gap-2" style={{ color: 'var(--text-3)' }}>
                    <KeyRound size={13} />
                  </div>
                  <input
                    value={label}
                    onChange={e => setKeyLabels(prev => ({ ...prev, [p.id]: e.target.value }))}
                    placeholder="Key label (e.g. production-sdk)…"
                    style={{ flex: 1, padding: '6px 12px', fontSize: 12.5 }}
                    onKeyDown={e => {
                      if (e.key === 'Enter' && label.trim())
                        createKey.mutate({ projectId: p.id, label: label.trim() })
                    }}
                  />
                  <button
                    onClick={() => {
                      if (label.trim()) createKey.mutate({ projectId: p.id, label: label.trim() })
                    }}
                    disabled={createKey.isPending || !label.trim()}
                    className="btn-ghost flex items-center gap-1.5"
                    style={{ fontSize: 12 }}
                  >
                    <Plus size={12} /> Generate Key
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
