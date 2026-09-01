export type AppView =
  | 'canvas'
  | 'business-domains'
  | 'value-streams'
  | 'invitations'
  | 'users'
  | 'settings'
  | 'strategic-fit'
  | 'my-edit-access'
  | 'one-pagers'
  | 'one-pager-quality';

export const ROUTES = {
  HOME: '/',
  CANVAS: '/canvas',
  BUSINESS_DOMAINS: '/business-domains',
  BUSINESS_DOMAIN_DETAIL: '/business-domains/:domainId',
  VALUE_STREAMS: '/value-streams',
  VALUE_STREAM_DETAIL: '/value-streams/:valueStreamId',
  STRATEGIC_FIT: '/strategic-fit',
  USERS: '/users',
  INVITATIONS: '/invitations',
  SETTINGS: '/settings',
  SETTINGS_MATURITY_SCALE: '/settings/maturity-scale',
  MY_EDIT_ACCESS: '/my-edit-access',
  ONE_PAGERS: '/one-pagers',
  ONE_PAGER_DETAIL: '/one-pagers/:subjectType/:subjectId',
  ONE_PAGER_QUALITY: '/one-pager-quality',
  LOGIN: '/login',
} as const;
