import type { HATEOASLinks } from '../../api/types';

export type ScopeKind = 'personal' | 'tenant' | 'empty';

export interface HomeDomainRef {
  id: string;
  name: string;
}

export interface HomeStewardship {
  domain?: HomeDomainRef;
  concern: string;
  concernLabel: string;
}

export interface HomeScope {
  kind: ScopeKind;
  stewardships: HomeStewardship[];
  architectedDomains?: HomeDomainRef[];
  architectedDomainCount: number;
  eaOwnedCapabilities: number;
  ownedApplications: number;
  editGrants: number;
}

export interface CapabilitiesTile {
  total: number;
  byStatus: { active: number; planned: number; deprecated: number };
}

export interface ApplicationsTile {
  total: number;
}

export interface DomainsTile {
  total: number;
  names?: string[];
}

export interface TimeShares {
  invest: number;
  tolerate: number;
  migrate: number;
  eliminate: number;
  notAssessed: number;
}

export interface TimeTile {
  total: number;
  shares?: TimeShares;
}

export interface HomePortfolio {
  capabilities?: CapabilitiesTile;
  applications?: ApplicationsTile;
  domains?: DomainsTile;
  time?: TimeTile;
}

export type MyWorkSubjectType = 'capability' | 'application';

export type MyWorkRelation = 'ea-owner' | 'owner' | 'nominated' | 'edit-grant';

export type DominantGrade = 'invest' | 'tolerate' | 'migrate' | 'eliminate';

export interface MyWorkItem {
  subjectType: MyWorkSubjectType;
  id: string;
  name: string;
  level?: string;
  relation: MyWorkRelation;
  grantExpiresOn?: string;
  dominantGrade?: DominantGrade;
  _links?: HATEOASLinks;
}

export interface MyWork {
  total: number;
  items: MyWorkItem[];
}

export interface HomeResponse {
  scope: HomeScope;
  portfolio?: HomePortfolio;
  myWork?: MyWork;
  _links: HATEOASLinks;
}
