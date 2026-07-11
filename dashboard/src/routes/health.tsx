import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { Activity, AlertCircle, CheckCircle2, Clock } from 'lucide-react'
import { api } from '#/lib/api'

export const Route = createFileRoute('/health')({ component: HealthPage })

function HealthPage() {
  const { data, isLoading, error, dataUpdatedAt } = useQuery({
    queryKey: ['health'],
    queryFn: api.health,
    refetchInterval: 10_000,
    retry: 1,
  })

  const ok = !error && data?.status === 'ok'

  return (
    <div className="max-w-xl">
      <div className="flex items-center gap-2 mb-6">
        <Activity size={18} style={{ color: '#2DD4BF' }} />
        <h1 style={{ color: 'var(--text-1)', fontSize: 20, fontWeight: 700 }}>System Health</h1>
      </div>

      {/* Status card */}
      <div
        className="card"
        style={{
          padding: '28px 32px',
          borderColor: ok ? 'rgba(74,222,128,0.25)' : 'rgba(248,113,113,0.25)',
          background: ok ? 'rgba(74,222,128,0.04)' : 'rgba(248,113,113,0.04)',
          display: 'flex',
          alignItems: 'center',
          gap: 20,
          marginBottom: 20,
        }}
      >
        {isLoading ? (
          <div className="skeleton" style={{ width: 48, height: 48, borderRadius: '50%' }} />
        ) : ok ? (
          <CheckCircle2 size={48} style={{ color: '#4ADE80', shrink: 0 }} />
        ) : (
          <AlertCircle size={48} style={{ color: '#F87171', shrink: 0 }} />
        )}
        <div>
          <p style={{ color: ok ? '#4ADE80' : '#F87171', fontSize: 22, fontWeight: 700, marginBottom: 4 }}>
            {isLoading ? 'Checking…' : ok ? 'All Systems Operational' : 'Service Unreachable'}
          </p>
          <p style={{ color: 'var(--text-2)', fontSize: 13 }}>
            {ok
              ? 'Custos Go server is responding normally.'
              : 'Cannot reach the Custos server. Check that it is running on port 8080.'}
          </p>
        </div>
      </div>

      {/* Detail rows */}
      <div className="card overflow-hidden">
        {[
          {
            label: 'API Server',
            detail: `http://localhost:8080`,
            status: ok,
          },
          {
            label: 'Database',
            detail: 'PostgreSQL via health check',
            status: ok,
          },
        ].map(({ label, detail, status }, i, arr) => (
          <div
            key={label}
            style={{
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              padding: '14px 20px',
              borderBottom: i < arr.length - 1 ? '1px solid var(--border)' : undefined,
            }}
          >
            <div>
              <p style={{ color: 'var(--text-1)', fontSize: 13.5, marginBottom: 3 }}>{label}</p>
              <p style={{ color: 'var(--text-3)', fontSize: 11.5 }}>{detail}</p>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <span
                style={{
                  width: 8, height: 8, borderRadius: '50%',
                  background: isLoading ? '#3D5470' : status ? '#4ADE80' : '#F87171',
                  display: 'inline-block',
                  boxShadow: !isLoading && status ? '0 0 8px rgba(74,222,128,0.5)' : undefined,
                }}
              />
              <span
                style={{
                  fontSize: 12,
                  color: isLoading ? 'var(--text-3)' : status ? '#4ADE80' : '#F87171',
                }}
              >
                {isLoading ? 'Checking' : status ? 'Healthy' : 'Down'}
              </span>
            </div>
          </div>
        ))}
      </div>

      {dataUpdatedAt > 0 && (
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 16, color: 'var(--text-3)', fontSize: 11 }}>
          <Clock size={11} />
          Last checked {new Date(dataUpdatedAt).toLocaleTimeString()} · auto-refreshes every 10s
        </div>
      )}
    </div>
  )
}
