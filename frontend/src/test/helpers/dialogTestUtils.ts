import { vi } from 'vitest';
import { useAppStore } from '../../store/appStore';

export function setupDialogTest() {
  const mockCreateComponent = vi.fn();
  const mockUpdateComponent = vi.fn();
  const mockCreateRelation = vi.fn();
  const mockUpdateRelation = vi.fn();

  vi.mocked(useAppStore).mockImplementation((selector: any) =>
    selector({
      createComponent: mockCreateComponent,
      updateComponent: mockUpdateComponent,
      createRelation: mockCreateRelation,
      updateRelation: mockUpdateRelation,
      components: [],
      relations: [],
      currentView: null,
      selectedNodeId: null,
      selectedEdgeId: null,
      isLoading: false,
      error: null,
      loadData: vi.fn(),
      updatePosition: vi.fn(),
      selectNode: vi.fn(),
      selectEdge: vi.fn(),
      clearSelection: vi.fn(),
      setError: vi.fn(),
    })
  );

  return {
    mockCreateComponent,
    mockUpdateComponent,
    mockCreateRelation,
    mockUpdateRelation,
  };
}
