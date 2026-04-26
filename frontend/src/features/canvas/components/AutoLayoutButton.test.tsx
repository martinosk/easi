import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { AutoLayoutButton } from './AutoLayoutButton';

const renderWithMantine = (ui: React.ReactElement) =>
  render(<MantineTestWrapper>{ui}</MantineTestWrapper>);

const mockApplyAutoLayout = vi.fn();
const mockUseCanvasNodes = vi.fn();
const mockUseCurrentView = vi.fn();

vi.mock('../hooks/useAutoLayout', () => ({
  useAutoLayout: () => ({
    applyAutoLayout: mockApplyAutoLayout,
    isLayouting: false,
  }),
}));

vi.mock('../hooks/useCanvasNodes', () => ({
  useCanvasNodes: () => mockUseCanvasNodes(),
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => mockUseCurrentView(),
}));

describe('AutoLayoutButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseCanvasNodes.mockReturnValue([{ id: 'comp-1' }]);
    mockUseCurrentView.mockReturnValue({
      currentViewId: 'view-1',
      currentView: {
        _links: {
          edit: { href: '/api/v1/views/view-1', method: 'PUT' },
        },
      },
    });
  });

  it('runs auto-layout immediately on click without showing a dialog', () => {
    renderWithMantine(<AutoLayoutButton />);

    fireEvent.click(screen.getByRole('button', { name: 'Auto layout canvas' }));

    expect(mockApplyAutoLayout).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('alertdialog')).toBeNull();
  });

  it('does not run auto-layout when there are no nodes', () => {
    mockUseCanvasNodes.mockReturnValue([]);
    renderWithMantine(<AutoLayoutButton />);

    fireEvent.click(screen.getByRole('button', { name: 'Auto layout canvas' }));

    expect(mockApplyAutoLayout).not.toHaveBeenCalled();
  });

  it('does not run auto-layout when the current view is not editable', () => {
    mockUseCurrentView.mockReturnValue({
      currentViewId: 'view-1',
      currentView: { _links: {} },
    });
    renderWithMantine(<AutoLayoutButton />);

    fireEvent.click(screen.getByRole('button', { name: 'Auto layout canvas' }));

    expect(mockApplyAutoLayout).not.toHaveBeenCalled();
  });
});
