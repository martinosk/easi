import { fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { FilterPopover } from './FilterPopover';

describe('FilterPopover', () => {
  function renderPopover() {
    return renderWithProviders(
      <FilterPopover
        artifactCreators={[{ aggregateId: 'a-1', creatorId: 'u-1' }]}
        users={[{ id: 'u-1', name: 'Alice', email: 'alice@example.com' }]}
        selectedCreatorIds={[]}
        onCreatorSelectionChange={vi.fn()}
        domains={[{ id: 'd-1', name: 'Sales' }]}
        selectedDomainIds={[]}
        onDomainSelectionChange={vi.fn()}
        hasActiveFilters={false}
        onClearAllFilters={vi.fn()}
      />,
      { withRouter: false },
    );
  }

  async function openPopover() {
    fireEvent.click(screen.getByLabelText('Toggle filters'));
    return screen.findByText('Filters');
  }

  async function expectClosed() {
    await waitFor(() => expect(screen.queryByText('Filters')).not.toBeInTheDocument());
  }

  it('is closed until the filter icon is clicked', () => {
    renderPopover();

    expect(screen.queryByText('Filters')).not.toBeInTheDocument();
  });

  it('opens the filter panel anchored to the icon', async () => {
    renderPopover();

    await openPopover();

    expect(screen.getByText('Created by')).toBeInTheDocument();
    expect(screen.getByText('Assigned to domain')).toBeInTheDocument();
    expect(screen.getByLabelText('Toggle filters')).toHaveAttribute('aria-expanded', 'true');
  });

  it('renders the panel through a portal outside the explorer', async () => {
    const { container } = renderPopover();

    const panel = await openPopover();

    expect(container).not.toContainElement(panel);
  });

  it('closes on click outside', async () => {
    renderPopover();
    await openPopover();

    fireEvent.mouseDown(document.body);

    await expectClosed();
  });

  it('closes on Escape', async () => {
    renderPopover();
    const panel = await openPopover();

    fireEvent.keyDown(panel, { key: 'Escape' });

    await expectClosed();
  });

  it('stays open when clicking inside the panel', async () => {
    renderPopover();
    const panel = await openPopover();

    fireEvent.mouseDown(panel);

    expect(screen.getByText('Filters')).toBeInTheDocument();
  });
});
