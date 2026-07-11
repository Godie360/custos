import { SEVERITY_BG, SEVERITY_COLORS } from '#/lib/utils'

export function SeverityBadge({ severity }: { severity: string }) {
  const color = SEVERITY_COLORS[severity] ?? '#6B7280'
  const bg = SEVERITY_BG[severity] ?? 'rgba(107,114,128,0.12)'
  return (
    <span
      style={{ color, background: bg, border: `1px solid ${color}30` }}
      className="px-2 py-0.5 rounded-md text-[11px] font-semibold uppercase tracking-wide"
    >
      {severity}
    </span>
  )
}

const STATUS_STYLES: Record<string, { color: string; bg: string }> = {
  open: { color: '#2DD4BF', bg: 'rgba(45,212,191,0.1)' },
  resolved: { color: '#4ADE80', bg: 'rgba(74,222,128,0.1)' },
  ignored: { color: '#4B5563', bg: 'rgba(75,85,99,0.1)' },
}

export function StatusBadge({ status }: { status: string }) {
  const s = STATUS_STYLES[status] ?? STATUS_STYLES['open']
  return (
    <span
      style={{ color: s.color, background: s.bg, border: `1px solid ${s.color}30` }}
      className="px-2 py-0.5 rounded-md text-[11px] font-medium"
    >
      {status}
    </span>
  )
}
