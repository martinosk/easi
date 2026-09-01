export const strategicFitAnalysisQueryKeys = {
  all: ['strategicFitAnalysis'] as const,
  byPillar: (pillarId: string) => [...strategicFitAnalysisQueryKeys.all, 'byPillar', pillarId] as const,
};
