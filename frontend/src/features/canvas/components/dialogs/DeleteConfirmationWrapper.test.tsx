import { fireEvent, screen, waitFor } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import type { CapabilityId, ComponentId } from '../../../../api/types';
import { renderWithProviders } from '../../../../test/helpers';
import { buildCapability, buildComponent } from '../../../../test/helpers/entityBuilders';
import { server } from '../../../../test/mocks/server';
import { DeleteConfirmationWrapper } from './DeleteConfirmationWrapper';

describe('DeleteConfirmationWrapper for a component', () => {
  it('names the parts that will be deleted and released with their parent', () => {
    const parent = buildComponent({
      id: 'crm' as ComponentId,
      name: 'CRM Suite',
      parts: [
        { id: 'quoting' as ComponentId, name: 'Quoting', kind: 'composition' },
        { id: 'billing' as ComponentId, name: 'Billing', kind: 'aggregation' },
      ],
    });

    renderWithProviders(
      <DeleteConfirmationWrapper
        deleteTarget={{ type: 'component-from-model', id: parent.id, name: parent.name }}
        deleteTargetComponent={parent}
        isDeleting={false}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
      { withRouter: false },
    );

    const dialog = screen.getByTestId('confirmation-dialog');
    expect(dialog).toHaveTextContent('The composed part "Quoting" will be deleted with it.');
    expect(dialog).toHaveTextContent('The aggregated part "Billing" will be released as standalone.');
  });

  it('confirms detaching a part with a detach action', () => {
    renderWithProviders(
      <DeleteConfirmationWrapper
        deleteTarget={{ type: 'containment', id: 'containment-crm-quoting', name: 'CRM Suite contains Quoting' }}
        isDeleting={false}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
      { withRouter: false },
    );

    const dialog = screen.getByTestId('confirmation-dialog');
    expect(dialog).toHaveTextContent('Detach Part');
    expect(dialog).toHaveTextContent('It becomes a standalone application');
    expect(screen.getByRole('button', { name: 'Detach' })).toBeInTheDocument();
  });

  it('keeps the plain message when the component is not known', () => {
    renderWithProviders(
      <DeleteConfirmationWrapper
        deleteTarget={{ type: 'component-from-model', id: 'solo', name: 'Solo' }}
        isDeleting={false}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
      { withRouter: false },
    );

    const dialog = screen.getByTestId('confirmation-dialog');
    expect(dialog).toHaveTextContent('delete ALL relations involving this component.');
    expect(dialog).not.toHaveTextContent('released');
  });
});

describe('DeleteConfirmationWrapper for a capability', () => {
  const finance = buildCapability({
    id: 'finance' as CapabilityId,
    name: 'Finance',
    _links: {
      self: { href: '/api/v1/capabilities/finance', method: 'GET' },
      delete: { href: '/api/v1/capabilities/finance', method: 'DELETE' },
    },
  });

  const serveImpactWithRealizedChild = () =>
    server.use(
      http.get('*/api/v1/capabilities/finance/delete-impact', () =>
        HttpResponse.json({
          capabilityId: 'finance',
          capabilityName: 'Finance',
          hasDescendants: true,
          affectedCapabilities: [{ id: 'invoicing', name: 'Invoicing', level: 'L2', parentId: 'finance' }],
          realizations: [
            {
              id: 'real-1',
              componentId: 'billing',
              componentName: 'Billing',
              capabilityId: 'invoicing',
              realizationLevel: 'Full',
              origin: 'Direct',
            },
          ],
          _links: {},
        }),
      ),
    );

  it('shows the capability delete impact instead of a generic confirmation', async () => {
    serveImpactWithRealizedChild();

    renderWithProviders(
      <DeleteConfirmationWrapper
        deleteTarget={{ type: 'capability-from-model', id: finance.id, name: finance.name }}
        deleteTargetCapability={finance}
        isDeleting={false}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
      { withRouter: false },
    );

    expect(await screen.findByTestId('cascade-warning')).toHaveTextContent('1 child capability');
    expect(screen.getByTestId('realization-to-remove')).toHaveTextContent('Billing stops realising Invoicing');
    expect(screen.queryByTestId('confirmation-dialog')).not.toBeInTheDocument();
  });

  it('cascade-deletes the capability on confirm and closes', async () => {
    serveImpactWithRealizedChild();
    let deleteBody: unknown;
    server.use(
      http.delete('*/api/v1/capabilities/finance', async ({ request }) => {
        deleteBody = await request.json();
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const onCancel = vi.fn();

    renderWithProviders(
      <DeleteConfirmationWrapper
        deleteTarget={{ type: 'capability-from-model', id: finance.id, name: finance.name }}
        deleteTargetCapability={finance}
        isDeleting={false}
        onConfirm={vi.fn()}
        onCancel={onCancel}
      />,
      { withRouter: false },
    );

    await screen.findByTestId('cascade-warning');
    fireEvent.click(screen.getByTestId('delete-capability-submit'));

    await waitFor(() => expect(onCancel).toHaveBeenCalled());
    expect(deleteBody).toEqual({ cascade: true });
  });
});
