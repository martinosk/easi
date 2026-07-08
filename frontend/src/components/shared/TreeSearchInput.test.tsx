import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { TreeSearchInput } from './TreeSearchInput';

describe('TreeSearchInput', () => {
  it('renders with placeholder and value', () => {
    renderWithProviders(<TreeSearchInput value="cust" onChange={() => {}} placeholder="Search capabilities..." />, {
      withRouter: false,
    });

    expect(screen.getByPlaceholderText('Search capabilities...')).toHaveValue('cust');
  });

  it('calls onChange with the typed value', async () => {
    const onChange = vi.fn();
    renderWithProviders(<TreeSearchInput value="" onChange={onChange} placeholder="Search..." />, {
      withRouter: false,
    });

    await userEvent.type(screen.getByPlaceholderText('Search...'), 'a');

    expect(onChange).toHaveBeenCalledWith('a');
  });

  it('shows a clear button only when there is a value and clears on click', async () => {
    const onChange = vi.fn();
    const { rerender } = renderWithProviders(
      <TreeSearchInput value="" onChange={onChange} placeholder="Search..." />,
      { withRouter: false },
    );

    expect(screen.queryByLabelText('Clear search')).not.toBeInTheDocument();

    rerender(<TreeSearchInput value="abc" onChange={onChange} placeholder="Search..." />);
    await userEvent.click(screen.getByLabelText('Clear search'));

    expect(onChange).toHaveBeenCalledWith('');
  });
});
