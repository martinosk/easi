import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId, toRealizationId } from '../../../api/types';
import { buildCapabilityRealization, renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { addComponent } from '../../../test/mocks/db';
import { buildStubJourney } from '../../../test/mocks/spec182/builders';
import { type StubJourney, seedSpec182Db } from '../../../test/mocks/spec182/store';
import { NO_HIERARCHY_JOURNEYS } from '../lens/hierarchyJourneys';
import { JourneySection } from './JourneySection';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function renderSection(realizationNames: Record<string, string> = {}) {
  const capability = cap('cap-1', 'Booking management', 'L2');
  const realizations = Object.entries(realizationNames).map(([componentId, componentName], index) => {
    addComponent({ id: toComponentId(componentId), name: componentName });
    return buildCapabilityRealization({
      id: toRealizationId(`real-${index + 1}`),
      componentId: toComponentId(componentId),
      componentName,
    });
  });
  return renderWithProviders(
    <JourneySection
      capability={capability}
      realizations={realizations}
      hierarchyJourneys={NO_HIERARCHY_JOURNEYS}
      onNavigateToCapability={vi.fn()}
    />,
    { withRouter: false },
  );
}

function seedActive(overrides: Partial<StubJourney> = {}) {
  seedSpec182Db({ canWrite: true, journeys: [buildStubJourney(overrides)] });
}

describe('JourneySection — status transitions', () => {
  it('starts a planned journey', async () => {
    seedActive();
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('start-journey-btn'));

    await waitFor(() => expect(screen.getByTestId('journey-status')).toHaveTextContent('in flight'));
  });

  it('completes an in-flight journey', async () => {
    seedActive({ status: 'in-flight', progress: 60 });
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('complete-journey-btn'));

    await waitFor(() => expect(screen.getByTestId('journey-status')).toHaveTextContent('done (100%)'));
  });

  it('abandons a journey only after confirmation', async () => {
    seedActive();
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('abandon-journey-btn'));

    const dialog = await screen.findByTestId('confirmation-dialog');
    expect(screen.getByTestId('journey-status')).toHaveTextContent('planned');

    await user.click(within(dialog).getByRole('button', { name: 'Abandon' }));

    await waitFor(() => expect(screen.getByTestId('journey-status')).toHaveTextContent('abandoned'));
  });

  it('keeps the journey untouched when abandon is cancelled', async () => {
    seedActive();
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('abandon-journey-btn'));
    const dialog = await screen.findByTestId('confirmation-dialog');
    await user.click(within(dialog).getByRole('button', { name: 'Cancel' }));

    expect(screen.queryByTestId('confirmation-dialog')).not.toBeInTheDocument();
    expect(screen.getByTestId('journey-status')).toHaveTextContent('planned');
  });
});

describe('JourneySection — progress editor', () => {
  it('updates progress through the progress editor', async () => {
    seedActive({ status: 'in-flight', progress: 20 });
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('update-progress-btn'));
    const input = screen.getByTestId('progress-input');
    await user.clear(input);
    await user.type(input, '60');
    await user.click(screen.getByTestId('save-progress-btn'));

    await waitFor(() => expect(screen.getByTestId('journey-status')).toHaveTextContent('in flight (60%)'));
  });
});

describe('JourneySection — details editing', () => {
  it('updates the note through the details form', async () => {
    seedActive();
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('edit-journey-btn'));
    const note = screen.getByTestId('details-note-input');
    await user.clear(note);
    await user.type(note, 'Revised plan summary');
    await user.click(screen.getByTestId('save-details-btn'));

    await waitFor(() => expect(screen.getByTestId('journey-note')).toHaveTextContent('Revised plan summary'));
  });
});

describe('JourneySection — milestones', () => {
  it('adds a milestone', async () => {
    seedActive();
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('add-milestone-btn'));
    await user.type(screen.getByTestId('milestone-label-input'), 'Phoenix booking API live');
    await user.click(screen.getByTestId('save-milestone-btn'));

    await waitFor(() => expect(screen.getByText('Phoenix booking API live')).toBeInTheDocument());
  });

  it('marks a milestone done through the edit dialog', async () => {
    seedActive({
      milestones: [
        { id: 'ms-1', label: 'North Sea routes', targetPeriod: { year: 2026, quarter: 4 }, status: 'in-flight' },
      ],
    });
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('edit-milestone-btn-ms-1'));
    const dialog = await screen.findByTestId('milestone-dialog');
    await user.click(within(dialog).getByRole('radio', { name: 'Done' }));
    await user.click(screen.getByTestId('save-milestone-btn'));

    await waitFor(() => expect(screen.getByTestId('milestone-when-ms-1')).toHaveTextContent('Done · Q4 2026'));
  });

  it('removes a milestone', async () => {
    seedActive({
      milestones: [{ id: 'ms-1', label: 'North Sea routes', targetPeriod: null, status: 'planned' }],
    });
    const user = userEvent.setup();
    renderSection();

    await user.click(await screen.findByTestId('remove-milestone-btn-ms-1'));

    await waitFor(() => expect(screen.queryByTestId('milestone-row-ms-1')).not.toBeInTheDocument());
  });
});

describe('JourneySection — change sources', () => {
  async function openSourcesAndPickCapacity(overrides: Partial<StubJourney> = {}) {
    seedActive(overrides);
    const user = userEvent.setup();
    renderSection({ 'comp-capacity': 'CapacityMgmt' });

    await user.click(await screen.findByTestId('change-sources-btn'));
    await user.click(screen.getByTestId('change-sources-input'));
    const option = (await screen.findAllByText('CapacityMgmt')).find((el) =>
      el.closest('.mantine-MultiSelect-dropdown'),
    );
    if (!option) throw new Error('CapacityMgmt option not found');
    await user.click(option);
    return user;
  }

  it('replaces the source applications', async () => {
    const user = await openSourcesAndPickCapacity();

    await user.click(screen.getByTestId('save-sources-btn'));

    await waitFor(() => expect(screen.getByTestId('journey-from-to')).toHaveTextContent('CapacityMgmt'));
  });

  it('blocks a source set violating the kind cardinality', async () => {
    await openSourcesAndPickCapacity({ kind: 'carve-out' });

    expect(await screen.findByTestId('change-sources-error')).toHaveTextContent(/exactly one/i);
    expect(screen.getByTestId('save-sources-btn')).toBeDisabled();
  });
});
