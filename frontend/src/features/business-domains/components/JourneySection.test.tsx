import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildStubJourney } from '../../../test/mocks/spec182/builders';
import { seedSpec182Db } from '../../../test/mocks/spec182/store';
import { NO_HIERARCHY_JOURNEYS } from '../lens/hierarchyJourneys';
import { JourneySection } from './JourneySection';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function renderSection() {
  const capability = cap('cap-1', 'Booking management', 'L2');
  return renderWithProviders(
    <JourneySection
      capability={capability}
      realizations={[]}
      hierarchyJourneys={NO_HIERARCHY_JOURNEYS}
      onNavigateToCapability={vi.fn()}
    />,
    { withRouter: false },
  );
}

describe('JourneySection — no journey', () => {
  it('shows "No change planned." and a Plan journey button when the caller can capture', async () => {
    seedSpec182Db({ canWrite: true });
    renderSection();

    expect(await screen.findByText('No change planned.')).toBeInTheDocument();
    expect(screen.getByTestId('plan-journey-btn')).toBeInTheDocument();
  });

  it('shows no capture affordance for a read-only caller', async () => {
    seedSpec182Db({ canWrite: false });
    renderSection();

    expect(await screen.findByText('No change planned.')).toBeInTheDocument();
    expect(screen.queryByTestId('plan-journey-btn')).not.toBeInTheDocument();
  });

  it('opens the capture form when Plan journey is clicked', async () => {
    seedSpec182Db({ canWrite: true });
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('plan-journey-btn'));

    expect(screen.getByTestId('capture-journey-form')).toBeInTheDocument();
  });

  it('renders the latest terminal journey as history beneath the capture affordance', async () => {
    seedSpec182Db({
      canWrite: true,
      journeys: [buildStubJourney({ status: 'done', progress: 100, completedAt: '2026-06-01T00:00:00Z' })],
    });
    renderSection();

    expect(await screen.findByTestId('journey-status')).toHaveTextContent('done (100%)');
    expect(screen.getByTestId('plan-journey-btn')).toBeInTheDocument();
  });
});

describe('JourneySection — transition table (mockup-literal)', () => {
  it('renders type, from → to, status, and target date for a planned migration', async () => {
    seedSpec182Db({ journeys: [buildStubJourney()] });
    renderSection();

    expect(await screen.findByTestId('journey-type')).toHaveTextContent('migration');
    expect(screen.getByTestId('journey-from-to')).toHaveTextContent('Seabook → Phoenix');
    expect(screen.getByTestId('journey-status')).toHaveTextContent('planned');
    expect(screen.getByTestId('journey-target-date')).toHaveTextContent('Q2 2027');
  });

  it('joins multiple sources with " + " for a consolidation', async () => {
    seedSpec182Db({
      journeys: [
        buildStubJourney({
          kind: 'consolidation',
          fromApplications: [
            { componentId: 'comp-mail', componentName: 'MailBlast', stale: false },
            { componentId: 'comp-sms', componentName: 'SMS-GW', stale: false },
          ],
          toApplication: { componentId: 'comp-hub', componentName: 'Notify Hub', stale: false },
        }),
      ],
    });
    renderSection();

    expect(await screen.findByTestId('journey-from-to')).toHaveTextContent('MailBlast + SMS-GW → Notify Hub');
  });

  it('shows "in flight (60%)" and an amber progress bar for an in-flight journey', async () => {
    seedSpec182Db({ journeys: [buildStubJourney({ status: 'in-flight', progress: 60 })] });
    renderSection();

    expect(await screen.findByTestId('journey-status')).toHaveTextContent('in flight (60%)');
    expect(screen.getByTestId('journey-progress-bar')).toBeInTheDocument();
    expect(screen.getByTestId('journey-progress-fill')).toHaveAttribute('data-fill', 'in-flight');
  });

  it('shows "done (100%)" and a green full progress bar for a done journey', async () => {
    seedSpec182Db({ journeys: [buildStubJourney({ status: 'done', progress: 60 })], canWrite: false });
    renderSection();

    expect(await screen.findByTestId('journey-status')).toHaveTextContent('done (100%)');
    expect(screen.getByTestId('journey-progress-fill')).toHaveAttribute('data-fill', 'done');
  });

  it('shows "New home" and "Target app" rows for a move instead of From → to', async () => {
    seedSpec182Db({
      journeys: [
        buildStubJourney({
          kind: 'move',
          fromApplications: [],
          toApplication: { componentId: 'comp-sap', componentName: 'SAP S/4', stale: false },
          move: {
            targetDomainId: 'domain-gf',
            targetDomainName: 'Group functions',
            targetDomainStale: false,
            targetParentId: null,
            targetParentName: '',
            targetParentStale: false,
            resultingName: 'Freight invoicing',
          },
        }),
      ],
    });
    renderSection();

    expect(await screen.findByTestId('journey-type')).toHaveTextContent('capability move');
    expect(screen.getByTestId('journey-new-home')).toHaveTextContent('Group functions → Freight invoicing');
    expect(screen.getByTestId('journey-target-app')).toHaveTextContent('SAP S/4');
    expect(screen.queryByTestId('journey-from-to')).not.toBeInTheDocument();
  });

  it('renders the plan summary note', async () => {
    seedSpec182Db({ journeys: [buildStubJourney()] });
    renderSection();

    expect(await screen.findByTestId('journey-note')).toHaveTextContent('Route-by-route migration');
    expect(screen.getByText('Plan summary')).toBeInTheDocument();
  });
});

describe('JourneySection — stale references (rule 11)', () => {
  it('marks stale from- and to-apps with a badge but renders the journey normally', async () => {
    seedSpec182Db({
      journeys: [
        buildStubJourney({
          fromApplications: [{ componentId: 'comp-seabook', componentName: 'Seabook', stale: true }],
          toApplication: { componentId: 'comp-phoenix', componentName: 'Phoenix', stale: true },
        }),
      ],
    });
    renderSection();

    expect(await screen.findByTestId('journey-from-to')).toHaveTextContent('Seabook');
    expect(screen.getAllByTestId('journey-stale-badge')).toHaveLength(2);
    expect(screen.getByTestId('journey-status')).toHaveTextContent('planned');
  });

  it('marks a stale move destination domain with a badge', async () => {
    seedSpec182Db({
      journeys: [
        buildStubJourney({
          kind: 'move',
          fromApplications: [],
          move: {
            targetDomainId: 'domain-gf',
            targetDomainName: 'Group functions',
            targetDomainStale: true,
            targetParentId: null,
            targetParentName: '',
            targetParentStale: false,
            resultingName: 'Freight invoicing',
          },
        }),
      ],
    });
    renderSection();

    expect(await screen.findByTestId('journey-new-home')).toHaveTextContent('Group functions');
    expect(screen.getAllByTestId('journey-stale-badge')).toHaveLength(1);
  });
});

describe('JourneySection — milestones (mockup-literal wording)', () => {
  it('renders milestone dots, labels and when-wording', async () => {
    seedSpec182Db({
      journeys: [
        buildStubJourney({
          milestones: [
            { id: 'ms-1', label: 'Phoenix booking API live', targetPeriod: { year: 2025, quarter: 4 }, status: 'done' },
            { id: 'ms-2', label: 'North Sea routes', targetPeriod: { year: 2026, quarter: 4 }, status: 'in-flight' },
            { id: 'ms-3', label: 'Seabook decommissioned', targetPeriod: null, status: 'planned' },
          ],
        }),
      ],
    });
    renderSection();

    expect(await screen.findByText('Milestones')).toBeInTheDocument();
    expect(screen.getByTestId('milestone-when-ms-1')).toHaveTextContent('Done · Q4 2025');
    expect(screen.getByTestId('milestone-when-ms-2')).toHaveTextContent('Q4 2026');
    expect(screen.getByTestId('milestone-dot-ms-1')).toHaveAttribute('data-status', 'done');
    expect(screen.getByTestId('milestone-dot-ms-2')).toHaveAttribute('data-status', 'in-flight');
    expect(screen.getByTestId('milestone-dot-ms-3')).toHaveAttribute('data-status', 'planned');
  });
});

describe('JourneySection — affordances by status and permission', () => {
  it('shows Start and Abandon for a writer on a planned journey', async () => {
    seedSpec182Db({ canWrite: true, journeys: [buildStubJourney()] });
    renderSection();

    expect(await screen.findByTestId('start-journey-btn')).toBeInTheDocument();
    expect(screen.getByTestId('abandon-journey-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('complete-journey-btn')).not.toBeInTheDocument();
  });

  it('shows Complete and Abandon plus editors for a writer on an in-flight journey', async () => {
    seedSpec182Db({ canWrite: true, journeys: [buildStubJourney({ status: 'in-flight', progress: 20 })] });
    renderSection();

    expect(await screen.findByTestId('complete-journey-btn')).toBeInTheDocument();
    expect(screen.getByTestId('abandon-journey-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('start-journey-btn')).not.toBeInTheDocument();
    expect(screen.getByTestId('update-progress-btn')).toBeInTheDocument();
    expect(screen.getByTestId('edit-journey-btn')).toBeInTheDocument();
    expect(screen.getByTestId('change-sources-btn')).toBeInTheDocument();
    expect(screen.getByTestId('add-milestone-btn')).toBeInTheDocument();
  });

  it('shows zero action buttons on a done journey even for a writer', async () => {
    seedSpec182Db({ canWrite: true, journeys: [buildStubJourney({ status: 'done' })] });
    renderSection();

    await screen.findByTestId('journey-status');
    for (const id of [
      'start-journey-btn',
      'complete-journey-btn',
      'abandon-journey-btn',
      'update-progress-btn',
      'edit-journey-btn',
      'change-sources-btn',
      'add-milestone-btn',
    ]) {
      expect(screen.queryByTestId(id)).not.toBeInTheDocument();
    }
  });

  it('shows an abandoned journey frozen with zero action buttons', async () => {
    seedSpec182Db({ canWrite: true, journeys: [buildStubJourney({ status: 'abandoned' })] });
    renderSection();

    expect(await screen.findByTestId('journey-status')).toHaveTextContent('abandoned');
    expect(screen.queryByTestId('start-journey-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('abandon-journey-btn')).not.toBeInTheDocument();
  });

  it('shows zero buttons for a read-only caller on an active journey', async () => {
    seedSpec182Db({
      canWrite: false,
      journeys: [
        buildStubJourney({
          status: 'in-flight',
          progress: 60,
          milestones: [{ id: 'ms-1', label: 'API live', targetPeriod: null, status: 'planned' }],
        }),
      ],
    });
    renderSection();

    await screen.findByTestId('journey-status');
    expect(screen.queryAllByRole('button')).toHaveLength(0);
  });
});
