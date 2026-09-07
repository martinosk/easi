import { fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { InlineDateField } from './InlineDateField';

function renderField(overrides: Partial<React.ComponentProps<typeof InlineDateField>> = {}) {
  const onSave = vi.fn().mockResolvedValue(undefined);
  renderWithProviders(
    <InlineDateField
      label="Acquisition Date"
      value="2024-03-15"
      canEdit
      onSave={onSave}
      editLabel="Edit acquisition date"
      emptyPrompt="Set an acquisition date"
      testId="acquisition-date"
      {...overrides}
    />,
    { withRouter: false },
  );
  return { onSave };
}

describe('InlineDateField', () => {
  it('renders the formatted date with an edit control when editable', () => {
    renderField();

    expect(screen.getByTestId('acquisition-date-value')).toHaveTextContent(new Date('2024-03-15').toLocaleDateString());
    expect(screen.getByRole('button', { name: 'Edit acquisition date' })).toBeInTheDocument();
  });

  it('renders the date with no edit control when not editable', () => {
    renderField({ canEdit: false });

    expect(screen.getByTestId('acquisition-date-value')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Edit acquisition date' })).not.toBeInTheDocument();
  });

  it('renders nothing for an empty value that cannot be edited', () => {
    renderField({ value: '', canEdit: false });

    expect(screen.queryByText('Acquisition Date')).not.toBeInTheDocument();
  });

  it('renders the empty prompt for an empty editable value', () => {
    renderField({ value: '' });

    expect(screen.getByRole('button', { name: 'Set an acquisition date' })).toBeInTheDocument();
  });

  it('saves the picked date on confirm and returns to read mode', async () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit acquisition date' }));
    fireEvent.change(screen.getByTestId('acquisition-date-input'), { target: { value: '2025-01-02' } });
    fireEvent.keyDown(screen.getByTestId('acquisition-date-input'), { key: 'Enter' });

    await waitFor(() => expect(onSave).toHaveBeenCalledWith('2025-01-02'));
    await waitFor(() => expect(screen.queryByTestId('acquisition-date-input')).not.toBeInTheDocument());
  });

  it('cancels without saving', () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit acquisition date' }));
    fireEvent.keyDown(screen.getByTestId('acquisition-date-input'), { key: 'Escape' });

    expect(onSave).not.toHaveBeenCalled();
    expect(screen.queryByTestId('acquisition-date-input')).not.toBeInTheDocument();
  });

  it('stays in edit mode with the error when saving fails', async () => {
    const onSave = vi.fn().mockRejectedValue(new Error('Date is in the future'));
    renderField({ onSave });

    fireEvent.click(screen.getByRole('button', { name: 'Edit acquisition date' }));
    fireEvent.change(screen.getByTestId('acquisition-date-input'), { target: { value: '2099-01-02' } });
    fireEvent.click(screen.getByTestId('acquisition-date-save'));

    expect(await screen.findByText('Date is in the future')).toBeInTheDocument();
    expect(screen.getByTestId('acquisition-date-input')).toBeInTheDocument();
  });
});
