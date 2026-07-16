import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createMantineTestWrapper } from '../../../test/helpers';
import type { EnterpriseCapability } from '../../enterprise-architecture/types';
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

vi.mock('../../enterprise-architecture/components/EnterpriseCapabilityDetailPanel', () => ({
  EnterpriseCapabilityDetailPanel: ({
    capability,
    onClose,
  }: {
    capability: EnterpriseCapability;
    onClose: () => void;
  }) => (
    <div data-testid="enterprise-capability-panel">
      {capability.name}
      <button type="button" onClick={onClose}>
        Close panel
      </button>
    </div>
  ),
}));

vi.mock('../../enterprise-architecture/hooks/useEnterpriseCapabilities', () => ({
  useEnterpriseCapability: vi.fn(),
}));

import { useEnterpriseCapability } from '../../enterprise-architecture/hooks/useEnterpriseCapabilities';

type EnterpriseCapabilityQuery = ReturnType<typeof useEnterpriseCapability>;

function mockEnterpriseCapabilityQuery(overrides: Partial<EnterpriseCapabilityQuery>) {
  vi.mocked(useEnterpriseCapability).mockReturnValue({
    data: undefined,
    isLoading: false,
    ...overrides,
  } as EnterpriseCapabilityQuery);
}

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
    mockEnterpriseCapabilityQuery({});
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

  it('hosts the enterprise capability panel once the capability is loaded', async () => {
    mockEnterpriseCapabilityQuery({
      data: { id: 'ec-1', name: 'Customer Insight' } as EnterpriseCapability,
    });

    renderDrawer('enterprise-capability', 'ec-1');

    expect(await screen.findByTestId('enterprise-capability-panel')).toHaveTextContent('Customer Insight');
  });

  it('shows a loading state while the enterprise capability loads', async () => {
    mockEnterpriseCapabilityQuery({ isLoading: true });

    renderDrawer('enterprise-capability', 'ec-1');

    expect(await screen.findByText('Loading...')).toBeInTheDocument();
    expect(screen.queryByTestId('enterprise-capability-panel')).not.toBeInTheDocument();
  });

  it('shows a failure state when the enterprise capability cannot be loaded', async () => {
    mockEnterpriseCapabilityQuery({});

    renderDrawer('enterprise-capability', 'ec-1');

    expect(await screen.findByText('Failed to load enterprise capability')).toBeInTheDocument();
  });

  it('closes the drawer when the enterprise panel close affordance is used', async () => {
    const user = userEvent.setup();
    mockEnterpriseCapabilityQuery({
      data: { id: 'ec-1', name: 'Customer Insight' } as EnterpriseCapability,
    });
    const onClose = renderDrawer('enterprise-capability', 'ec-1');

    await user.click(await screen.findByRole('button', { name: 'Close panel' }));

    expect(onClose).toHaveBeenCalled();
  });
});
