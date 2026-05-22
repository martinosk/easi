import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Capability, CapabilityId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers/renderWithProviders';
import { NestedCapabilityGrid } from './NestedCapabilityGrid';

function render(ui: React.ReactElement) {
  return renderWithProviders(ui, { withRouter: false });
}

describe('NestedCapabilityGrid', () => {
  const createCapability = (
    id: string,
    name: string,
    level: 'L1' | 'L2' | 'L3' | 'L4',
    parentId?: string,
  ): Capability => ({
    id: id as CapabilityId,
    name,
    level,
    parentId: parentId as CapabilityId | undefined,
    createdAt: '2024-01-01',
    _links: { self: { href: `/api/v1/capabilities/${id}`, method: 'GET' } },
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
      render(<NestedCapabilityGrid capabilities={mockCapabilities} depth={1} onCapabilityClick={onCapabilityClick} />);

      expect(screen.getByText('Finance')).toBeInTheDocument();
      expect(screen.getByText('HR')).toBeInTheDocument();
      expect(screen.queryByText('Accounting')).not.toBeInTheDocument();
      expect(screen.queryByText('Treasury')).not.toBeInTheDocument();
    });
  });

  describe('depth level 2', () => {
    it('should show L1 and L2 capabilities', () => {
      const onCapabilityClick = vi.fn();
      render(<NestedCapabilityGrid capabilities={mockCapabilities} depth={2} onCapabilityClick={onCapabilityClick} />);

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
      render(<NestedCapabilityGrid capabilities={mockCapabilities} depth={3} onCapabilityClick={onCapabilityClick} />);

      expect(screen.getByText('Finance')).toBeInTheDocument();
      expect(screen.getByText('Accounting')).toBeInTheDocument();
      expect(screen.getByText('General Ledger')).toBeInTheDocument();
      expect(screen.queryByText('Journal Entries')).not.toBeInTheDocument();
    });
  });

  describe('depth level 4', () => {
    it('should show all capabilities', () => {
      const onCapabilityClick = vi.fn();
      render(<NestedCapabilityGrid capabilities={mockCapabilities} depth={4} onCapabilityClick={onCapabilityClick} />);

      expect(screen.getByText('Finance')).toBeInTheDocument();
      expect(screen.getByText('Accounting')).toBeInTheDocument();
      expect(screen.getByText('General Ledger')).toBeInTheDocument();
      expect(screen.getByText('Journal Entries')).toBeInTheDocument();
    });
  });

  describe('level encoding', () => {
    it('tags each tile with its capability level', () => {
      const onCapabilityClick = vi.fn();
      render(<NestedCapabilityGrid capabilities={mockCapabilities} depth={4} onCapabilityClick={onCapabilityClick} />);

      expect(screen.getByTestId('capability-l1-1')).toHaveAttribute('data-level', 'L1');
      expect(screen.getByTestId('capability-l2-1')).toHaveAttribute('data-level', 'L2');
      expect(screen.getByTestId('capability-l3-1')).toHaveAttribute('data-level', 'L3');
      expect(screen.getByTestId('capability-l4-1')).toHaveAttribute('data-level', 'L4');
    });
  });

  describe('click handler', () => {
    it('should call onCapabilityClick when capability is clicked', () => {
      const onCapabilityClick = vi.fn();
      render(<NestedCapabilityGrid capabilities={mockCapabilities} depth={2} onCapabilityClick={onCapabilityClick} />);

      screen.getByText('Accounting').click();

      expect(onCapabilityClick).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'l2-1', name: 'Accounting' }),
        expect.anything(),
      );
    });
  });

  describe('nesting structure', () => {
    it('should nest L2 inside L1 container', () => {
      const onCapabilityClick = vi.fn();
      const { container } = render(
        <NestedCapabilityGrid capabilities={mockCapabilities} depth={2} onCapabilityClick={onCapabilityClick} />,
      );

      const l1Container = container.querySelector('[data-testid="capability-l1-1"]');
      const l2Element = l1Container?.querySelector('[data-testid="capability-l2-1"]');

      expect(l2Element).toBeInTheDocument();
    });
  });

  describe('drop zone', () => {
    it('should mark grid when isDragOver is true', () => {
      const onCapabilityClick = vi.fn();
      render(
        <NestedCapabilityGrid
          capabilities={mockCapabilities}
          depth={2}
          onCapabilityClick={onCapabilityClick}
          isDragOver={true}
        />,
      );

      expect(screen.getByTestId('nested-capability-grid')).toHaveAttribute('data-drag-over', 'true');
    });
  });
});
