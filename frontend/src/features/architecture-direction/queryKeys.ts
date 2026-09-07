export const timeAssessmentQueryKeys = {
  all: ['timeAssessments'] as const,
  byCapabilityIds: (capabilityIds: string[]) =>
    [...timeAssessmentQueryKeys.all, 'byCapabilityIds', [...capabilityIds].sort()] as const,
  collection: () => [...timeAssessmentQueryKeys.all, 'collection'] as const,
  rollups: (componentIds: string[]) => [...timeAssessmentQueryKeys.all, 'rollups', [...componentIds].sort()] as const,
};

export const realizationRoleQueryKeys = {
  all: ['realizationRoles'] as const,
  byCapabilityIds: (capabilityIds: string[]) =>
    [...realizationRoleQueryKeys.all, 'byCapabilityIds', [...capabilityIds].sort()] as const,
  collection: () => [...realizationRoleQueryKeys.all, 'collection'] as const,
};
