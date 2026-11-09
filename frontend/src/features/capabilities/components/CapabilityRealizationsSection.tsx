import { ActionIcon, Button, Group, Stack, Text } from '@mantine/core';
import { IconPlus, IconTrash } from '@tabler/icons-react';
import { useMemo, useState } from 'react';
import type { BusinessDomainId, Capability, CapabilityRealization, ComponentId } from '../../../api/types';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { useAppStore } from '../../../store/appStore';
import { canDelete, hasLink } from '../../../utils/hateoas';
import { AppChip } from '../../business-domains/components/AppChip';
import { RealizationAssessment } from '../../business-domains/components/RealizationAssessment';
import { RealizationRoleControl } from '../../business-domains/components/RealizationRoleControl';
import {
  type CapabilityAssessments,
  useCapabilityAssessments,
} from '../../business-domains/hooks/useCapabilityAssessments';
import { type CapabilityRoles, useCapabilityRoles } from '../../business-domains/hooks/useCapabilityRoles';
import { useStrategyImportanceByCapability } from '../../business-domains/hooks/useStrategyImportance';
import { useCapabilityRealizations, useDeleteRealization } from '../hooks/useCapabilities';
import { AddRealizationDialog } from './AddRealizationDialog';
import classes from './CapabilityRealizationsSection.module.css';
import { RealizationFitContext } from './RealizationFitContext';

function RemoveRealizationControl({ realization }: { realization: CapabilityRealization }) {
  const [confirming, setConfirming] = useState(false);
  const deleteRealization = useDeleteRealization();
  const name = realization.componentName ?? realization.componentId;

  const handleConfirm = () => {
    deleteRealization.mutate(realization, { onSuccess: () => setConfirming(false) });
  };

  return (
    <>
      <ActionIcon
        variant="subtle"
        color="gray"
        size="sm"
        aria-label={`Remove ${name}`}
        onClick={() => setConfirming(true)}
      >
        <IconTrash size="1rem" />
      </ActionIcon>
      {confirming && (
        <ConfirmationDialog
          title="Remove realising application"
          message="Remove this application from the capability's realising applications?"
          itemName={name}
          confirmText="Remove"
          onConfirm={handleConfirm}
          onCancel={() => setConfirming(false)}
          isLoading={deleteRealization.isPending}
          error={deleteRealization.error ? deleteRealization.error.message : null}
        />
      )}
    </>
  );
}

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
    timeGrade: assessments.getAssessment(realization.componentId)?.grade ?? undefined,
    role: roles.getRole(realization.componentId)?.role,
  };

  return (
    <div className={classes.realizationRow} data-testid={`drawer-realization-${realization.id}`}>
      <Group gap="xs" wrap="wrap" justify="space-between">
        <Group gap="xs" wrap="wrap">
          <AppChip realization={assessed} onClick={onApplicationClick} />
          <Text size="sm">{realization.realizationLevel}</Text>
        </Group>
        {canDelete(realization) && <RemoveRealizationControl realization={realization} />}
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

function AddRealizationControl({
  capability,
  realizations,
}: {
  capability: Capability;
  realizations: CapabilityRealization[];
}) {
  const [adding, setAdding] = useState(false);
  if (!hasLink(capability, 'x-add-realization')) return null;

  return (
    <Group justify="flex-start">
      <Button variant="subtle" size="xs" leftSection={<IconPlus size="1rem" />} onClick={() => setAdding(true)}>
        Add application
      </Button>
      {adding && (
        <AddRealizationDialog capability={capability} realizations={realizations} onClose={() => setAdding(false)} />
      )}
    </Group>
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
    <Stack gap="xs">
      {realizations.length === 0 && <Text className={classes.empty}>no realising application mapped</Text>}
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
      <AddRealizationControl capability={capability} realizations={realizations} />
    </Stack>
  );
}
