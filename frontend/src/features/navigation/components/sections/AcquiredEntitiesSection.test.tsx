import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AcquiredEntitiesSection } from './AcquiredEntitiesSection';
import type { AcquiredEntity, AcquiredEntityId, IntegrationStatus, HATEOASLinks, ViewId, ComponentId } from '../../../../api/types';

vi.mock('../../../canvas/context/CanvasLayoutContext', () => ({
  useCanvasLayoutContext: vi.fn(() => ({
    positions: {},
    isLoading: false,
    error: null,
    updateComponentPosition: vi.fn(),
    updateCapabilityPosition: vi.fn(),
    updateOriginEntityPosition: vi.fn(),
    batchUpdatePositions: vi.fn(),
    getPositionForElement: vi.fn(),
    refetch: vi.fn(),
  })),
}));

import { useCanvasLayoutContext } from '../../../canvas/context/CanvasLayoutContext';

describe('AcquiredEntitiesSection', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/test', method: 'GET' } };

  const createMockEntity = (overrides = {}): AcquiredEntity => ({
    id: 'ae-123' as AcquiredEntityId,
    name: 'TechCorp',
    acquisitionDate: '2021-03-15',
    integrationStatus: 'InProgress' as IntegrationStatus,
    notes: 'Cloud infrastructure company',
    componentCount: 5,
    createdAt: '2021-01-01T00:00:00Z',
    _links: mockLinks,
    ...overrides,
  });

  const defaultProps = {
    acquiredEntities: [],
    currentView: null,
    selectedEntityId: null,
    isExpanded: true,
    onToggle: vi.fn(),
    onAddEntity: vi.fn(),
    onEntitySelect: vi.fn(),
    onEntityContextMenu: vi.fn(),
  };

  beforeEach(() => {
    vi.mocked(useCanvasLayoutContext).mockReturnValue({
      positions: {},
      isLoading: false,
      error: null,
      updateComponentPosition: vi.fn(),
      updateCapabilityPosition: vi.fn(),
      updateOriginEntityPosition: vi.fn(),
      batchUpdatePositions: vi.fn(),
      getPositionForElement: vi.fn(),
      refetch: vi.fn(),
    });
  });

  describe('rendering', () => {
    it('should display section label with count', () => {
      const entities = [createMockEntity(), createMockEntity({ id: 'ae-456', name: 'AcmeCo' })];
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={entities} />);

      expect(screen.getByText('Acquired Entities')).toBeInTheDocument();
      expect(screen.getByText('2')).toBeInTheDocument();
    });

    it('should display empty message when no entities exist', () => {
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={[]} />);

      expect(screen.getByText('No acquired entities')).toBeInTheDocument();
    });

    it('should display entity names', () => {
      const entities = [
        createMockEntity({ id: 'ae-1', name: 'TechCorp' }),
        createMockEntity({ id: 'ae-2', name: 'AcmeCo' }),
      ];
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={entities} />);

      expect(screen.getByText(/TechCorp/)).toBeInTheDocument();
      expect(screen.getByText(/AcmeCo/)).toBeInTheDocument();
    });

    it('should display acquisition year in parentheses when date is available', () => {
      const entity = createMockEntity({ acquisitionDate: '2021-03-15' });
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={[entity]} />);

      expect(screen.getByText(/TechCorp.*\(2021\)/)).toBeInTheDocument();
    });

    it('should not display year when acquisition date is undefined', () => {
      const entity = createMockEntity({ acquisitionDate: undefined });
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={[entity]} />);

      expect(screen.getByText('TechCorp')).toBeInTheDocument();
      expect(screen.queryByText(/\(2021\)/)).not.toBeInTheDocument();
    });

    it('should not render children when collapsed', () => {
      const entity = createMockEntity();
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={[entity]} isExpanded={false} />);

      expect(screen.queryByText(/TechCorp/)).not.toBeInTheDocument();
    });

    it('should render children when expanded', () => {
      const entity = createMockEntity();
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={[entity]} isExpanded={true} />);

      expect(screen.getByText(/TechCorp/)).toBeInTheDocument();
    });
  });

  describe('search functionality', () => {
    it('should render search input', () => {
      render(<AcquiredEntitiesSection {...defaultProps} />);

      expect(screen.getByPlaceholderText('Search acquired entities...')).toBeInTheDocument();
    });

    it('should filter entities by name', () => {
      const entities = [
        createMockEntity({ id: 'ae-1', name: 'TechCorp' }),
        createMockEntity({ id: 'ae-2', name: 'AcmeCo' }),
        createMockEntity({ id: 'ae-3', name: 'DataInc' }),
      ];
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={entities} />);

      const searchInput = screen.getByPlaceholderText('Search acquired entities...');
      fireEvent.change(searchInput, { target: { value: 'tech' } });

      expect(screen.getByText(/TechCorp/)).toBeInTheDocument();
      expect(screen.queryByText(/AcmeCo/)).not.toBeInTheDocument();
      expect(screen.queryByText(/DataInc/)).not.toBeInTheDocument();
    });

    it('should filter entities by notes', () => {
      const entities = [
        createMockEntity({ id: 'ae-1', name: 'TechCorp', notes: 'Cloud infrastructure' }),
        createMockEntity({ id: 'ae-2', name: 'AcmeCo', notes: 'Finance software' }),
      ];
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={entities} />);

      const searchInput = screen.getByPlaceholderText('Search acquired entities...');
      fireEvent.change(searchInput, { target: { value: 'cloud' } });

      expect(screen.getByText(/TechCorp/)).toBeInTheDocument();
      expect(screen.queryByText(/AcmeCo/)).not.toBeInTheDocument();
    });

    it('should show no matches message when search yields no results', () => {
      const entities = [createMockEntity({ name: 'TechCorp' })];
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={entities} />);

      const searchInput = screen.getByPlaceholderText('Search acquired entities...');
      fireEvent.change(searchInput, { target: { value: 'nonexistent' } });

      expect(screen.getByText('No matches')).toBeInTheDocument();
    });

    it('should show clear button when search has text', () => {
      render(<AcquiredEntitiesSection {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText('Search acquired entities...');
      fireEvent.change(searchInput, { target: { value: 'test' } });

      expect(screen.getByLabelText('Clear search')).toBeInTheDocument();
    });

    it('should clear search when clear button is clicked', () => {
      const entities = [createMockEntity()];
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={entities} />);

      const searchInput = screen.getByPlaceholderText('Search acquired entities...');
      fireEvent.change(searchInput, { target: { value: 'test' } });
      fireEvent.click(screen.getByLabelText('Clear search'));

      expect(searchInput).toHaveValue('');
    });

    it('should be case insensitive', () => {
      const entities = [createMockEntity({ name: 'TechCorp' })];
      render(<AcquiredEntitiesSection {...defaultProps} acquiredEntities={entities} />);

      const searchInput = screen.getByPlaceholderText('Search acquired entities...');
      fireEvent.change(searchInput, { target: { value: 'TECHCORP' } });

      expect(screen.getByText(/TechCorp/)).toBeInTheDocument();
    });
  });

  describe('selection', () => {
    it('should call onEntitySelect when entity is clicked', () => {
      const onEntitySelect = vi.fn();
      const entity = createMockEntity({ id: 'ae-123' as AcquiredEntityId });
      render(
        <AcquiredEntitiesSection
          {...defaultProps}
          acquiredEntities={[entity]}
          onEntitySelect={onEntitySelect}
        />
      );

      fireEvent.click(screen.getByTitle('TechCorp'));

      expect(onEntitySelect).toHaveBeenCalledWith('ae-123');
    });

    it('should apply selected class when entity is selected', () => {
      const entity = createMockEntity({ id: 'ae-123' as AcquiredEntityId });
      render(
        <AcquiredEntitiesSection
          {...defaultProps}
          acquiredEntities={[entity]}
          selectedEntityId="ae-123"
        />
      );

      const entityButton = screen.getByTitle('TechCorp');
      expect(entityButton).toHaveClass('selected');
    });

    it('should not apply selected class when entity is not selected', () => {
      const entity = createMockEntity({ id: 'ae-123' as AcquiredEntityId });
      render(
        <AcquiredEntitiesSection
          {...defaultProps}
          acquiredEntities={[entity]}
          selectedEntityId="ae-456"
        />
      );

      const entityButton = screen.getByTitle('TechCorp');
      expect(entityButton).not.toHaveClass('selected');
    });
  });

  describe('context menu', () => {
    it('should call onEntityContextMenu on right click', () => {
      const onEntityContextMenu = vi.fn();
      const entity = createMockEntity();
      render(
        <AcquiredEntitiesSection
          {...defaultProps}
          acquiredEntities={[entity]}
          onEntityContextMenu={onEntityContextMenu}
        />
      );

      fireEvent.contextMenu(screen.getByTitle('TechCorp'));

      expect(onEntityContextMenu).toHaveBeenCalledWith(expect.any(Object), entity);
    });
  });

  describe('drag and drop', () => {
    it('should always be draggable', () => {
      const entity = createMockEntity({ id: 'ae-123' as AcquiredEntityId });
      render(
        <AcquiredEntitiesSection
          {...defaultProps}
          acquiredEntities={[entity]}
        />
      );

      const entityButton = screen.getByTitle('TechCorp');
      expect(entityButton).toHaveAttribute('draggable', 'true');
    });

    it('should set acquiredEntityId on drag start', () => {
      const entity = createMockEntity({ id: 'ae-123' as AcquiredEntityId });
      render(
        <AcquiredEntitiesSection
          {...defaultProps}
          acquiredEntities={[entity]}
        />
      );

      const entityButton = screen.getByTitle('TechCorp');
      const mockDataTransfer = {
        setData: vi.fn(),
        effectAllowed: '',
      };

      fireEvent.dragStart(entityButton, { dataTransfer: mockDataTransfer });

      expect(mockDataTransfer.setData).toHaveBeenCalledWith('acquiredEntityId', 'ae-123');
      expect(mockDataTransfer.effectAllowed).toBe('copy');
    });
  });

  describe('add button', () => {
    it('should call onAddEntity when add button is clicked', () => {
      const onAddEntity = vi.fn();
      render(<AcquiredEntitiesSection {...defaultProps} onAddEntity={onAddEntity} />);

      fireEvent.click(screen.getByTestId('create-acquired-entity-button'));

      expect(onAddEntity).toHaveBeenCalled();
    });

    it('should have correct title for add button', () => {
      render(<AcquiredEntitiesSection {...defaultProps} />);

      expect(screen.getByTitle('Create new acquired entity')).toBeInTheDocument();
    });
  });

  describe('toggle', () => {
    it('should call onToggle when header is clicked', () => {
      const onToggle = vi.fn();
      render(<AcquiredEntitiesSection {...defaultProps} onToggle={onToggle} />);

      fireEvent.click(screen.getByText('Acquired Entities'));

      expect(onToggle).toHaveBeenCalled();
    });
  });

  describe('on-canvas status', () => {
    const createMockView = (id: string, componentIds: string[]) => ({
      id: id as ViewId,
      name: 'Test View',
      isDefault: false,
      isPrivate: false,
      components: componentIds.map(cid => ({ componentId: cid as ComponentId, x: 0, y: 0 })),
      capabilities: [],
      createdAt: '2021-01-01T00:00:00Z',
      _links: mockLinks,
    });

    it('should show entity as on-canvas when it has a layout position', () => {
      vi.mocked(useCanvasLayoutContext).mockReturnValue({
        positions: { 'acq-ae-123': { x: 100, y: 200 } },
        isLoading: false,
        error: null,
        updateComponentPosition: vi.fn(),
        updateCapabilityPosition: vi.fn(),
        updateOriginEntityPosition: vi.fn(),
        batchUpdatePositions: vi.fn(),
        getPositionForElement: vi.fn(),
        refetch: vi.fn(),
      });

      const entity = createMockEntity({ id: 'ae-123' as AcquiredEntityId, name: 'TechCorp' });
      const currentView = createMockView('view-1', ['comp-456']);

      render(
        <AcquiredEntitiesSection
          {...defaultProps}
          acquiredEntities={[entity]}
          currentView={currentView}
        />
      );

      const entityButton = screen.getByTitle('TechCorp');
      expect(entityButton).not.toHaveClass('not-in-view');
    });

    it('should show entity as not-on-canvas when it has no layout position', () => {
      vi.mocked(useCanvasLayoutContext).mockReturnValue({
        positions: {},
        isLoading: false,
        error: null,
        updateComponentPosition: vi.fn(),
        updateCapabilityPosition: vi.fn(),
        updateOriginEntityPosition: vi.fn(),
        batchUpdatePositions: vi.fn(),
        getPositionForElement: vi.fn(),
        refetch: vi.fn(),
      });

      const entity = createMockEntity({ id: 'ae-123' as AcquiredEntityId, name: 'TechCorp' });
      const currentView = createMockView('view-1', ['comp-456']);

      render(
        <AcquiredEntitiesSection
          {...defaultProps}
          acquiredEntities={[entity]}
          currentView={currentView}
        />
      );

      const entityButton = screen.getByTitle('TechCorp (not on canvas)');
      expect(entityButton).toHaveClass('not-in-view');
    });

    it('should show all entities as on-canvas when no current view is selected', () => {
      vi.mocked(useCanvasLayoutContext).mockReturnValue({
        positions: {},
        isLoading: false,
        error: null,
        updateComponentPosition: vi.fn(),
        updateCapabilityPosition: vi.fn(),
        updateOriginEntityPosition: vi.fn(),
        batchUpdatePositions: vi.fn(),
        getPositionForElement: vi.fn(),
        refetch: vi.fn(),
      });

      const entity = createMockEntity({ id: 'ae-123' as AcquiredEntityId, name: 'TechCorp' });

      render(
        <AcquiredEntitiesSection
          {...defaultProps}
          acquiredEntities={[entity]}
          currentView={null}
        />
      );

      const entityButton = screen.getByTitle('TechCorp');
      expect(entityButton).not.toHaveClass('not-in-view');
    });
  });
});
