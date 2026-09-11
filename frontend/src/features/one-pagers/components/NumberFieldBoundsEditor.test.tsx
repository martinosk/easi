import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { SubjectAttribute } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { NumberFieldBoundsEditor } from './NumberFieldBoundsEditor';

function boundsAttribute(min?: number, max?: number): SubjectAttribute {
  return {
    id: 'field-1',
    name: 'Capacity',
    type: 'number',
    helpText: '',
    active: true,
    min,
    max,
    _links: {
      'x-set-bounds': {
        href: '/api/v1/meta-model/subject-types/application/attributes/field-1/bounds',
        method: 'PUT',
      },
    },
  };
}

describe('NumberFieldBoundsEditor', () => {
  it('resyncs the inputs when the attribute bounds change', async () => {
    const { rerender } = renderWithProviders(
      <NumberFieldBoundsEditor attribute={boundsAttribute(1, 10)} onSave={vi.fn()} />,
    );

    expect(screen.getByTestId('one-pager-bounds-min-field-1')).toHaveValue('1');
    expect(screen.getByTestId('one-pager-bounds-max-field-1')).toHaveValue('10');

    rerender(<NumberFieldBoundsEditor attribute={boundsAttribute(2, 20)} onSave={vi.fn()} />);

    await waitFor(() => expect(screen.getByTestId('one-pager-bounds-min-field-1')).toHaveValue('2'));
    expect(screen.getByTestId('one-pager-bounds-max-field-1')).toHaveValue('20');
  });

  it('disables save again once the refetched attribute matches the edited bounds', async () => {
    const user = userEvent.setup();
    const { rerender } = renderWithProviders(
      <NumberFieldBoundsEditor attribute={boundsAttribute(1, 10)} onSave={vi.fn()} />,
    );

    const minInput = screen.getByTestId('one-pager-bounds-min-field-1');
    await user.clear(minInput);
    await user.type(minInput, '5');
    expect(screen.getByTestId('one-pager-bounds-save-field-1')).toBeEnabled();

    rerender(<NumberFieldBoundsEditor attribute={boundsAttribute(5, 10)} onSave={vi.fn()} />);

    await waitFor(() => expect(screen.getByTestId('one-pager-bounds-save-field-1')).toBeDisabled());
  });

  it('renders nothing without the x-set-bounds affordance', async () => {
    const attribute = { ...boundsAttribute(1, 10), _links: {} };

    renderWithProviders(<NumberFieldBoundsEditor attribute={attribute} onSave={vi.fn()} />);

    await waitFor(() => expect(screen.queryByTestId('one-pager-bounds-form-field-1')).not.toBeInTheDocument());
  });
});
