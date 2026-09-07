import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createMantineTestWrapper } from '../../../test/helpers';
import type { OnePagerSubjectType } from '../types';
import { SubjectDrawer } from './SubjectDrawer';

vi.mock('../../capabilities/components/CapabilityDetailsPanel', () => ({
  CapabilityDetailsPanel: ({ capabilityId }: { capabilityId: string }) => (
    <div data-testid="capability-panel">{capabilityId}</div>
  ),
}));

vi.mock('../../components/components/ComponentDetailsPanel', () => ({
  ComponentDetailsPanel: ({ componentId }: { componentId: string }) => (
    <div data-testid="component-panel">{componentId}</div>
  ),
}));

vi.mock('../../origin-entities/components/AcquiredEntityDetailsPanel', () => ({
  AcquiredEntityDetailsPanel: ({ entityId }: { entityId: string }) => (
    <div data-testid="acquired-entity-panel">{entityId}</div>
  ),
}));

vi.mock('../../origin-entities/components/VendorDetailsPanel', () => ({
  VendorDetailsPanel: ({ entityId }: { entityId: string }) => <div data-testid="vendor-panel">{entityId}</div>,
}));

vi.mock('../../origin-entities/components/InternalTeamDetailsPanel', () => ({
  InternalTeamDetailsPanel: ({ entityId }: { entityId: string }) => (
    <div data-testid="internal-team-panel">{entityId}</div>
  ),
}));

function renderDrawer(subjectType: OnePagerSubjectType, subjectId: string, onClose = vi.fn()) {
  const { Wrapper } = createMantineTestWrapper();
  render(<SubjectDrawer opened onClose={onClose} subjectType={subjectType} subjectId={subjectId} />, {
    wrapper: Wrapper,
  });
  return onClose;
}

describe('SubjectDrawer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const panelCases: [OnePagerSubjectType, string][] = [
    ['capability', 'capability-panel'],
    ['application', 'component-panel'],
    ['acquired-entity', 'acquired-entity-panel'],
    ['vendor', 'vendor-panel'],
    ['internal-team', 'internal-team-panel'],
  ];

  it.each(panelCases)('hosts the %s detail panel for the subject id', async (subjectType, panelTestId) => {
    renderDrawer(subjectType, 'subject-1');

    expect(await screen.findByTestId(panelTestId)).toHaveTextContent('subject-1');
  });

  it('titles the drawer with the subject type label', async () => {
    renderDrawer('acquired-entity', 'subject-1');

    expect(await screen.findByText('Acquired Entity')).toBeInTheDocument();
  });
});
