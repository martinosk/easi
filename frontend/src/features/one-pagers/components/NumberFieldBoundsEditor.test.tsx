import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { CustomField } from '../types';
import { NumberFieldBoundsEditor } from './NumberFieldBoundsEditor';

function boundsField(min?: number, max?: number): CustomField {
  return {
    id: 'field-1',
    name: 'Capacity',
    type: 'number',
    required: false,
    helpText: '',
    active: true,
    min,
    max,
    _links: {
      'x-set-bounds': {
        href: '/api/v1/one-pagers/configurations/application/custom-fields/field-1/bounds',
        method: 'PUT',
      },
    },
  };
}

describe('NumberFieldBoundsEditor', () => {
  it('resyncs the inputs when the field bounds change', async () => {
    const { rerender } = renderWithProviders(<NumberFieldBoundsEditor field={boundsField(1, 10)} onSave={vi.fn()} />);

    expect(screen.getByTestId('one-pager-bounds-min-field-1')).toHaveValue('1');
    expect(screen.getByTestId('one-pager-bounds-max-field-1')).toHaveValue('10');

    rerender(<NumberFieldBoundsEditor field={boundsField(2, 20)} onSave={vi.fn()} />);

    await waitFor(() => expect(screen.getByTestId('one-pager-bounds-min-field-1')).toHaveValue('2'));
    expect(screen.getByTestId('one-pager-bounds-max-field-1')).toHaveValue('20');
  });

  it('disables save again once the refetched field matches the edited bounds', async () => {
    const user = userEvent.setup();
    const { rerender } = renderWithProviders(<NumberFieldBoundsEditor field={boundsField(1, 10)} onSave={vi.fn()} />);

    const minInput = screen.getByTestId('one-pager-bounds-min-field-1');
    await user.clear(minInput);
    await user.type(minInput, '5');
    expect(screen.getByTestId('one-pager-bounds-save-field-1')).toBeEnabled();

    rerender(<NumberFieldBoundsEditor field={boundsField(5, 10)} onSave={vi.fn()} />);

    await waitFor(() => expect(screen.getByTestId('one-pager-bounds-save-field-1')).toBeDisabled());
  });
});
