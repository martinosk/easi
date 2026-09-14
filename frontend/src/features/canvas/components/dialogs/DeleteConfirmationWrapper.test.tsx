import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { ComponentId } from '../../../../api/types';
import { renderWithProviders } from '../../../../test/helpers';
import { buildComponent } from '../../../../test/helpers/entityBuilders';
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
