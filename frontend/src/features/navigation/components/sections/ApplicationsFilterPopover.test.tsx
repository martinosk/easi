import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../../test/helpers';
import { ApplicationsFilterPopover } from './ApplicationsFilterPopover';

describe('ApplicationsFilterPopover', () => {
  function renderPopover(overrides: Partial<React.ComponentProps<typeof ApplicationsFilterPopover>> = {}) {
    const props = {
      ownership: null,
      onOwnershipChange: vi.fn(),
      hosting: null,
      onHostingChange: vi.fn(),
      ...overrides,
    };
    renderWithProviders(<ApplicationsFilterPopover {...props} />, { withRouter: false });
    return props;
  }

  function open() {
    fireEvent.click(screen.getByLabelText('Toggle application filters'));
    return screen.findByText('Filters');
  }

  it('is closed until the filter icon is clicked', () => {
    renderPopover();

    expect(screen.queryByText('Ownership')).not.toBeInTheDocument();
  });

  it('shows ownership and hosting filter groups when opened', async () => {
    renderPopover();

    await open();

    expect(screen.getByText('Ownership')).toBeInTheDocument();
    expect(screen.getByText('Hosting')).toBeInTheDocument();
    expect(screen.getByText('Owned')).toBeInTheDocument();
    expect(screen.getByText('On-premises')).toBeInTheDocument();
  });

  it('reports the chosen ownership state', async () => {
    const props = renderPopover();
    await open();

    fireEvent.click(screen.getByText('Nominated'));

    expect(props.onOwnershipChange).toHaveBeenCalledWith('nominated');
  });

  it('reports the chosen hosting classification', async () => {
    const props = renderPopover();
    await open();

    fireEvent.click(screen.getByText('Cloud'));

    expect(props.onHostingChange).toHaveBeenCalledWith('cloud');
  });

  it('shows the number of active filters on the icon', () => {
    renderPopover({ ownership: 'owned', hosting: 'saas' });

    expect(screen.getByText('2')).toBeInTheDocument();
  });

  it('clears both filters with clear all', async () => {
    const props = renderPopover({ ownership: 'owned', hosting: 'saas' });
    await open();

    fireEvent.click(screen.getByText('Clear all'));

    expect(props.onOwnershipChange).toHaveBeenCalledWith(null);
    expect(props.onHostingChange).toHaveBeenCalledWith(null);
  });

  it('clears a single filter group', async () => {
    const props = renderPopover({ ownership: 'owned' });
    await open();

    fireEvent.click(screen.getByLabelText('Clear filter'));

    expect(props.onOwnershipChange).toHaveBeenCalledWith(null);
    expect(props.onHostingChange).not.toHaveBeenCalled();
  });
});
