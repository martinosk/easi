import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { HATEOASLinks, ViewId } from '../../../../api/types';
import type { NodeContextMenu as NodeContextMenuType } from '../../hooks/useContextMenu';
import { type GenerateViewTarget, NodeContextMenu } from './NodeContextMenu';

vi.mock('../../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({ currentViewId: 'view-1' as ViewId }),
}));

vi.mock('../../../views/hooks/useViews', () => ({
  useRemoveComponentFromView: () => ({ mutate: vi.fn() }),
  useRemoveCapabilityFromView: () => ({ mutate: vi.fn() }),
  useRemoveOriginEntityFromView: () => ({ mutate: vi.fn() }),
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

describe('NodeContextMenu - Generate View', () => {
  it('shows generate view item for component when canCreateView is true', () => {
    renderMenu({ menu: makeMenu(), canCreateView: true, onRequestGenerateView: vi.fn() });
    expect(screen.getByText('Generate View for Test Component')).toBeDefined();
  });

  it('shows generate view item for capability node', () => {
    renderMenu({
      menu: makeMenu({ nodeType: 'capability', nodeId: 'cap-1', nodeName: 'My Capability' }),
      canCreateView: true,
      onRequestGenerateView: vi.fn(),
    });
    expect(screen.getByText('Generate View for My Capability')).toBeDefined();
  });

  it('shows generate view item for origin entity node', () => {
    renderMenu({
      menu: makeMenu({ nodeType: 'originEntity', nodeId: 'vendor-1', nodeName: 'Acme Corp' }),
      canCreateView: true,
      onRequestGenerateView: vi.fn(),
    });
    expect(screen.getByText('Generate View for Acme Corp')).toBeDefined();
  });

  it('hides generate view item when canCreateView is false', () => {
    renderMenu({ menu: makeMenu(), canCreateView: false, onRequestGenerateView: vi.fn() });
    expect(screen.queryByText(/Generate View/)).toBeNull();
  });

  it('hides generate view item when onRequestGenerateView is not provided', () => {
    renderMenu({ menu: makeMenu(), canCreateView: true });
    expect(screen.queryByText(/Generate View/)).toBeNull();
  });

  it('truncates long entity names at 30 chars with ellipsis', () => {
    const longName = 'A Very Long Entity Name That Exceeds Thirty Characters';
    renderMenu({
      menu: makeMenu({ nodeName: longName }),
      canCreateView: true,
      onRequestGenerateView: vi.fn(),
    });
    const menuItem = screen.getByText(/Generate View for/);
    expect(menuItem.textContent).toContain('\u2026');
    expect(menuItem.textContent!.length).toBeLessThan(`Generate View for ${longName}`.length);
  });

  it('calls onRequestGenerateView with correct target on click', () => {
    const onGenerateView = vi.fn();
    renderMenu({
      menu: makeMenu({ nodeId: 'comp-42', nodeName: 'My System', nodeType: 'component' }),
      canCreateView: true,
      onRequestGenerateView: onGenerateView,
    });

    fireEvent.click(screen.getByText('Generate View for My System'));

    expect(onGenerateView).toHaveBeenCalledWith({
      entityRef: { id: 'comp-42', type: 'component' },
      entityName: 'My System',
    });
  });
});
