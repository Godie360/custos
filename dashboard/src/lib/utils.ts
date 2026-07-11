export function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const s = Math.floor(diff / 1000)
  if (s < 60) return `${s}s ago`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
}

export function fmtDuration(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)}s`
  if (seconds < 3600) return `${Math.round(seconds / 60)}m`
  if (seconds < 86400) return `${(seconds / 3600).toFixed(1)}h`
  return `${(seconds / 86400).toFixed(1)}d`
}

export function fmtNum(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`
  return String(n)
}

export const SEVERITY_COLORS: Record<string, string> = {
  critical: '#F87171',
  high: '#FB923C',
  medium: '#FCD34D',
  low: '#6B7280',
  error: '#FB923C',
}

export const SEVERITY_BG: Record<string, string> = {
  critical: 'rgba(248,113,113,0.12)',
  high: 'rgba(251,146,60,0.12)',
  medium: 'rgba(252,211,77,0.12)',
  low: 'rgba(107,114,128,0.12)',
  error: 'rgba(251,146,60,0.12)',
}
