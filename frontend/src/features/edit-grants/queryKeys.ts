export const editGrantsQueryKeys = {
  all: ['editGrants'] as const,
  mine: () => [...editGrantsQueryKeys.all, 'mine'] as const,
  forArtifact: (artifactType: string, artifactId: string) =>
    [...editGrantsQueryKeys.all, 'artifact', artifactType, artifactId] as const,
  detail: (id: string) => [...editGrantsQueryKeys.all, 'detail', id] as const,
};
