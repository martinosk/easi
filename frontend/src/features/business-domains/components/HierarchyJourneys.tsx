import { Stack, UnstyledButton } from '@mantine/core';
import type { CapabilityJourney } from '../../journeys/types';
import { formatTargetPeriod, journeyKindLabel, journeyStatusLabel } from '../../journeys/utils/journeyFormat';
import { DrawerSectionHeader } from './DrawerSectionHeader';
import classes from './HierarchyJourneys.module.css';

export interface HierarchyJourneysProps {
  journeys: CapabilityJourney[];
  onNavigate: (capabilityId: string) => void;
}

interface JourneyRowProps {
  journey: CapabilityJourney;
  onNavigate: (capabilityId: string) => void;
}

function StatusPill({ journey }: { journey: CapabilityJourney }) {
  return (
    <span className={classes.pill} data-status={journey.status}>
      {journeyStatusLabel(journey)}
    </span>
  );
}

function TargetPeriod({ journey }: { journey: CapabilityJourney }) {
  return <span className={classes.when}>{formatTargetPeriod(journey.targetPeriod)}</span>;
}

function SubCapabilityJourneyRow({ journey, onNavigate }: JourneyRowProps) {
  return (
    <UnstyledButton
      component="button"
      className={classes.row}
      onClick={() => onNavigate(journey.capabilityId)}
      data-testid={`sub-capability-journey-${journey.capabilityId}`}
    >
      <span className={classes.dot} data-status={journey.status} />
      <span className={classes.body}>
        <span className={classes.name}>{journey.capabilityName}</span>
        <span className={classes.route}>
          {journeyKindLabel(journey.kind)} → {journey.toApplication.componentName}
        </span>
      </span>
      <span className={classes.right}>
        <StatusPill journey={journey} />
        <TargetPeriod journey={journey} />
      </span>
    </UnstyledButton>
  );
}

export function SubCapabilityJourneys({ journeys, onNavigate }: HierarchyJourneysProps) {
  if (journeys.length === 0) return null;

  return (
    <>
      <DrawerSectionHeader>Sub-capability journeys</DrawerSectionHeader>
      <Stack gap="xs" data-testid="sub-capability-journeys">
        {journeys.map((journey) => (
          <SubCapabilityJourneyRow key={journey.id} journey={journey} onNavigate={onNavigate} />
        ))}
      </Stack>
    </>
  );
}

function AncestorJourneyRow({ journey, onNavigate }: JourneyRowProps) {
  return (
    <UnstyledButton
      component="button"
      className={classes.ancestorRow}
      onClick={() => onNavigate(journey.capabilityId)}
      data-testid={`ancestor-journey-${journey.capabilityId}`}
    >
      <span className={classes.body}>
        <span className={classes.name}>{journey.capabilityName}</span>
        <span className={classes.route}>→ {journey.toApplication.componentName}</span>
      </span>
      <span className={classes.right}>
        <StatusPill journey={journey} />
        <TargetPeriod journey={journey} />
      </span>
    </UnstyledButton>
  );
}

export function AncestorJourneys({ journeys, onNavigate }: HierarchyJourneysProps) {
  if (journeys.length === 0) return null;

  return (
    <>
      <DrawerSectionHeader>Part of</DrawerSectionHeader>
      <Stack gap="xs" data-testid="ancestor-journeys">
        {journeys.map((journey) => (
          <AncestorJourneyRow key={journey.id} journey={journey} onNavigate={onNavigate} />
        ))}
      </Stack>
    </>
  );
}
