import type {
  CapabilityId,
  ComponentId,
  HATEOASLink,
  HATEOASLinks,
  TimeGrade,
} from '../../api/types';

export type { TimeGrade };

export type TimeSuggestionConfidence = 'LOW' | 'MEDIUM' | 'HIGH';

export interface TimeSuggestion {
  grade: TimeGrade | null;
  confidence: TimeSuggestionConfidence;
  technicalGap: number | null;
  functionalGap: number | null;
}

export interface TimeAssessment {
  id: string;
  capabilityId: CapabilityId | string;
  capabilityName: string;
  componentId: ComponentId | string;
  componentName: string;
  grade: TimeGrade | null;
  rationale: string;
  assessedBy: string;
  assessedByName?: string;
  assessedAt: string | null;
  stale: boolean;
  suggestion: TimeSuggestion | null;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    edit?: HATEOASLink;
    delete?: HATEOASLink;
  };
}

export interface TimeAssessmentsResponse {
  data: TimeAssessment[];
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    'x-assess'?: HATEOASLink;
  };
}

export interface TimeAssessmentGradeCounts {
  Invest: number;
  Tolerate: number;
  Migrate: number;
  Eliminate: number;
}

export interface TimeAssessmentRollup {
  componentId: ComponentId | string;
  counts: TimeAssessmentGradeCounts;
}

export interface TimeAssessmentRollupsResponse {
  data: TimeAssessmentRollup[];
  _links: HATEOASLinks;
}

export interface AssessRealizationRequest {
  grade: TimeGrade;
  rationale?: string;
}

export type RealizationRole = 'standard' | 'legacy';

export interface RealizationRoleAssignment {
  capabilityId: CapabilityId | string;
  capabilityName: string;
  componentId: ComponentId | string;
  componentName: string;
  role: RealizationRole;
  assignedBy: string;
  assignedAt: string;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    edit?: HATEOASLink;
    delete?: HATEOASLink;
  };
}

export interface RealizationRolesResponse {
  data: RealizationRoleAssignment[];
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    'x-assign'?: HATEOASLink;
  };
}

export interface AssignRealizationRoleRequest {
  role: RealizationRole;
}
