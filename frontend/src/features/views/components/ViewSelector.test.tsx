import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { toViewId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { ViewSelector } from './ViewSelector';

const mockUseViews = vi.fn();
const mockUseActiveUsers = vi.fn();

vi.mock('../hooks/useViews', () => ({
  useViews: () => mockUseViews(),
}));

vi.mock('../../users/hooks/useUsers', () => ({
  useActiveUsers: () => mockUseActiveUsers(),
}));

vi.mock('../hooks/useCurrentView', () => ({
  useCurrentView: () => ({ currentView: { id: useAppStore.getState().currentViewId } }),
}));

const v1 = toViewId('view-1');
const v2 = toViewId('view-2');
const v3 = toViewId('view-3');

const views = [
  { id: v1, name: 'View One', isDefault: false, isPrivate: false, description: '' },
  { id: v2, name: 'View Two', isDefault: false, isPrivate: false, description: '' },
  { id: v3, name: 'View Three', isDefault: false, isPrivate: false, description: '' },
];

const allFiltersEnabled = {
  edges: { relation: true, realization: true, parentage: true, origin: true },
  types: { component: true, capability: true, originEntity: true },
};

function makeDirtyDraftFor(viewId: string) {
  useAppStore.setState({
    draftsByView: {
      [viewId]: {
        original: { entities: [], positions: {} },
        entities: [{ id: 'X', type: 'component' }],
        positions: { X: { x: 0, y: 0 } },
        filters: allFiltersEnabled,
        relations: [],
      },
    },
  });
}

describe('ViewSelector', () => {
  beforeEach(() => {
    mockUseViews.mockReturnValue({ data: views });
    mockUseActiveUsers.mockReturnValue({ data: [] });
    useAppStore.setState({
      currentViewId: v1,
      openViewIds: [v1, v2],
      dynamicOriginal: null,
      dynamicViewId: null,
      dynamicEntities: [],
      dynamicPositions: {},
      draftsByView: {},
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders only views in openViewIds, in order', () => {
    render(<ViewSelector />);

    const tabs = screen.getAllByRole('button', { name: /^Switch to/ });
    expect(tabs).toHaveLength(2);
    expect(tabs[0]).toHaveTextContent('View One');
    expect(tabs[1]).toHaveTextContent('View Two');
  });

  it('shows a dirty indicator only on dirty views', () => {
    makeDirtyDraftFor(v2);

    render(<ViewSelector />);

    const dirtyDots = screen.getAllByLabelText('Unsaved changes');
    expect(dirtyDots).toHaveLength(1);
  });

  it('disables the close button when only one tab is open', () => {
    useAppStore.setState({ openViewIds: [v1] });
    render(<ViewSelector />);

    const closeBtn = screen.getByRole('button', { name: 'Close View One' });
    expect(closeBtn).toBeDisabled();
  });

  it('closes a clean tab immediately without showing a dialog', () => {
    render(<ViewSelector />);
    fireEvent.click(screen.getByRole('button', { name: 'Close View Two' }));

    expect(screen.queryByRole('alertdialog')).toBeNull();
    expect(useAppStore.getState().openViewIds).toEqual([v1]);
  });

  it('shows confirmation dialog and discards on confirm for a dirty tab', () => {
    makeDirtyDraftFor(v2);
    render(<ViewSelector />);

    fireEvent.click(screen.getByRole('button', { name: 'Close View Two' }));
    expect(screen.getByRole('alertdialog')).toBeDefined();
    fireEvent.click(screen.getByRole('button', { name: 'Discard & close' }));

    expect(useAppStore.getState().openViewIds).toEqual([v1]);
    expect(useAppStore.getState().draftsByView[v2]).toBeUndefined();
  });

  it('keeps editing on cancel from confirmation dialog', () => {
    makeDirtyDraftFor(v2);
    render(<ViewSelector />);

    fireEvent.click(screen.getByRole('button', { name: 'Close View Two' }));
    fireEvent.click(screen.getByRole('button', { name: 'Keep editing' }));

    expect(useAppStore.getState().openViewIds).toEqual([v1, v2]);
    expect(useAppStore.getState().draftsByView[v2]).toBeDefined();
  });

  it('switches to a neighbor when closing the active tab', () => {
    useAppStore.setState({ currentViewId: v2, openViewIds: [v1, v2, v3] });
    render(<ViewSelector />);

    fireEvent.click(screen.getByRole('button', { name: 'Close View Two' }));

    expect(useAppStore.getState().currentViewId).toBe(v3);
    expect(useAppStore.getState().openViewIds).toEqual([v1, v3]);
  });

  it('switches to the left neighbor when closing the rightmost active tab', () => {
    useAppStore.setState({ currentViewId: v3, openViewIds: [v1, v2, v3] });
    render(<ViewSelector />);

    fireEvent.click(screen.getByRole('button', { name: 'Close View Three' }));

    expect(useAppStore.getState().currentViewId).toBe(v2);
  });
});
