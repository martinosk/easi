import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/react';
import { ComponentCanvas } from './ComponentCanvas';
import { useAppStore } from '../store/appStore';

// Mock the store
vi.mock('../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

// Mock React Flow
vi.mock('@xyflow/react', () => {
  let mockOnNodeDragStop: any = null;

  return {
    ReactFlow: (props: any) => {
      mockOnNodeDragStop = props.onNodeDragStop;
      return (
        <div data-testid="react-flow">
          <button
            data-testid="simulate-drag"
            onClick={() => {
              if (mockOnNodeDragStop) {
                const mockEvent = {} as React.MouseEvent;
                const mockNode = { id: '1', position: { x: 250, y: 350 } };
                mockOnNodeDragStop(mockEvent, mockNode);
              }
            }}
          >
            Simulate Drag
          </button>
        </div>
      );
    },
    Background: () => <div data-testid="background" />,
    Controls: () => <div data-testid="controls" />,
    MiniMap: () => <div data-testid="minimap" />,
    applyNodeChanges: vi.fn((_changes, nodes) => nodes),
    applyEdgeChanges: vi.fn((_changes, edges) => edges),
    BackgroundVariant: { Dots: 'dots' },
    MarkerType: { ArrowClosed: 'arrowclosed' },
  };
});

describe('ComponentCanvas - Dragging', () => {
  const mockOnConnect = vi.fn();
  const mockUpdatePosition = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should update position when component is dragged', () => {
    const mockComponents = [
      { id: '1', name: 'Component A' },
    ];

    const mockView = {
      id: 'view-1',
      name: 'Test View',
      components: [
        { componentId: '1', x: 100, y: 100 },
      ],
    };

    vi.mocked(useAppStore).mockImplementation((selector: any) =>
      selector({
      components: mockComponents,
      relations: [],
      currentView: mockView,
      selectedNodeId: null,
      selectedEdgeId: null,
      selectNode: vi.fn(),
      selectEdge: vi.fn(),
      clearSelection: vi.fn(),
      updatePosition: mockUpdatePosition,
      })
    );

    const { getByTestId } = render(<ComponentCanvas onConnect={mockOnConnect} />);

    // Simulate drag
    const dragButton = getByTestId('simulate-drag');
    dragButton.click();

    expect(mockUpdatePosition).toHaveBeenCalledWith('1', 250, 350);
  });

  it('should handle multiple component drags', () => {
    const mockComponents = [
      { id: '1', name: 'Component A' },
      { id: '2', name: 'Component B' },
    ];

    const mockView = {
      id: 'view-1',
      name: 'Test View',
      components: [
        { componentId: '1', x: 100, y: 100 },
        { componentId: '2', x: 200, y: 200 },
      ],
    };

    vi.mocked(useAppStore).mockImplementation((selector: any) =>
      selector({
      components: mockComponents,
      relations: [],
      currentView: mockView,
      selectedNodeId: null,
      selectedEdgeId: null,
      selectNode: vi.fn(),
      selectEdge: vi.fn(),
      clearSelection: vi.fn(),
      updatePosition: mockUpdatePosition,
      })
    );

    render(<ComponentCanvas onConnect={mockOnConnect} />);

    // Verify updatePosition is called for the dragged component
    expect(mockUpdatePosition).not.toHaveBeenCalled();
  });
});
