import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
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

describe('DomainVisualizationPage - Keyboard Shortcuts (Slice 4)', () => {
  const mockDomains: BusinessDomain[] = [
    createDomain('domain-1', 'Finance'),
  ];

  const mockCapabilities: Capability[] = [
    createCapability('l1-1', 'Financial Management', 'L1'),
    createCapability('l2-1', 'Accounting', 'L2', 'l1-1'),
    createCapability('l3-1', 'General Ledger', 'L3', 'l2-1'),
    createCapability('l4-1', 'Journal Entries', 'L4', 'l3-1'),
    createCapability('l1-2', 'Treasury', 'L1'),
    createCapability('l2-2', 'Cash Management', 'L2', 'l1-2'),
    createCapability('l1-3', 'Reporting', 'L1'),
    createCapability('l2-3', 'Financial Reporting', 'L2', 'l1-3'),
    createCapability('l1-4', 'Compliance', 'L1'),
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

  describe('Ctrl+A selects all L1 capabilities', () => {
    it('should select all L1 capabilities when pressing Ctrl+A', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-3')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-4')).toHaveClass('selected');
      });
    });

    it('should prevent default browser behavior when pressing Ctrl+A', async () => {
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      const event = new KeyboardEvent('keydown', { key: 'a', ctrlKey: true, bubbles: true });
      const preventDefaultSpy = vi.spyOn(event, 'preventDefault');

      act(() => {
        grid.dispatchEvent(event);
      });

      expect(preventDefaultSpy).toHaveBeenCalled();
    });

    it('should select all L1 capabilities with Cmd+A on Mac', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Meta>}a{/Meta}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-3')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-4')).toHaveClass('selected');
      });
    });

    it('should only select L1 capabilities, not L2-L4', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).toHaveClass('selected');
      });

      expect(screen.getByTestId('capability-l2-1')).not.toHaveClass('selected');
      expect(screen.getByTestId('capability-l2-2')).not.toHaveClass('selected');
      expect(screen.getByTestId('capability-l3-1')).not.toHaveClass('selected');
      expect(screen.getByTestId('capability-l4-1')).not.toHaveClass('selected');
    });

    it('should show visual feedback for all selected L1 capabilities', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        const l1Cap1 = screen.getByTestId('capability-l1-1');
        const l1Cap2 = screen.getByTestId('capability-l1-2');
        const l1Cap3 = screen.getByTestId('capability-l1-3');
        const l1Cap4 = screen.getByTestId('capability-l1-4');

        expect(l1Cap1).toHaveClass('selected');
        expect(l1Cap2).toHaveClass('selected');
        expect(l1Cap3).toHaveClass('selected');
        expect(l1Cap4).toHaveClass('selected');

        const styles1 = window.getComputedStyle(l1Cap1);
        const styles2 = window.getComputedStyle(l1Cap2);
        const styles3 = window.getComputedStyle(l1Cap3);
        const styles4 = window.getComputedStyle(l1Cap4);

        expect(styles1.border).toContain('3px');
        expect(styles2.border).toContain('3px');
        expect(styles3.border).toContain('3px');
        expect(styles4.border).toContain('3px');
      });
    });
  });

  describe('Ctrl+A with existing selections', () => {
    it('should replace partial selection with all L1 capabilities when pressing Ctrl+A', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(cap1).toHaveClass('selected');
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-3')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-4')).toHaveClass('selected');
      });
    });

    it('should select all L1 capabilities even when L2 capabilities were manually selected', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Accounting')).toBeInTheDocument();
      });

      const l2Cap = screen.getByTestId('capability-l2-1');
      await user.keyboard('{Shift>}');
      await user.click(l2Cap);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(l2Cap).toHaveClass('selected');
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-3')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-4')).toHaveClass('selected');
      });
    });
  });

  describe('Escape key clears selection', () => {
    it('should clear all selections when pressing Escape', async () => {
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

      const grid = screen.getByRole('main');
      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(cap1).not.toHaveClass('selected');
        expect(cap2).not.toHaveClass('selected');
      });
    });

    it('should clear all selections after Ctrl+A when pressing Escape', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-3')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-4')).toHaveClass('selected');
      });

      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).not.toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).not.toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-3')).not.toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-4')).not.toHaveClass('selected');
      });
    });

    it('should clear single capability selection when pressing Escape', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(cap1).toHaveClass('selected');
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(cap1).not.toHaveClass('selected');
      });
    });

    it('should do nothing when pressing Escape with no selection', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      expect(cap1).not.toHaveClass('selected');

      const grid = screen.getByRole('main');
      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(cap1).not.toHaveClass('selected');
      });
    });
  });

  describe('context menu operations with Ctrl+A selection', () => {
    it('should apply remove operation to all L1 capabilities selected by Ctrl+A', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const removeOption = await screen.findByRole('menuitem', {
        name: /remove from business domain/i,
      });
      await user.click(removeOption);

      expect(mockDissociateCapability).toHaveBeenCalledTimes(4);
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-1' })
      );
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-2' })
      );
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-3' })
      );
      expect(mockDissociateCapability).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l1-4' })
      );
    });

    it('should apply delete operation to all L1 capabilities selected by Ctrl+A', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmDialog = await screen.findByRole('dialog');
      expect(confirmDialog).toHaveTextContent(/4 capabilities/i);

      const confirmButton = await screen.findByRole('button', { name: /confirm|delete/i });
      await user.click(confirmButton);

      expect(mockDeleteCapability).toHaveBeenCalledTimes(4);
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-1');
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-2');
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-3');
      expect(mockDeleteCapability).toHaveBeenCalledWith('l1-4');
    });

    it('should show correct count in delete confirmation dialog after Ctrl+A', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const deleteOption = await screen.findByRole('menuitem', {
        name: /delete from model/i,
      });
      await user.click(deleteOption);

      const confirmDialog = await screen.findByRole('dialog');
      expect(confirmDialog).toHaveTextContent(/4 capabilities/i);
    });
  });

  describe('keyboard shortcuts interaction with grid focus', () => {
    it('should work when focus is on the main grid container', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      grid.focus();

      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).toHaveClass('selected');
      });
    });

    it('should work when a capability is focused', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      cap1.focus();

      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).toHaveClass('selected');
      });
    });
  });

  describe('keyboard shortcuts do not interfere with other functionality', () => {
    it('should not trigger selection when pressing A without Ctrl', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('a');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).not.toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).not.toHaveClass('selected');
      });
    });

    it('should not trigger selection when pressing Ctrl with other keys', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{b}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).not.toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).not.toHaveClass('selected');
      });
    });

    it('should still allow shift-click selection after using Ctrl+A and Escape', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
      });

      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).not.toHaveClass('selected');
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      await user.keyboard('{Shift>}');
      await user.click(cap1);
      await user.keyboard('{/Shift}');

      await waitFor(() => {
        expect(cap1).toHaveClass('selected');
      });
    });

    it('should still allow normal click to open detail panel after Ctrl+A and Escape', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).toHaveClass('selected');
      });

      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).not.toHaveClass('selected');
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      await user.click(cap1);

      const detailPanel = await screen.findByText('Capability Details');
      expect(detailPanel).toBeInTheDocument();
    });

    it('should not interfere with Escape closing context menu', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const cap1 = screen.getByTestId('capability-l1-1');
      await user.pointer({ keys: '[MouseRight>]', target: cap1 });

      const contextMenu = await screen.findByRole('menu');
      expect(contextMenu).toBeInTheDocument();

      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument();
      });
    });
  });

  describe('edge cases', () => {
    it('should handle Ctrl+A when no domain is selected', async () => {
      const user = userEvent.setup();
      render(<DomainVisualizationPage />);

      await waitFor(() => {
        expect(screen.getByText('Select a domain from the left sidebar')).toBeInTheDocument();
      });

      const container = screen.getByText('Select a domain from the left sidebar').closest('div');
      await user.keyboard('{a}');

      expect(screen.queryByTestId('capability-l1-1')).not.toBeInTheDocument();
    });

    it('should handle Ctrl+A when domain has no capabilities', async () => {

      const user = userEvent.setup();
      const { useDomainCapabilities } = await import('../hooks/useDomainCapabilities');
      vi.mocked(useDomainCapabilities).mockReturnValue({
        capabilities: [],
        isLoading: false,
        error: null,
        refetch: mockRefetch,
        associateCapability: vi.fn(),
        dissociateCapability: mockDissociateCapability,
      });

      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Finance' })).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      expect(screen.queryByTestId('capability-l1-1')).not.toBeInTheDocument();
    });

    it('should handle Ctrl+A when domain has only L2-L4 capabilities', async () => {

      const user = userEvent.setup();
      const l2OnlyCapabilities = [
        createCapability('l2-1', 'Accounting', 'L2', 'l1-1'),
        createCapability('l3-1', 'General Ledger', 'L3', 'l2-1'),
      ];

      const { useDomainCapabilities } = await import('../hooks/useDomainCapabilities');
      vi.mocked(useDomainCapabilities).mockReturnValue({
        capabilities: l2OnlyCapabilities,
        isLoading: false,
        error: null,
        refetch: mockRefetch,
        associateCapability: vi.fn(),
        dissociateCapability: mockDissociateCapability,
      });

      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Finance' })).toBeInTheDocument();
      });

      const l2Cap = screen.queryByTestId('capability-l2-1');
      const l3Cap = screen.queryByTestId('capability-l3-1');

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');

      await waitFor(() => {
        if (l2Cap) {
          expect(l2Cap).not.toHaveClass('selected');
        }
        if (l3Cap) {
          expect(l3Cap).not.toHaveClass('selected');
        }
      });
    });

    it('should handle rapid Ctrl+A and Escape key presses', async () => {

      const user = userEvent.setup();
      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Financial Management')).toBeInTheDocument();
      });

      const grid = screen.getByRole('main');
      await user.keyboard('{Control>}a{/Control}');
      await user.keyboard('{Escape}');
      await user.keyboard('{a}');
      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(screen.getByTestId('capability-l1-1')).not.toHaveClass('selected');
        expect(screen.getByTestId('capability-l1-2')).not.toHaveClass('selected');
      });
    });

    it('should handle Ctrl+A when capabilities are being loaded', async () => {

      const user = userEvent.setup();
      const { useDomainCapabilities } = await import('../hooks/useDomainCapabilities');
      vi.mocked(useDomainCapabilities).mockReturnValue({
        capabilities: [],
        isLoading: true,
        error: null,
        refetch: mockRefetch,
        associateCapability: vi.fn(),
        dissociateCapability: mockDissociateCapability,
      });

      render(<DomainVisualizationPage initialDomainId={'domain-1' as any} />);

      await waitFor(() => {
        expect(screen.getByText('Loading capabilities...')).toBeInTheDocument();
      });

      const loadingDiv = screen.getByText('Loading capabilities...').closest('div');
      await user.keyboard('{a}');

      expect(screen.queryByTestId('capability-l1-1')).not.toBeInTheDocument();
    });
  });
});
