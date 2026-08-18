import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { buildCapabilityJourney, renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildStubJourney } from '../../../test/mocks/spec182/builders';
import { seedSpec182Db } from '../../../test/mocks/spec182/store';
import type { CapabilityJourney } from '../../journeys/types';
import type { CapabilityHierarchyJourneys } from '../lens/hierarchyJourneys';
import { JourneySection } from './JourneySection';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function journeyOn(
  capabilityId: string,
  capabilityName: string,
  targetApplication: string,
  overrides: Partial<CapabilityJourney> = {},
): CapabilityJourney {
  return buildCapabilityJourney({
    id: `journey-${capabilityId}`,
    capabilityId,
    capabilityName,
    kind: 'migration',
    toApplication: { componentId: `comp-${capabilityId}`, componentName: targetApplication, stale: false },
    ...overrides,
  });
}

const HAZARDOUS_JOURNEY = journeyOn('hazardous', 'Hazardous', 'Control Tower Hazardous', {
  status: 'planned',
  progress: null,
  targetPeriod: { year: 2026, quarter: 3 },
});

const FERRY_BOOKING_JOURNEY = journeyOn('ferry-booking', 'Ferry booking', 'Control Tower', {
  status: 'in-flight',
  progress: 25,
  targetPeriod: { year: 2028, quarter: 4 },
});

function renderDrawerSection(hierarchyJourneys: CapabilityHierarchyJourneys, onNavigate = vi.fn()) {
  const capability = cap('ferry-booking', 'Ferry booking', 'L1');
  renderWithProviders(
    <JourneySection
      capability={capability}
      realizations={[]}
      hierarchyJourneys={hierarchyJourneys}
      onNavigateToCapability={onNavigate}
    />,
    { withRouter: false },
  );
  return onNavigate;
}

function descendants(journeys: CapabilityJourney[]): CapabilityHierarchyJourneys {
  return { descendants: journeys, ancestors: [] };
}

function ancestors(journeys: CapabilityJourney[]): CapabilityHierarchyJourneys {
  return { descendants: [], ancestors: journeys };
}

describe('JourneySection — sub-capability journeys', () => {
  it('lists a descendant journey with its kind, status, target period and target application', async () => {
    seedSpec182Db({
      journeys: [buildStubJourney({ capabilityId: 'ferry-booking', status: 'in-flight', progress: 25 })],
    });
    renderDrawerSection(descendants([HAZARDOUS_JOURNEY]));

    const row = await screen.findByTestId('sub-capability-journey-hazardous');
    expect(row).toHaveTextContent('Hazardous');
    expect(row).toHaveTextContent('migration');
    expect(row).toHaveTextContent('Control Tower Hazardous');
    expect(row).toHaveTextContent('planned');
    expect(row).toHaveTextContent('Q3 2026');
  });

  it('lists descendant journeys when the capability has no journey of its own', async () => {
    seedSpec182Db({ journeys: [] });
    renderDrawerSection(descendants([HAZARDOUS_JOURNEY]));

    expect(await screen.findByText('No change planned.')).toBeInTheDocument();
    expect(screen.getByTestId('sub-capability-journey-hazardous')).toBeInTheDocument();
  });

  it('shows a completed descendant journey as done', async () => {
    seedSpec182Db({ journeys: [] });
    renderDrawerSection(descendants([{ ...HAZARDOUS_JOURNEY, status: 'done', progress: 100 }]));

    const row = await screen.findByTestId('sub-capability-journey-hazardous');
    expect(row).toHaveTextContent('done');
  });

  it('shows the descendant progress for an in-flight journey', async () => {
    seedSpec182Db({ journeys: [] });
    renderDrawerSection(descendants([{ ...HAZARDOUS_JOURNEY, status: 'in-flight', progress: 40 }]));

    expect(await screen.findByTestId('sub-capability-journey-hazardous')).toHaveTextContent('in flight (40%)');
  });

  it('renders no list when there are no descendant journeys', async () => {
    seedSpec182Db({ journeys: [] });
    renderDrawerSection(descendants([]));

    await screen.findByText('No change planned.');
    expect(screen.queryByTestId('sub-capability-journeys')).not.toBeInTheDocument();
    expect(screen.queryByText('Sub-capability journeys')).not.toBeInTheDocument();
  });

  it('navigates to the descendant capability when its row is activated', async () => {
    seedSpec182Db({ journeys: [] });
    const onNavigate = renderDrawerSection(descendants([HAZARDOUS_JOURNEY]));

    await userEvent.click(await screen.findByTestId('sub-capability-journey-hazardous'));

    expect(onNavigate).toHaveBeenCalledWith('hazardous');
  });

  it('renders the list with navigation only for a read-only caller', async () => {
    seedSpec182Db({ canWrite: false, journeys: [] });
    renderDrawerSection(descendants([HAZARDOUS_JOURNEY]));

    expect(await screen.findByTestId('sub-capability-journey-hazardous')).toBeInTheDocument();
    expect(screen.queryByTestId('plan-journey-btn')).not.toBeInTheDocument();
    expect(screen.getAllByRole('button')).toHaveLength(1);
  });
});

describe('JourneySection — ancestor journeys', () => {
  it('names the ancestor journey the capability is part of, with status and target period', async () => {
    seedSpec182Db({ journeys: [] });
    renderDrawerSection(ancestors([FERRY_BOOKING_JOURNEY]));

    const line = await screen.findByTestId('ancestor-journey-ferry-booking');
    expect(screen.getByText('Part of')).toBeInTheDocument();
    expect(line).toHaveTextContent('Ferry booking');
    expect(line).toHaveTextContent('Control Tower');
    expect(line).toHaveTextContent('in flight (25%)');
    expect(line).toHaveTextContent('Q4 2028');
  });

  it('renders no ancestor line when there is no ancestor journey', async () => {
    seedSpec182Db({ journeys: [] });
    renderDrawerSection(ancestors([]));

    await screen.findByText('No change planned.');
    expect(screen.queryByTestId('ancestor-journeys')).not.toBeInTheDocument();
    expect(screen.queryByText('Part of')).not.toBeInTheDocument();
  });

  it('navigates to the ancestor capability when the line is activated', async () => {
    seedSpec182Db({ journeys: [] });
    const onNavigate = renderDrawerSection(ancestors([FERRY_BOOKING_JOURNEY]));

    await userEvent.click(await screen.findByTestId('ancestor-journey-ferry-booking'));

    expect(onNavigate).toHaveBeenCalledWith('ferry-booking');
  });
});
