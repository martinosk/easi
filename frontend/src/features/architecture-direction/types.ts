import type { CapabilityId, EnterpriseCapabilityId, BusinessDomainId, HATEOASLink, HATEOASLinks } from '../../api/types';

export type DirectionId = string;

export type DirectionType = 'consolidate' | 'decompose' | 'stay';
export type DirectionStatus = 'draft' | 'proposed' | 'agreed' | 'rejected';
export type Horizon = 'now' | 'next' | 'later';

export interface DirectionPlacement {
  targetBusinessDomainId: BusinessDomainId | string;
  resultingName?: string;
}

export interface DirectionSourceCapability {
  id: CapabilityId | string;
  stale: boolean;
  name: string | null;
  businessDomainId?: BusinessDomainId | string;
  businessDomainName?: string | null;
}

export interface Direction {
  id: DirectionId;
  enterpriseCapabilityId: EnterpriseCapabilityId;
  type: DirectionType;
  status: DirectionStatus;
  horizon: Horizon;
  narrative?: string;
  sourceCapabilities: DirectionSourceCapability[];
  placements: DirectionPlacement[];
  createdAt: string;
  updatedAt?: string;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    up?: HATEOASLink;
    edit?: HATEOASLink;
    'x-propose'?: HATEOASLink;
    'x-agree'?: HATEOASLink;
    'x-reject'?: HATEOASLink;
  };
}

export interface ECDirectionResponse {
  direction: Direction | null;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    up?: HATEOASLink;
    'x-capture-direction'?: HATEOASLink;
  };
}

export interface PlacementInput {
  targetBusinessDomainId: string;
  resultingName?: string;
}

export interface CaptureDirectionRequest {
  type: DirectionType;
  sourceCapabilityIds: string[];
  placements: PlacementInput[];
  horizon: Horizon;
  narrative?: string;
}

export interface UpdateDirectionRequest {
  sourceCapabilityIds?: string[];
  placements?: PlacementInput[];
  horizon?: Horizon;
  narrative?: string;
}
