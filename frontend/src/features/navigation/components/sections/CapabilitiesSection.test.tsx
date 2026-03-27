import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CapabilitiesSection } from './CapabilitiesSection';
import type { Capability, CapabilityId, HATEOASLinks } from '../../../../api/types';

vi.mock('../../../../hooks/useMaturityColorScale', () => ({
  useMaturityColorScale: () => ({
    getColorForValue: () => '#ccc',
    getSectionNameForValue: () => 'Medium',
  }),
}));

describe('CapabilitiesSection', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/test', method: 'GET' } };

  const createCapability = (overrides: Partial<Capability> = {}): Capability => ({
    id: 'cap-1' as CapabilityId,
    name: 'Capability One',
    level: 'L1',
    createdAt: '2024-01-01T00:00:00Z',
    _links: mockLinks,
    ...overrides,
  });

  const mockMultiSelect = {
    isMultiSelected: () => false,
    handleItemClick: () => 'single' as const,
    handleContextMenu: () => false,
    handleDragStart: () => false,
    selectedItems: [],
  };

  const defaultProps = {
    capabilities: [],
    currentView: null,
    isExpanded: true,
    onToggle: vi.fn(),
    onAddCapability: vi.fn(),
    onCapabilitySelect: vi.fn(),
    onCapabilityContextMenu: vi.fn(),
    expandedCapabilities: new Set<string>(),
    toggleCapabilityExpanded: vi.fn(),
    selectedCapabilityId: null,
    setSelectedCapabilityId: vi.fn(),
    multiSelect: mockMultiSelect,
  };

  describe('search functionality', () => {
    it('should render search input', () => {
      render(<CapabilitiesSection {...defaultProps} />);

      expect(screen.getByPlaceholderText('Search capabilities...')).toBeInTheDocument();
    });

    it('should filter capabilities by name', () => {
      const capabilities = [
        createCapability({ id: 'cap-1' as CapabilityId, name: 'Payment Processing' }),
        createCapability({ id: 'cap-2' as CapabilityId, name: 'Order Management' }),
        createCapability({ id: 'cap-3' as CapabilityId, name: 'Customer Support' }),
      ];
      render(<CapabilitiesSection {...defaultProps} capabilities={capabilities} />);

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: 'payment' },
      });

      expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      expect(screen.queryByText('Order Management')).not.toBeInTheDocument();
      expect(screen.queryByText('Customer Support')).not.toBeInTheDocument();
    });

    it('should filter capabilities by description', () => {
      const capabilities = [
        createCapability({ id: 'cap-1' as CapabilityId, name: 'Alpha', description: 'Handles invoicing' }),
        createCapability({ id: 'cap-2' as CapabilityId, name: 'Beta', description: 'Manages orders' }),
      ];
      render(<CapabilitiesSection {...defaultProps} capabilities={capabilities} />);

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: 'invoicing' },
      });

      expect(screen.getByText('Alpha')).toBeInTheDocument();
      expect(screen.queryByText('Beta')).not.toBeInTheDocument();
    });

    it('should show nested child when child matches search', () => {
      const capabilities = [
        createCapability({ id: 'cap-parent' as CapabilityId, name: 'Business', level: 'L1' }),
        createCapability({ id: 'cap-child' as CapabilityId, name: 'Invoicing', level: 'L2', parentId: 'cap-parent' as CapabilityId }),
      ];
      render(<CapabilitiesSection {...defaultProps} capabilities={capabilities} />);

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: 'invoicing' },
      });

      expect(screen.getByText('Business')).toBeInTheDocument();
      expect(screen.getByText('Invoicing')).toBeInTheDocument();
    });

    it('should auto-expand parents when nested child matches search', () => {
      const capabilities = [
        createCapability({ id: 'cap-parent' as CapabilityId, name: 'Business', level: 'L1' }),
        createCapability({ id: 'cap-child' as CapabilityId, name: 'Invoicing', level: 'L2', parentId: 'cap-parent' as CapabilityId }),
      ];
      render(
        <CapabilitiesSection
          {...defaultProps}
          capabilities={capabilities}
          expandedCapabilities={new Set()}
        />
      );

      expect(screen.queryByText('Invoicing')).not.toBeInTheDocument();

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: 'invoicing' },
      });

      expect(screen.getByText('Invoicing')).toBeInTheDocument();
    });

    it('should show deeply nested child when it matches search', () => {
      const capabilities = [
        createCapability({ id: 'cap-l1' as CapabilityId, name: 'Enterprise', level: 'L1' }),
        createCapability({ id: 'cap-l2' as CapabilityId, name: 'Finance', level: 'L2', parentId: 'cap-l1' as CapabilityId }),
        createCapability({ id: 'cap-l3' as CapabilityId, name: 'Tax Reporting', level: 'L3', parentId: 'cap-l2' as CapabilityId }),
      ];
      render(
        <CapabilitiesSection
          {...defaultProps}
          capabilities={capabilities}
          expandedCapabilities={new Set()}
        />
      );

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: 'tax' },
      });

      expect(screen.getByText('Enterprise')).toBeInTheDocument();
      expect(screen.getByText('Finance')).toBeInTheDocument();
      expect(screen.getByText('Tax Reporting')).toBeInTheDocument();
    });

    it('should hide non-matching branches entirely', () => {
      const capabilities = [
        createCapability({ id: 'cap-a' as CapabilityId, name: 'Branch A', level: 'L1' }),
        createCapability({ id: 'cap-a1' as CapabilityId, name: 'Match Here', level: 'L2', parentId: 'cap-a' as CapabilityId }),
        createCapability({ id: 'cap-b' as CapabilityId, name: 'Branch B', level: 'L1' }),
        createCapability({ id: 'cap-b1' as CapabilityId, name: 'No Match', level: 'L2', parentId: 'cap-b' as CapabilityId }),
      ];
      render(<CapabilitiesSection {...defaultProps} capabilities={capabilities} />);

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: 'match here' },
      });

      expect(screen.getByText('Branch A')).toBeInTheDocument();
      expect(screen.getByText('Match Here')).toBeInTheDocument();
      expect(screen.queryByText('Branch B')).not.toBeInTheDocument();
      expect(screen.queryByText('No Match')).not.toBeInTheDocument();
    });

    it('should show parent match without requiring children to match', () => {
      const capabilities = [
        createCapability({ id: 'cap-parent' as CapabilityId, name: 'Matching Parent', level: 'L1' }),
        createCapability({ id: 'cap-child' as CapabilityId, name: 'Unrelated Child', level: 'L2', parentId: 'cap-parent' as CapabilityId }),
      ];
      render(<CapabilitiesSection {...defaultProps} capabilities={capabilities} />);

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: 'matching parent' },
      });

      expect(screen.getByText('Matching Parent')).toBeInTheDocument();
    });

    it('should be case-insensitive', () => {
      const capabilities = [
        createCapability({ id: 'cap-1' as CapabilityId, name: 'Payment Processing' }),
      ];
      render(<CapabilitiesSection {...defaultProps} capabilities={capabilities} />);

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: 'PAYMENT' },
      });

      expect(screen.getByText('Payment Processing')).toBeInTheDocument();
    });

    it('should show no matches message when search yields no results', () => {
      const capabilities = [createCapability({ name: 'Something' })];
      render(<CapabilitiesSection {...defaultProps} capabilities={capabilities} />);

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: 'nonexistent' },
      });

      expect(screen.getByText('No matches')).toBeInTheDocument();
    });

    it('should show no capabilities message when list is empty', () => {
      render(<CapabilitiesSection {...defaultProps} capabilities={[]} />);

      expect(screen.getByText('No capabilities')).toBeInTheDocument();
    });

    it('should clear search when clear button is clicked', () => {
      const capabilities = [createCapability()];
      render(<CapabilitiesSection {...defaultProps} capabilities={capabilities} />);

      const searchInput = screen.getByPlaceholderText('Search capabilities...');
      fireEvent.change(searchInput, { target: { value: 'test' } });
      fireEvent.click(screen.getByLabelText('Clear search'));

      expect(searchInput).toHaveValue('');
    });

    it('should restore manual expand state when search is cleared', () => {
      const capabilities = [
        createCapability({ id: 'cap-parent' as CapabilityId, name: 'Parent', level: 'L1' }),
        createCapability({ id: 'cap-child' as CapabilityId, name: 'Child', level: 'L2', parentId: 'cap-parent' as CapabilityId }),
      ];
      render(
        <CapabilitiesSection
          {...defaultProps}
          capabilities={capabilities}
          expandedCapabilities={new Set()}
        />
      );

      expect(screen.queryByText('Child')).not.toBeInTheDocument();

      const searchInput = screen.getByPlaceholderText('Search capabilities...');
      fireEvent.change(searchInput, { target: { value: 'child' } });
      expect(screen.getByText('Child')).toBeInTheDocument();

      fireEvent.change(searchInput, { target: { value: '' } });
      expect(screen.queryByText('Child')).not.toBeInTheDocument();
    });

    it('should show all capabilities when search is whitespace only', () => {
      const capabilities = [
        createCapability({ id: 'cap-1' as CapabilityId, name: 'Alpha' }),
        createCapability({ id: 'cap-2' as CapabilityId, name: 'Beta' }),
      ];
      render(<CapabilitiesSection {...defaultProps} capabilities={capabilities} />);

      fireEvent.change(screen.getByPlaceholderText('Search capabilities...'), {
        target: { value: '   ' },
      });

      expect(screen.getByText('Alpha')).toBeInTheDocument();
      expect(screen.getByText('Beta')).toBeInTheDocument();
    });
  });
});
