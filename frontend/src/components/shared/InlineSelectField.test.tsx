import { fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { InlineSelectField } from './InlineSelectField';

const options = [
  { value: 'Active', label: 'Active' },
  { value: 'Planned', label: 'Planned' },
];

function renderField(overrides: Partial<React.ComponentProps<typeof InlineSelectField>> = {}) {
  const onSave = vi.fn().mockResolvedValue(undefined);
  renderWithProviders(
    <InlineSelectField
      label="Status"
      value="Active"
      options={options}
      canEdit
      onSave={onSave}
      editLabel="Edit status"
      testId="status"
      {...overrides}
    />,
    { withRouter: false },
  );
  return { onSave };
}

describe('InlineSelectField', () => {
  it('renders the current option label with an edit control when editable', () => {
    renderField();

    expect(screen.getByTestId('status-value')).toHaveTextContent('Active');
    expect(screen.getByRole('button', { name: 'Edit status' })).toBeInTheDocument();
  });

  it('renders the value with no edit control when not editable', () => {
    renderField({ canEdit: false });

    expect(screen.getByTestId('status-value')).toHaveTextContent('Active');
    expect(screen.queryByRole('button', { name: 'Edit status' })).not.toBeInTheDocument();
  });

  it('renders nothing for an empty value that cannot be edited', () => {
    renderField({ value: '', canEdit: false });

    expect(screen.queryByText('Status')).not.toBeInTheDocument();
  });

  it('renders the empty prompt for an empty editable value', () => {
    renderField({ value: '', emptyPrompt: 'Set a status' });

    expect(screen.getByRole('button', { name: 'Set a status' })).toBeInTheDocument();
  });

  it('renders a custom read view when one is given', () => {
    renderField({ renderValue: (value) => <span data-testid="custom">badge:{value}</span> });

    expect(screen.getByTestId('custom')).toHaveTextContent('badge:Active');
  });

  it('saves the picked option and returns to read mode', async () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit status' }));
    const input = screen.getByTestId('status-input');
    fireEvent.click(input);
    fireEvent.click(await screen.findByRole('option', { name: 'Planned' }));
    fireEvent.click(screen.getByTestId('status-save'));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith('Planned'));
    await waitFor(() => expect(screen.queryByTestId('status-input')).not.toBeInTheDocument());
  });

  it('cancels without saving', () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit status' }));
    fireEvent.click(screen.getByTestId('status-cancel'));

    expect(onSave).not.toHaveBeenCalled();
    expect(screen.queryByTestId('status-input')).not.toBeInTheDocument();
  });

  it('does not save when the value is unchanged', async () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit status' }));
    fireEvent.click(screen.getByTestId('status-save'));

    await waitFor(() => expect(screen.queryByTestId('status-input')).not.toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it('stays in edit mode and shows the error when saving fails', async () => {
    const onSave = vi.fn().mockRejectedValue(new Error('EA owner is ambiguous'));
    renderField({ onSave });

    fireEvent.click(screen.getByRole('button', { name: 'Edit status' }));
    const input = screen.getByTestId('status-input');
    fireEvent.click(input);
    fireEvent.click(await screen.findByRole('option', { name: 'Planned' }));
    fireEvent.click(screen.getByTestId('status-save'));

    expect(await screen.findByText('EA owner is ambiguous')).toBeInTheDocument();
    expect(screen.getByTestId('status-input')).toBeInTheDocument();
  });

  it('escape cancels while editing', () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit status' }));
    fireEvent.keyDown(screen.getByTestId('status-input'), { key: 'Escape' });

    expect(onSave).not.toHaveBeenCalled();
    expect(screen.queryByTestId('status-input')).not.toBeInTheDocument();
  });
});
