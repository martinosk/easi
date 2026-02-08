import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MultiSelectContextMenu } from './MultiSelectContextMenu';
import type { MultiSelectMenuState } from '../../hooks/useMultiSelectContextMenu';
import type { NodeContextMenu } from '../../hooks/useNodeContextMenu';

function makeNode(overrides: Partial<NodeContextMenu> = {}): NodeContextMenu {
  const nodeId = overrides.nodeId ?? 'test-id';
  return {
    x: 0,
    y: 0,
    nodeId,
    nodeName: 'Test Node',
    nodeType: 'component',
    ...overrides,
  };
}

function makeMenu(overrides: Partial<MultiSelectMenuState> = {}): MultiSelectMenuState {
  return {
    x: 100,
    y: 200,
    selectedNodes: [makeNode({ nodeId: '1' }), makeNode({ nodeId: '2' })],
    actions: [
      { type: 'removeFromView', label: 'Remove from View (2 items)', isDanger: false },
      { type: 'deleteFromModel', label: 'Delete from Model (2 items)', isDanger: true },
    ],
    ...overrides,
  };
}

describe('MultiSelectContextMenu', () => {
  it('renders nothing when menu is null', () => {
    const { container } = render(
      <MultiSelectContextMenu
        menu={null}
        onClose={vi.fn()}
        onRequestBulkOperation={vi.fn()}
      />
    );
    expect(container.innerHTML).toBe('');
  });

  it('renders menu items with correct labels', () => {
    render(
      <MultiSelectContextMenu
        menu={makeMenu()}
        onClose={vi.fn()}
        onRequestBulkOperation={vi.fn()}
      />
    );

    expect(screen.getByText('Remove from View (2 items)')).toBeDefined();
    expect(screen.getByText('Delete from Model (2 items)')).toBeDefined();
  });

  it('calls onRequestBulkOperation with removeFromView on click', () => {
    const onRequestBulkOperation = vi.fn();
    const onClose = vi.fn();

    render(
      <MultiSelectContextMenu
        menu={makeMenu()}
        onClose={onClose}
        onRequestBulkOperation={onRequestBulkOperation}
      />
    );

    fireEvent.click(screen.getByText('Remove from View (2 items)'));

    expect(onRequestBulkOperation).toHaveBeenCalledWith({
      type: 'removeFromView',
      nodes: expect.any(Array),
    });
  });

  it('calls onRequestBulkOperation with deleteFromModel on click', () => {
    const onRequestBulkOperation = vi.fn();
    const onClose = vi.fn();

    render(
      <MultiSelectContextMenu
        menu={makeMenu()}
        onClose={onClose}
        onRequestBulkOperation={onRequestBulkOperation}
      />
    );

    fireEvent.click(screen.getByText('Delete from Model (2 items)'));

    expect(onRequestBulkOperation).toHaveBeenCalledWith({
      type: 'deleteFromModel',
      nodes: expect.any(Array),
    });
  });

  it('renders only available actions', () => {
    const menu = makeMenu({
      actions: [
        { type: 'removeFromView', label: 'Remove from View (3 items)', isDanger: false },
      ],
    });

    render(
      <MultiSelectContextMenu
        menu={menu}
        onClose={vi.fn()}
        onRequestBulkOperation={vi.fn()}
      />
    );

    expect(screen.getByText('Remove from View (3 items)')).toBeDefined();
    expect(screen.queryByText(/Delete from Model/)).toBeNull();
  });
});
