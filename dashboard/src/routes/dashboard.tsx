import { useQuery } from '@tanstack/react-query'
import { createFileRoute, Link } from '@tanstack/react-router'
import {
  Area, AreaChart, Bar, BarChart, CartesianGrid, Cell, Pie, PieChart,
  ResponsiveContainer, Tooltip, XAxis, YAxis,
} from 'recharts'
import { AlertTriangle, ArrowUpRight, TrendingDown, TrendingUp } from 'lucide-react'
import { api } from '#/lib/api'
import { SeverityBadge } from '#/components/Badge'
import { SparkArea } from '#/components/SparkArea'
import { fmtDuration, fmtNum, SEVERITY_COLORS, timeAgo } from '#/lib/utils'

export const Route = createFileRoute('/dashboard')({ component: Dashboard })

const CHART_TOOLTIP_STYLE = {
  contentStyle: { background: '#0D1117', border: '1px solid #1A2535', borderRadius: 8, fontSize: 12 },
  labelStyle: { color: '#7A9DB8' },
  itemStyle: { color: '#2DD4BF' },
}

function StatCard({
  label,
  value,
  sub,
  trend,
  sparkData,
  sparkColor = '#2DD4BF',
}: {
  label: string
  value: string | number
  sub?: string
  trend?: { pct: number; up: boolean }
  sparkData?: Array<{ value: number }>
  sparkColor?: string
}) {
  return (
    <div className="card flex flex-col" style={{ padding: '18px 20px 12px' }}>
      <div className="flex items-start justify-between mb-3">
        <div>
          <p style={{ color: 'var(--text-2)', fontSize: 12, marginBottom: 6 }}>{label}</p>
          <p style={{ color: 'var(--text-1)', fontSize: 28, fontWeight: 700, lineHeight: 1, letterSpacing: '-0.5px' }}>
            {typeof value === 'number' ? fmtNum(value) : value}
          </p>
        </div>
        {trend && (
          <div
            style={{
              display: 'flex', alignItems: 'center', gap: 4, fontSize: 12, fontWeight: 600,
              color: trend.up ? '#4ADE80' : '#F87171',
              background: trend.up ? 'rgba(74,222,128,0.1)' : 'rgba(248,113,113,0.1)',
              padding: '3px 8px', borderRadius: 20,
            }}
          >
            {trend.up ? <TrendingUp size={11} /> : <TrendingDown size={11} />}
            {trend.pct.toFixed(1)}%
          </div>
        )}
      </div>
      {sub && <p style={{ color: 'var(--text-3)', fontSize: 11, marginBottom: 8 }}>{sub}</p>}
      {sparkData && <SparkArea data={sparkData} color={sparkColor} height={44} />}
    </div>
  )
}

function SectionHeader({ title, action }: { title: string; action?: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between mb-4">
      <h2 style={{ color: 'var(--text-1)', fontSize: 14, fontWeight: 600 }}>{title}</h2>
      {action}
    </div>
  )
}

function Dashboard() {
  const { data: analytics, isLoading: loadingAnalytics } = useQuery({
    queryKey: ['analytics'],
    queryFn: api.analytics.summary,
    refetchInterval: 30_000,
  })

  const { data: recentIssues = [], isLoading: loadingIssues } = useQuery({
    queryKey: ['issues', { limit: 6 }],
    queryFn: () => api.issues.list({ limit: 6 }),
    refetchInterval: 15_000,
  })

  const sparkData = analytics?.issues_over_time?.map(d => ({ value: d.count })) ?? []

  const serviceData = Object.entries(analytics?.issues_by_service ?? {})
    .sort((a, b) => b[1] - a[1])
    .slice(0, 6)
    .map(([name, count]) => ({ name, count }))

  const maxService = serviceData[0]?.count ?? 1

  const severityData = Object.entries(analytics?.issues_by_severity ?? {}).map(([name, value]) => ({
    name, value,
  }))

  if (loadingAnalytics) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="card" style={{ height: 120 }}>
              <div className="skeleton" style={{ height: '100%', borderRadius: 10 }} />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-5">
      {/* KPI Cards */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          label="Total Issues"
          value={analytics?.total_issues ?? 0}
          sparkData={sparkData}
          sparkColor="#2DD4BF"
        />
        <StatCard
          label="Open Issues"
          value={analytics?.open_issues ?? 0}
          sub="Requires attention"
          sparkData={sparkData.slice(-10)}
          sparkColor="#5EEAD4"
        />
        <StatCard
          label="Critical"
          value={analytics?.critical_issues ?? 0}
          sub="High-priority errors"
          sparkData={sparkData.slice(-10).map(d => ({ value: Math.round(d.value * 0.3) }))}
          sparkColor="#F87171"
        />
        <StatCard
          label="MTTD"
          value={fmtDuration(analytics?.mean_time_to_detect_seconds ?? 0)}
          sub="Mean time to detect"
          sparkData={sparkData.slice(-10).map(d => ({ value: d.value }))}
          sparkColor="#FCD34D"
        />
      </div>

      {/* Issues over time — big chart */}
      <div className="card" style={{ padding: '20px 24px' }}>
        <SectionHeader
          title="Issues Over Time"
          action={
            <span style={{ fontSize: 11, color: 'var(--text-3)' }}>
              Last {analytics?.issues_over_time?.length ?? 0} days
            </span>
          }
        />
        {sparkData.length > 0 ? (
          <ResponsiveContainer width="100%" height={220}>
            <AreaChart data={analytics?.issues_over_time ?? []} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
              <defs>
                <linearGradient id="tealGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#2DD4BF" stopOpacity={0.25} />
                  <stop offset="95%" stopColor="#2DD4BF" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="4 4" stroke="#1A2535" vertical={false} />
              <XAxis
                dataKey="date"
                tick={{ fill: '#3D5470', fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                tickFormatter={v => v?.slice(5) ?? v}
              />
              <YAxis tick={{ fill: '#3D5470', fontSize: 11 }} tickLine={false} axisLine={false} />
              <Tooltip {...CHART_TOOLTIP_STYLE} />
              <Area
                type="monotone"
                dataKey="count"
                name="Issues"
                stroke="#2DD4BF"
                strokeWidth={2}
                fill="url(#tealGrad)"
                dot={false}
                activeDot={{ r: 4, fill: '#2DD4BF', stroke: '#07090A', strokeWidth: 2 }}
              />
            </AreaChart>
          </ResponsiveContainer>
        ) : (
          <div style={{ height: 220, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <p style={{ color: 'var(--text-3)', fontSize: 13 }}>No time-series data yet — errors will appear here as they arrive.</p>
          </div>
        )}
      </div>

      {/* Bottom row */}
      <div className="grid grid-cols-3 gap-4">
        {/* Top services */}
        <div className="card" style={{ padding: '20px' }}>
          <SectionHeader title="Top Failing Services" />
          {serviceData.length > 0 ? (
            <div className="space-y-3">
              {serviceData.map(({ name, count }) => (
                <div key={name}>
                  <div className="flex justify-between mb-1">
                    <span style={{ color: 'var(--text-1)', fontSize: 12.5 }}>{name}</span>
                    <span style={{ color: 'var(--text-2)', fontSize: 12 }}>{count}</span>
                  </div>
                  <div style={{ height: 4, background: 'var(--border)', borderRadius: 4 }}>
                    <div
                      style={{
                        height: '100%',
                        borderRadius: 4,
                        width: `${(count / maxService) * 100}%`,
                        background: 'linear-gradient(90deg, #2DD4BF, #0D9488)',
                        transition: 'width 0.4s',
                      }}
                    />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p style={{ color: 'var(--text-3)', fontSize: 12 }}>No service data yet.</p>
          )}
        </div>

        {/* Severity breakdown */}
        <div className="card" style={{ padding: '20px' }}>
          <SectionHeader title="Severity Breakdown" />
          {severityData.length > 0 ? (
            <>
              <ResponsiveContainer width="100%" height={130}>
                <PieChart>
                  <Pie
                    data={severityData}
                    dataKey="value"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    innerRadius={38}
                    outerRadius={58}
                    strokeWidth={0}
                  >
                    {severityData.map(({ name }) => (
                      <Cell key={name} fill={SEVERITY_COLORS[name] ?? '#6B7280'} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{ background: '#0D1117', border: '1px solid #1A2535', borderRadius: 8, fontSize: 12 }}
                    itemStyle={{ color: '#E8F4FF' }}
                  />
                </PieChart>
              </ResponsiveContainer>
              <div className="flex flex-wrap gap-2 mt-2">
                {severityData.map(({ name, value }) => (
                  <div key={name} className="flex items-center gap-1.5">
                    <span
                      style={{ width: 8, height: 8, borderRadius: '50%', background: SEVERITY_COLORS[name] ?? '#6B7280', display: 'inline-block' }}
                    />
                    <span style={{ color: 'var(--text-2)', fontSize: 11 }}>{name} ({value})</span>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <p style={{ color: 'var(--text-3)', fontSize: 12 }}>No issues yet.</p>
          )}
        </div>

        {/* Recent issues */}
        <div className="card" style={{ padding: '20px' }}>
          <SectionHeader
            title="Recent Issues"
            action={
              <Link to="/issues" style={{ color: 'var(--teal)', fontSize: 12, display: 'flex', alignItems: 'center', gap: 4 }}>
                View all <ArrowUpRight size={12} />
              </Link>
            }
          />
          {loadingIssues ? (
            <div className="space-y-2">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="skeleton" style={{ height: 44 }} />
              ))}
            </div>
          ) : recentIssues.length === 0 ? (
            <div style={{ textAlign: 'center', paddingTop: 32 }}>
              <AlertTriangle size={24} style={{ color: 'var(--text-3)', margin: '0 auto 8px' }} />
              <p style={{ color: 'var(--text-3)', fontSize: 12 }}>No issues yet.</p>
            </div>
          ) : (
            <div className="space-y-2">
              {recentIssues.map(issue => (
                <Link
                  key={issue.id}
                  to="/issues/$id"
                  params={{ id: issue.id }}
                  style={{ borderBottom: '1px solid var(--border)', paddingBottom: 8, display: 'block' }}
                  className="hover:opacity-80 transition-opacity"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p style={{ color: 'var(--text-1)', fontSize: 12.5, marginBottom: 2 }} className="truncate">
                        {issue.error_type}
                      </p>
                      <p style={{ color: 'var(--text-3)', fontSize: 11 }} className="truncate">{issue.service}</p>
                    </div>
                    <div className="shrink-0 flex flex-col items-end gap-1">
                      <SeverityBadge severity={issue.severity} />
                      <span style={{ color: 'var(--text-3)', fontSize: 10 }}>{timeAgo(issue.last_seen)}</span>
                    </div>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Services bar chart */}
      {serviceData.length > 0 && (
        <div className="card" style={{ padding: '20px 24px' }}>
          <SectionHeader title="Issue Volume by Service" />
          <ResponsiveContainer width="100%" height={180}>
            <BarChart data={serviceData} margin={{ left: -10, bottom: 0 }}>
              <CartesianGrid strokeDasharray="4 4" stroke="#1A2535" horizontal={true} vertical={false} />
              <XAxis dataKey="name" tick={{ fill: '#3D5470', fontSize: 11 }} tickLine={false} axisLine={false} />
              <YAxis tick={{ fill: '#3D5470', fontSize: 11 }} tickLine={false} axisLine={false} />
              <Tooltip {...CHART_TOOLTIP_STYLE} cursor={{ fill: 'rgba(45,212,191,0.05)' }} />
              <Bar dataKey="count" name="Issues" radius={[4, 4, 0, 0]}>
                {serviceData.map((_, i) => (
                  <Cell
                    key={i}
                    fill={i === 0 ? '#2DD4BF' : i === 1 ? '#0D9488' : '#1A3A35'}
                  />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  )
}
