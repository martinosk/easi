import { deepLinkParams } from './registry';

export function generateViewShareUrl(viewId: string): string {
  const url = new URL(window.location.origin);
  url.searchParams.set(deepLinkParams.VIEW.param, viewId);
  return url.toString();
}

export function generateDomainShareUrl(domainId: string): string {
  const url = new URL(window.location.origin);
  url.pathname = '/business-domains';
  url.searchParams.set(deepLinkParams.DOMAIN.param, domainId);
  return url.toString();
}
