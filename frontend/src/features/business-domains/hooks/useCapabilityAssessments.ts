import { useMemo } from 'react';
import type { Capability, CapabilityRealization, ComponentId } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import {
  useTimeAssessmentRollups,
  useTimeAssessmentsByCapabilityIds,
} from '../../architecture-direction/hooks/useTimeAssessments';
import type { TimeAssessment, TimeAssessmentGradeCounts } from '../../architecture-direction/types';

export interface CapabilityAssessments {
  getAssessment: (componentId: ComponentId | string) => TimeAssessment | undefined;
  getRollup: (componentId: ComponentId | string) => TimeAssessmentGradeCounts | undefined;
  canAssess: boolean;
}

function directComponentIds(realizations: CapabilityRealization[]): string[] {
  return [...new Set(realizations.filter((r) => r.origin === 'Direct').map((r) => String(r.componentId)))];
}

export function useCapabilityAssessments(
  capability: Capability | null,
  realizations: CapabilityRealization[],
): CapabilityAssessments {
  const capabilityIds = useMemo(() => (capability ? [String(capability.id)] : []), [capability]);
  const componentIds = useMemo(() => directComponentIds(realizations), [realizations]);

  const assessmentsQuery = useTimeAssessmentsByCapabilityIds(capabilityIds);
  const rollupsQuery = useTimeAssessmentRollups(componentIds);

  const assessmentByComponentId = useMemo(() => {
    const map = new Map<string, TimeAssessment>();
    for (const assessment of assessmentsQuery.data?.data ?? []) map.set(String(assessment.componentId), assessment);
    return map;
  }, [assessmentsQuery.data]);

  const rollupByComponentId = useMemo(() => {
    const map = new Map<string, TimeAssessmentGradeCounts>();
    for (const rollup of rollupsQuery.data?.data ?? []) map.set(String(rollup.componentId), rollup.counts);
    return map;
  }, [rollupsQuery.data]);

  return {
    getAssessment: (componentId) => assessmentByComponentId.get(String(componentId)),
    getRollup: (componentId) => rollupByComponentId.get(String(componentId)),
    canAssess: hasLink(assessmentsQuery.data, 'x-assess'),
  };
}
