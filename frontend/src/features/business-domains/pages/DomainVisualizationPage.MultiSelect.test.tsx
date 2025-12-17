import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DomainVisualizationPage } from './DomainVisualizationPage';
import type { BusinessDomain, Capability, CapabilityId } from '../../../api/types';

vi.mock('../hooks/useBusinessDomains');
vi.mock('../hooks/useDomainCapabilities');
vi.mock('../hooks/useCapabilityTree');
vi.mock('../hooks/useGridPositions');
vi.mock('../hooks/useDragHandlers');
vi.mock('../hooks/usePersistedDepth');
vi.mock('../../../store/appStore');

const createDomain = (id: string, name: string): BusinessDomain => ({
  id: id as any,
  name,
  description: '',
  createdAt: '2024-01-01',
  _links: {
    self: { href: `/api/v1/business-domains/${id}` },
    capabilities: `/api/v1/business-domains/${id}/capabilities`,
  },
});

const createCapability = (
  id: string,
  name: string,
  level: 'L1' | 'L2' | 'L3' | 'L4',
  parentId?: string
): Capability => ({
  id: id as CapabilityId,
  name,
  level,
  parentId: parentId as CapabilityId | undefined,
  createdAt: '2024-01-01',
  _links: {
    self: { href: `/api/v1/capabilities/${id}` },
    dissociate: `/api/v1/business-domains/domain-1/capabilities/${id}`,
  },
});

describe('DomainVisualizationPage - Multi-Select (Slice 3)', () => {
  const mockDomains: BusinessDomain[] = [
    createDomain('domain-1', 'Finance'),
  ];

  const mockCapabilities: Capability[] = [
    createCapability('l1-1', 'Financial Management', 'L1'),
    createCapability('l2-1', 'Accounting', 'L2', 'l1-1'),
    createCapability('l3-1', 'General Ledger', 'L3', 'l2-1'),
    createCapability('l1-2', 'Treasury', 'L1'),
    createCapability('l2-2', 'Cash Management', 'L2', 'l1-2'),
    createCapability('l1-3', 'Reporting', 'L1'),
  ];

  let mockRefetch: ReturnType<typeof vi.fn>;
  let mockDissociateCapability: ReturnType<typeof vi.fn>;
  let mockDeleteCapability: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    vi.clearAllMocks();

    mockRefetch = vi.fn().mockResolvedValue(undefined);
    mockDissociateCapability = vi.fn().mockResolvedValue(undefined);
    mockDeleteCapability = vi.fn().mockResolvedValue(undefined);

    const { useBusinessDomains } = await import('../hooks/useBusinessDomains');
    vi.mocked(useBusinessDomains).mockReturnValue({
      domains: mockDomains,
      isLoading: false,
      error: null,
    });

    const { useDomainCapabilities } = await import('../hooks/useDomainCapabilities');
    vi.mocked(useDomainCapabilities).mockReturnValue({
      capabilities: mockCapabilities,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
      associateCapability: vi.fn(),
      dissociateCapability: mockDissociateCapability,
    });

    const { useCapabilityTree } = await import('../hooks/useCapabilityTree');
    vi.mocked(useCapabilityTree).mockReturnValue({
      tree: [],
      isLoading: false,
    });

    const { useGridPositions } = await import('../hooks/useGridPositions');
    vi.mocked(useGridPositions).mockReturnValue({
      positions: {},
      updatePosition: vi.fn(),
    });

    const { useDragHandlers } = await import('../hooks/useDragHandlers');
    vi.mocked(useDragHandlers).mockReturnValue({
      isDragOver: false,
      handleDragOver: vi.fn(),
      handleDragLeave: vi.fn(),
      handleDrop: vi.fn(),
      handleDragStart: vi.fn(),
      handleDragEnd: vi.fn(),
    });

    const { usePersistedDepth } = await import('../hooks/usePersistedDepth');
    vi.mocked(usePersistedDepth).mockReturnValue([4, vi.fn()]);

    const { useAppStore } = await import('../../../store/appStore');
    vi.mocked(useAppStore).mockImplementation((selector: any) => {
      if (selector) {
        return selector({ deleteCapability: mockDeleteCapability });
      }
      return { deleteCapability: mockDeleteCapability };
    });
  });

  describe('selection state management', () => {
    it('should toggle selection when shift-clicking a capability', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');

      await user.keyboard('{Shift>}');
      await user.click(capability);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(capability).toHaveClass('selected');
      });
    });

    it('should deselect capability when shift-clicking already selected capability', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');

      await user.keyboard('{Shift>}');
      await user.click(capability);
      await user.keyboard('{/Shift}');
      await waitFor(() => {
        expect(capability).toHaveClass('selected');
      });

      await user.keyboard('{Shift>}');
      await user.click(capability);
      await user.keyboard('{/Shift}');
      await waitFor(() => {
        expect(capability).not.toHaveClass('selected');
      });
    });

    it('should allow selecting multiple capabilities with shift-click', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');
      const cap3 = screen.getByTestId('capability-l1-3');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap3);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(cap1).toHaveClass('selected');
        expect(cap2).toHaveClass('selected');
        expect(cap3).toHaveClass('selected');
      });
    });

    it('should clear selection when clicking capability without shift key', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(cap1).toHaveClass('selected');
        expect(cap2).toHaveClass('selected');
      });

      await user.click(cap1);

      await waitFor(() => {
        expect(cap1).not.toHaveClass('selected');
        expect(cap2).not.toHaveClass('selected');
      });
    });

    it('should open detail panel when clicking without shift key', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.click(capability);

      const detailPanel = await screen.findByText('Capability Details');
      expect(detailPanel).toBeInTheDocument();
    });

    it('should not open detail panel when shift-clicking', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.keyboard('{Shift>}');
      await user.click(capability);
      await user.keyboard('{/Shift}');

      expect(screen.queryByText('Capability Details')).not.toBeInTheDocument();
    });
  });

  describe('visual indicators for selected capabilities', () => {
    it('should apply selected styling to selected capability', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.keyboard('{Shift>}');
      await user.click(capability);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(capability).toHaveClass('selected');
      });
    });

    it('should show distinct border for selected capabilities', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.keyboard('{Shift>}');
      await user.click(capability);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        const styles = window.getComputedStyle(capability);
        expect(styles.border).toContain('3px');
      });
    });

    it('should visually distinguish multiple selected capabilities', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(cap1).toHaveClass('selected');
        expect(cap2).toHaveClass('selected');
      });
    });
  });

  describe('context menu on selected capabilities', () => {
    it('should show context menu when right-clicking selected capability', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.keyboard('{Shift>}');
      await user.click(capability);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(capability).toHaveClass('selected');
      });

      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();
    });

    it('should show both remove and delete options in context menu', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });

      expect(removeOption).toBeInTheDocument();
      expect(deleteOption).toBeInTheDocument();
    });
  });

  describe('remove from business domain on multi-select', () => {
    it('should dissociate all selected L1 capabilities when clicking remove', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(2);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-1' })
      );
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-2' })
      );
    });

    it('should resolve L2 capability to L1 ancestor when selected', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Accounting')).toBeInTheDocument();
      });

      const l2Cap = screen.getByTestId('capability-l2-1');
      const l1Cap = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(l2Cap);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(l1Cap);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: l2Cap });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(2);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-1', level: 'L1' })
      );
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-2', level: 'L1' })
      );
    });

    it('should resolve L3 capability to L1 ancestor when selected', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('General Ledger')).toBeInTheDocument();
      });

      const l3Cap = screen.getByTestId('capability-l3-1');
      const l1Cap = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(l3Cap);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(l1Cap);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: l3Cap });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(2);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-1', level: 'L1' })
      );
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-2', level: 'L1' })
      );
    });

    it('should handle duplicate L1 ancestors correctly', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const l1Cap = screen.getByTestId('capability-l1-1');
      const l2Cap = screen.getByTestId('capability-l2-1');
      const l3Cap = screen.getByTestId('capability-l3-1');

      await user.keyboard('{Shift>}');
      await user.click(l1Cap);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(l2Cap);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(l3Cap);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: l1Cap });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(1);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-1', level: 'L1' })
      );
    });

    it('should call refetch after dissociating all selected capabilities', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      await waitFor(() => {
        expect(mockRefetch).toHaveBeenCalledTimes(1);
      });
    });

    it('should close context menu after remove action', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();

      const removeOption = screen.getByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });
  });

  describe('delete from model on multi-select', () => {
    it('should open confirmation dialog when clicking delete on multiple selected', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmDialog = await screen.findByRole('dialog');
      expect(confirmDialog).toBeInTheDocument();
    });

    it('should show count of capabilities being deleted in confirmation dialog', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');
      const cap3 = screen.getByTestId('capability-l1-3');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap3);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmDialog = await screen.findByRole('dialog');
      expect(confirmDialog).toHaveTextContent(/3 capabilities/i);
    });

    it('should delete all selected L1 capabilities when confirming', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      expect(mockDeleteCapability).toHaveBeenCalledTimes(2);
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-1');
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-2');
    });

    it('should resolve child capabilities to L1 ancestors before deleting', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Accounting')).toBeInTheDocument();
      });

      const l2Cap = screen.getByTestId('capability-l2-1');
      const l3Cap = screen.getByTestId('capability-l3-1');

      await user.keyboard('{Shift>}');
      await user.click(l2Cap);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(l3Cap);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: l2Cap });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      expect(mockDeleteCapability).toHaveBeenCalledTimes(1);
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-1');
    });

    it('should handle duplicate L1 ancestors when deleting', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const l1Cap = screen.getByTestId('capability-l1-1');
      const l2Cap = screen.getByTestId('capability-l2-1');
      const l3Cap = screen.getByTestId('capability-l3-1');
      const l1Cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(l1Cap);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(l2Cap);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(l3Cap);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(l1Cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: l1Cap });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmDialog = await screen.findByRole('dialog');
      expect(confirmDialog).toHaveTextContent(/2 capabilities/i);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      expect(mockDeleteCapability).toHaveBeenCalledTimes(2);
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-1');
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-2');
    });

    it('should not delete when canceling confirmation dialog', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const cancelButton = await screen.findByRole('button', { name: /cancel/i });
      await user.click(cancelButton);

      expect(mockDeleteCapability).not.toHaveBeenCalled();
    });

    it('should call refetch after successfully deleting all selected capabilities', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockRefetch).toHaveBeenCalledTimes(1);
      });
    });

    it('should close context menu after opening delete confirmation', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();

      const deleteOption = screen.getByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });
  });

  describe('clearing selection', () => {
    it('should clear all selections after successful remove operation', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(cap1).toHaveClass('selected');
        expect(cap2).toHaveClass('selected');
      });

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      await waitFor(() => {
        expect(cap1).not.toHaveClass('selected');
        expect(cap2).not.toHaveClass('selected');
      });
    });

    it('should clear all selections after successful delete operation', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      const cap2 = screen.getByTestId('capability-l1-2');

      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');
      await user.keyboard('{Shift>}');
      await user.click(cap2);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(cap1).toHaveClass('selected');
        expect(cap2).toHaveClass('selected');
      });

      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      await waitFor(() => {
        expect(cap1).not.toHaveClass('selected');
        expect(cap2).not.toHaveClass('selected');
      });
    });
  });

  describe('edge cases', () => {
    it('should handle empty selection when opening context menu on non-selected capability', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();

      const removeOption = screen.getByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(1);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-1' })
      );
    });

    it('should select single capability when selecting only one', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.keyboard('{Shift>}');
      await user.click(capability);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(capability).toHaveClass('selected');
      });

      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmDialog = await screen.findByRole('dialog');
      expect(confirmDialog).toHaveTextContent(/1 capability|Financial Management/i);
    });
  });
});
