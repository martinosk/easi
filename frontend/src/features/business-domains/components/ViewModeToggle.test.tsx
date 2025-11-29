import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ViewModeToggle } from './ViewModeToggle';

describe('ViewModeToggle', () => {
  it('should render all view mode options', () => {
    const onModeChange = vi.fn();
    render(<ViewModeToggle mode="treemap" onModeChange={onModeChange} />);

    expect(screen.getByRole('button', { name: /treemap/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /tree/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /grid/i })).toBeInTheDocument();
  });

  it('should highlight the selected mode', () => {
    const onModeChange = vi.fn();
    render(<ViewModeToggle mode="grid" onModeChange={onModeChange} />);

    const gridButton = screen.getByRole('button', { name: /grid/i });
    expect(gridButton).toHaveClass('bg-blue-600', 'text-white');
  });

  it('should call onModeChange when a different mode is clicked', () => {
    const onModeChange = vi.fn();
    render(<ViewModeToggle mode="treemap" onModeChange={onModeChange} />);

    const gridButton = screen.getByRole('button', { name: /grid/i });
    gridButton.click();

    expect(onModeChange).toHaveBeenCalledWith('grid');
  });

  it('should not highlight non-selected modes', () => {
    const onModeChange = vi.fn();
    render(<ViewModeToggle mode="grid" onModeChange={onModeChange} />);

    const treemapButton = screen.getByRole('button', { name: /treemap/i });
    const treeButton = screen.getByRole('button', { name: /tree/i });

    expect(treemapButton).not.toHaveClass('bg-blue-600', 'text-white');
    expect(treeButton).not.toHaveClass('bg-blue-600', 'text-white');
  });

  it('should render in correct order: treemap, tree, grid', () => {
    const onModeChange = vi.fn();
    render(<ViewModeToggle mode="treemap" onModeChange={onModeChange} />);

    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(3);
    expect(buttons[0]).toHaveTextContent(/treemap/i);
    expect(buttons[1]).toHaveTextContent(/tree/i);
    expect(buttons[2]).toHaveTextContent(/grid/i);
  });
});
