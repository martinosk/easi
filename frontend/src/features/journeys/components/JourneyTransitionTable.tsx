import { Badge, Table } from '@mantine/core';
import { Fragment, type ReactNode } from 'react';
import type { CapabilityJourney } from '../types';
import { formatTargetPeriod, journeyKindLabel, journeyStatusLabel } from '../utils/journeyFormat';
import classes from './JourneyTransitionTable.module.css';

function StaleBadge() {
  return (
    <Badge color="orange" variant="light" size="xs" data-testid="journey-stale-badge">
      stale
    </Badge>
  );
}

function RefName({ name, stale }: { name: string; stale: boolean }) {
  return (
    <>
      {name}
      {stale && (
        <>
          {' '}
          <StaleBadge />
        </>
      )}
    </>
  );
}

function FromToCell({ journey }: { journey: CapabilityJourney }) {
  return (
    <>
      {journey.fromApplications.map((app, index) => (
        <Fragment key={app.componentId}>
          {index > 0 && ' + '}
          <RefName name={app.componentName} stale={app.stale} />
        </Fragment>
      ))}
      {' → '}
      <RefName name={journey.toApplication.componentName} stale={journey.toApplication.stale} />
    </>
  );
}

function Row({ label, testId, children }: { label: string; testId: string; children: ReactNode }) {
  return (
    <Table.Tr>
      <Table.Td className={classes.label}>{label}</Table.Td>
      <Table.Td data-testid={testId}>{children}</Table.Td>
    </Table.Tr>
  );
}

function MoveRows({ journey }: { journey: CapabilityJourney }) {
  const move = journey.move;
  if (!move) return null;
  return (
    <>
      <Row label="New home" testId="journey-new-home">
        <RefName name={move.targetDomainName} stale={move.targetDomainStale} />
        {' → '}
        {move.resultingName}
      </Row>
      <Row label="Target app" testId="journey-target-app">
        <RefName name={journey.toApplication.componentName} stale={journey.toApplication.stale} />
      </Row>
    </>
  );
}

export function JourneyTransitionTable({ journey }: { journey: CapabilityJourney }) {
  return (
    <Table className={classes.table} data-testid="journey-transition-table">
      <Table.Tbody>
        <Row label="Type" testId="journey-type">
          {journeyKindLabel(journey.kind)}
        </Row>
        {journey.kind === 'move' ? (
          <MoveRows journey={journey} />
        ) : (
          <Row label="From → to" testId="journey-from-to">
            <FromToCell journey={journey} />
          </Row>
        )}
        <Row label="Status" testId="journey-status">
          {journeyStatusLabel(journey)}
        </Row>
        <Row label="Target date" testId="journey-target-date">
          {formatTargetPeriod(journey.targetPeriod)}
        </Row>
      </Table.Tbody>
    </Table>
  );
}
