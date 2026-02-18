export interface Config {
  timezone: string;
  sync: SyncConfig;
  notifications: Notifications;
  webhook: WebhookConfig;
  calendar_sync: CalendarSyncConfig;
  property_mapping: PropertyMapping;
  content_rules: ContentRules;
  snooze_until: string;
}

export interface SyncConfig {
  check_interval: number;
}

export interface Notifications {
  advance: AdvanceNotification[];
  periodic: PeriodicNotification[];
  manual: string;
}

export interface AdvanceNotification {
  enabled: boolean;
  minutes_before: number;
  message: string;
  conditions: AdvanceConditions;
}

export interface AdvanceConditions {
  days_of_week: number[];
  property_filters: PropertyFilter[];
}

export interface PropertyFilter {
  property: string;
  operator: string;
  value: string;
}

export interface PeriodicNotification {
  enabled: boolean;
  days_of_week: number[];
  time: string;
  days_ahead: number;
  message: string;
}

export interface WebhookConfig {
  is_test: boolean;
  notification: WebhookTarget;
  internal_notification: WebhookTarget;
}

export interface WebhookTarget {
  content_type: string;
  payload_template: string;
}

export interface CalendarSyncConfig {
  enabled: boolean;
  interval_hours: number;
  lookahead_days: number;
}

export interface PropertyMapping {
  title: string;
  date: string;
  location: string;
  attendees: string;
  attendees_enabled: boolean;
  custom: CustomMapping[];
}

export interface CustomMapping {
  variable: string;
  property: string;
}

export interface ContentRules {
  start_heading: string;
  include_start_heading: boolean;
  stop_at_next_heading: boolean;
  stop_at_delimiter: boolean;
  delimiter_text: string;
}

export interface DashboardData {
  today_count: number;
  next_sync: string;
  next_sync_in: string;
  last_sync: string;
  last_sync_count: number;
  last_sync_error?: string;
  snooze_active: boolean;
  snooze_until?: string;
}

export interface UpcomingEvent {
  notion_page_id: string;
  title: string;
  start_date: string;
  start_time: string;
  end_date?: string;
  end_time?: string;
  is_all_day: boolean;
  location?: string;
  url?: string;
  calendar_state: 'disabled' | 'needs_sync' | 'synced' | 'error';
}

export interface HistoryItem {
  id: number;
  type: string;
  status: string;
  message: string;
  error?: string;
  sent_at: string;
}

export interface ApiError {
  error: string;
  details?: Record<string, string>;
}

export interface NotificationPreviewRequest {
  template: string;
  from_date?: string;
  to_date?: string;
  minutes_before?: number;
}

export interface ManualNotificationRequest {
  template: string;
  from_date: string;
  to_date: string;
}

export function buildManualNotificationRequest(
  template: string,
  fromDate: string,
  toDate: string,
): ManualNotificationRequest {
  return {
    template,
    from_date: fromDate,
    to_date: toDate,
  };
}

export function buildPreviewNotificationRequest(
  template: string,
  options: {
    minutesBefore?: number;
    daysAhead?: number;
    fromDate?: string;
    toDate?: string;
    now?: Date;
  } = {},
): NotificationPreviewRequest {
  if (options.minutesBefore && options.minutesBefore > 0) {
    return {
      template,
      minutes_before: options.minutesBefore,
    };
  }
  if (options.fromDate && options.toDate) {
    return buildManualNotificationRequest(template, options.fromDate, options.toDate);
  }
  if (options.daysAhead && options.daysAhead > 0) {
    const now = options.now ?? new Date();
    const to = new Date(now.getTime() + options.daysAhead * 24 * 60 * 60 * 1000);
    return buildManualNotificationRequest(
      template,
      now.toISOString().split('T')[0],
      to.toISOString().split('T')[0],
    );
  }
  return { template };
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw errorData as ApiError;
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}

export const api = {
  getConfig: () => request<Config>('/api/config'),
  updateConfig: (cfg: Config) => request<Config>('/api/config', { method: 'PUT', body: JSON.stringify(cfg) }),
  getDashboard: () => request<DashboardData>('/api/dashboard'),
  getUpcomingEvents: () => request<UpcomingEvent[]>('/api/events/upcoming'),
  getHistory: () => request<HistoryItem[]>('/api/history'),
  syncNotion: () => request<{ count: number }>('/api/sync', { method: 'POST' }),
  syncCalendar: (fromDate?: string, toDate?: string) => 
    request<{ count: number }>('/api/calendar/sync', { method: 'POST', body: JSON.stringify({ from_date: fromDate, to_date: toDate }) }),
  clearCalendarSync: () => request<void>('/api/calendar/clear', { method: 'POST' }),
  clearHistory: () => request<void>('/api/history/clear', { method: 'POST' }),
  previewNotification: (req: NotificationPreviewRequest) =>
    request<{ message: string }>('/api/notifications/preview', { method: 'POST', body: JSON.stringify(req) }),
  sendManualNotification: (req: ManualNotificationRequest) =>
    request<{ message: string }>('/api/notifications/manual', { method: 'POST', body: JSON.stringify(req) }),
  getDefaultTemplates: () => request<Record<string, string>>('/api/templates/defaults'),
};
