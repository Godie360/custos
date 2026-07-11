import { Link, Outlet, createRootRoute } from '@tanstack/react-router'
import {
  AlertTriangle,
  BarChart2,
  Filter,
  FolderKey,
  ShieldAlert,
} from 'lucide-react'
import '../styles.css'

export const Route = createRootRoute({
  component: RootComponent,
})

const nav = [
  { to: '/issues', label: 'Issues', icon: AlertTriangle },
  { to: '/analytics', label: 'Analytics', icon: BarChart2 },
  { to: '/projects', label: 'Projects', icon: FolderKey },
  { to: '/filters', label: 'Filters', icon: Filter },
]

function RootComponent() {
  return (
    <div className="flex min-h-screen bg-gray-950 text-gray-100">
      <aside className="w-56 shrink-0 border-r border-gray-800 flex flex-col">
        <div className="flex items-center gap-2 px-5 py-4 border-b border-gray-800">
          <ShieldAlert className="text-indigo-400" size={20} />
          <span className="font-semibold text-white text-sm tracking-wide">Custos</span>
        </div>
        <nav className="flex flex-col gap-1 p-3 flex-1">
          {nav.map(({ to, label, icon: Icon }) => (
            <Link
              key={to}
              to={to}
              className="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-gray-400 hover:bg-gray-800 hover:text-white transition-colors"
              activeProps={{ className: 'bg-gray-800 text-white' }}
            >
              <Icon size={15} />
              {label}
            </Link>
          ))}
        </nav>
      </aside>
      <main className="flex-1 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
