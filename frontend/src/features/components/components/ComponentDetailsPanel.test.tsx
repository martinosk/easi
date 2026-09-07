import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import { describe, expect, it } from 'vitest';
import type { Component, ComponentId, HATEOASLinks } from '../../../api/types';
import { buildComponent, buildExpert, renderWithProviders, seedDb, server } from '../../../test/helpers';
import { ComponentDetailsPanel } from './ComponentDetailsPanel';

const API_BASE = 'http://localhost:8080';

function readOnlyLinks(id: string): HATEOASLinks {
  return { self: { href: `/api/v1/components/${id}`, method: 'GET' } };
}

function editableLinks(id: string): HATEOASLinks {
  return {
    self: { href: `/api/v1/components/${id}`, method: 'GET' },
    edit: { href: `/api/v1/components/${id}`, method: 'PUT' },
    'x-add-expert': { href: `/api/v1/components/${id}/experts`, method: 'POST' },
  };
}

function renderPanel(component: Component, extra: React.ReactNode = undefined) {
  seedDb({ components: [component], capabilities: [] });
  return renderWithProviders(<ComponentDetailsPanel componentId={component.id} viewMembership={extra} />);
}

describe('ComponentDetailsPanel', () => {
  it('renders the application name as the heading and every section in order', async () => {
    renderPanel(
      buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Billing Service',
        description: 'Invoices customers',
        experts: [buildExpert({ name: 'Jane' })],
        _links: editableLinks('comp-1'),
      }),
    );

    expect(await screen.findByRole('heading', { name: 'Billing Service' })).toBeInTheDocument();
    expect(screen.queryByText('Application Details')).not.toBeInTheDocument();

    const labels = screen
      .getAllByText(/^(Description|Ownership|Hosting|Experts|Created|Type)$/)
      .map((element) => element.textContent);
    expect(labels).toEqual(['Description', 'Ownership', 'Hosting', 'Experts', 'Created', 'Type']);
    expect(screen.getByText('Jane', { exact: false })).toBeInTheDocument();
  });

  it('never offers a whole-record Edit action', async () => {
    renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service', _links: editableLinks('comp-1') }));

    await screen.findByRole('heading', { name: 'Billing Service' });
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
  });

  it('renames the application in place when the edit link is present', async () => {
    renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service', _links: editableLinks('comp-1') }));

    fireEvent.click(await screen.findByRole('button', { name: 'Edit name' }));
    fireEvent.change(screen.getByTestId('component-name-input'), { target: { value: 'Billing Platform' } });
    fireEvent.keyDown(screen.getByTestId('component-name-input'), { key: 'Enter' });

    expect(await screen.findByRole('heading', { name: 'Billing Platform' })).toBeInTheDocument();
    expect(screen.queryByTestId('component-name-input')).not.toBeInTheDocument();
  });

  it('edits the description in place and keeps the name', async () => {
    renderPanel(
      buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Billing Service',
        description: 'Old text',
        _links: editableLinks('comp-1'),
      }),
    );

    fireEvent.click(await screen.findByRole('button', { name: 'Edit description' }));
    fireEvent.change(screen.getByTestId('component-description-input'), { target: { value: 'New text' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(screen.queryByTestId('component-description-input')).not.toBeInTheDocument());
    await waitFor(() => expect(screen.getByTestId('component-description-value')).toHaveTextContent('New text'));
    expect(screen.getByRole('heading', { name: 'Billing Service' })).toBeInTheDocument();
  });

  it('invites a description when none exists and the edit link is present', async () => {
    renderPanel(
      buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Billing Service',
        description: undefined,
        _links: editableLinks('comp-1'),
      }),
    );

    expect(await screen.findByRole('button', { name: 'Add a description' })).toBeInTheDocument();
  });

  it('renders read-only when the application carries no edit links', async () => {
    renderPanel(
      buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Billing Service',
        description: undefined,
        _links: readOnlyLinks('comp-1'),
      }),
    );

    await screen.findByRole('heading', { name: 'Billing Service' });
    expect(screen.queryByRole('button', { name: 'Edit name' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Add a description' })).not.toBeInTheDocument();
    expect(screen.queryByText('Description')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '+ Add Expert' })).not.toBeInTheDocument();
    expect(screen.getByTestId('hosting-badge')).toBeInTheDocument();
  });

  it('offers Add Expert only when the application carries the x-add-expert link', async () => {
    renderPanel(
      buildComponent({
        id: 'comp-1' as ComponentId,
        experts: [buildExpert({ name: 'Jane' })],
        _links: editableLinks('comp-1'),
      }),
    );

    expect(await screen.findByRole('button', { name: '+ Add Expert' })).toBeInTheDocument();
  });

  it('shows no view-membership section unless the host supplies one', async () => {
    renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }));

    await screen.findByRole('heading', { name: 'Billing Service' });
    expect(screen.queryByText('In this view')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Remove from View' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
  });

  it('renders the view-membership section supplied by the host after the application sections', async () => {
    renderPanel(
      buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }),
      <div data-testid="host-slot">In this view</div>,
    );

    const panel = await screen.findByTestId('component-details-panel');
    const slot = within(panel).getByTestId('host-slot');
    const typeLabel = within(panel).getByText('Type');
    expect(typeLabel.compareDocumentPosition(slot) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it('loads the application through the detail query when it is not in the list cache', async () => {
    server.use(
      http.get(`${API_BASE}/api/v1/components`, () =>
        HttpResponse.json({ data: [], _links: { self: '/api/v1/components' } }),
      ),
    );
    renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }));

    expect(await screen.findByRole('heading', { name: 'Billing Service' })).toBeInTheDocument();
  });

  it('reports a failure when the application cannot be found', async () => {
    seedDb({ components: [], capabilities: [] });
    renderWithProviders(<ComponentDetailsPanel componentId="missing" />);

    await waitFor(() => expect(screen.getByText('Failed to load application')).toBeInTheDocument());
  });
});
