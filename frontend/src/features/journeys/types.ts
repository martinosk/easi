import type { HATEOASLink, HATEOASLinks } from '../../api/types';

export type JourneyId = string;

export type JourneyKind = 'migration' | 'consolidation' | 'carve-out' | 'move';

export type JourneyStatus = 'planned' | 'in-flight' | 'done' | 'abandoned';

export type MilestoneStatus = 'planned' | 'in-flight' | 'done';

export interface TargetPeriod {
  year: number;
  quarter: number;
}

export interface JourneyApplicationRef {
  componentId: string;
  componentName: string;
  stale: boolean;
}

export interface JourneyMove {
  targetDomainId: string;
  targetDomainName: string;
  targetDomainStale: boolean;
  targetParentId: string | null;
  targetParentName: string;
  targetParentStale: boolean;
  resultingName: string;
}

export interface JourneyMilestone {
  id: string;
  label: string;
  targetPeriod: TargetPeriod | null;
  status: MilestoneStatus;
  _links: HATEOASLinks & {
    edit?: HATEOASLink;
    delete?: HATEOASLink;
  };
}

export interface CapabilityJourney {
  id: JourneyId;
  capabilityId: string;
  capabilityName: string;
  capabilityStale: boolean;
  kind: JourneyKind;
  status: JourneyStatus;
  progress: number | null;
  targetPeriod: TargetPeriod | null;
  note: string;
  fromApplications: JourneyApplicationRef[];
  toApplication: JourneyApplicationRef;
  move?: JourneyMove;
  milestones: JourneyMilestone[];
  plannedBy: string;
  plannedByName: string;
  plannedAt: string;
  updatedAt: string;
  startedAt: string | null;
  completedAt: string | null;
  abandonedAt: string | null;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    'x-history'?: HATEOASLink;
    'x-start'?: HATEOASLink;
    'x-complete'?: HATEOASLink;
    'x-abandon'?: HATEOASLink;
    edit?: HATEOASLink;
    'x-progress'?: HATEOASLink;
    'x-change-sources'?: HATEOASLink;
    'x-add-milestone'?: HATEOASLink;
  };
}

export interface CapabilityJourneyResponse {
  journey: CapabilityJourney | null;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    'x-capture'?: HATEOASLink;
  };
}

export interface CapabilityJourneyHistoryResponse {
  data: CapabilityJourney[];
  _links: HATEOASLinks & { self?: HATEOASLink };
}

export interface CapabilityJourneysBulkResponse {
  data: CapabilityJourney[];
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    'x-capture'?: HATEOASLink;
  };
}

export interface TargetPeriodInput {
  year: number;
  quarter: number;
}

export interface CaptureJourneyRequest {
  kind: JourneyKind;
  fromComponentIds: string[];
  toComponentId: string;
  note?: string;
  targetPeriod?: TargetPeriodInput | null;
  targetDomainId?: string;
  targetParentId?: string | null;
  resultingName?: string;
}

export interface UpdateJourneyDetailsRequest {
  note: string;
  targetPeriod: TargetPeriodInput | null;
  resultingName?: string;
}

export interface UpdateJourneyProgressRequest {
  progress: number;
}

export interface ChangeJourneySourceApplicationsRequest {
  componentIds: string[];
}

export interface AddJourneyMilestoneRequest {
  label: string;
  targetPeriod?: TargetPeriodInput | null;
  status?: MilestoneStatus;
}

export interface UpdateJourneyMilestoneRequest {
  label: string;
  targetPeriod?: TargetPeriodInput | null;
  status: MilestoneStatus;
}
