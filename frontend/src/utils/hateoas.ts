import type { HATEOASLink, HATEOASLinks, HttpMethod } from '../api/types';

export interface ResourceWithLinks {
  _links?: HATEOASLinks;
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
