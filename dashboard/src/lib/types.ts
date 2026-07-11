export interface Issue {
  id: string
  fingerprint: string
  project_id: string
  service: string
  environment: string
  error_type: string
  message: string
  stack_trace: string[]
  severity: 'low' | 'medium' | 'high' | 'critical' | 'error'
  status: 'open' | 'resolved' | 'ignored'
  occurrence_count: number
  first_seen: string
  last_seen: string
  ai_explanation?: string
  likely_cause?: string
  suggested_checks?: string[]
}

export interface Project {
  id: string
  name: string
  slug: string
  created_at: string
}

export interface APIKey {
  id: string
  project_id: string
  label: string
  key_prefix: string
  created_at: string
  last_used_at?: string
}

export interface AnalyticsSummary {
  total_issues: number
  open_issues: number
  critical_issues: number
  issues_by_severity: Record<string, number>
  issues_by_service: Record<string, number>
  issues_over_time: Array<{ date: string; count: number }>
  mean_time_to_detect_seconds: number
}

export interface FilterRule {
  id: string
  project_id: string
  field: 'error_type' | 'message' | 'service' | 'environment'
  operator: 'equals' | 'contains' | 'starts_with'
  value: string
  created_at: string
}

export interface ListIssuesParams {
  service?: string
  environment?: string
  severity?: string
  limit?: number
  offset?: number
}
