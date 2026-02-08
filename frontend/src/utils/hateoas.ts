import type { HATEOASLink, HATEOASLinks, HttpMethod } from '../api/types';

export interface ResourceWithLinks {
  _links?: HATEOASLinks;
}

export interface LinkRequest<T = unknown> {
  url: string;
  method: HttpMethod;
  data?: T;
}

export function hasLink(resource: ResourceWithLinks | null | undefined, linkName: string): boolean {
  return resource?._links?.[linkName] !== undefined;
}

export function getLink(resource: ResourceWithLinks | null | undefined, linkName: string): string | undefined {
  return resource?._links?.[linkName]?.href;
}

export function getLinkMethod(resource: ResourceWithLinks | null | undefined, linkName: string): HttpMethod | undefined {
  return resource?._links?.[linkName]?.method;
}

export function getLinkObject(resource: ResourceWithLinks | null | undefined, linkName: string): HATEOASLink | undefined {
  return resource?._links?.[linkName];
}

export function followLink(resource: ResourceWithLinks | null | undefined, linkName: string): string {
  const href = getLink(resource, linkName);
  if (!href) {
    throw new Error(`Link '${linkName}' not found on resource`);
  }
  return href;
}

export function buildLinkRequest<T = unknown>(
  resource: ResourceWithLinks | null | undefined,
  linkName: string,
  data?: T
): LinkRequest<T> {
  const link = getLinkObject(resource, linkName);
  if (!link) {
    throw new Error(`Link '${linkName}' not found on resource`);
  }
  return {
    url: link.href,
    method: link.method,
    data,
  };
}

export function canEdit(resource: ResourceWithLinks | null | undefined): boolean {
  return hasLink(resource, 'edit');
}

export function canDelete(resource: ResourceWithLinks | null | undefined): boolean {
  return hasLink(resource, 'delete');
}

export function canRemove(resource: ResourceWithLinks | null | undefined): boolean {
  return hasLink(resource, 'x-remove');
}

export function canCreate(resource: ResourceWithLinks | null | undefined): boolean {
  return hasLink(resource, 'create');
}

export function canInviteToEdit(resource: ResourceWithLinks | null | undefined): boolean {
  return hasLink(resource, 'x-edit-grants');
}
