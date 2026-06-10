import type {
  BusinessDomainId,
  CapabilityId,
  CapabilityLevel,
  EnterpriseCapabilityId,
  HATEOASLink,
  HATEOASLinks,
  PaginationInfo,
} from '../../api/types';
import type { CarvedOutBy, IncludedCapabilityRole } from '../enterprise-architecture/types';

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
    'x-add-source'?: HATEOASLink;
  };
}

export interface ECDirectionResponse {
  direction: Direction | null;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    up?: HATEOASLink;
    'x-capture-direction'?: HATEOASLink;
    'x-composition'?: HATEOASLink;
  };
}

export interface CaptureDirectionRequest {
  type: DirectionType;
  sourceCapabilityIds: string[];
  horizon: Horizon;
  narrative?: string;
}

export interface UpdateDirectionRequest {
  horizon?: Horizon;
  narrative?: string;
}

export interface AddSourceRequest {
  capabilityId: string;
}

export interface ConflictingEnterpriseCapability {
  id: string;
  name: string;
}

export interface SourceCandidate {
  capabilityId: string;
  name: string;
  level: CapabilityLevel;
  parentId?: string | null;
  businessDomainId?: string | null;
  businessDomainName?: string | null;
  eligible: boolean;
  ineligibilityReason?: string | null;
  conflictingEnterpriseCapability?: ConflictingEnterpriseCapability | null;
  _links?: HATEOASLinks & { self?: HATEOASLink; 'x-conflicting-ec'?: HATEOASLink };
}

export interface SourceCandidatesResponse {
  data: SourceCandidate[];
  pagination: PaginationInfo;
  _links: HATEOASLinks;
}

export interface SourceCandidatesQuery {
  q: string;
  domainId?: string;
  limit?: number;
}

export interface SourceEligibility {
  capabilityId: string;
  eligible: boolean;
  ineligibilityReason?: string | null;
  conflictingEnterpriseCapability?: ConflictingEnterpriseCapability | null;
}

export interface PreviewIncludedCapability {
  capabilityId: string;
  name: string;
  level: CapabilityLevel;
  businessDomainId?: string | null;
  businessDomainName?: string | null;
  role: IncludedCapabilityRole;
  carvedOutBy?: CarvedOutBy | null;
}

export interface CompositionPreviewRequest {
  sourceCapabilityIds: string[];
}

export interface CompositionPreviewResponse {
  includedCapabilities: PreviewIncludedCapability[];
  sourceEligibility: SourceEligibility[];
  meta: {
    sourceCount: number;
    includedCount: number;
    carvedOutCount: number;
  };
  _links: HATEOASLinks;
}
