import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import { describe, expect, it } from 'vitest';
import type { Capability, CapabilityId, HATEOASLinks } from '../../../api/types';
import { metadataQueryKeys } from '../../../lib/appQueryKeys';
import { buildCapability, buildExpert, renderWithProviders, seedDb, server } from '../../../test/helpers';
import { CapabilityDetailsPanel } from './CapabilityDetailsPanel';

const API_BASE = 'http://localhost:8080';

function readOnlyLinks(id: string): HATEOASLinks {
  return { self: { href: `/api/v1/capabilities/${id}`, method: 'GET' } };
}

function editableLinks(id: string): HATEOASLinks {
  return {
    self: { href: `/api/v1/capabilities/${id}`, method: 'GET' },
    edit: { href: `/api/v1/capabilities/${id}`, method: 'PUT' },
    'x-update-metadata': { href: `/api/v1/capabilities/${id}/metadata`, method: 'PUT' },
    'x-add-tag': { href: `/api/v1/capabilities/${id}/tags`, method: 'POST' },
    'x-add-expert': { href: `/api/v1/capabilities/${id}/experts`, method: 'POST' },
  };
}

async function findSelectOption(label: string): Promise<Element> {
  return waitFor(() => {
    const option = Array.from(document.querySelectorAll('[data-combobox-option]')).find(
      (element) => element.textContent === label,
    );
    if (!option) throw new Error(`option ${label} not rendered`);
    return option;
  });
}

function fullCapability(links: HATEOASLinks): Capability {
  return buildCapability({
    id: 'cap-1' as CapabilityId,
    name: 'Order Management',
    description: 'Takes and tracks orders',
    level: 'L2',
    status: 'Active',
    maturityValue: 30,
    ownershipModel: 'TeamOwned',
    primaryOwner: 'Sales Platform Team',
    eaOwner: 'user-1',
    eaOwnerName: 'Alice Smith',
    tags: ['core'],
    experts: [buildExpert({ name: 'Jane' })],
    _links: links,
  });
}

interface RenderOptions {
  viewMembership?: React.ReactNode;
  domainContext?: React.ReactNode;
}

function renderPanel(capability: Capability, slots: RenderOptions = {}) {
  seedDb({ capabilities: [capability], components: [] });
  return renderWithProviders(<CapabilityDetailsPanel capabilityId={capability.id} {...slots} />);
}

describe('CapabilityDetailsPanel', () => {
  it('renders the capability name as the heading and every section in order', async () => {
    renderPanel(fullCapability(editableLinks('cap-1')));

    expect(await screen.findByRole('heading', { name: 'Order Management' })).toBeInTheDocument();
    expect(screen.queryByText('Capability Details')).not.toBeInTheDocument();

    const labels = screen
      .getAllByText(
        /^(Description|Level|Status|Maturity|Ownership Model|Primary Owner|EA Owner|Tags|Experts|Created|Realising applications)$/,
      )
      .map((element) => element.textContent);
    expect(labels).toEqual([
      'Description',
      'Level',
      'Status',
      'Maturity',
      'Ownership Model',
      'Primary Owner',
      'EA Owner',
      'Tags',
      'Experts',
      'Created',
      'Realising applications',
    ]);
    expect(screen.getByText('Jane', { exact: false })).toBeInTheDocument();
    expect(screen.getByText('Alice Smith')).toBeInTheDocument();
    expect(screen.queryByText('user-1')).not.toBeInTheDocument();
  });

  it('never offers a whole-record Edit action', async () => {
    renderPanel(fullCapability(editableLinks('cap-1')));

    await screen.findByRole('heading', { name: 'Order Management' });
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
  });

  it('renames the capability in place through the edit link, sending the description unchanged', async () => {
    let received: unknown;
    server.use(
      http.put(`${API_BASE}/api/v1/capabilities/cap-1`, async ({ request }) => {
        received = await request.json();
        return HttpResponse.json({ ...fullCapability(editableLinks('cap-1')), name: 'Order Handling' });
      }),
    );
    renderPanel(fullCapability(editableLinks('cap-1')));

    fireEvent.click(await screen.findByRole('button', { name: 'Edit name' }));
    fireEvent.change(screen.getByTestId('capability-name-input'), { target: { value: 'Order Handling' } });
    fireEvent.keyDown(screen.getByTestId('capability-name-input'), { key: 'Enter' });

    await waitFor(() => expect(received).toEqual({ name: 'Order Handling', description: 'Takes and tracks orders' }));
    await waitFor(() => expect(screen.queryByTestId('capability-name-input')).not.toBeInTheDocument());
  });

  it('shows an add-description prompt only when the edit link is present', async () => {
    const capability = { ...fullCapability(editableLinks('cap-1')), description: undefined };
    const { unmount } = renderPanel(capability);
    expect(await screen.findByRole('button', { name: 'Add a description' })).toBeInTheDocument();
    unmount();

    renderPanel({ ...capability, _links: readOnlyLinks('cap-1') });
    await screen.findByRole('heading', { name: 'Order Management' });
    expect(screen.queryByRole('button', { name: 'Add a description' })).not.toBeInTheDocument();
    expect(screen.queryByText('Description')).not.toBeInTheDocument();
  });

  it('changes status in place through x-update-metadata, sending the other metadata unchanged', async () => {
    let received: unknown;
    server.use(
      http.put(`${API_BASE}/api/v1/capabilities/cap-1/metadata`, async ({ request }) => {
        received = await request.json();
        return HttpResponse.json({ ...fullCapability(editableLinks('cap-1')), status: 'Inactive' });
      }),
    );
    const { queryClient } = renderPanel(fullCapability(editableLinks('cap-1')));

    fireEvent.click(await screen.findByRole('button', { name: 'Edit status' }));
    await waitFor(() => expect(queryClient.getQueryState(metadataQueryKeys.statuses())?.status).toBe('success'));
    fireEvent.click(screen.getByTestId('capability-status-input'));
    fireEvent.click(await findSelectOption('Inactive'));
    fireEvent.click(screen.getByTestId('capability-status-save'));

    await waitFor(() =>
      expect(received).toEqual({
        status: 'Inactive',
        maturityValue: 30,
        ownershipModel: 'TeamOwned',
        primaryOwner: 'Sales Platform Team',
        eaOwner: 'user-1',
      }),
    );
  });

  it('adds a tag in place through x-add-tag', async () => {
    let received: unknown;
    server.use(
      http.post(`${API_BASE}/api/v1/capabilities/cap-1/tags`, async ({ request }) => {
        received = await request.json();
        return new HttpResponse(null, { status: 204 });
      }),
    );
    renderPanel(fullCapability(editableLinks('cap-1')));

    fireEvent.click(await screen.findByRole('button', { name: 'Add a tag' }));
    fireEvent.change(screen.getByTestId('capability-tag-input'), { target: { value: 'freight' } });
    fireEvent.keyDown(screen.getByTestId('capability-tag-input'), { key: 'Enter' });

    await waitFor(() => expect(received).toEqual({ tag: 'freight' }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('renders every field read-only with no controls when no links are present', async () => {
    renderPanel(fullCapability(readOnlyLinks('cap-1')));

    await screen.findByRole('heading', { name: 'Order Management' });
    expect(screen.getByText('Takes and tracks orders')).toBeInTheDocument();
    expect(screen.getByText('Active')).toBeInTheDocument();
    expect(screen.getByText('core')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /^Edit / })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Add a tag' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '+ Add Expert' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('remove-capability-expert-0')).not.toBeInTheDocument();
  });

  it('falls back to the stored EA owner value when no display name is provided', async () => {
    renderPanel({ ...fullCapability(readOnlyLinks('cap-1')), eaOwnerName: undefined, eaOwner: 'Bob Jones' });

    expect(await screen.findByText('Bob Jones')).toBeInTheDocument();
  });

  it('offers Add Expert only through the x-add-expert link and mounts one experts list', async () => {
    renderPanel(fullCapability(editableLinks('cap-1')));

    expect(await screen.findByRole('button', { name: '+ Add Expert' })).toBeInTheDocument();
    expect(screen.getAllByText('Experts')).toHaveLength(1);
  });

  it('renders the view-membership slot after realising applications and the domain slot above the fields', async () => {
    renderPanel(fullCapability(editableLinks('cap-1')), {
      viewMembership: <div data-testid="view-slot">In this view</div>,
      domainContext: <div data-testid="domain-slot">Sales</div>,
    });

    await screen.findByRole('heading', { name: 'Order Management' });
    const heading = screen.getByRole('heading', { name: 'Order Management' });
    const domainSlot = screen.getByTestId('domain-slot');
    const viewSlot = screen.getByTestId('view-slot');
    const realising = screen.getByText('Realising applications');
    expect(domainSlot.compareDocumentPosition(heading) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(realising.compareDocumentPosition(viewSlot) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it('does not offer view-scoped actions without a view-membership slot', async () => {
    renderPanel(fullCapability(editableLinks('cap-1')));

    await screen.findByRole('heading', { name: 'Order Management' });
    expect(screen.queryByRole('button', { name: 'Remove from View' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
    expect(within(document.body).queryByText('In this view')).not.toBeInTheDocument();
  });
});
