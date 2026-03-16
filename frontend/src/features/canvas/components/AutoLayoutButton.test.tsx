import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AutoLayoutButton } from './AutoLayoutButton';

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

  it('shows warning dialog when auto-layout is clicked', () => {
    render(<AutoLayoutButton />);

    fireEvent.click(screen.getByRole('button', { name: 'Auto layout canvas' }));

    expect(screen.getByRole('alertdialog')).toBeDefined();
    expect(
      screen.getByText(
        'Auto layout is an experimental feature that will completely re-arrange your view.'
      )
    ).toBeDefined();
    expect(mockApplyAutoLayout).not.toHaveBeenCalled();
  });

  it('cancels auto-layout when Cancel is clicked in warning dialog', () => {
    render(<AutoLayoutButton />);

    fireEvent.click(screen.getByRole('button', { name: 'Auto layout canvas' }));
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));

    expect(screen.queryByRole('alertdialog')).toBeNull();
    expect(mockApplyAutoLayout).not.toHaveBeenCalled();
  });

  it('runs auto-layout when OK is clicked in warning dialog', () => {
    render(<AutoLayoutButton />);

    fireEvent.click(screen.getByRole('button', { name: 'Auto layout canvas' }));
    fireEvent.click(screen.getByRole('button', { name: 'OK' }));

    expect(mockApplyAutoLayout).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('alertdialog')).toBeNull();
  });
});
