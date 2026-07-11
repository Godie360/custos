import { Link, Outlet, createRootRoute, useRouterState } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import {
  Activity, BarChart2, Bug, Filter, FolderOpen, LayoutDashboard, Search, Wifi, WifiOff,
} from 'lucide-react'
import { useState } from 'react'
import { api } from '#/lib/api'
import '../styles.css'

export const Route = createRootRoute({ component: Shell })

const NAV = [
  {
    group: '',
    items: [
      { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
      { to: '/issues', label: 'Issues', icon: Bug, badge: true },
    ],
  },
  {
    group: 'Monitoring',
    items: [
      { to: '/analytics', label: 'Analytics', icon: BarChart2 },
      { to: '/projects', label: 'Projects', icon: FolderOpen },
    ],
  },
  {
    group: 'Configuration',
    items: [
      { to: '/filters', label: 'Filters', icon: Filter },
    ],
  },
  {
    group: 'System',
    items: [
      { to: '/health', label: 'Health', icon: Activity },
    ],
  },
]

function useOpenCount() {
  const { data } = useQuery({
    queryKey: ['analytics'],
    queryFn: api.analytics.summary,
    refetchInterval: 30_000,
    staleTime: 15_000,
  })
  return data?.open_issues ?? 0
}

function useHealth() {
  const { data, isError } = useQuery({
    queryKey: ['health'],
    queryFn: api.health,
    refetchInterval: 30_000,
    retry: 1,
  })
  return isError ? 'down' : data?.status === 'ok' ? 'ok' : 'checking'
}

function NavItem({
  to,
  label,
  icon: Icon,
  badge,
  openCount,
}: {
  to: string
  label: string
  icon: React.ElementType
  badge?: boolean
  openCount?: number
}) {
  const location = useRouterState({ select: s => s.location.pathname })
  const active = location === to || location.startsWith(to + '/')
  return (
    <Link to={to} className={`nav-item ${active ? 'active' : ''}`}>
      <Icon size={15} strokeWidth={1.8} />
      <span className="flex-1">{label}</span>
      {badge && openCount ? (
        <span
          style={{ background: 'rgba(45,212,191,0.15)', color: '#2DD4BF', fontSize: 11 }}
          className="px-1.5 py-0.5 rounded-full font-semibold"
        >
          {openCount}
        </span>
      ) : null}
    </Link>
  )
}

function Sidebar({ openCount }: { openCount: number }) {
  return (
    <aside
      style={{ background: 'var(--bg-sidebar)', borderRight: '1px solid var(--border)', width: 220 }}
      className="flex flex-col shrink-0 h-screen sticky top-0"
    >
      {/* Logo */}
      <div
        style={{ borderBottom: '1px solid var(--border)' }}
        className="flex items-center gap-3 px-4 py-4"
      >
        <img src="/logo.png" alt="Custos" className="w-7 h-7 object-contain" />
        <span style={{ color: 'var(--text-1)', fontSize: 16, fontWeight: 700, letterSpacing: '-0.3px' }}>
          Custos
        </span>
      </div>

      {/* Nav */}
      <nav className="flex flex-col gap-0.5 p-3 flex-1 overflow-y-auto">
        {NAV.map(({ group, items }) => (
          <div key={group} className="mb-3">
            {group && (
              <p
                style={{ color: 'var(--text-3)', fontSize: 10.5, letterSpacing: '0.08em', fontWeight: 600 }}
                className="uppercase px-2 mb-1.5"
              >
                {group}
              </p>
            )}
            {items.map(item => (
              <NavItem key={item.to} {...item} openCount={openCount} />
            ))}
          </div>
        ))}
      </nav>

      {/* Footer */}
      <div style={{ borderTop: '1px solid var(--border)', padding: '12px 16px' }}>
        <p style={{ color: 'var(--text-3)', fontSize: 11 }}>v0.1.0 · MIT License</p>
      </div>
    </aside>
  )
}

function Header() {
  const [search, setSearch] = useState('')
  const health = useHealth()

  return (
    <header
      style={{ borderBottom: '1px solid var(--border)', background: 'rgba(7,9,10,0.85)' }}
      className="sticky top-0 z-10 flex items-center gap-4 px-6 h-14 backdrop-blur-sm"
    >
      {/* Search */}
      <div className="flex-1 max-w-md relative">
        <Search
          size={13}
          style={{ color: 'var(--text-3)' }}
          className="absolute left-3 top-1/2 -translate-y-1/2"
        />
        <input
          value={search}
          onChange={e => setSearch(e.target.value)}
          placeholder="Search issues, services…"
          style={{ paddingLeft: 32, fontSize: 13, height: 34, width: '100%' }}
        />
      </div>

      <div className="flex items-center gap-4 ml-auto">
        {/* Server health */}
        <div className="flex items-center gap-1.5">
          {health === 'ok' ? (
            <Wifi size={13} style={{ color: '#4ADE80' }} />
          ) : (
            <WifiOff size={13} style={{ color: '#F87171' }} />
          )}
          <span style={{ fontSize: 12, color: health === 'ok' ? '#4ADE80' : '#F87171' }}>
            {health === 'ok' ? 'Connected' : health === 'checking' ? 'Checking…' : 'Offline'}
          </span>
        </div>

        {/* Live indicator */}
        <div className="flex items-center gap-1.5">
          <span className="relative flex h-2 w-2">
            <span
              style={{ background: '#2DD4BF' }}
              className="animate-ping absolute inline-flex h-full w-full rounded-full opacity-60"
            />
            <span style={{ background: '#2DD4BF' }} className="relative inline-flex rounded-full h-2 w-2" />
          </span>
          <span style={{ fontSize: 12, color: 'var(--text-2)' }}>Live</span>
        </div>
      </div>
    </header>
  )
}

function Shell() {
  const openCount = useOpenCount()
  return (
    <div className="flex" style={{ minHeight: '100vh', background: 'var(--bg)' }}>
      <Sidebar openCount={openCount} />
      <div className="flex flex-col flex-1 min-w-0">
        <Header />
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
