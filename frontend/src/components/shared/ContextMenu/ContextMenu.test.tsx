import { fireEvent, screen, waitFor } from '@testing-library/react';
import type { ReactElement } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { ContextMenu } from './ContextMenu';
import type { ContextMenuItem, ContextMenuVariant } from './types';

const render = (ui: ReactElement) => renderWithProviders(ui, { withRouter: false });

function makeItems(count: number): ContextMenuItem[] {
  return Array.from({ length: count }, (_, i) => ({
    label: `Action ${i + 1}`,
    description: `Does action ${i + 1}`,
    onClick: vi.fn(),
  }));
}

const menuRoot = () => screen.getByTestId('context-menu');

describe('ContextMenu', () => {
  it('returns null when items are empty', () => {
    render(<ContextMenu x={50} y={50} items={[]} onClose={vi.fn()} />);
    expect(screen.queryByRole('menu')).toBeNull();
  });

  it.each<{ scenario: string; count: number; variant?: ContextMenuVariant; expected: string }>([
    { scenario: 'radial for ≤6 items in auto mode', count: 3, expected: 'radial' },
    { scenario: 'linear when items exceed the radial cap', count: 8, expected: 'linear' },
    { scenario: 'linear when explicitly overridden', count: 2, variant: 'linear', expected: 'linear' },
  ])('renders $scenario', ({ count, variant, expected }) => {
    render(<ContextMenu x={100} y={100} items={makeItems(count)} variant={variant} onClose={vi.fn()} />);
    expect(menuRoot()).toHaveAttribute('data-variant', expected);
    expect(menuRoot()).toHaveAttribute('role', 'menu');
    expect(screen.getAllByRole('menuitem')).toHaveLength(count);
  });

  it('shows the title in the radial hub when nothing is hovered', () => {
    render(<ContextMenu x={100} y={100} items={makeItems(3)} title="My Component" onClose={vi.fn()} />);
    expect(screen.getByText('My Component')).toBeDefined();
  });

  it('updates the radial hub label and description on hover', () => {
    render(<ContextMenu x={100} y={100} items={makeItems(3)} title="My Component" onClose={vi.fn()} />);
    fireEvent.mouseEnter(screen.getAllByRole('menuitem')[0]);
    expect(screen.getAllByText('Action 1').length).toBeGreaterThan(0);
    expect(screen.getByText('Does action 1')).toBeDefined();
  });

  it('renders item labels and descriptions in the linear variant', () => {
    render(<ContextMenu x={100} y={100} items={makeItems(8)} onClose={vi.fn()} />);
    expect(screen.getByText('Action 7')).toBeDefined();
    expect(screen.getByText('Does action 7')).toBeDefined();
  });

  describe.each<{ variant: ContextMenuVariant; count: number }>([
    { variant: 'radial', count: 3 },
    { variant: 'linear', count: 3 },
  ])('$variant variant dismissal', ({ variant, count }) => {
    it('invokes onClick and onClose once when an item is clicked', () => {
      const onClose = vi.fn();
      const items = makeItems(count);
      render(<ContextMenu x={100} y={100} items={items} variant={variant} onClose={onClose} />);
      fireEvent.click(screen.getAllByRole('menuitem')[1]);
      expect(items[1].onClick).toHaveBeenCalledTimes(1);
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it('does not invoke a disabled item', () => {
      const onClose = vi.fn();
      const items = makeItems(count);
      items[0].disabled = true;
      render(<ContextMenu x={100} y={100} items={items} variant={variant} onClose={onClose} />);
      fireEvent.click(screen.getAllByRole('menuitem')[0]);
      expect(items[0].onClick).not.toHaveBeenCalled();
      expect(onClose).not.toHaveBeenCalled();
    });

    it('closes on click outside', () => {
      const onClose = vi.fn();
      render(<ContextMenu x={100} y={100} items={makeItems(count)} variant={variant} onClose={onClose} />);
      fireEvent.mouseDown(document.body);
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it('does not close on click inside', () => {
      const onClose = vi.fn();
      render(<ContextMenu x={100} y={100} items={makeItems(count)} variant={variant} onClose={onClose} />);
      fireEvent.mouseDown(screen.getAllByRole('menuitem')[0]);
      expect(onClose).not.toHaveBeenCalled();
    });

    it('moves keyboard focus into the menu on open', async () => {
      render(<ContextMenu x={100} y={100} items={makeItems(count)} variant={variant} onClose={vi.fn()} />);
      await waitFor(() => expect(menuRoot()).toContainElement(document.activeElement as HTMLElement));
    });

    it('closes on Escape', () => {
      const onClose = vi.fn();
      render(<ContextMenu x={100} y={100} items={makeItems(count)} variant={variant} onClose={onClose} />);
      fireEvent.keyDown(menuRoot(), { key: 'Escape' });
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });
});
