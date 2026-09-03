import { fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { z } from 'zod';
import { renderWithProviders } from '../../test/helpers';
import { InlineTextField } from './InlineTextField';

const nameSchema = z
  .string()
  .transform((v) => v.trim())
  .refine((v) => v.length > 0, 'Name is required');

function renderField(overrides: Partial<React.ComponentProps<typeof InlineTextField>> = {}) {
  const onSave = vi.fn().mockResolvedValue(undefined);
  renderWithProviders(
    <InlineTextField
      label="Name"
      value="Billing"
      canEdit
      schema={nameSchema}
      onSave={onSave}
      editLabel="Edit name"
      testId="name"
      {...overrides}
    />,
    { withRouter: false },
  );
  return { onSave };
}

describe('InlineTextField', () => {
  it('renders the value as text with an edit control when editable', () => {
    renderField();

    expect(screen.getByText('Billing')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Edit name' })).toBeInTheDocument();
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
  });

  it('renders plain text with no edit control when not editable', () => {
    renderField({ canEdit: false });

    expect(screen.getByText('Billing')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Edit name' })).not.toBeInTheDocument();
  });

  it('renders nothing for an empty value that cannot be edited', () => {
    renderWithProviders(
      <InlineTextField
        label="Description"
        value=""
        canEdit={false}
        schema={z.string()}
        onSave={vi.fn()}
        editLabel="Edit description"
        testId="description"
      />,
      { withRouter: false },
    );

    expect(screen.queryByText('Description')).not.toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('offers the empty prompt for an empty editable value and starts editing from it', () => {
    renderField({ value: '', emptyPrompt: 'Add a description' });

    fireEvent.click(screen.getByRole('button', { name: 'Add a description' }));

    expect(screen.getByTestId('name-input')).toBeInTheDocument();
  });

  it('saves the parsed value on confirm and returns to read mode', async () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit name' }));
    fireEvent.change(screen.getByTestId('name-input'), { target: { value: '  Billing Platform  ' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith('Billing Platform'));
    await waitFor(() => expect(screen.queryByTestId('name-input')).not.toBeInTheDocument());
  });

  it('confirms on Enter for a single-line field', async () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit name' }));
    fireEvent.change(screen.getByTestId('name-input'), { target: { value: 'Ledger' } });
    fireEvent.keyDown(screen.getByTestId('name-input'), { key: 'Enter' });

    await waitFor(() => expect(onSave).toHaveBeenCalledWith('Ledger'));
  });

  it('cancels on Escape without saving and restores the previous value', () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit name' }));
    fireEvent.change(screen.getByTestId('name-input'), { target: { value: 'Ledger' } });
    fireEvent.keyDown(screen.getByTestId('name-input'), { key: 'Escape' });

    expect(onSave).not.toHaveBeenCalled();
    expect(screen.queryByTestId('name-input')).not.toBeInTheDocument();
    expect(screen.getByText('Billing')).toBeInTheDocument();
  });

  it('cancels from the cancel control without saving', () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit name' }));
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));

    expect(onSave).not.toHaveBeenCalled();
    expect(screen.queryByTestId('name-input')).not.toBeInTheDocument();
  });

  it('does not save when the value is unchanged', async () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit name' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(screen.queryByTestId('name-input')).not.toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it('keeps editing and shows the validation message for an invalid value', async () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit name' }));
    fireEvent.change(screen.getByTestId('name-input'), { target: { value: '   ' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(await screen.findByText('Name is required')).toBeInTheDocument();
    expect(screen.getByTestId('name-input')).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it('keeps editing and shows the error when the save fails', async () => {
    const onSave = vi.fn().mockRejectedValue(new Error('Name already taken'));
    renderField({ onSave });

    fireEvent.click(screen.getByRole('button', { name: 'Edit name' }));
    fireEvent.change(screen.getByTestId('name-input'), { target: { value: 'Ledger' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(await screen.findByText('Name already taken')).toBeInTheDocument();
    expect(screen.getByTestId('name-input')).toHaveValue('Ledger');
  });

  it('confirms a multiline field with Ctrl+Enter and not with Enter', async () => {
    const { onSave } = renderField({ multiline: true, editLabel: 'Edit description', testId: 'description' });

    fireEvent.click(screen.getByRole('button', { name: 'Edit description' }));
    fireEvent.change(screen.getByTestId('description-input'), { target: { value: 'Two lines' } });
    fireEvent.keyDown(screen.getByTestId('description-input'), { key: 'Enter' });
    expect(onSave).not.toHaveBeenCalled();

    fireEvent.keyDown(screen.getByTestId('description-input'), { key: 'Enter', ctrlKey: true });
    await waitFor(() => expect(onSave).toHaveBeenCalledWith('Two lines'));
  });

  it('renders the value as a heading when no label is given', () => {
    renderField({ label: undefined });

    expect(screen.getByRole('heading', { name: 'Billing' })).toBeInTheDocument();
  });
});
