import { ROUTES } from '../../routes/routePaths';
import type { DeepLinkParam } from './types';

export const deepLinkParams = {
  VIEW: { param: 'view', routes: [ROUTES.CANVAS] } as DeepLinkParam,
  DOMAIN: { param: 'domain', routes: ['/business-domains'] } as DeepLinkParam,
  CAPABILITY: { param: 'capability', routes: ['/business-domains'] } as DeepLinkParam,
} as const;

export function getParamValue(param: string): string | null {
  const params = new URLSearchParams(window.location.search);
  return params.get(param);
}

function currentRoutePath(): string {
  const base = import.meta.env.BASE_URL.replace(/\/$/, '');
  const { pathname } = window.location;
  return base && pathname.startsWith(base) ? pathname.slice(base.length) || '/' : pathname;
}

function isOnRoute(route: string): boolean {
  const path = currentRoutePath();
  return path === route || path.startsWith(`${route}/`);
}

export function readDeepLink({ param, routes }: DeepLinkParam): string | null {
  return routes.some(isOnRoute) ? getParamValue(param) : null;
}

export function clearParams(paramsToClear: string[]): void {
  const url = new URL(window.location.href);
  for (const param of paramsToClear) {
    url.searchParams.delete(param);
  }
  window.history.replaceState({}, '', url.toString());
}
