import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { DndContext } from '@dnd-kit/core';
import { NestedCapabilityGrid } from './NestedCapabilityGrid';
import type { Capability } from '../../../api/types';

const renderWithDnd = (component: React.ReactNode) => {
  return render(<DndContext>{component}</DndContext>);
};

describe('NestedCapabilityGrid', () => {
  const createCapability = (
    id: string,
    name: string,
    level: 'L1' | 'L2' | 'L3' | 'L4',
    parentId?: string
  ): Capability => ({
    id: id as any,
    name,
    level,
    parentId: parentId as any,
    createdAt: '2024-01-01',
    _links: { self: { href: `/api/v1/capabilities/${id}` } },
  });

  const mockCapabilities: Capability[] = [
    createCapability('l1-1', 'Finance', 'L1'),
    createCapability('l2-1', 'Accounting', 'L2', 'l1-1'),
    createCapability('l2-2', 'Treasury', 'L2', 'l1-1'),
    createCapability('l3-1', 'General Ledger', 'L3', 'l2-1'),
    createCapability('l4-1', 'Journal Entries', 'L4', 'l3-1'),
    createCapability('l1-2', 'HR', 'L1'),
    createCapability('l2-3', 'Recruitment', 'L2', 'l1-2'),
  ];

  describe('depth level 1', () => {
    it('should only show L1 capabilities', () => {
      const onCapabilityClick = vi.fn();
      renderWithDnd(
        <NestedCapabilityGrid
          capabilities={mockCapabilities}
          depth={1}
          onCapabilityClick={onCapabilityClick}
        />
      );

      expect(screen.getByText('Finance')).toBeInTheDocument();
      expect(screen.getByText('HR')).toBeInTheDocument();
      expect(screen.queryByText('Accounting')).not.toBeInTheDocument();
      expect(screen.queryByText('Treasury')).not.toBeInTheDocument();
    });
  });

  describe('depth level 2', () => {
    it('should show L1 and L2 capabilities', () => {
      const onCapabilityClick = vi.fn();
      renderWithDnd(
        <NestedCapabilityGrid
          capabilities={mockCapabilities}
          depth={2}
          onCapabilityClick={onCapabilityClick}
        />
      );

      expect(screen.getByText('Finance')).toBeInTheDocument();
      expect(screen.getByText('Accounting')).toBeInTheDocument();
      expect(screen.getByText('Treasury')).toBeInTheDocument();
      expect(screen.getByText('Recruitment')).toBeInTheDocument();
      expect(screen.queryByText('General Ledger')).not.toBeInTheDocument();
    });
  });

  describe('depth level 3', () => {
    it('should show L1, L2 and L3 capabilities', () => {
      const onCapabilityClick = vi.fn();
      renderWithDnd(
        <NestedCapabilityGrid
          capabilities={mockCapabilities}
          depth={3}
          onCapabilityClick={onCapabilityClick}
        />
      );

      expect(screen.getByText('Finance')).toBeInTheDocument();
      expect(screen.getByText('Accounting')).toBeInTheDocument();
      expect(screen.getByText('General Ledger')).toBeInTheDocument();
      expect(screen.queryByText('Journal Entries')).not.toBeInTheDocument();
    });
  });

  describe('depth level 4', () => {
    it('should show all capabilities', () => {
      const onCapabilityClick = vi.fn();
      renderWithDnd(
        <NestedCapabilityGrid
          capabilities={mockCapabilities}
          depth={4}
          onCapabilityClick={onCapabilityClick}
        />
      );

      expect(screen.getByText('Finance')).toBeInTheDocument();
      expect(screen.getByText('Accounting')).toBeInTheDocument();
      expect(screen.getByText('General Ledger')).toBeInTheDocument();
      expect(screen.getByText('Journal Entries')).toBeInTheDocument();
    });
  });

  describe('color scheme', () => {
    it('should apply correct colors for each level', () => {
      const onCapabilityClick = vi.fn();
      renderWithDnd(
        <NestedCapabilityGrid
          capabilities={mockCapabilities}
          depth={4}
          onCapabilityClick={onCapabilityClick}
        />
      );

      const l1Element = screen.getByTestId('capability-l1-1');
      const l2Element = screen.getByTestId('capability-l2-1');
      const l3Element = screen.getByTestId('capability-l3-1');
      const l4Element = screen.getByTestId('capability-l4-1');

      expect(l1Element).toHaveStyle({ backgroundColor: 'rgb(59, 130, 246)' });
      expect(l2Element).toHaveStyle({ backgroundColor: 'rgb(139, 92, 246)' });
      expect(l3Element).toHaveStyle({ backgroundColor: 'rgb(236, 72, 153)' });
      expect(l4Element).toHaveStyle({ backgroundColor: 'rgb(249, 115, 22)' });
    });
  });

  describe('click handler', () => {
    it('should call onCapabilityClick when capability is clicked', () => {
      const onCapabilityClick = vi.fn();
      renderWithDnd(
        <NestedCapabilityGrid
          capabilities={mockCapabilities}
          depth={2}
          onCapabilityClick={onCapabilityClick}
        />
      );

      screen.getByText('Accounting').click();

      expect(onCapabilityClick).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l2-1', name: 'Accounting' })
      );
    });
  });

  describe('nesting structure', () => {
    it('should nest L2 inside L1 container', () => {
      const onCapabilityClick = vi.fn();
      const { container } = renderWithDnd(
        <NestedCapabilityGrid
          capabilities={mockCapabilities}
          depth={2}
          onCapabilityClick={onCapabilityClick}
        />
      );

      const l1Container = container.querySelector('[data-testid="capability-l1-1"]');
      const l2Element = l1Container?.querySelector('[data-testid="capability-l2-1"]');

      expect(l2Element).toBeInTheDocument();
    });
  });
});
