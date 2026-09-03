import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createRef, forwardRef } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentCanvasRef } from '../../features/canvas';
import { renderWithProviders } from '../../test/helpers';
import { CanvasWorkspace } from './CanvasWorkspace';

vi.mock('../../features/canvas', async () => {
  const actual = await vi.importActual<typeof import('../../features/canvas')>('../../features/canvas');
  return {
    ...actual,
    ComponentCanvas: forwardRef((_props, ref) => <div ref={ref as never} data-testid="component-canvas-mock" />),
  };
});

vi.mock('../../features/navigation', () => ({
  NavigationTree: ({ onEditComponent }: { onEditComponent?: (componentId: string) => void }) => (
    <div data-testid="navigation-tree-mock">
      <button type="button" data-testid="tree-edit-mock" onClick={() => onEditComponent?.('comp-1')}>
        Edit
      </button>
    </div>
  ),
}));

vi.mock('../../features/views', async () => {
  const actual = await vi.importActual<typeof import('../../features/views')>('../../features/views');
  return {
    ...actual,
    ViewSelector: () => <div data-testid="view-selector-mock" />,
  };
});

vi.mock('../shared/DetailContentRenderer', () => ({
  DetailContentRendererWithPlaceholder: () => <div data-testid="detail-content-mock" />,
}));

const STORAGE_KEY = 'easi-canvas-panels';

function renderWorkspace(onComponentSelect = vi.fn()) {
  const canvasRef = createRef<ComponentCanvasRef>();
  return renderWithProviders(
    <CanvasWorkspace
      canvasRef={canvasRef}
      selectedNodeId={null}
      selectedEdgeId={null}
      onConnect={vi.fn()}
      onComponentDrop={vi.fn()}
      onComponentSelect={onComponentSelect}
      onCapabilitySelect={vi.fn()}
      onViewSelect={vi.fn()}
      onEditRelation={vi.fn()}
      onEditCapability={vi.fn()}
      onRemoveFromView={vi.fn()}
    />,
  );
}

describe('CanvasWorkspace', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('renders the explorer, canvas, and details panes', () => {
    renderWorkspace();

    expect(screen.getByTestId('explorer-pane')).toBeInTheDocument();
    expect(screen.getByTestId('canvas-pane')).toBeInTheDocument();
    expect(screen.getByTestId('details-pane')).toBeInTheDocument();
    expect(screen.getByTestId('navigation-tree-mock')).toBeInTheDocument();
    expect(screen.getByTestId('view-selector-mock')).toBeInTheDocument();
    expect(screen.getByTestId('component-canvas-mock')).toBeInTheDocument();
    expect(screen.getByTestId('detail-content-mock')).toBeInTheDocument();
    expect(within(screen.getByTestId('explorer-pane')).getByText('Explorer')).toBeInTheDocument();
    expect(within(screen.getByTestId('details-pane')).getByText('Details')).toBeInTheDocument();
  });

  it("routes the tree's Edit action to component selection so the details pane shows it", async () => {
    const onComponentSelect = vi.fn();
    renderWorkspace(onComponentSelect);

    await userEvent.click(screen.getByTestId('tree-edit-mock'));

    expect(onComponentSelect).toHaveBeenCalledWith('comp-1');
  });

  it('hides the explorer pane via its header collapse button, and reopens it from the floating button', async () => {
    renderWorkspace();

    await userEvent.click(screen.getByRole('button', { name: 'Hide explorer panel' }));
    expect(screen.queryByTestId('explorer-pane')).not.toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Show explorer panel' }));
    expect(screen.getByTestId('explorer-pane')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Show explorer panel' })).not.toBeInTheDocument();
  });

  it('hides the details pane via its header collapse button and offers a floating reopen button', async () => {
    renderWorkspace();

    await userEvent.click(screen.getByRole('button', { name: 'Hide details panel' }));
    expect(screen.queryByTestId('details-pane')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Show details panel' })).toBeInTheDocument();
  });

  it('persists panel visibility to localStorage', async () => {
    renderWorkspace();

    await userEvent.click(screen.getByRole('button', { name: 'Hide explorer panel' }));

    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '{}');
    expect(stored).toEqual({ navigation: false, details: true });
  });

  it('restores panel visibility from localStorage on mount', () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ navigation: false, details: true }));

    renderWorkspace();

    expect(screen.queryByTestId('explorer-pane')).not.toBeInTheDocument();
    expect(screen.getByTestId('details-pane')).toBeInTheDocument();
  });

  it('defaults both panels to visible when nothing is persisted', () => {
    renderWorkspace();

    expect(screen.getByTestId('explorer-pane')).toBeInTheDocument();
    expect(screen.getByTestId('details-pane')).toBeInTheDocument();
  });
});
