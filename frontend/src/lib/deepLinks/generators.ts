import { generatePath } from 'react-router-dom';
import { ROUTES } from '../../routes/routePaths';
import { deepLinkParams } from './registry';

export function generateViewPath(viewId: string): string {
  const search = new URLSearchParams({ [deepLinkParams.VIEW.param]: viewId });
  return `${ROUTES.CANVAS}?${search}`;
}

export function generateViewShareUrl(viewId: string): string {
  return new URL(generateViewPath(viewId), window.location.origin).toString();
}

export function generateDomainShareUrl(domainId: string): string {
  const url = new URL(window.location.origin);
  url.pathname = '/business-domains';
  url.searchParams.set(deepLinkParams.DOMAIN.param, domainId);
  return url.toString();
}

export function generateOnePagerShareUrl(subjectType: string, subjectId: string): string {
  const url = new URL(window.location.origin);
  url.pathname = generatePath(ROUTES.ONE_PAGER_DETAIL, { subjectType, subjectId });
  return url.toString();
}
