import type { HomeScope, HomeStewardship } from '../types';

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

export function countOf(count: number, singular: string, plural: string): string {
  return `${count} ${count === 1 ? singular : plural}`;
}

export function formatCalendarDate(isoDate: string): string {
  const [year, month, day] = isoDate.split('-').map(Number);
  return `${day} ${MONTHS[month - 1]} ${year}`;
}

export interface ScopeStatement {
  key: string;
  text: string;
}

interface StewardshipGroup {
  concernLabel: string;
  domainName?: string;
  domains: number;
}

function groupStewardships(stewardships: HomeStewardship[]): Map<string, StewardshipGroup> {
  const groups = new Map<string, StewardshipGroup>();
  for (const { concern, concernLabel, domain } of stewardships) {
    const key = domain ? `stewardship:${concern}:${domain.id}` : `stewardship:${concern}`;
    const group = groups.get(key) ?? { concernLabel, domainName: domain?.name, domains: 0 };
    groups.set(key, { ...group, domains: group.domains + 1 });
  }
  return groups;
}

function stewardshipText({ concernLabel, domainName, domains }: StewardshipGroup): string {
  if (domainName) return `Steward of ${concernLabel} in ${domainName}`;
  return domains > 1 ? `Steward of ${concernLabel} in ${domains} domains` : `Steward of ${concernLabel}`;
}

function stewardshipStatements(stewardships: HomeStewardship[]): ScopeStatement[] {
  return [...groupStewardships(stewardships)].map(([key, group]) => ({ key, text: stewardshipText(group) }));
}

function architectText(scope: HomeScope): string | null {
  if (scope.architectedDomains && scope.architectedDomains.length > 0) {
    return `Domain architect of ${scope.architectedDomains.map((d) => d.name).join(', ')}`;
  }
  if (scope.architectedDomainCount > 0) {
    return `Domain architect of ${countOf(scope.architectedDomainCount, 'domain', 'domains')}`;
  }
  return null;
}

function countText(count: number, text: (n: number) => string): string | null {
  return count > 0 ? text(count) : null;
}

function anchorStatements(scope: HomeScope): ScopeStatement[] {
  const texts: Record<string, string | null> = {
    architect: architectText(scope),
    'ea-owner': countText(scope.eaOwnedCapabilities, (n) => `EA owner of ${countOf(n, 'capability', 'capabilities')}`),
    owner: countText(scope.ownedApplications, (n) => `Owner of ${countOf(n, 'application', 'applications')}`),
    'edit-grant': countText(
      scope.editGrants,
      (n) => `Edit access to ${countOf(n, 'capability or application', 'capabilities or applications')}`,
    ),
  };
  return Object.entries(texts).flatMap(([key, text]) => (text === null ? [] : [{ key, text }]));
}

export function scopeStatements(scope: HomeScope): ScopeStatement[] {
  return [...stewardshipStatements(scope.stewardships), ...anchorStatements(scope)];
}
