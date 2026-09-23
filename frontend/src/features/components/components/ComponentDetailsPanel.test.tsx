import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import { beforeEach, describe, expect, it } from 'vitest';
import type { Component, ComponentId, HATEOASLinks } from '../../../api/types';
import {
  buildComponent,
  buildExpert,
  detailGroupIds,
  fieldLabelsIn,
  renderWithProviders,
  seedDb,
  server,
} from '../../../test/helpers';
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

const FIELD_LABEL = /^(Description|Type|Ownership|Hosting|Experts|Created)$/;

function renderPanel(component: Component, extra: React.ReactNode = undefined) {
  seedDb({ components: [component], capabilities: [] });
  return renderWithProviders(<ComponentDetailsPanel componentId={component.id} viewMembership={extra} />);
}

describe('ComponentDetailsPanel', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('renders the application name above the groups and arranges the sections into the default groups', async () => {
    renderPanel(
      buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Billing Service',
        description: 'Invoices customers',
        experts: [buildExpert({ name: 'Jane' })],
        _links: editableLinks('comp-1'),
      }),
    );

    const heading = await screen.findByRole('heading', { name: 'Billing Service' });
    expect(screen.queryByText('Application Details')).not.toBeInTheDocument();
    expect(detailGroupIds()).toEqual([
      'detail-group-description',
      'detail-group-ownership',
      'detail-group-composition',
      'detail-group-metadata',
      'detail-group-realisations',
      'detail-group-origins',
      'detail-group-fit',
    ]);
    const description = screen.getByTestId('detail-group-description');
    expect(heading.compareDocumentPosition(description) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(within(description).queryByRole('heading')).not.toBeInTheDocument();
    expect(fieldLabelsIn(within(description), FIELD_LABEL)).toEqual(['Description', 'Type']);
    expect(fieldLabelsIn(within(screen.getByTestId('detail-group-ownership')), FIELD_LABEL)).toEqual([
      'Ownership',
      'Hosting',
    ]);
    expect(
      within(screen.getByTestId('detail-group-composition')).getByTestId('containment-section'),
    ).toBeInTheDocument();
    expect(fieldLabelsIn(within(screen.getByTestId('detail-group-metadata')), FIELD_LABEL)).toEqual([
      'Experts',
      'Created',
    ]);
    expect(screen.getByText('Jane', { exact: false })).toBeInTheDocument();
    expect(screen.getAllByText('Composition', { selector: 'span' })).toHaveLength(1);
  });

  it('renders an empty state in the list groups instead of omitting them', async () => {
    renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }));

    await screen.findByRole('heading', { name: 'Billing Service' });
    expect(
      within(screen.getByTestId('detail-group-realisations')).getByText('realises no capability'),
    ).toBeInTheDocument();
    expect(
      await within(screen.getByTestId('detail-group-origins')).findByText('no origin recorded'),
    ).toBeInTheDocument();
    expect(
      await within(screen.getByTestId('detail-group-fit')).findByText('no strategic pillars configured'),
    ).toBeInTheDocument();
  });

  it('renders the one-pager action and history below the groups', async () => {
    renderPanel(
      buildComponent({
        id: 'comp-1' as ComponentId,
        name: 'Billing Service',
        _links: {
          ...editableLinks('comp-1'),
          'x-one-pager': { href: '/api/v1/one-pagers/application/comp-1', method: 'GET' },
        },
      }),
    );

    const lastGroup = (await screen.findAllByTestId(/^detail-group-/)).at(-1) as HTMLElement;
    const onePager = screen.getByRole('button', { name: 'One-Pager' });
    const history = screen.getByRole('button', { name: /^History/ });
    expect(lastGroup.compareDocumentPosition(onePager) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(onePager.compareDocumentPosition(history) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it('never offers a whole-record Edit action', async () => {
    renderPanel(
      buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service', _links: editableLinks('comp-1') }),
    );

    await screen.findByRole('heading', { name: 'Billing Service' });
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
  });

  it('renames the application in place when the edit link is present', async () => {
    renderPanel(
      buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service', _links: editableLinks('comp-1') }),
    );

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
    expect(screen.queryByText('Description', { selector: 'label' })).not.toBeInTheDocument();
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

  it('shows no view group unless the host supplies view membership', async () => {
    renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }));

    await screen.findByRole('heading', { name: 'Billing Service' });
    expect(screen.queryByTestId('detail-group-view')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Remove from View' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
  });

  it('renders the view membership supplied by the host as an "In this view" group after Fit scores', async () => {
    renderPanel(
      buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }),
      <div data-testid="host-slot">colour</div>,
    );

    await screen.findByRole('heading', { name: 'Billing Service' });
    expect(detailGroupIds().slice(-2)).toEqual(['detail-group-fit', 'detail-group-view']);
    const view = within(screen.getByTestId('detail-group-view'));
    expect(view.getByRole('button', { name: 'In this view' })).toBeInTheDocument();
    expect(view.getByTestId('host-slot')).toBeInTheDocument();
  });

  it('remembers a collapsed group and a moved group across panels', async () => {
    const first = renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }));
    await screen.findByRole('heading', { name: 'Billing Service' });
    fireEvent.click(screen.getByRole('button', { name: 'Metadata' }));
    fireEvent.click(screen.getByRole('button', { name: 'Move Fit scores up' }));
    first.unmount();

    renderPanel(buildComponent({ id: 'comp-1' as ComponentId, name: 'Billing Service' }));
    await screen.findByRole('heading', { name: 'Billing Service' });
    expect(detailGroupIds().slice(-2)).toEqual(['detail-group-fit', 'detail-group-origins']);
    expect(screen.getByRole('button', { name: 'Metadata' })).toHaveAttribute('aria-expanded', 'false');
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
