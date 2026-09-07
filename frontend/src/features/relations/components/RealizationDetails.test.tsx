import { fireEvent, screen, waitFor } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import type { CapabilityRealization, HATEOASLinks } from '../../../api/types';
import { toCapabilityId, toComponentId, toRealizationId } from '../../../api/types';
import type { AppStore } from '../../../store/appStore';
import { useAppStore } from '../../../store/appStore';
import {
  buildCapability,
  buildCapabilityRealization,
  buildComponent,
  renderWithProviders,
  seedDb,
  server,
} from '../../../test/helpers';
import { RealizationDetails } from './RealizationDetails';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

const API_BASE = 'http://localhost:8080';

function links(editable: boolean): HATEOASLinks {
  return {
    self: { href: '/api/v1/capability-realizations/real-1', method: 'GET' },
    ...(editable ? { edit: { href: '/api/v1/capability-realizations/real-1', method: 'PUT' } } : {}),
  };
}

function realization(editable: boolean, overrides: Partial<CapabilityRealization> = {}): CapabilityRealization {
  return buildCapabilityRealization({
    id: toRealizationId('real-1'),
    capabilityId: toCapabilityId('cap-1'),
    componentId: toComponentId('comp-1'),
    realizationLevel: 'Partial',
    origin: 'Direct',
    notes: 'Covers invoicing only',
    _links: links(editable),
    ...overrides,
  });
}

function renderPane(real: CapabilityRealization) {
  seedDb({
    capabilityRealizations: [real],
    capabilities: [buildCapability({ id: toCapabilityId('cap-1'), name: 'Order Management' })],
    components: [buildComponent({ id: toComponentId('comp-1'), name: 'Billing Engine' })],
  });
  vi.mocked(useAppStore).mockImplementation((selector: (state: AppStore) => unknown) =>
    selector({ selectedEdgeId: 'realization-real-1' } as unknown as AppStore),
  );
  return renderWithProviders(<RealizationDetails />);
}

function captureUpdate(response: CapabilityRealization): () => unknown {
  let received: unknown;
  server.use(
    http.put(`${API_BASE}/api/v1/capability-realizations/real-1`, async ({ request }) => {
      received = await request.json();
      return HttpResponse.json(response);
    }),
  );
  return () => received;
}

async function pickOption(label: string): Promise<Element> {
  return waitFor(() => {
    const option = Array.from(document.querySelectorAll('[data-combobox-option]')).find(
      (element) => element.textContent === label,
    );
    if (!option) throw new Error(`option ${label} not rendered`);
    return option;
  });
}

describe('RealizationDetails', () => {
  it('renders every section in order without an Edit action', async () => {
    renderPane(realization(true));

    expect(await screen.findByText('Order Management')).toBeInTheDocument();
    expect(screen.queryByText('Realization Details')).not.toBeInTheDocument();
    const labels = screen
      .getAllByText(/^(Capability|Application|Realization Level|Origin|Notes|Linked)$/)
      .map((element) => element.textContent);
    expect(labels).toEqual(['Capability', 'Application', 'Realization Level', 'Origin', 'Notes', 'Linked']);
    expect(screen.getByText('Billing Engine')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
  });

  it('changes the level in place, sending the notes unchanged', async () => {
    const received = captureUpdate({ ...realization(true), realizationLevel: 'Full' });
    renderPane(realization(true));

    fireEvent.click(await screen.findByRole('button', { name: 'Edit realization level' }));
    fireEvent.click(screen.getByTestId('realization-level-input'));
    fireEvent.click(await pickOption('Full (100%)'));
    fireEvent.click(screen.getByTestId('realization-level-save'));

    await waitFor(() => expect(received()).toEqual({ realizationLevel: 'Full', notes: 'Covers invoicing only' }));
  });

  it('adds notes in place, sending the level unchanged', async () => {
    const received = captureUpdate({ ...realization(true), notes: 'Pilot only' });
    renderPane(realization(true, { notes: undefined }));

    fireEvent.click(await screen.findByRole('button', { name: 'Add notes' }));
    fireEvent.change(screen.getByTestId('realization-notes-input'), { target: { value: 'Pilot only' } });
    fireEvent.click(screen.getByTestId('realization-notes-save'));

    await waitFor(() => expect(received()).toEqual({ realizationLevel: 'Partial', notes: 'Pilot only' }));
  });

  it('renders an inherited realization read-only even when an edit link is present', async () => {
    renderPane(realization(true, { origin: 'Inherited', sourceCapabilityName: 'Sales' }));

    expect(await screen.findByText('Order Management')).toBeInTheDocument();
    expect(screen.getByText('Inherited')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /^(Edit|Add) / })).not.toBeInTheDocument();
  });

  it('renders read-only with no controls when the edit link is absent', async () => {
    renderPane(realization(false, { notes: undefined }));

    expect(await screen.findByText('Order Management')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /^(Edit|Add) / })).not.toBeInTheDocument();
    expect(screen.queryByText('Notes')).not.toBeInTheDocument();
  });
});
