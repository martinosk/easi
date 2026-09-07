import { fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { InlineMaturityField } from './InlineMaturityField';

function renderField(overrides: Partial<React.ComponentProps<typeof InlineMaturityField>> = {}) {
  const onSave = vi.fn().mockResolvedValue(undefined);
  renderWithProviders(<InlineMaturityField value={12} canEdit onSave={onSave} {...overrides} />, {
    withRouter: false,
  });
  return { onSave };
}

describe('InlineMaturityField', () => {
  it('renders the section name and value as a badge with an edit control', () => {
    renderField();

    expect(screen.getByTestId('capability-maturity-value')).toHaveTextContent('Genesis (12)');
    expect(screen.getByRole('button', { name: 'Edit maturity' })).toBeInTheDocument();
  });

  it('renders the badge with no edit control when not editable', () => {
    renderField({ canEdit: false });

    expect(screen.getByTestId('capability-maturity-value')).toHaveTextContent('Genesis (12)');
    expect(screen.queryByRole('button', { name: 'Edit maturity' })).not.toBeInTheDocument();
  });

  it('saves the slider value on confirm and returns to read mode', async () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit maturity' }));
    const slider = await screen.findByTestId('maturity-slider');
    fireEvent.keyDown(slider, { key: 'ArrowRight' });
    fireEvent.click(screen.getByTestId('capability-maturity-save'));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith(13));
    await waitFor(() => expect(screen.queryByTestId('maturity-slider')).not.toBeInTheDocument());
  });

  it('cancels without saving', () => {
    const { onSave } = renderField();

    fireEvent.click(screen.getByRole('button', { name: 'Edit maturity' }));
    fireEvent.click(screen.getByTestId('capability-maturity-cancel'));

    expect(onSave).not.toHaveBeenCalled();
    expect(screen.queryByTestId('maturity-slider')).not.toBeInTheDocument();
  });

  it('stays in edit mode with the error when saving fails', async () => {
    const onSave = vi.fn().mockRejectedValue(new Error('Maturity out of range'));
    renderField({ onSave });

    fireEvent.click(screen.getByRole('button', { name: 'Edit maturity' }));
    fireEvent.keyDown(await screen.findByTestId('maturity-slider'), { key: 'ArrowRight' });
    fireEvent.click(screen.getByTestId('capability-maturity-save'));

    expect(await screen.findByText('Maturity out of range')).toBeInTheDocument();
    expect(screen.getByTestId('maturity-slider')).toBeInTheDocument();
  });
});
