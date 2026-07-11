import { useQuery } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { api } from '#/lib/api'

export const Route = createFileRoute('/analytics')({ component: Analytics })

const SEVERITY_COLORS: Record<string, string> = {
  low: '#6b7280',
  medium: '#ca8a04',
  high: '#ea580c',
  critical: '#dc2626',
  error: '#ea580c',
}

function StatCard({ label, value }: { label: string; value: number | string }) {
  return (
    <div className="bg-gray-900 border border-gray-800 rounded-lg px-5 py-4">
      <p className="text-xs text-gray-500 mb-1">{label}</p>
      <p className="text-2xl font-semibold text-white">{value}</p>
    </div>
  )
}

function Analytics() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['analytics'],
    queryFn: api.analytics.summary,
    refetchInterval: 30_000,
  })

  if (isLoading) return <div className="p-6 text-gray-500 text-sm">Loading…</div>
  if (error || !data) return <div className="p-6 text-red-400 text-sm">Failed to load analytics.</div>

  const mttd = data.mean_time_to_detect_seconds
  const mttdStr =
    mttd < 60
      ? `${Math.round(mttd)}s`
      : mttd < 3600
        ? `${Math.round(mttd / 60)}m`
        : `${(mttd / 3600).toFixed(1)}h`

  const severityData = Object.entries(data.issues_by_severity).map(([name, value]) => ({
    name,
    value,
  }))

  const serviceData = Object.entries(data.issues_by_service)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 8)
    .map(([name, count]) => ({ name, count }))

  return (
    <div className="p-6">
      <h1 className="text-xl font-semibold text-white mb-6">Analytics</h1>

      {/* KPI cards */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        <StatCard label="Total issues" value={data.total_issues.toLocaleString()} />
        <StatCard label="Open" value={data.open_issues.toLocaleString()} />
        <StatCard label="Critical" value={data.critical_issues.toLocaleString()} />
        <StatCard label="MTTD" value={mttdStr} />
      </div>

      <div className="grid grid-cols-2 gap-6 mb-6">
        {/* Issues over time */}
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
          <p className="text-sm font-medium text-gray-300 mb-4">Issues over time</p>
          {data.issues_over_time?.length > 0 ? (
            <ResponsiveContainer width="100%" height={180}>
              <LineChart data={data.issues_over_time}>
                <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" />
                <XAxis dataKey="date" tick={{ fill: '#6b7280', fontSize: 11 }} />
                <YAxis tick={{ fill: '#6b7280', fontSize: 11 }} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#111827', border: '1px solid #374151', borderRadius: 6 }}
                  labelStyle={{ color: '#d1d5db' }}
                  itemStyle={{ color: '#818cf8' }}
                />
                <Line type="monotone" dataKey="count" stroke="#818cf8" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-gray-600 text-sm">No time-series data yet.</p>
          )}
        </div>

        {/* Severity breakdown */}
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
          <p className="text-sm font-medium text-gray-300 mb-4">Severity breakdown</p>
          {severityData.length > 0 ? (
            <ResponsiveContainer width="100%" height={180}>
              <PieChart>
                <Pie
                  data={severityData}
                  dataKey="value"
                  nameKey="name"
                  cx="50%"
                  cy="50%"
                  outerRadius={70}
                  label={({ name, percent }) =>
                    `${name} ${(percent * 100).toFixed(0)}%`
                  }
                  labelLine={false}
                >
                  {severityData.map(({ name }) => (
                    <Cell key={name} fill={SEVERITY_COLORS[name] ?? '#6b7280'} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{ backgroundColor: '#111827', border: '1px solid #374151', borderRadius: 6 }}
                  itemStyle={{ color: '#d1d5db' }}
                />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-gray-600 text-sm">No issues yet.</p>
          )}
        </div>
      </div>

      {/* Top failing services */}
      <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
        <p className="text-sm font-medium text-gray-300 mb-4">Top failing services</p>
        {serviceData.length > 0 ? (
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={serviceData} layout="vertical" margin={{ left: 8 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" horizontal={false} />
              <XAxis type="number" tick={{ fill: '#6b7280', fontSize: 11 }} />
              <YAxis dataKey="name" type="category" tick={{ fill: '#9ca3af', fontSize: 11 }} width={120} />
              <Tooltip
                contentStyle={{ backgroundColor: '#111827', border: '1px solid #374151', borderRadius: 6 }}
                labelStyle={{ color: '#d1d5db' }}
                itemStyle={{ color: '#818cf8' }}
              />
              <Bar dataKey="count" fill="#4f46e5" radius={[0, 3, 3, 0]} />
            </BarChart>
          </ResponsiveContainer>
        ) : (
          <p className="text-gray-600 text-sm">No service data yet.</p>
        )}
      </div>
    </div>
  )
}
