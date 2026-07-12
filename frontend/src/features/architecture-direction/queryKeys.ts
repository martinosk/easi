export const directionQueryKeys = {
  all: ['directions'] as const,
  byEnterpriseCapability: (id: string) => [...directionQueryKeys.all, 'byEC', id] as const,
  sourceCandidates: (id: string, query: { q: string; domainId?: string }) =>
    [...directionQueryKeys.all, 'sourceCandidates', id, query] as const,
  compositionPreview: (id: string, sourceCapabilityIds: string[]) =>
    [...directionQueryKeys.all, 'compositionPreview', id, [...sourceCapabilityIds].sort()] as const,
};

export const timeAssessmentQueryKeys = {
  all: ['timeAssessments'] as const,
  byCapabilityIds: (capabilityIds: string[]) =>
    [...timeAssessmentQueryKeys.all, 'byCapabilityIds', [...capabilityIds].sort()] as const,
  rollups: (componentIds: string[]) => [...timeAssessmentQueryKeys.all, 'rollups', [...componentIds].sort()] as const,
};

export const realizationRoleQueryKeys = {
  all: ['realizationRoles'] as const,
  byCapabilityIds: (capabilityIds: string[]) =>
    [...realizationRoleQueryKeys.all, 'byCapabilityIds', [...capabilityIds].sort()] as const,
};
