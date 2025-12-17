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

describe('DomainVisualizationPage - Context Menu (Slice 1)', () => {
  const mockDomains: BusinessDomain[] = [
    createDomain('domain-1', 'Finance'),
    createDomain('domain-2', 'HR'),
  ];

  const mockCapabilities: Capability[] = [
    createCapability('l1-1', 'Financial Management', 'L1'),
    createCapability('l2-1', 'Accounting', 'L2', 'l1-1'),
    createCapability('l3-1', 'General Ledger', 'L3', 'l2-1'),
    createCapability('l4-1', 'Journal Entries', 'L4', 'l3-1'),
    createCapability('l1-2', 'Treasury', 'L1'),
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

  describe('opening context menu', () => {
    it('should open context menu at correct position when right-clicking L1 capability', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability, coords: { clientX: 200, clientY: 300 } });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();
      expect(contextMenu).toHaveStyle({ left: '200px', top: '300px' });
    });

    it('should open context menu when right-clicking L2 capability', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Accounting')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l2-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability, coords: { clientX: 150, clientY: 250 } });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();
    });

    it('should open context menu when right-clicking L3 capability', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('General Ledger')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l3-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();
    });

    it('should open context menu when right-clicking L4 capability', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Journal Entries')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l4-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();
    });
  });

  describe('context menu options', () => {
    it('should show "Remove from Business Domain" option', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      expect(removeOption).toBeInTheDocument();
    });

    it('should show "Delete from Model" option', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      expect(deleteOption).toBeInTheDocument();
    });

    it('should show both options in correct order', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const menuItems = await screen.findAllByRole('menuitem');
      expect(menuItems).toHaveLength(2);
      expect(menuItems[0]).toHaveTextContent(/remove from business domain/i);
      expect(menuItems[1]).toHaveTextContent(/delete from model/i);
    });
  });

  describe('remove from business domain action', () => {
    it('should call dissociateCapability with L1 capability when clicking remove on L1', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(1);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'l1-1',
          name: 'Financial Management',
          level: 'L1',
        })
      );
    });

    it('should call dissociateCapability with L1 ancestor when clicking remove on L2', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Accounting')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l2-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(1);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'l1-1',
          name: 'Financial Management',
          level: 'L1',
        })
      );
    });

    it('should call dissociateCapability with L1 ancestor when clicking remove on L3', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('General Ledger')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l3-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(1);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'l1-1',
          level: 'L1',
        })
      );
    });

    it('should call dissociateCapability with L1 ancestor when clicking remove on L4', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Journal Entries')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l4-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(1);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'l1-1',
          level: 'L1',
        })
      );
    });

    it('should call refetch after successful dissociate', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      await waitFor(() => {
        expect(mockRefetch).toHaveBeenCalledTimes(1);
      });
    });

    it('should close context menu after remove action completes', async () => {

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

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });
  });

  describe('delete from model action', () => {
    it('should open confirmation dialog when clicking delete', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmDialog = await screen.findByRole('dialog');
      expect(confirmDialog).toBeInTheDocument();
      expect(confirmDialog).toHaveTextContent(/delete capability/i);
    });

    it('should show capability name in confirmation dialog', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmDialog = await screen.findByRole('dialog');
      expect(confirmDialog).toHaveTextContent('Financial Management');
    });

    it('should call deleteCapability with L1 ancestor when confirming delete on L1', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      expect(mockDeleteCapability).toHaveBeenCalledTimes(1);
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-1');
    });

    it('should call deleteCapability with L1 ancestor when confirming delete on L2', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Accounting')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l2-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      expect(mockDeleteCapability).toHaveBeenCalledTimes(1);
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-1');
    });

    it('should call deleteCapability with L1 ancestor when confirming delete on L3', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('General Ledger')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l3-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      expect(mockDeleteCapability).toHaveBeenCalledTimes(1);
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-1');
    });

    it('should call deleteCapability with L1 ancestor when confirming delete on L4', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Journal Entries')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l4-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      expect(mockDeleteCapability).toHaveBeenCalledTimes(1);
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-1');
    });

    it('should not call deleteCapability when canceling confirmation', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const cancelButton = await screen.findByRole('button', { name: /cancel/i });
      await user.click(cancelButton);

      expect(mockDeleteCapability).not.toHaveBeenCalled();
    });

    it('should call refetch after successful delete', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

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

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

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

  describe('closing context menu', () => {
    it('should close context menu when clicking outside', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();

      await user.pointer({ keys: '[MouseLeft>]', target: document.body });

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });

    it('should close context menu when pressing Escape key', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();

      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });

    it('should not close context menu when clicking inside menu', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();

      await user.pointer({ keys: '[MouseLeft>]', target: contextMenu });

      expect(screen.getByRole('menu')).toBeInTheDocument();
    });
  });

  describe('L1 ancestor resolution', () => {
    it('should resolve L1 ancestor for L2 capability', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Accounting')).toBeInTheDocument();
      });

      const l2Capability = screen.getByTestId('capability-l2-1');
      await user.pointer({ keys: '[MouseRight>]', target: l2Capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'l1-1',
          level: 'L1',
        })
      );
    });

    it('should resolve L1 ancestor for L3 capability traversing through L2', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('General Ledger')).toBeInTheDocument();
      });

      const l3Capability = screen.getByTestId('capability-l3-1');
      await user.pointer({ keys: '[MouseRight>]', target: l3Capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'l1-1',
          level: 'L1',
        })
      );
    });

    it('should resolve L1 ancestor for L4 capability traversing through L3 and L2', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Journal Entries')).toBeInTheDocument();
      });

      const l4Capability = screen.getByTestId('capability-l4-1');
      await user.pointer({ keys: '[MouseRight>]', target: l4Capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'l1-1',
          level: 'L1',
        })
      );
    });

    it('should return same capability if already L1', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const l1Capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: l1Capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'l1-1',
          level: 'L1',
        })
      );
    });
  });

  describe('context menu state management', () => {
    it('should track context menu position', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability, coords: { clientX: 123, clientY: 456 } });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toHaveStyle({ left: '123px', top: '456px' });
    });

    it('should track target capability when opening context menu', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'l1-1',
          name: 'Financial Management',
        })
      );
    });

    it('should clear context menu state when closed', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: capability });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();

      await user.pointer({ keys: '[MouseLeft>]', target: document.body });

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });

      await user.pointer({ keys: '[MouseRight>]', target: capability, coords: { clientX: 999, clientY: 888 } });

      const newContextMenu = await screen.findByRole('menu');
      expect(newContextMenu).toBeInTheDocument();
      expect(newContextMenu).toHaveStyle({ left: '999px' });
    });

    it('should open new context menu if right-clicking different capability', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const firstCapability = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: firstCapability, coords: { clientX: 100, clientY: 100 } });

      let contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toHaveStyle({ left: '100px', top: '100px' });

      const secondCapability = screen.getByTestId('capability-l1-2');
      await user.pointer({ keys: '[MouseRight>]', target: secondCapability, coords: { clientX: 200, clientY: 200 } });

      contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toHaveStyle({ left: '200px', top: '200px' });
    });
  });

  describe('context menu does not interfere with normal capability click', () => {
    it('should still open detail panel when left-clicking capability', async () => {
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

    it('should not open context menu on left click', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const capability = screen.getByTestId('capability-l1-1');
      await user.click(capability);

      expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    });
  });
});
