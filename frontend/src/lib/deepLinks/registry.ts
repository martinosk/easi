import type { DeepLinkParam } from './types';

export const deepLinkParams = {
  VIEW: { param: 'view', routes: ['*'] } as DeepLinkParam,
  DOMAIN: { param: 'domain', routes: ['/business-domains'] } as DeepLinkParam,
  CAPABILITY: { param: 'capability', routes: ['/business-domains'] } as DeepLinkParam,
} as const;

export function getParamValue(param: string): string | null {
  const params = new URLSearchParams(window.location.search);
  return params.get(param);
}

export function clearParams(paramsToClear: string[]): void {
  const url = new URL(window.location.href);
  paramsToClear.forEach(param => url.searchParams.delete(param));
  window.history.replaceState({}, '', url.toString());
}
