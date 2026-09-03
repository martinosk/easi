import { Group, Stack, Text } from '@mantine/core';
import { useMemo } from 'react';
import type { BusinessDomainId, Capability, CapabilityRealization, ComponentId } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { useAppStore } from '../../../store/appStore';
import { AppChip } from '../../business-domains/components/AppChip';
import { RealizationAssessment } from '../../business-domains/components/RealizationAssessment';
import { RealizationRoleControl } from '../../business-domains/components/RealizationRoleControl';
import {
  type CapabilityAssessments,
  useCapabilityAssessments,
} from '../../business-domains/hooks/useCapabilityAssessments';
import { type CapabilityRoles, useCapabilityRoles } from '../../business-domains/hooks/useCapabilityRoles';
import { useStrategyImportanceByCapability } from '../../business-domains/hooks/useStrategyImportance';
import { useCapabilityRealizations } from '../hooks/useCapabilities';
import classes from './CapabilityRealizationsSection.module.css';
import { RealizationFitContext } from './RealizationFitContext';

interface RealizationRowProps {
  realization: CapabilityRealization;
  domainIds: BusinessDomainId[];
  assessments: CapabilityAssessments;
  roles: CapabilityRoles;
  onApplicationClick: (componentId: ComponentId) => void;
}

function RealizationRow({ realization, domainIds, assessments, roles, onApplicationClick }: RealizationRowProps) {
  const isDirect = realization.origin === 'Direct';
  const assessed = {
    ...realization,
    timeGrade: assessments.getAssessment(realization.componentId)?.grade,
    role: roles.getRole(realization.componentId)?.role,
  };

  return (
    <div className={classes.realizationRow} data-testid={`drawer-realization-${realization.id}`}>
      <Group gap="xs" wrap="wrap">
        <AppChip realization={assessed} onClick={onApplicationClick} />
        <Text size="sm">{realization.realizationLevel}</Text>
      </Group>
      {realization.origin === 'Inherited' && (
        <Text className={classes.realizationMeta}>
          Inherited{realization.sourceCapabilityName ? ` from ${realization.sourceCapabilityName}` : ''}
        </Text>
      )}
      {realization.notes && <Text className={classes.realizationMeta}>{realization.notes}</Text>}
      {isDirect && (
        <RealizationAssessment
          capabilityId={realization.capabilityId}
          componentId={realization.componentId}
          assessment={assessments.getAssessment(realization.componentId)}
          rollup={assessments.getRollup(realization.componentId)}
          canAssess={assessments.canAssess}
          suggestion={assessments.getSuggestion(realization.componentId)}
        />
      )}
      {isDirect && (
        <RealizationRoleControl
          capabilityId={realization.capabilityId}
          componentId={realization.componentId}
          role={roles.getRole(realization.componentId)}
          canAssign={roles.canAssign}
        />
      )}
      {domainIds.map((domainId) => (
        <RealizationFitContext
          key={domainId}
          componentId={realization.componentId}
          capabilityId={realization.capabilityId}
          businessDomainId={domainId}
        />
      ))}
    </div>
  );
}

export interface CapabilityRealizationsSectionProps {
  capability: Capability;
  onApplicationClick?: (componentId: ComponentId) => void;
}

export function CapabilityRealizationsSection({ capability, onApplicationClick }: CapabilityRealizationsSectionProps) {
  const selectNode = useAppStore((state) => state.selectNode);
  const { data: realizations = [] } = useCapabilityRealizations(capability.id);
  const { data: importanceRatings = [] } = useStrategyImportanceByCapability(capability.id);
  const assessments = useCapabilityAssessments(capability, realizations);
  const roles = useCapabilityRoles(capability);
  const domainIds = useMemo(
    () => Array.from(new Set(importanceRatings.map((rating) => rating.businessDomainId))),
    [importanceRatings],
  );
  const handleClick = onApplicationClick ?? selectNode;

  return (
    <DetailField label="Realising applications">
      {realizations.length === 0 ? (
        <Text className={classes.empty}>no realising application mapped</Text>
      ) : (
        <Stack gap="xs">
          {realizations.map((realization) => (
            <RealizationRow
              key={realization.id}
              realization={realization}
              domainIds={domainIds}
              assessments={assessments}
              roles={roles}
              onApplicationClick={handleClick}
            />
          ))}
        </Stack>
      )}
    </DetailField>
  );
}
