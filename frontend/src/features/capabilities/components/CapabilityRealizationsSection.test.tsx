import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Capability, CapabilityRealization, Component } from '../../../api/types';
import { toCapabilityId, toComponentId } from '../../../api/types';
import {
  buildCapability,
  buildCapabilityRealization,
  buildComponent,
  renderWithProviders,
} from '../../../test/helpers';
import { getCapabilityRealizations, getComponents, seedDb } from '../../../test/mocks/db';
import { CapabilityRealizationsSection } from './CapabilityRealizationsSection';

vi.mock('../../business-domains/hooks/useStrategyImportance', () => ({
  useStrategyImportanceByCapability: () => ({ data: [], isLoading: false }),
}));

vi.mock('../../business-domains/hooks/useCapabilityAssessments', () => ({
  useCapabilityAssessments: () => ({
    getAssessment: () => undefined,
    getRollup: () => undefined,
    getSuggestion: () => null,
    canAssess: false,
  }),
}));

vi.mock('../../business-domains/hooks/useCapabilityRoles', () => ({
  useCapabilityRoles: () => ({ getRole: () => undefined, canAssign: false }),
}));

const CAP_ID = toCapabilityId('cap-om');
const ADD_LINK = { href: `/api/v1/capabilities/${CAP_ID}/systems`, method: 'POST' as const };

const crm = buildComponent({ id: toComponentId('comp-crm'), name: 'CRM' });
const sap = buildComponent({ id: toComponentId('comp-sap'), name: 'SAP ERP' });
const billing = buildComponent({ id: toComponentId('comp-billing'), name: 'Billing' });

const CREATE_REALIZING_COMPONENT = {
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (realization)',
  targetType: 'component',
  relationType: 'capability-realization',
};

function capability(withAddLink: boolean): Capability {
  const base = buildCapability({ id: CAP_ID, name: 'Order Management' });
  return withAddLink ? { ...base, _links: { ...base._links, 'x-add-realization': ADD_LINK } } : base;
}

function capabilityThatCanCreateApplications(): Capability {
  const base = capability(true);
  return {
    ...base,
    _links: { ...base._links, 'x-related': [CREATE_REALIZING_COMPONENT] } as unknown as Capability['_links'],
  };
}

function realizationOf(component: Component, overrides: Partial<CapabilityRealization> = {}) {
  return buildCapabilityRealization({
    capabilityId: CAP_ID,
    componentId: component.id,
    componentName: component.name,
    ...overrides,
  });
}

function renderSection(cap: Capability, realizations: CapabilityRealization[]) {
  seedDb({ capabilities: [cap], components: [crm, sap, billing], capabilityRealizations: realizations });
  return renderWithProviders(<CapabilityRealizationsSection capability={cap} />, { withRouter: false });
}

describe('CapabilityRealizationsSection — add and remove realising applications', () => {
  beforeEach(() => {
    seedDb({});
  });

  it('adds a realising application with the chosen level through the x-add-realization link', async () => {
    const user = userEvent.setup();
    renderSection(capability(true), []);

    await user.click(await screen.findByRole('button', { name: 'Add application' }));
    await user.click(await screen.findByTestId('add-realization-application'));
    await user.click(await screen.findByRole('option', { name: 'SAP ERP', hidden: true }));
    await user.click(screen.getByTestId('add-realization-level'));
    await user.click(await screen.findByRole('option', { name: 'Partial', hidden: true }));
    await user.click(screen.getByRole('button', { name: 'Add' }));

    const row = await screen.findByText('SAP ERP');
    const rowElement = row.closest('[data-testid^="drawer-realization-"]') as HTMLElement;
    expect(within(rowElement).getByText('Partial')).toBeInTheDocument();
    expect(getCapabilityRealizations()).toHaveLength(1);
  });

  it('defaults the realization level to Full', async () => {
    const user = userEvent.setup();
    renderSection(capability(true), []);

    await user.click(await screen.findByRole('button', { name: 'Add application' }));

    expect(await screen.findByTestId('add-realization-level')).toHaveValue('Full');
  });

  it('does not offer applications that already directly realise the capability', async () => {
    const user = userEvent.setup();
    renderSection(capability(true), [realizationOf(crm), realizationOf(billing, { origin: 'Inherited', _links: {} })]);

    await user.click(await screen.findByRole('button', { name: 'Add application' }));
    await user.click(await screen.findByTestId('add-realization-application'));

    expect(await screen.findByRole('option', { name: 'SAP ERP', hidden: true })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Billing', hidden: true })).toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'CRM', hidden: true })).not.toBeInTheDocument();
  });

  it('shows no add control without the x-add-realization link', async () => {
    renderSection(capability(false), [realizationOf(crm)]);

    await screen.findByText('CRM');
    expect(screen.queryByRole('button', { name: 'Add application' })).not.toBeInTheDocument();
  });

  it('keeps the add control in the empty state when the link is present', async () => {
    renderSection(capability(true), []);

    expect(await screen.findByText('no realising application mapped')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add application' })).toBeInTheDocument();
  });

  it('removes a direct realization through its delete link after confirmation', async () => {
    const user = userEvent.setup();
    renderSection(capability(true), [realizationOf(crm)]);

    await user.click(await screen.findByRole('button', { name: 'Remove CRM' }));
    await user.click(await screen.findByRole('button', { name: 'Remove' }));

    await waitFor(() => expect(screen.queryByText('CRM')).not.toBeInTheDocument());
    expect(getCapabilityRealizations()).toHaveLength(0);
  });

  it('does not remove when the confirmation is cancelled', async () => {
    const user = userEvent.setup();
    renderSection(capability(true), [realizationOf(crm)]);

    await user.click(await screen.findByRole('button', { name: 'Remove CRM' }));
    await user.click(await screen.findByRole('button', { name: 'Cancel' }));

    expect(screen.getByText('CRM')).toBeInTheDocument();
    expect(getCapabilityRealizations()).toHaveLength(1);
  });

  it('shows no remove control on an inherited realization', async () => {
    renderSection(capability(true), [
      realizationOf(billing, { origin: 'Inherited', sourceCapabilityName: 'Invoicing', _links: {} }),
    ]);

    expect(await screen.findByText('Inherited from Invoicing')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Remove Billing' })).not.toBeInTheDocument();
  });

  it('shows no remove control on a direct realization without a delete link', async () => {
    const readOnly = realizationOf(crm);
    renderSection(capability(false), [{ ...readOnly, _links: { self: readOnly._links.self } }]);

    await screen.findByText('CRM');
    expect(screen.queryByRole('button', { name: 'Remove CRM' })).not.toBeInTheDocument();
  });

  it('creates a new application and adds it as realising with the chosen level', async () => {
    const user = userEvent.setup();
    renderSection(capabilityThatCanCreateApplications(), []);

    await user.click(await screen.findByRole('button', { name: 'Add application' }));
    await user.click(await screen.findByTestId('add-realization-level'));
    await user.click(await screen.findByRole('option', { name: 'Planned', hidden: true }));
    await user.click(screen.getByRole('button', { name: 'Create new application' }));
    await user.type(await screen.findByTestId('component-name-input'), 'Fleet Planner');
    await user.click(screen.getByTestId('create-component-submit'));

    const row = await screen.findByText('Fleet Planner');
    const rowElement = row.closest('[data-testid^="drawer-realization-"]') as HTMLElement;
    expect(within(rowElement).getByText('Planned')).toBeInTheDocument();
    expect(getComponents().map((c) => c.name)).toContain('Fleet Planner');
    expect(getCapabilityRealizations()).toHaveLength(1);
    await waitFor(() => expect(screen.queryByTestId('create-component-dialog')).not.toBeInTheDocument());
    expect(screen.queryByTestId('add-realization-application')).not.toBeInTheDocument();
  });

  it('offers no create-new control without the capability-realization related link', async () => {
    const user = userEvent.setup();
    renderSection(capability(true), []);

    await user.click(await screen.findByRole('button', { name: 'Add application' }));

    await screen.findByTestId('add-realization-application');
    expect(screen.queryByRole('button', { name: 'Create new application' })).not.toBeInTheDocument();
  });
});
