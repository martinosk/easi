import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ContextMenu } from './ContextMenu';
import type { ContextMenuItem } from './types';

function makeItems(count: number): ContextMenuItem[] {
  return Array.from({ length: count }, (_, i) => ({
    label: `Action ${i + 1}`,
    description: `Does action ${i + 1}`,
    onClick: vi.fn(),
  }));
}

describe('ContextMenu', () => {
  it('returns null when items are empty', () => {
    const { container } = render(<ContextMenu x={50} y={50} items={[]} onClose={vi.fn()} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders the radial variant for ≤6 items in auto mode', () => {
    render(<ContextMenu x={100} y={100} items={makeItems(3)} onClose={vi.fn()} />);
    const menu = screen.getByRole('menu');
    expect(menu.className).toContain('ctx-menu--radial');
    expect(screen.getAllByRole('menuitem')).toHaveLength(3);
  });

  it('falls back to linear when items exceed the radial cap', () => {
    render(<ContextMenu x={100} y={100} items={makeItems(8)} onClose={vi.fn()} />);
    const menu = screen.getByRole('menu');
    expect(menu.className).toContain('ctx-menu--linear');
  });

  it('honours an explicit linear variant override', () => {
    render(<ContextMenu x={100} y={100} items={makeItems(2)} variant="linear" onClose={vi.fn()} />);
    const menu = screen.getByRole('menu');
    expect(menu.className).toContain('ctx-menu--linear');
  });

  it('shows the title in the radial hub when nothing is hovered', () => {
    render(<ContextMenu x={100} y={100} items={makeItems(3)} title="My Component" onClose={vi.fn()} />);
    expect(screen.getByText('My Component')).toBeDefined();
  });

  it('updates the radial hub label and description on hover', () => {
    render(<ContextMenu x={100} y={100} items={makeItems(3)} title="My Component" onClose={vi.fn()} />);
    const firstAction = screen.getAllByRole('menuitem')[0];
    fireEvent.mouseEnter(firstAction);
    expect(screen.getAllByText('Action 1').length).toBeGreaterThan(0);
    expect(screen.getByText('Does action 1')).toBeDefined();
  });

  it('invokes onClick and onClose when a petal is clicked', () => {
    const onClose = vi.fn();
    const items = makeItems(3);
    render(<ContextMenu x={100} y={100} items={items} onClose={onClose} />);
    fireEvent.click(screen.getAllByRole('menuitem')[1]);
    expect(items[1].onClick).toHaveBeenCalledTimes(1);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('closes on Escape', () => {
    const onClose = vi.fn();
    render(<ContextMenu x={100} y={100} items={makeItems(3)} onClose={onClose} />);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalled();
  });
});
