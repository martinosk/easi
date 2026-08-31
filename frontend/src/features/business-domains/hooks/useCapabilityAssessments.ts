import { useMemo } from 'react';
import type { Capability, CapabilityRealization, ComponentId } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import {
  useTimeAssessmentRollups,
  useTimeAssessmentsByCapabilityIds,
} from '../../architecture-direction/hooks/useTimeAssessments';
import type {
  TimeAssessment,
  TimeAssessmentGradeCounts,
  TimeAssessmentRollup,
  TimeSuggestion,
} from '../../architecture-direction/types';

export interface CapabilityAssessments {
  getAssessment: (componentId: ComponentId | string) => TimeAssessment | undefined;
  getRollup: (componentId: ComponentId | string) => TimeAssessmentGradeCounts | undefined;
  getSuggestion: (componentId: ComponentId | string) => TimeSuggestion | null;
  canAssess: boolean;
}

function directComponentIds(realizations: CapabilityRealization[]): string[] {
  return [...new Set(realizations.filter((r) => r.origin === 'Direct').map((r) => String(r.componentId)))];
}

function byComponentId<T>(entries: TimeAssessment[], pick: (entry: TimeAssessment) => T | null): Map<string, T> {
  const map = new Map<string, T>();
  for (const entry of entries) {
    const value = pick(entry);
    if (value !== null) map.set(String(entry.componentId), value);
  }
  return map;
}

function rollupsByComponentId(rollups: TimeAssessmentRollup[]): Map<string, TimeAssessmentGradeCounts> {
  return new Map(rollups.map((rollup) => [String(rollup.componentId), rollup.counts]));
}

export function useCapabilityAssessments(
  capability: Capability | null,
  realizations: CapabilityRealization[],
): CapabilityAssessments {
  const capabilityIds = useMemo(() => (capability ? [String(capability.id)] : []), [capability]);
  const componentIds = useMemo(() => directComponentIds(realizations), [realizations]);

  const assessmentsQuery = useTimeAssessmentsByCapabilityIds(capabilityIds);
  const rollupsQuery = useTimeAssessmentRollups(componentIds);

  const entries = useMemo(() => assessmentsQuery.data?.data ?? [], [assessmentsQuery.data]);

  const assessed = useMemo(() => byComponentId(entries, (entry) => (entry.grade ? entry : null)), [entries]);
  const suggested = useMemo(() => byComponentId(entries, (entry) => entry.suggestion), [entries]);
  const rollups = useMemo(() => rollupsByComponentId(rollupsQuery.data?.data ?? []), [rollupsQuery.data]);

  return {
    getAssessment: (componentId) => assessed.get(String(componentId)),
    getRollup: (componentId) => rollups.get(String(componentId)),
    getSuggestion: (componentId) => suggested.get(String(componentId)) ?? null,
    canAssess: hasLink(assessmentsQuery.data, 'x-assess'),
  };
}
