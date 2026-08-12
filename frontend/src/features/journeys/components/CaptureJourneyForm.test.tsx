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
    expect(screen.getByTestId('journey-kind-description')).toHaveTextContent(/merge onto one/i);

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

  it('surfaces the cardinality error once the sources are edited below the minimum', async () => {
    const seabook = addComponent({ name: 'Seabook' });
    addComponent({ name: 'Phoenix' });
    const user = userEvent.setup();
    renderForm({ realizations: [realizationOf(seabook)] });

    await user.click(screen.getByTestId('journey-from-apps'));
    await pickFromAppOption('Seabook');

    await waitFor(() => expect(screen.getByTestId('capture-submit-error')).toHaveTextContent(/at least one/i));
    expect(screen.getByTestId('capture-journey-submit')).toBeDisabled();
  });

  it('enables submit for a carve-out with exactly one source and a target', async () => {
    const user = await renderWithSeabookSource('Carve-out');

    await user.click(screen.getByTestId('journey-to-app'));
    await pickToAppOption('Pricing Engine');

    await waitFor(() => expect(screen.getByTestId('capture-journey-submit')).not.toBeDisabled());
  });
});

describe('CaptureJourneyForm — implicit sources (specs 193 & 194)', () => {
  const KIND_SETUPS = { Move: { needsDomain: true }, Consolidation: { needsDomain: false } } as const;
  type ImplicitKind = keyof typeof KIND_SETUPS;

  function seedCatalog(names: readonly string[]) {
    return new Map(names.map((name) => [name, addComponent({ name })]));
  }

  function inToAppDropdown(name: string) {
    return screen.queryAllByText(name).some((el) => el.closest('.mantine-Select-dropdown') !== null);
  }

  async function capturedJourney(capabilityId: string | number) {
    const result = await journeyApi.getForCapability(String(capabilityId));
    return result.journey;
  }

  async function pickTarget(user: ReturnType<typeof userEvent.setup>, name: string) {
    await user.click(screen.getByTestId('journey-to-app'));
    await pickToAppOption(name);
  }

  async function submitAs(user: ReturnType<typeof userEvent.setup>, kind: ImplicitKind) {
    if (KIND_SETUPS[kind].needsDomain) {
      await user.click(screen.getByTestId('journey-target-domain'));
      await pickToAppOption('Group functions');
    }
    fireEvent.click(screen.getByTestId('capture-journey-submit'));
  }

  const captureCases: ReadonlyArray<{
    name: string;
    kind: ImplicitKind;
    realisers: readonly string[];
    target: string;
    sources: readonly string[];
  }> = [
    {
      name: 'a move onto a realiser with the other realisers as implicit sources',
      kind: 'Move',
      realisers: ['Seabook', 'Phoenix'],
      target: 'Phoenix',
      sources: ['Seabook'],
    },
    {
      name: 'a move onto the sole realiser with no sources',
      kind: 'Move',
      realisers: ['Seabook'],
      target: 'Seabook',
      sources: [],
    },
    {
      name: 'a consolidation onto a realiser with the other realisers as implicit sources',
      kind: 'Consolidation',
      realisers: ['TrackIt', 'CargoEye', 'Phoenix'],
      target: 'Phoenix',
      sources: ['TrackIt', 'CargoEye'],
    },
    {
      name: 'a consolidation onto a new application with all realisers as implicit sources',
      kind: 'Consolidation',
      realisers: ['TrackIt', 'CargoEye'],
      target: 'Unity',
      sources: ['TrackIt', 'CargoEye'],
    },
    {
      name: 'a two-realiser consolidation onto one of them with a single source',
      kind: 'Consolidation',
      realisers: ['TrackIt', 'Phoenix'],
      target: 'Phoenix',
      sources: ['TrackIt'],
    },
  ];

  for (const c of captureCases) {
    it(`captures ${c.name}`, async () => {
      const catalog = seedCatalog([...new Set([...c.realisers, c.target])]);
      seedDomains();
      const user = userEvent.setup();
      const onCaptured = vi.fn();
      const { capability } = renderForm({
        onCaptured,
        realizations: c.realisers.map((name) => realizationOf(catalog.get(name)!)),
      });

      await user.click(screen.getByRole('radio', { name: c.kind }));
      await pickTarget(user, c.target);
      await submitAs(user, c.kind);

      await waitFor(() => expect(onCaptured).toHaveBeenCalled());
      const journey = await capturedJourney(capability.id);
      expect(journey?.toApplication.componentId).toBe(catalog.get(c.target)!.id);
      expect(journey?.fromApplications.map((app) => app.componentId)).toEqual(
        c.sources.map((name) => catalog.get(name)!.id),
      );
    });
  }

  for (const kind of ['Move', 'Consolidation'] as const) {
    it(`offers current realisers as ${kind.toLowerCase()} targets`, async () => {
      const catalog = seedCatalog(['Seabook', 'Phoenix']);
      seedDomains();
      const user = userEvent.setup();
      renderForm({ realizations: [...catalog.values()].map(realizationOf) });

      await user.click(screen.getByRole('radio', { name: kind }));
      await user.click(screen.getByTestId('journey-to-app'));

      await waitFor(() => expect(inToAppDropdown('Seabook')).toBe(true));
      expect(inToAppDropdown('Phoenix')).toBe(true);
    });

    it(`recomputes the implicit sources when the ${kind.toLowerCase()} target changes`, async () => {
      const catalog = seedCatalog(['Seabook', 'Phoenix']);
      seedDomains();
      const user = userEvent.setup();
      const onCaptured = vi.fn();
      const { capability } = renderForm({ onCaptured, realizations: [...catalog.values()].map(realizationOf) });

      await user.click(screen.getByRole('radio', { name: kind }));
      await pickTarget(user, 'Phoenix');
      await pickTarget(user, 'Seabook');
      await submitAs(user, kind);

      await waitFor(() => expect(onCaptured).toHaveBeenCalled());
      const journey = await capturedJourney(capability.id);
      expect(journey?.toApplication.componentId).toBe(catalog.get('Seabook')!.id);
      expect(journey?.fromApplications.map((app) => app.componentId)).toEqual([catalog.get('Phoenix')!.id]);
    });
  }

  it('hides the from-apps picker for a consolidation', async () => {
    const catalog = seedCatalog(['Seabook', 'Phoenix']);
    const user = userEvent.setup();
    renderForm({ realizations: [...catalog.values()].map(realizationOf) });

    await user.click(screen.getByRole('radio', { name: 'Consolidation' }));

    expect(screen.queryByTestId('journey-from-apps')).not.toBeInTheDocument();
  });

  it('blocks a consolidation when the capability has fewer than two realisers', async () => {
    const catalog = seedCatalog(['Seabook', 'Phoenix']);
    const user = userEvent.setup();
    renderForm({ realizations: [realizationOf(catalog.get('Seabook')!)] });

    await user.click(screen.getByRole('radio', { name: 'Consolidation' }));

    expect(screen.getByTestId('consolidation-gate')).toHaveTextContent(/at least two/i);
    await pickTarget(user, 'Phoenix');
    expect(screen.getByTestId('capture-journey-submit')).toBeDisabled();
  });

  it('keeps selected sources excluded from targets for kinds with explicit sources', async () => {
    const catalog = seedCatalog(['Seabook', 'Phoenix']);
    const user = userEvent.setup();
    renderForm({ realizations: [realizationOf(catalog.get('Seabook')!)] });

    await user.click(screen.getByTestId('journey-to-app'));

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
