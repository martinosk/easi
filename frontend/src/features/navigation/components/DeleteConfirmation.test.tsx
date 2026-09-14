import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { ComponentId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { DeleteConfirmation } from './DeleteConfirmation';

describe('DeleteConfirmation for an application', () => {
  const render = (component: ReturnType<typeof buildComponent>) =>
    renderWithProviders(
      <DeleteConfirmation
        deleteTarget={{ type: 'component', component }}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
        isLoading={false}
      />,
      { withRouter: false },
    );

  it('names the composed parts that will be deleted and the aggregated parts that will be released', () => {
    render(
      buildComponent({
        id: 'crm' as ComponentId,
        name: 'CRM Suite',
        parts: [
          { id: 'quoting' as ComponentId, name: 'Quoting', kind: 'composition' },
          { id: 'billing' as ComponentId, name: 'Billing', kind: 'aggregation' },
        ],
      }),
    );

    const dialog = screen.getByTestId('confirmation-dialog');
    expect(dialog).toHaveTextContent('The composed part "Quoting" will be deleted with it.');
    expect(dialog).toHaveTextContent('The aggregated part "Billing" will be released as standalone.');
  });

  it('keeps the plain message for a standalone application', () => {
    render(buildComponent({ id: 'solo' as ComponentId, name: 'Solo' }));

    const dialog = screen.getByTestId('confirmation-dialog');
    expect(dialog).toHaveTextContent('delete ALL relations involving this application.');
    expect(dialog).not.toHaveTextContent('will be deleted with it');
    expect(dialog).not.toHaveTextContent('released');
  });
});
