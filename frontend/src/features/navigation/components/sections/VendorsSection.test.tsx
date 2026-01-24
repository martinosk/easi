import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { VendorsSection } from './VendorsSection';
import type { Vendor, VendorId, HATEOASLinks } from '../../../../api/types';

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

describe('VendorsSection', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/test', method: 'GET' } };

  const createMockVendor = (overrides = {}): Vendor => ({
    id: 'v-123' as VendorId,
    name: 'SAP',
    implementationPartner: 'Accenture',
    notes: 'Enterprise ERP vendor',
    componentCount: 3,
    createdAt: '2021-01-01T00:00:00Z',
    _links: mockLinks,
    ...overrides,
  });

  const defaultProps = {
    vendors: [],
    currentView: null,
    selectedVendorId: null,
    isExpanded: true,
    onToggle: vi.fn(),
    onAddVendor: vi.fn(),
    onVendorSelect: vi.fn(),
    onVendorContextMenu: vi.fn(),
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
      const vendors = [createMockVendor(), createMockVendor({ id: 'v-456', name: 'Microsoft' })];
      render(<VendorsSection {...defaultProps} vendors={vendors} />);

      expect(screen.getByText('Vendors')).toBeInTheDocument();
      expect(screen.getByText('2')).toBeInTheDocument();
    });

    it('should display empty message when no vendors exist', () => {
      render(<VendorsSection {...defaultProps} vendors={[]} />);

      expect(screen.getByText('No vendors')).toBeInTheDocument();
    });

    it('should display vendor names', () => {
      const vendors = [
        createMockVendor({ id: 'v-1', name: 'SAP' }),
        createMockVendor({ id: 'v-2', name: 'Microsoft' }),
      ];
      render(<VendorsSection {...defaultProps} vendors={vendors} />);

      expect(screen.getByText('SAP')).toBeInTheDocument();
      expect(screen.getByText('Microsoft')).toBeInTheDocument();
    });

    it('should not render children when collapsed', () => {
      const vendor = createMockVendor();
      render(<VendorsSection {...defaultProps} vendors={[vendor]} isExpanded={false} />);

      expect(screen.queryByText('SAP')).not.toBeInTheDocument();
    });

    it('should render children when expanded', () => {
      const vendor = createMockVendor();
      render(<VendorsSection {...defaultProps} vendors={[vendor]} isExpanded={true} />);

      expect(screen.getByText('SAP')).toBeInTheDocument();
    });
  });

  describe('search functionality', () => {
    it('should render search input', () => {
      render(<VendorsSection {...defaultProps} />);

      expect(screen.getByPlaceholderText('Search vendors...')).toBeInTheDocument();
    });

    it('should filter vendors by name', () => {
      const vendors = [
        createMockVendor({ id: 'v-1', name: 'SAP' }),
        createMockVendor({ id: 'v-2', name: 'Microsoft' }),
        createMockVendor({ id: 'v-3', name: 'Salesforce' }),
      ];
      render(<VendorsSection {...defaultProps} vendors={vendors} />);

      const searchInput = screen.getByPlaceholderText('Search vendors...');
      fireEvent.change(searchInput, { target: { value: 'sap' } });

      expect(screen.getByText('SAP')).toBeInTheDocument();
      expect(screen.queryByText('Microsoft')).not.toBeInTheDocument();
      expect(screen.queryByText('Salesforce')).not.toBeInTheDocument();
    });

    it('should filter vendors by implementation partner', () => {
      const vendors = [
        createMockVendor({ id: 'v-1', name: 'SAP', implementationPartner: 'Accenture' }),
        createMockVendor({ id: 'v-2', name: 'Microsoft', implementationPartner: 'Deloitte' }),
      ];
      render(<VendorsSection {...defaultProps} vendors={vendors} />);

      const searchInput = screen.getByPlaceholderText('Search vendors...');
      fireEvent.change(searchInput, { target: { value: 'accenture' } });

      expect(screen.getByText('SAP')).toBeInTheDocument();
      expect(screen.queryByText('Microsoft')).not.toBeInTheDocument();
    });

    it('should filter vendors by notes', () => {
      const vendors = [
        createMockVendor({ id: 'v-1', name: 'SAP', notes: 'Enterprise resource planning' }),
        createMockVendor({ id: 'v-2', name: 'Salesforce', notes: 'Customer relationship management' }),
      ];
      render(<VendorsSection {...defaultProps} vendors={vendors} />);

      const searchInput = screen.getByPlaceholderText('Search vendors...');
      fireEvent.change(searchInput, { target: { value: 'enterprise' } });

      expect(screen.getByText('SAP')).toBeInTheDocument();
      expect(screen.queryByText('Salesforce')).not.toBeInTheDocument();
    });

    it('should show no matches message when search yields no results', () => {
      const vendors = [createMockVendor({ name: 'SAP' })];
      render(<VendorsSection {...defaultProps} vendors={vendors} />);

      const searchInput = screen.getByPlaceholderText('Search vendors...');
      fireEvent.change(searchInput, { target: { value: 'nonexistent' } });

      expect(screen.getByText('No matches')).toBeInTheDocument();
    });

    it('should clear search when clear button is clicked', () => {
      const vendors = [createMockVendor()];
      render(<VendorsSection {...defaultProps} vendors={vendors} />);

      const searchInput = screen.getByPlaceholderText('Search vendors...');
      fireEvent.change(searchInput, { target: { value: 'test' } });
      fireEvent.click(screen.getByLabelText('Clear search'));

      expect(searchInput).toHaveValue('');
    });
  });

  describe('selection', () => {
    it('should call onVendorSelect when vendor is clicked', () => {
      const onVendorSelect = vi.fn();
      const vendor = createMockVendor({ id: 'v-123' as VendorId });
      render(
        <VendorsSection
          {...defaultProps}
          vendors={[vendor]}
          onVendorSelect={onVendorSelect}
        />
      );

      fireEvent.click(screen.getByTitle('SAP'));

      expect(onVendorSelect).toHaveBeenCalledWith('v-123');
    });

    it('should apply selected class when vendor is selected', () => {
      const vendor = createMockVendor({ id: 'v-123' as VendorId });
      render(
        <VendorsSection
          {...defaultProps}
          vendors={[vendor]}
          selectedVendorId="v-123"
        />
      );

      const vendorButton = screen.getByTitle('SAP');
      expect(vendorButton).toHaveClass('selected');
    });
  });

  describe('context menu', () => {
    it('should call onVendorContextMenu on right click', () => {
      const onVendorContextMenu = vi.fn();
      const vendor = createMockVendor();
      render(
        <VendorsSection
          {...defaultProps}
          vendors={[vendor]}
          onVendorContextMenu={onVendorContextMenu}
        />
      );

      fireEvent.contextMenu(screen.getByTitle('SAP'));

      expect(onVendorContextMenu).toHaveBeenCalledWith(expect.any(Object), vendor);
    });
  });

  describe('drag and drop', () => {
    it('should always be draggable', () => {
      const vendor = createMockVendor({ id: 'v-123' as VendorId });
      render(
        <VendorsSection
          {...defaultProps}
          vendors={[vendor]}
        />
      );

      const vendorButton = screen.getByTitle('SAP');
      expect(vendorButton).toHaveAttribute('draggable', 'true');
    });

    it('should set vendorId on drag start', () => {
      const vendor = createMockVendor({ id: 'v-123' as VendorId });
      render(
        <VendorsSection
          {...defaultProps}
          vendors={[vendor]}
        />
      );

      const vendorButton = screen.getByTitle('SAP');
      const mockDataTransfer = {
        setData: vi.fn(),
        effectAllowed: '',
      };

      fireEvent.dragStart(vendorButton, { dataTransfer: mockDataTransfer });

      expect(mockDataTransfer.setData).toHaveBeenCalledWith('vendorId', 'v-123');
      expect(mockDataTransfer.effectAllowed).toBe('copy');
    });
  });

  describe('add button', () => {
    it('should call onAddVendor when add button is clicked', () => {
      const onAddVendor = vi.fn();
      render(<VendorsSection {...defaultProps} onAddVendor={onAddVendor} />);

      fireEvent.click(screen.getByTestId('create-vendor-button'));

      expect(onAddVendor).toHaveBeenCalled();
    });

    it('should have correct title for add button', () => {
      render(<VendorsSection {...defaultProps} />);

      expect(screen.getByTitle('Create new vendor')).toBeInTheDocument();
    });
  });

  describe('toggle', () => {
    it('should call onToggle when header is clicked', () => {
      const onToggle = vi.fn();
      render(<VendorsSection {...defaultProps} onToggle={onToggle} />);

      fireEvent.click(screen.getByText('Vendors'));

      expect(onToggle).toHaveBeenCalled();
    });
  });
});
