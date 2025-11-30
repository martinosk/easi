import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { DndContext } from '@dnd-kit/core';
import { DomainGrid } from './DomainGrid';
import type { Capability, Position } from '../../../api/types';

const renderWithDnd = (component: React.ReactNode) => {
  return render(<DndContext>{component}</DndContext>);
};

describe('DomainGrid', () => {
  const createCapability = (
    id: string,
    name: string,
    level: 'L1' | 'L2' | 'L3' | 'L4' = 'L1'
  ): Capability => ({
    id: id as any,
    name,
    level,
    createdAt: '2024-01-01',
    _links: { self: { href: `/api/v1/capabilities/${id}` } },
  });

  const createPosition = (x: number, y: number): Position => ({ x, y });

  const mockL1Capabilities: Capability[] = [
    createCapability('cap-1', 'Zebra Management', 'L1'),
    createCapability('cap-2', 'Alpha Processing', 'L1'),
    createCapability('cap-3', 'Delta Analytics', 'L1'),
  ];

  it('should render all L1 capabilities', () => {
    const onCapabilityClick = vi.fn();
    renderWithDnd(
      <DomainGrid capabilities={mockL1Capabilities} onCapabilityClick={onCapabilityClick} />
    );

    expect(screen.getByText('Zebra Management')).toBeInTheDocument();
    expect(screen.getByText('Alpha Processing')).toBeInTheDocument();
    expect(screen.getByText('Delta Analytics')).toBeInTheDocument();
  });

  it('should render capabilities sorted alphabetically by name', () => {
    const onCapabilityClick = vi.fn();
    renderWithDnd(
      <DomainGrid capabilities={mockL1Capabilities} onCapabilityClick={onCapabilityClick} />
    );

    const capabilityNames = screen
      .getAllByRole('button')
      .map((btn) => btn.textContent);

    expect(capabilityNames).toEqual([
      'Alpha Processing',
      'Delta Analytics',
      'Zebra Management',
    ]);
  });

  it('should apply L1 blue color to capability items', () => {
    const onCapabilityClick = vi.fn();
    renderWithDnd(
      <DomainGrid capabilities={mockL1Capabilities} onCapabilityClick={onCapabilityClick} />
    );

    const firstCapability = screen.getByText('Alpha Processing').closest('button');
    expect(firstCapability).toHaveStyle({ backgroundColor: 'rgb(59, 130, 246)' });
  });

  it('should only render L1 capabilities and ignore others', () => {
    const onCapabilityClick = vi.fn();
    const mixedCapabilities: Capability[] = [
      createCapability('cap-1', 'L1 Capability', 'L1'),
      createCapability('cap-2', 'L2 Capability', 'L2'),
      createCapability('cap-3', 'L3 Capability', 'L3'),
      createCapability('cap-4', 'L4 Capability', 'L4'),
    ];

    renderWithDnd(<DomainGrid capabilities={mixedCapabilities} onCapabilityClick={onCapabilityClick} />);

    expect(screen.getByText('L1 Capability')).toBeInTheDocument();
    expect(screen.queryByText('L2 Capability')).not.toBeInTheDocument();
    expect(screen.queryByText('L3 Capability')).not.toBeInTheDocument();
    expect(screen.queryByText('L4 Capability')).not.toBeInTheDocument();
  });

  it('should call onCapabilityClick when a capability is clicked', () => {
    const onCapabilityClick = vi.fn();
    renderWithDnd(
      <DomainGrid capabilities={mockL1Capabilities} onCapabilityClick={onCapabilityClick} />
    );

    const capability = screen.getByText('Alpha Processing').closest('button');
    capability?.click();

    expect(onCapabilityClick).toHaveBeenCalledWith(mockL1Capabilities[1]);
  });

  it('should render empty grid when no capabilities provided', () => {
    const onCapabilityClick = vi.fn();
    const { container } = renderWithDnd(
      <DomainGrid capabilities={[]} onCapabilityClick={onCapabilityClick} />
    );

    expect(container.querySelector('.domain-grid')).toBeInTheDocument();
    expect(screen.queryAllByRole('button')).toHaveLength(0);
  });

  it('should have droppable area for drag and drop', () => {
    const onCapabilityClick = vi.fn();
    const { container } = renderWithDnd(
      <DomainGrid capabilities={mockL1Capabilities} onCapabilityClick={onCapabilityClick} />
    );

    expect(container.querySelector('[data-dnd-context]')).toBeInTheDocument();
  });

  describe('with positions', () => {
    it('should order capabilities by position when provided', () => {
      const onCapabilityClick = vi.fn();
      const positions = {
        'cap-1': createPosition(2, 0),
        'cap-2': createPosition(0, 0),
        'cap-3': createPosition(1, 0),
      };

      renderWithDnd(
        <DomainGrid
          capabilities={mockL1Capabilities}
          onCapabilityClick={onCapabilityClick}
          positions={positions}
        />
      );

      const capabilityNames = screen
        .getAllByRole('button')
        .map((btn) => btn.textContent);

      expect(capabilityNames).toEqual([
        'Alpha Processing',
        'Delta Analytics',
        'Zebra Management',
      ]);
    });

    it('should place capabilities without positions at the end', () => {
      const onCapabilityClick = vi.fn();
      const positions = {
        'cap-1': createPosition(0, 0),
      };

      renderWithDnd(
        <DomainGrid
          capabilities={mockL1Capabilities}
          onCapabilityClick={onCapabilityClick}
          positions={positions}
        />
      );

      const capabilityNames = screen
        .getAllByRole('button')
        .map((btn) => btn.textContent);

      expect(capabilityNames[0]).toBe('Zebra Management');
    });

    it('should make items sortable when positions are provided', () => {
      const onCapabilityClick = vi.fn();
      const positions = {
        'cap-1': createPosition(0, 0),
        'cap-2': createPosition(1, 0),
      };

      const { container } = renderWithDnd(
        <DomainGrid
          capabilities={mockL1Capabilities}
          onCapabilityClick={onCapabilityClick}
          positions={positions}
        />
      );

      const sortableItems = container.querySelectorAll('[data-sortable]');
      expect(sortableItems.length).toBeGreaterThan(0);
    });
  });
});
