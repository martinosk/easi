import { Badge, Button, Drawer, Group, Stack, Text, Title } from '@mantine/core';
import { useState } from 'react';
import type {
  BusinessDomain,
  Capability,
  CapabilityId,
  CapabilityRealization,
  ComponentId,
} from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { canEdit as canEditResource, hasLink } from '../../../utils/hateoas';
import { AddExpertDialog } from '../../capabilities/components/AddExpertDialog';
import { CapabilityExpertsList } from '../../capabilities/components/CapabilityExpertsList';
import { EditCapabilityDialog } from '../../capabilities/components/EditCapabilityDialog';
import { OnePagerActionButton } from '../../one-pagers';
import { type CapabilityAssessments, useCapabilityAssessments } from '../hooks/useCapabilityAssessments';
import { type CapabilityRoles, useCapabilityRoles } from '../hooks/useCapabilityRoles';
import type { CapabilityHierarchyJourneys } from '../lens/hierarchyJourneys';
import { AppChip } from './AppChip';
import classes from './CapabilityDrawer.module.css';
import { DrawerSectionHeader } from './DrawerSectionHeader';
import { JourneySection } from './JourneySection';
import { RealizationAssessment } from './RealizationAssessment';
import { RealizationRoleControl } from './RealizationRoleControl';
import { StrategicImportanceSection } from './StrategicImportanceSection';

export interface CapabilityDrawerProps {
  capability: Capability | null;
  domain: BusinessDomain | null;
  l1Name: string | null;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  hierarchyJourneys: CapabilityHierarchyJourneys;
  onClose: () => void;
  onChipClick: (componentId: ComponentId) => void;
  onNavigateToCapability: (capabilityId: string) => void;
}

interface RealizationRowProps {
  realization: CapabilityRealization;
  onChipClick: (componentId: ComponentId) => void;
  assessments: CapabilityAssessments;
  roles: CapabilityRoles;
}

function RealizationRow({ realization, onChipClick, assessments, roles }: RealizationRowProps) {
  return (
    <div className={classes.realizationRow} data-testid={`drawer-realization-${realization.id}`}>
      <Group gap="xs" wrap="wrap">
        <AppChip realization={realization} onClick={onChipClick} />
        <Text size="sm">{realization.realizationLevel}</Text>
      </Group>
      {realization.origin === 'Inherited' && (
        <Text className={classes.realizationMeta}>
          Inherited{realization.sourceCapabilityName ? ` from ${realization.sourceCapabilityName}` : ''}
        </Text>
      )}
      {realization.notes && <Text className={classes.realizationMeta}>{realization.notes}</Text>}
      {realization.origin === 'Direct' && (
        <RealizationAssessment
          capabilityId={realization.capabilityId}
          componentId={realization.componentId}
          assessment={assessments.getAssessment(realization.componentId)}
          rollup={assessments.getRollup(realization.componentId)}
          canAssess={assessments.canAssess}
          suggestion={assessments.getSuggestion(realization.componentId)}
        />
      )}
      {realization.origin === 'Direct' && (
        <RealizationRoleControl
          capabilityId={realization.capabilityId}
          componentId={realization.componentId}
          role={roles.getRole(realization.componentId)}
          canAssign={roles.canAssign}
        />
      )}
    </div>
  );
}

interface RealisingApplicationsSectionProps {
  capability: Capability;
  realizations: CapabilityRealization[];
  onChipClick: (componentId: ComponentId) => void;
}

function RealisingApplicationsSection({ capability, realizations, onChipClick }: RealisingApplicationsSectionProps) {
  const assessments = useCapabilityAssessments(capability, realizations);
  const roles = useCapabilityRoles(capability);

  return (
    <Stack gap="xs">
      <DrawerSectionHeader>Realising applications</DrawerSectionHeader>
      {realizations.length === 0 ? (
        <Text className={classes.emptyRealizations}>no realising application mapped</Text>
      ) : (
        <Stack gap="xs">
          {realizations.map((realization) => (
            <RealizationRow
              key={realization.id}
              realization={realization}
              onChipClick={onChipClick}
              assessments={assessments}
              roles={roles}
            />
          ))}
        </Stack>
      )}
    </Stack>
  );
}

interface DetailsSectionProps {
  capability: Capability;
  onAddExpertClick: () => void;
}

function DetailsSection({ capability, onAddExpertClick }: DetailsSectionProps) {
  return (
    <Stack gap="sm">
      <DrawerSectionHeader>Details</DrawerSectionHeader>
      {capability.description && <DetailField label="Description">{capability.description}</DetailField>}
      {capability.ownershipModel && <DetailField label="Ownership Model">{capability.ownershipModel}</DetailField>}
      {capability.primaryOwner && <DetailField label="Primary Owner">{capability.primaryOwner}</DetailField>}
      {capability.eaOwner && <DetailField label="EA Owner">{capability.eaOwnerName ?? capability.eaOwner}</DetailField>}
      {capability.tags && capability.tags.length > 0 && (
        <DetailField label="Tags">
          <Group gap="xs">
            {capability.tags.map((tag) => (
              <Badge key={tag} variant="light">
                {tag}
              </Badge>
            ))}
          </Group>
        </DetailField>
      )}
      <CapabilityExpertsList
        capabilityId={capability.id}
        experts={capability.experts}
        canAddExpert={hasLink(capability, 'x-add-expert')}
        onAddClick={onAddExpertClick}
      />
    </Stack>
  );
}

export function CapabilityDrawer({
  capability,
  domain,
  l1Name,
  getRealizationsForCapability,
  hierarchyJourneys,
  onClose,
  onChipClick,
  onNavigateToCapability,
}: CapabilityDrawerProps) {
  const [editOpen, setEditOpen] = useState(false);
  const [addExpertOpen, setAddExpertOpen] = useState(false);

  return (
    <Drawer
      opened={capability !== null}
      onClose={onClose}
      position="right"
      size="md"
      data-testid="capability-drawer"
      title={
        domain && l1Name ? (
          <Text className={classes.breadcrumb}>
            {domain.name} · {l1Name}
          </Text>
        ) : undefined
      }
    >
      {capability && (
        <Stack gap="md">
          <Title order={3}>{capability.name}</Title>
          <Group gap="xs">
            <Badge variant="filled" color="dark">
              {capability.level}
            </Badge>
            {capability.maturityLevel && <Badge variant="light">{capability.maturityLevel}</Badge>}
          </Group>

          <RealisingApplicationsSection
            capability={capability}
            realizations={getRealizationsForCapability(capability.id)}
            onChipClick={onChipClick}
          />

          <JourneySection
            capability={capability}
            realizations={getRealizationsForCapability(capability.id)}
            hierarchyJourneys={hierarchyJourneys}
            onNavigateToCapability={onNavigateToCapability}
          />

          {domain && <StrategicImportanceSection domain={domain} capabilityId={capability.id} />}

          <DetailsSection capability={capability} onAddExpertClick={() => setAddExpertOpen(true)} />

          <Group gap="sm">
            {canEditResource(capability) && (
              <Button variant="default" size="xs" onClick={() => setEditOpen(true)}>
                Edit
              </Button>
            )}
            <OnePagerActionButton subject={capability} subjectType="capability" subjectId={capability.id} />
          </Group>

          <EditCapabilityDialog isOpen={editOpen} onClose={() => setEditOpen(false)} capability={capability} />
          <AddExpertDialog
            isOpen={addExpertOpen}
            onClose={() => setAddExpertOpen(false)}
            capabilityId={capability.id}
          />
        </Stack>
      )}
    </Drawer>
  );
}
