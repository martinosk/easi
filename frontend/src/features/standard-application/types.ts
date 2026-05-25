import type {
  ComponentId,
  EnterpriseCapabilityId,
  HATEOASLink,
  HATEOASLinks,
  StandardApplicationId,
} from '../../api/types';

export type { StandardApplicationId };

export interface StandardApplication {
  id: StandardApplicationId;
  enterpriseCapabilityId: EnterpriseCapabilityId;
  applicationId: ComponentId | string;
  applicationStale: boolean;
  narrative: string;
  setAt: string;
  updatedAt?: string;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    up?: HATEOASLink;
    edit?: HATEOASLink;
    'x-history'?: HATEOASLink;
  };
}

export interface ECStandardApplicationResponse {
  standard: StandardApplication | null;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    up?: HATEOASLink;
    edit?: HATEOASLink;
    'x-set-standard'?: HATEOASLink;
    'x-history'?: HATEOASLink;
  };
}

export interface StandardApplicationHistoryEntry {
  applicationId: ComponentId | string;
  previousApplicationId?: ComponentId | string;
  narrative: string;
  setAt: string;
}

export interface StandardApplicationHistory {
  standardApplicationId: StandardApplicationId;
  enterpriseCapabilityId: EnterpriseCapabilityId;
  entries: StandardApplicationHistoryEntry[];
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    up?: HATEOASLink;
  };
}

export interface SetStandardApplicationRequest {
  applicationId: string;
  narrative: string;
}
