import type { HATEOASLinks } from '../../api/types';

export type Concern = 'ownership' | 'assessment' | 'documentation' | 'planning' | 'structure';

export interface StewardshipPerson {
  id: string;
  name: string | null;
}

export interface ConcernStewardship {
  concern: Concern;
  label: string;
  description: string;
  steward: StewardshipPerson | null;
  assignedBy: string | null;
  assignedAt: string | null;
  _links: HATEOASLinks;
}

export interface DomainStewardships {
  domainId: string;
  domainName: string;
  fallback: StewardshipPerson | null;
  data: ConcernStewardship[];
  _links: HATEOASLinks;
}

export interface StewardsTarget {
  domainId: string;
  domainName: string;
  stewardshipsHref: string;
}
