import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { capabilitiesApi } from '../../capabilities/api';
import { CapabilitySidebar } from './CapabilitySidebar';

vi.mock('../../capabilities/api', () => ({
  capabilitiesApi: {
    getAll: vi.fn(),
  },
}));

const capabilities = [
  cap('a', 'Customer Management', 'L1'),
  cap('a1', 'Onboarding', 'L2', 'a'),
  cap('b', 'Billing', 'L1'),
];

const renderSidebar = async (props: Partial<React.ComponentProps<typeof CapabilitySidebar>> = {}) => {
  vi.mocked(capabilitiesApi.getAll).mockResolvedValue(capabilities);
  renderWithProviders(<CapabilitySidebar mappedCapabilityIds={new Set(['b'])} {...props} />, { withRouter: false });
  await waitFor(() => expect(screen.getByTestId('cap-tree-a')).toBeInTheDocument());
};

describe('CapabilitySidebar', () => {
  it('renders header, filter, and collapsed root rows with preserved test ids', async () => {
    await renderSidebar();

    expect(screen.getByTestId('capability-sidebar')).toBeInTheDocument();
    expect(screen.getByText('Capabilities')).toBeInTheDocument();
    expect(screen.getByTestId('capability-filter')).toBeInTheDocument();
    expect(screen.getByTestId('cap-tree-b')).toBeInTheDocument();
    expect(screen.queryByTestId('cap-tree-a1')).not.toBeInTheDocument();

    const rootRow = screen.getByTestId('cap-tree-a');
    await userEvent.click(within(rootRow).getByLabelText('Expand'));
    expect(screen.getByTestId('cap-tree-a1')).toBeInTheDocument();
  });

  it('marks mapped capabilities with a badge and disables their drag', async () => {
    await renderSidebar();

    expect(screen.getByText('Mapped')).toBeInTheDocument();
    expect(screen.getByTestId('cap-tree-b')).toHaveAttribute('draggable', 'false');
    expect(screen.getByTestId('cap-tree-a')).toHaveAttribute('draggable', 'true');
  });

  it('serializes the capability as JSON on drag start', async () => {
    const onDragCapability = vi.fn();
    await renderSidebar({ onDragCapability });

    const setData = vi.fn();
    fireEvent.dragStart(screen.getByTestId('cap-tree-a'), { dataTransfer: { setData, effectAllowed: '' } });

    expect(setData).toHaveBeenCalledWith('application/json', JSON.stringify(capabilities[0]));
    expect(onDragCapability).toHaveBeenCalledWith(capabilities[0]);
  });

  it('filters by name and bolds the matched substring', async () => {
    await renderSidebar();

    await userEvent.type(screen.getByTestId('capability-filter'), 'bill');

    expect(screen.queryByTestId('cap-tree-a')).not.toBeInTheDocument();
    const mark = screen.getByText('Bill');
    expect(mark.tagName).toBe('MARK');
    expect(mark).toHaveStyle({ fontWeight: 700 });
  });

  it('shows the preserved no-match message', async () => {
    await renderSidebar();

    await userEvent.type(screen.getByTestId('capability-filter'), 'zzz');

    expect(screen.getByText('No capabilities match your filter')).toBeInTheDocument();
  });

  it('shows the preserved empty message', async () => {
    vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);
    renderWithProviders(<CapabilitySidebar mappedCapabilityIds={new Set()} />, { withRouter: false });

    await waitFor(() => expect(screen.getByText('No capabilities found')).toBeInTheDocument());
  });
});
