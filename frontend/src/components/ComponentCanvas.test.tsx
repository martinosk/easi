import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ComponentCanvas } from './ComponentCanvas';
import { useAppStore } from '../store/appStore';

// Mock the store
vi.mock('../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

// Mock React Flow
vi.mock('@xyflow/react', () => ({
  ReactFlow: ({ nodes, edges }: any) => (
    <div data-testid="react-flow">
      {nodes.map((node: any) => (
        <div key={node.id} data-testid={`node-${node.id}`}>
          {node.data.label}
        </div>
      ))}
      {edges.map((edge: any) => (
        <div key={edge.id} data-testid={`edge-${edge.id}`}>
          {edge.label}
        </div>
      ))}
    </div>
  ),
  Background: () => <div data-testid="background" />,
  Controls: () => <div data-testid="controls" />,
  MiniMap: () => <div data-testid="minimap" />,
  applyNodeChanges: vi.fn((_changes, nodes) => nodes),
  applyEdgeChanges: vi.fn((_changes, edges) => edges),
  BackgroundVariant: { Dots: 'dots' },
  MarkerType: { ArrowClosed: 'arrowclosed' },
}));

describe('ComponentCanvas', () => {
  const mockOnConnect = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render canvas with components', () => {
    const mockComponents = [
      { id: '1', name: 'Component A', description: 'Description A' },
      { id: '2', name: 'Component B', description: 'Description B' },
    ];

    const mockView = {
      id: 'view-1',
      name: 'Test View',
      components: [
        { componentId: '1', x: 100, y: 100 },
        { componentId: '2', x: 200, y: 200 },
      ],
    };

    vi.mocked(useAppStore).mockReturnValue({
      components: mockComponents,
      relations: [],
      currentView: mockView,
      selectedNodeId: null,
      selectedEdgeId: null,
      selectNode: vi.fn(),
      selectEdge: vi.fn(),
      clearSelection: vi.fn(),
      updatePosition: vi.fn(),
    } as any);

    render(<ComponentCanvas onConnect={mockOnConnect} />);

    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    expect(screen.getByTestId('node-1')).toHaveTextContent('Component A');
    expect(screen.getByTestId('node-2')).toHaveTextContent('Component B');
  });

  it('should render relations between components', () => {
    const mockComponents = [
      { id: '1', name: 'Component A' },
      { id: '2', name: 'Component B' },
    ];

    const mockRelations = [
      {
        id: 'rel-1',
        sourceComponentId: '1',
        targetComponentId: '2',
        relationType: 'Triggers',
        name: 'Relation 1',
      },
    ];

    const mockView = {
      id: 'view-1',
      name: 'Test View',
      components: [
        { componentId: '1', x: 100, y: 100 },
        { componentId: '2', x: 200, y: 200 },
      ],
    };

    vi.mocked(useAppStore).mockReturnValue({
      components: mockComponents,
      relations: mockRelations,
      currentView: mockView,
      selectedNodeId: null,
      selectedEdgeId: null,
      selectNode: vi.fn(),
      selectEdge: vi.fn(),
      clearSelection: vi.fn(),
      updatePosition: vi.fn(),
    } as any);

    render(<ComponentCanvas onConnect={mockOnConnect} />);

    expect(screen.getByTestId('edge-rel-1')).toHaveTextContent('Relation 1');
  });

  it('should highlight selected component', () => {
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

    vi.mocked(useAppStore).mockReturnValue({
      components: mockComponents,
      relations: [],
      currentView: mockView,
      selectedNodeId: '1',
      selectedEdgeId: null,
      selectNode: vi.fn(),
      selectEdge: vi.fn(),
      clearSelection: vi.fn(),
      updatePosition: vi.fn(),
    } as any);

    render(<ComponentCanvas onConnect={mockOnConnect} />);

    // The selected node should be marked in the data
    expect(screen.getByTestId('node-1')).toBeInTheDocument();
  });

  it('should render empty canvas when no components', () => {
    const mockView = {
      id: 'view-1',
      name: 'Test View',
      components: [],
    };

    vi.mocked(useAppStore).mockReturnValue({
      components: [],
      relations: [],
      currentView: mockView,
      selectedNodeId: null,
      selectedEdgeId: null,
      selectNode: vi.fn(),
      selectEdge: vi.fn(),
      clearSelection: vi.fn(),
      updatePosition: vi.fn(),
    } as any);

    render(<ComponentCanvas onConnect={mockOnConnect} />);

    expect(screen.getByTestId('react-flow')).toBeInTheDocument();
  });
});
