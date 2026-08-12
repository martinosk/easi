import { fireEvent, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { HttpResponse, http } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import type { Component } from '../../../api/types';
import { buildCapabilityRealization, renderWithProviders } from '../../../test/helpers';
import { addCapability, addComponent } from '../../../test/mocks/db';
import { server } from '../../../test/mocks/server';
import { journeyApi } from '../api/journeyApi';
import { CaptureJourneyForm } from './CaptureJourneyForm';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function seedDomains() {
  server.use(
    http.get('*/api/v1/business-domains', () => {
      return HttpResponse.json({
        data: [{ id: 'domain-gf', name: 'Group functions', description: '', capabilityCount: 0, createdAt: '2026-01-01T00:00:00Z', _links: {} }],
        _links: { self: '/api/v1/business-domains' },
      });
    }),
  );
}

function renderForm(overrides: Partial<Parameters<typeof CaptureJourneyForm>[0]> = {}) {
  const capability = addCapability({ name: 'Booking management' });
  const props = {
    capability,
    realizations: [],
    onCaptured: vi.fn(),
    onCancel: vi.fn(),
    ...overrides,
  };
  return { ...renderWithProviders(<CaptureJourneyForm {...props} />, { withRouter: false }), capability };
}

async function pickOption(dropdownClass: string, name: string) {
  const matches = await screen.findAllByText(name);
  const target = matches.find((el) => el.closest(`.${dropdownClass}`));
  if (!target) throw new Error(`Option "${name}" not found within .${dropdownClass}`);
  await userEvent.click(target);
}

const pickFromAppOption = (name: string) => pickOption('mantine-MultiSelect-dropdown', name);
const pickToAppOption = (name: string) => pickOption('mantine-Select-dropdown', name);

function realizationOf(component: Component) {
  return buildCapabilityRealization({ componentId: component.id, componentName: component.name });
}

describe('CaptureJourneyForm — kind-driven fields', () => {
  it('defaults to migration and shows the from-apps picker', () => {
    renderForm();
    expect(screen.getByTestId('journey-kind')).toBeInTheDocument();
    expect(screen.getByTestId('journey-from-apps')).toBeInTheDocument();
  });

  it('hides the from-apps picker for a move (implicit sources)', async () => {
    const user = userEvent.setup();
    renderForm();

    await user.click(screen.getByRole('radio', { name: 'Move' }));

    expect(screen.queryByTestId('journey-from-apps')).not.toBeInTheDocument();
  });

  it('shows move-only fields (domain, parent, resulting name) only for a move', async () => {
    seedDomains();
    const user = userEvent.setup();
    const { capability } = renderForm();

    expect(screen.queryByTestId('journey-target-domain')).not.toBeInTheDocument();

    await user.click(screen.getByRole('radio', { name: 'Move' }));

    expect(screen.getByTestId('journey-target-domain')).toBeInTheDocument();
    expect(screen.getByTestId('journey-target-parent')).toBeInTheDocument();
    expect(screen.getByTestId('journey-resulting-name')).toHaveValue(capability.name);
  });

  it('explains the selected kind and updates the description when the kind changes', async () => {
    const user = userEvent.setup();
    renderForm();

    expect(screen.getByTestId('journey-kind-description')).toHaveTextContent(/at least one source/i);

    await user.click(screen.getByRole('radio', { name: 'Consolidation' }));
    expect(screen.getByTestId('journey-kind-description')).toHaveTextContent(/at least two source/i);

    await user.click(screen.getByRole('radio', { name: 'Carve-out' }));
    expect(screen.getByTestId('journey-kind-description')).toHaveTextContent(/exactly one source/i);

    await user.click(screen.getByRole('radio', { name: 'Move' }));
    expect(screen.getByTestId('journey-kind-description')).toHaveTextContent(/relocates/i);
  });

  it('prefills From applications with the current realisations from the start (not only after a Move round-trip)', () => {
    const seabook = addComponent({ name: 'Seabook' });
    const phoenix = addComponent({ name: 'Phoenix' });
    renderForm({ realizations: [realizationOf(seabook), realizationOf(phoenix)] });

    const isSelectedPill = (name: string) =>
      screen.getAllByText(name).some((el) => el.closest('[class*="Pill"]') !== null);
    expect(isSelectedPill('Seabook')).toBe(true);
    expect(isSelectedPill('Phoenix')).toBe(true);
  });
});

describe('CaptureJourneyForm — cardinality validation (rule 3)', () => {
  it('disables submit for a migration with zero sources', () => {
    addComponent({ name: 'Seabook' });
    renderForm({ realizations: [] });
    expect(screen.getByTestId('capture-journey-submit')).toBeDisabled();
  });

  async function renderWithSeabookSource(kindLabel: string) {
    const seabook = addComponent({ name: 'Seabook' });
    addComponent({ name: 'Pricing Engine' });
    const user = userEvent.setup();
    renderForm({ realizations: [realizationOf(seabook)] });
    await user.click(screen.getByRole('radio', { name: kindLabel }));
    return user;
  }

  it('disables submit for a consolidation with too few prefilled sources', async () => {
    await renderWithSeabookSource('Consolidation');

    expect(screen.getByTestId('capture-journey-submit')).toBeDisabled();
  });

  it('surfaces the cardinality error once the sources are edited below the minimum', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    const phoenix = addComponent({ name: 'Phoenix' });
    const user = userEvent.setup();
    renderForm({ realizations: [realizationOf(seabook), realizationOf(phoenix)] });

    await user.click(screen.getByRole('radio', { name: 'Consolidation' }));
    await user.click(screen.getByTestId('journey-from-apps'));
    await pickFromAppOption('Seabook');

    await waitFor(() => expect(screen.getByTestId('capture-submit-error')).toHaveTextContent(/at least two/i));
    expect(screen.getByTestId('capture-journey-submit')).toBeDisabled();
  });

  it('enables submit for a carve-out with exactly one source and a target', async () => {
    const user = await renderWithSeabookSource('Carve-out');

    await user.click(screen.getByTestId('journey-to-app'));
    await pickToAppOption('Pricing Engine');

    await waitFor(() => expect(screen.getByTestId('capture-journey-submit')).not.toBeDisabled());
  });
});

describe('CaptureJourneyForm — move target among realisations (spec 193)', () => {
  async function captureMove(target: string) {
    const user = userEvent.setup();
    await user.click(screen.getByRole('radio', { name: 'Move' }));
    await user.click(screen.getByTestId('journey-to-app'));
    await pickToAppOption(target);
    await user.click(screen.getByTestId('journey-target-domain'));
    await pickToAppOption('Group functions');
    fireEvent.click(screen.getByTestId('capture-journey-submit'));
  }

  async function capturedJourney(capabilityId: string | number) {
    const result = await journeyApi.getForCapability(String(capabilityId));
    return result.journey;
  }

  it('offers current realisers as move targets', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    const phoenix = addComponent({ name: 'Phoenix' });
    seedDomains();
    const user = userEvent.setup();
    renderForm({ realizations: [realizationOf(seabook), realizationOf(phoenix)] });

    await user.click(screen.getByRole('radio', { name: 'Move' }));
    await user.click(screen.getByTestId('journey-to-app'));

    const inToAppDropdown = (name: string) =>
      screen.queryAllByText(name).some((el) => el.closest('.mantine-Select-dropdown') !== null);
    await waitFor(() => expect(inToAppDropdown('Seabook')).toBe(true));
    expect(inToAppDropdown('Phoenix')).toBe(true);
  });

  it('captures a move onto a realiser with the other realisers as implicit sources', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    const phoenix = addComponent({ name: 'Phoenix' });
    seedDomains();
    const onCaptured = vi.fn();
    const { capability } = renderForm({ onCaptured, realizations: [realizationOf(seabook), realizationOf(phoenix)] });

    await captureMove('Phoenix');

    await waitFor(() => expect(onCaptured).toHaveBeenCalled());
    const journey = await capturedJourney(capability.id);
    expect(journey?.toApplication.componentId).toBe(phoenix.id);
    expect(journey?.fromApplications.map((app) => app.componentId)).toEqual([seabook.id]);
  });

  it('captures a move onto the sole realiser with no sources', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    seedDomains();
    const onCaptured = vi.fn();
    const { capability } = renderForm({ onCaptured, realizations: [realizationOf(seabook)] });

    await captureMove('Seabook');

    await waitFor(() => expect(onCaptured).toHaveBeenCalled());
    const journey = await capturedJourney(capability.id);
    expect(journey?.toApplication.componentId).toBe(seabook.id);
    expect(journey?.fromApplications).toEqual([]);
  });

  it('recomputes the implicit sources when the move target changes', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    const phoenix = addComponent({ name: 'Phoenix' });
    seedDomains();
    const user = userEvent.setup();
    const onCaptured = vi.fn();
    const { capability } = renderForm({ onCaptured, realizations: [realizationOf(seabook), realizationOf(phoenix)] });

    await user.click(screen.getByRole('radio', { name: 'Move' }));
    await user.click(screen.getByTestId('journey-to-app'));
    await pickToAppOption('Phoenix');
    await user.click(screen.getByTestId('journey-to-app'));
    await pickToAppOption('Seabook');
    await user.click(screen.getByTestId('journey-target-domain'));
    await pickToAppOption('Group functions');
    fireEvent.click(screen.getByTestId('capture-journey-submit'));

    await waitFor(() => expect(onCaptured).toHaveBeenCalled());
    const journey = await capturedJourney(capability.id);
    expect(journey?.toApplication.componentId).toBe(seabook.id);
    expect(journey?.fromApplications.map((app) => app.componentId)).toEqual([phoenix.id]);
  });

  it('keeps selected sources excluded from targets for non-move kinds', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    addComponent({ name: 'Phoenix' });
    const user = userEvent.setup();
    renderForm({ realizations: [realizationOf(seabook)] });

    await user.click(screen.getByTestId('journey-to-app'));

    const inToAppDropdown = (name: string) =>
      screen.queryAllByText(name).some((el) => el.closest('.mantine-Select-dropdown') !== null);
    await waitFor(() => expect(inToAppDropdown('Phoenix')).toBe(true));
    expect(inToAppDropdown('Seabook')).toBe(false);
  });
});

describe('CaptureJourneyForm — capture', () => {
  it('captures a migration journey and calls onCaptured', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    const phoenix = addComponent({ name: 'Phoenix' });
    const user = userEvent.setup();
    const onCaptured = vi.fn();
    const { capability } = renderForm({ onCaptured, realizations: [realizationOf(seabook)] });

    await user.click(screen.getByTestId('journey-to-app'));
    await pickToAppOption('Phoenix');

    fireEvent.click(screen.getByTestId('capture-journey-submit'));

    await waitFor(() => expect(onCaptured).toHaveBeenCalled());
    const result = await journeyApi.getForCapability(String(capability.id));
    expect(result.journey?.kind).toBe('migration');
    expect(result.journey?.toApplication.componentId).toBe(phoenix.id);
  });
});
