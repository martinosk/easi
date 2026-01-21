import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { InternalTeamsSection } from './InternalTeamsSection';
import type { InternalTeam, InternalTeamId, HATEOASLinks } from '../../../../api/types';

describe('InternalTeamsSection', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/test', method: 'GET' } };

  const createMockTeam = (overrides = {}): InternalTeam => ({
    id: 'it-123' as InternalTeamId,
    name: 'Platform Engineering',
    department: 'Technology',
    contactPerson: 'John Doe',
    notes: 'Core platform team',
    componentCount: 10,
    createdAt: '2021-01-01T00:00:00Z',
    _links: mockLinks,
    ...overrides,
  });

  const defaultProps = {
    internalTeams: [],
    selectedTeamId: null,
    isExpanded: true,
    onToggle: vi.fn(),
    onAddTeam: vi.fn(),
    onTeamSelect: vi.fn(),
    onTeamContextMenu: vi.fn(),
  };

  describe('rendering', () => {
    it('should display section label with count', () => {
      const teams = [createMockTeam(), createMockTeam({ id: 'it-456', name: 'Data Team' })];
      render(<InternalTeamsSection {...defaultProps} internalTeams={teams} />);

      expect(screen.getByText('Internal Teams')).toBeInTheDocument();
      expect(screen.getByText('2')).toBeInTheDocument();
    });

    it('should display empty message when no teams exist', () => {
      render(<InternalTeamsSection {...defaultProps} internalTeams={[]} />);

      expect(screen.getByText('No internal teams')).toBeInTheDocument();
    });

    it('should display team names', () => {
      const teams = [
        createMockTeam({ id: 'it-1', name: 'Platform Engineering' }),
        createMockTeam({ id: 'it-2', name: 'Data Team' }),
      ];
      render(<InternalTeamsSection {...defaultProps} internalTeams={teams} />);

      expect(screen.getByText('Platform Engineering')).toBeInTheDocument();
      expect(screen.getByText('Data Team')).toBeInTheDocument();
    });

    it('should not render children when collapsed', () => {
      const team = createMockTeam();
      render(<InternalTeamsSection {...defaultProps} internalTeams={[team]} isExpanded={false} />);

      expect(screen.queryByText('Platform Engineering')).not.toBeInTheDocument();
    });

    it('should render children when expanded', () => {
      const team = createMockTeam();
      render(<InternalTeamsSection {...defaultProps} internalTeams={[team]} isExpanded={true} />);

      expect(screen.getByText('Platform Engineering')).toBeInTheDocument();
    });
  });

  describe('search functionality', () => {
    it('should render search input', () => {
      render(<InternalTeamsSection {...defaultProps} />);

      expect(screen.getByPlaceholderText('Search internal teams...')).toBeInTheDocument();
    });

    it('should filter teams by name', () => {
      const teams = [
        createMockTeam({ id: 'it-1', name: 'Platform Engineering', department: 'Engineering', contactPerson: 'Jane Doe', notes: 'Core services' }),
        createMockTeam({ id: 'it-2', name: 'Data Team', department: 'Analytics', contactPerson: 'John Smith', notes: 'Data services' }),
        createMockTeam({ id: 'it-3', name: 'Finance IT', department: 'Finance', contactPerson: 'Bob Wilson', notes: 'Financial systems' }),
      ];
      render(<InternalTeamsSection {...defaultProps} internalTeams={teams} />);

      const searchInput = screen.getByPlaceholderText('Search internal teams...');
      fireEvent.change(searchInput, { target: { value: 'platform' } });

      expect(screen.getByText('Platform Engineering')).toBeInTheDocument();
      expect(screen.queryByText('Data Team')).not.toBeInTheDocument();
      expect(screen.queryByText('Finance IT')).not.toBeInTheDocument();
    });

    it('should filter teams by department', () => {
      const teams = [
        createMockTeam({ id: 'it-1', name: 'Platform Engineering', department: 'Technology' }),
        createMockTeam({ id: 'it-2', name: 'Finance IT', department: 'Finance' }),
      ];
      render(<InternalTeamsSection {...defaultProps} internalTeams={teams} />);

      const searchInput = screen.getByPlaceholderText('Search internal teams...');
      fireEvent.change(searchInput, { target: { value: 'technology' } });

      expect(screen.getByText('Platform Engineering')).toBeInTheDocument();
      expect(screen.queryByText('Finance IT')).not.toBeInTheDocument();
    });

    it('should filter teams by contact person', () => {
      const teams = [
        createMockTeam({ id: 'it-1', name: 'Platform Engineering', contactPerson: 'John Doe' }),
        createMockTeam({ id: 'it-2', name: 'Data Team', contactPerson: 'Jane Smith' }),
      ];
      render(<InternalTeamsSection {...defaultProps} internalTeams={teams} />);

      const searchInput = screen.getByPlaceholderText('Search internal teams...');
      fireEvent.change(searchInput, { target: { value: 'john' } });

      expect(screen.getByText('Platform Engineering')).toBeInTheDocument();
      expect(screen.queryByText('Data Team')).not.toBeInTheDocument();
    });

    it('should filter teams by notes', () => {
      const teams = [
        createMockTeam({ id: 'it-1', name: 'Platform Engineering', notes: 'Core platform services' }),
        createMockTeam({ id: 'it-2', name: 'Data Team', notes: 'Analytics and BI' }),
      ];
      render(<InternalTeamsSection {...defaultProps} internalTeams={teams} />);

      const searchInput = screen.getByPlaceholderText('Search internal teams...');
      fireEvent.change(searchInput, { target: { value: 'analytics' } });

      expect(screen.getByText('Data Team')).toBeInTheDocument();
      expect(screen.queryByText('Platform Engineering')).not.toBeInTheDocument();
    });

    it('should show no matches message when search yields no results', () => {
      const teams = [createMockTeam({ name: 'Platform Engineering' })];
      render(<InternalTeamsSection {...defaultProps} internalTeams={teams} />);

      const searchInput = screen.getByPlaceholderText('Search internal teams...');
      fireEvent.change(searchInput, { target: { value: 'nonexistent' } });

      expect(screen.getByText('No matches')).toBeInTheDocument();
    });

    it('should clear search when clear button is clicked', () => {
      const teams = [createMockTeam()];
      render(<InternalTeamsSection {...defaultProps} internalTeams={teams} />);

      const searchInput = screen.getByPlaceholderText('Search internal teams...');
      fireEvent.change(searchInput, { target: { value: 'test' } });
      fireEvent.click(screen.getByLabelText('Clear search'));

      expect(searchInput).toHaveValue('');
    });
  });

  describe('selection', () => {
    it('should call onTeamSelect when team is clicked', () => {
      const onTeamSelect = vi.fn();
      const team = createMockTeam({ id: 'it-123' as InternalTeamId });
      render(
        <InternalTeamsSection
          {...defaultProps}
          internalTeams={[team]}
          onTeamSelect={onTeamSelect}
        />
      );

      fireEvent.click(screen.getByTitle('Platform Engineering'));

      expect(onTeamSelect).toHaveBeenCalledWith('it-123');
    });

    it('should apply selected class when team is selected', () => {
      const team = createMockTeam({ id: 'it-123' as InternalTeamId });
      render(
        <InternalTeamsSection
          {...defaultProps}
          internalTeams={[team]}
          selectedTeamId="it-123"
        />
      );

      const teamButton = screen.getByTitle('Platform Engineering');
      expect(teamButton).toHaveClass('selected');
    });
  });

  describe('context menu', () => {
    it('should call onTeamContextMenu on right click', () => {
      const onTeamContextMenu = vi.fn();
      const team = createMockTeam();
      render(
        <InternalTeamsSection
          {...defaultProps}
          internalTeams={[team]}
          onTeamContextMenu={onTeamContextMenu}
        />
      );

      fireEvent.contextMenu(screen.getByTitle('Platform Engineering'));

      expect(onTeamContextMenu).toHaveBeenCalledWith(expect.any(Object), team);
    });
  });

  describe('drag and drop', () => {
    it('should always be draggable', () => {
      const team = createMockTeam({ id: 'it-123' as InternalTeamId });
      render(
        <InternalTeamsSection
          {...defaultProps}
          internalTeams={[team]}
        />
      );

      const teamButton = screen.getByTitle('Platform Engineering');
      expect(teamButton).toHaveAttribute('draggable', 'true');
    });

    it('should set internalTeamId on drag start', () => {
      const team = createMockTeam({ id: 'it-123' as InternalTeamId });
      render(
        <InternalTeamsSection
          {...defaultProps}
          internalTeams={[team]}
        />
      );

      const teamButton = screen.getByTitle('Platform Engineering');
      const mockDataTransfer = {
        setData: vi.fn(),
        effectAllowed: '',
      };

      fireEvent.dragStart(teamButton, { dataTransfer: mockDataTransfer });

      expect(mockDataTransfer.setData).toHaveBeenCalledWith('internalTeamId', 'it-123');
      expect(mockDataTransfer.effectAllowed).toBe('copy');
    });
  });

  describe('add button', () => {
    it('should call onAddTeam when add button is clicked', () => {
      const onAddTeam = vi.fn();
      render(<InternalTeamsSection {...defaultProps} onAddTeam={onAddTeam} />);

      fireEvent.click(screen.getByTestId('create-internal-team-button'));

      expect(onAddTeam).toHaveBeenCalled();
    });

    it('should have correct title for add button', () => {
      render(<InternalTeamsSection {...defaultProps} />);

      expect(screen.getByTitle('Create new internal team')).toBeInTheDocument();
    });
  });

  describe('toggle', () => {
    it('should call onToggle when header is clicked', () => {
      const onToggle = vi.fn();
      render(<InternalTeamsSection {...defaultProps} onToggle={onToggle} />);

      fireEvent.click(screen.getByText('Internal Teams'));

      expect(onToggle).toHaveBeenCalled();
    });
  });
});
