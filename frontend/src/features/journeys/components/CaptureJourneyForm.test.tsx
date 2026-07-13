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
    await user.click(screen.getByTestId('journey-from-apps'));
    await pickFromAppOption('Seabook');
    return user;
  }

  it('disables submit for a consolidation with only one source selected', async () => {
    await renderWithSeabookSource('Consolidation');

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

describe('CaptureJourneyForm — capture', () => {
  it('captures a migration journey and calls onCaptured', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    const phoenix = addComponent({ name: 'Phoenix' });
    const user = userEvent.setup();
    const onCaptured = vi.fn();
    const { capability } = renderForm({ onCaptured, realizations: [realizationOf(seabook)] });

    await user.click(screen.getByTestId('journey-from-apps'));
    await pickFromAppOption('Seabook');
    await user.click(screen.getByTestId('journey-to-app'));
    await pickToAppOption('Phoenix');

    fireEvent.click(screen.getByTestId('capture-journey-submit'));

    await waitFor(() => expect(onCaptured).toHaveBeenCalled());
    const result = await journeyApi.getForCapability(String(capability.id));
    expect(result.journey?.kind).toBe('migration');
    expect(result.journey?.toApplication.componentId).toBe(phoenix.id);
  });
});
