import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { HATEOASLinks, ViewId } from '../../../../api/types';
import { useAppStore } from '../../../../store/appStore';
import type { NodeContextMenu as NodeContextMenuType } from '../../hooks/useContextMenu';
import { type GenerateViewTarget, NodeContextMenu } from './NodeContextMenu';

vi.mock('../../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({ currentViewId: 'view-1' as ViewId }),
}));

vi.mock('../../hooks/useDraftRemoveFromView', () => ({
  useDraftRemoveFromView: () => vi.fn(),
}));

const editableLinks: HATEOASLinks = {
  self: { href: '/test', method: 'GET' },
  delete: { href: '/test', method: 'DELETE' },
  'x-remove': { href: '/test', method: 'DELETE' },
};

function makeMenu(overrides: Partial<NodeContextMenuType> = {}): NodeContextMenuType {
  return {
    x: 100,
    y: 200,
    nodeId: 'comp-1',
    nodeName: 'Test Component',
    nodeType: 'component',
    modelLinks: editableLinks,
    viewElementLinks: editableLinks,
    ...overrides,
  };
}

function renderMenu(props: {
  menu: NodeContextMenuType | null;
  canCreateView?: boolean;
  onRequestGenerateView?: (target: GenerateViewTarget) => void;
}) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <NodeContextMenu
        menu={props.menu}
        onClose={vi.fn()}
        onRequestDelete={vi.fn()}
        canCreateView={props.canCreateView}
        onRequestGenerateView={props.onRequestGenerateView}
      />
    </QueryClientProvider>,
  );
}

describe('NodeContextMenu - Create dynamic view', () => {
  it('shows create dynamic view item for component when canCreateView is true', () => {
    renderMenu({ menu: makeMenu(), canCreateView: true, onRequestGenerateView: vi.fn() });
    expect(screen.getByText('Create dynamic view from Test Component')).toBeDefined();
  });

  it('shows create dynamic view item for capability node', () => {
    renderMenu({
      menu: makeMenu({ nodeType: 'capability', nodeId: 'cap-1', nodeName: 'My Capability' }),
      canCreateView: true,
      onRequestGenerateView: vi.fn(),
    });
    expect(screen.getByText('Create dynamic view from My Capability')).toBeDefined();
  });

  it('shows create dynamic view item for origin entity node', () => {
    renderMenu({
      menu: makeMenu({ nodeType: 'originEntity', nodeId: 'vendor-1', nodeName: 'Acme Corp' }),
      canCreateView: true,
      onRequestGenerateView: vi.fn(),
    });
    expect(screen.getByText('Create dynamic view from Acme Corp')).toBeDefined();
  });

  it('hides create dynamic view item when canCreateView is false', () => {
    renderMenu({ menu: makeMenu(), canCreateView: false, onRequestGenerateView: vi.fn() });
    expect(screen.queryByText(/Create dynamic view/)).toBeNull();
  });

  it('hides create dynamic view item when onRequestGenerateView is not provided', () => {
    renderMenu({ menu: makeMenu(), canCreateView: true });
    expect(screen.queryByText(/Create dynamic view/)).toBeNull();
  });

  it('truncates long entity names at 30 chars with ellipsis', () => {
    const longName = 'A Very Long Entity Name That Exceeds Thirty Characters';
    renderMenu({
      menu: makeMenu({ nodeName: longName }),
      canCreateView: true,
      onRequestGenerateView: vi.fn(),
    });
    const menuItem = screen.getByText(/Create dynamic view from/);
    expect(menuItem.textContent).toContain('\u2026');
    expect(menuItem.textContent!.length).toBeLessThan(`Create dynamic view from ${longName}`.length);
  });

  it('shows Remove from View for drafted entities even without view-element links', () => {
    act(() => {
      useAppStore.setState({
        dynamicViewId: 'view-1' as ViewId,
        dynamicEntities: [{ id: 'comp-42', type: 'component' }],
      });
    });

    renderMenu({
      menu: makeMenu({
        nodeId: 'comp-42',
        nodeName: 'Drafted Component',
        viewElementLinks: undefined,
      }),
    });

    expect(screen.getByText('Remove from View')).toBeDefined();

    act(() => {
      useAppStore.setState({ dynamicViewId: null, dynamicEntities: [] });
    });
  });

  it('calls onRequestGenerateView with correct target on click', () => {
    const onGenerateView = vi.fn();
    renderMenu({
      menu: makeMenu({ nodeId: 'comp-42', nodeName: 'My System', nodeType: 'component' }),
      canCreateView: true,
      onRequestGenerateView: onGenerateView,
    });

    fireEvent.click(screen.getByText('Create dynamic view from My System'));

    expect(onGenerateView).toHaveBeenCalledWith({
      entityRef: { id: 'comp-42', type: 'component' },
      entityName: 'My System',
    });
  });
});
