import { fireEvent, screen, waitFor } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import type { HATEOASLinks, Relation } from '../../../api/types';
import { toComponentId, toRelationId } from '../../../api/types';
import type { AppStore } from '../../../store/appStore';
import { useAppStore } from '../../../store/appStore';
import { buildComponent, buildRelation, renderWithProviders, seedDb, server } from '../../../test/helpers';
import { RelationDetails } from './RelationDetails';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

const API_BASE = 'http://localhost:8080';

function links(editable: boolean): HATEOASLinks {
  return {
    self: { href: '/api/v1/relations/rel-1', method: 'GET' },
    describedby: { href: '/api/v1/reference/relations/generic', method: 'GET' },
    ...(editable ? { edit: { href: '/api/v1/relations/rel-1', method: 'PUT' } } : {}),
  };
}

function relation(editable: boolean, overrides: Partial<Relation> = {}): Relation {
  return buildRelation({
    id: toRelationId('rel-1'),
    sourceComponentId: toComponentId('comp-1'),
    targetComponentId: toComponentId('comp-2'),
    relationType: 'Serves',
    name: 'Sends invoices',
    description: 'Monthly invoice feed',
    _links: links(editable),
    ...overrides,
  });
}

function renderPane(rel: Relation) {
  seedDb({
    relations: [rel],
    components: [
      buildComponent({ id: toComponentId('comp-1'), name: 'Billing Engine' }),
      buildComponent({ id: toComponentId('comp-2'), name: 'Order Management' }),
    ],
  });
  vi.mocked(useAppStore).mockImplementation((selector: (state: AppStore) => unknown) =>
    selector({ selectedEdgeId: 'rel-1' } as unknown as AppStore),
  );
  return renderWithProviders(<RelationDetails />);
}

describe('RelationDetails', () => {
  it('renders the name as the heading and every section in order without an Edit action', async () => {
    renderPane(relation(true));

    expect(await screen.findByRole('heading', { name: 'Sends invoices' })).toBeInTheDocument();
    expect(screen.queryByText('Relation Details')).not.toBeInTheDocument();
    const labels = screen
      .getAllByText(/^(Type|Source|Target|Description|Created)$/)
      .map((element) => element.textContent);
    expect(labels).toEqual(['Type', 'Source', 'Target', 'Description', 'Created']);
    expect(screen.getByText('Billing Engine')).toBeInTheDocument();
    expect(screen.getByText('Order Management')).toBeInTheDocument();
    expect(screen.getByText('Reference Documentation')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
  });

  it('renames the relation in place through the edit link, sending the description unchanged', async () => {
    let received: unknown;
    server.use(
      http.put(`${API_BASE}/api/v1/relations/rel-1`, async ({ request }) => {
        received = await request.json();
        return HttpResponse.json({ ...relation(true), name: 'Sends credit notes' });
      }),
    );
    renderPane(relation(true));

    fireEvent.click(await screen.findByRole('button', { name: 'Edit name' }));
    fireEvent.change(screen.getByTestId('relation-name-input'), { target: { value: 'Sends credit notes' } });
    fireEvent.keyDown(screen.getByTestId('relation-name-input'), { key: 'Enter' });

    await waitFor(() => expect(received).toEqual({ name: 'Sends credit notes', description: 'Monthly invoice feed' }));
  });

  it('prompts to add a description only when editable and saves it in place', async () => {
    let received: unknown;
    server.use(
      http.put(`${API_BASE}/api/v1/relations/rel-1`, async ({ request }) => {
        received = await request.json();
        return HttpResponse.json({ ...relation(true), description: 'Nightly feed' });
      }),
    );
    renderPane(relation(true, { description: undefined }));

    fireEvent.click(await screen.findByRole('button', { name: 'Add a description' }));
    fireEvent.change(screen.getByTestId('relation-description-input'), { target: { value: 'Nightly feed' } });
    fireEvent.click(screen.getByTestId('relation-description-save'));

    await waitFor(() => expect(received).toEqual({ name: 'Sends invoices', description: 'Nightly feed' }));
  });

  it('renders read-only with no controls when the edit link is absent', async () => {
    renderPane(relation(false, { description: undefined }));

    expect(await screen.findByRole('heading', { name: 'Sends invoices' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /^(Edit|Add) / })).not.toBeInTheDocument();
    expect(screen.queryByText('Description')).not.toBeInTheDocument();
  });
});
