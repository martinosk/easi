import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { DndContext } from '@dnd-kit/core';
import { CapabilityExplorer } from './CapabilityExplorer';
import type { Capability, CapabilityId } from '../../../api/types';

const createCapability = (
  id: string,
  name: string,
  level: 'L1' | 'L2' | 'L3' | 'L4' = 'L1',
  parentId?: string
): Capability => ({
  id: id as CapabilityId,
  name,
  level,
  parentId: parentId as CapabilityId | undefined,
  createdAt: '2024-01-01',
  _links: { self: { href: `/api/v1/capabilities/${id}` } },
});

describe('CapabilityExplorer', () => {
  const mockCapabilities: Capability[] = [
    createCapability('cap-1', 'Customer Management', 'L1'),
    createCapability('cap-2', 'Customer Onboarding', 'L2', 'cap-1'),
    createCapability('cap-3', 'Customer Verification', 'L3', 'cap-2'),
    createCapability('cap-4', 'Order Processing', 'L1'),
    createCapability('cap-5', 'Order Validation', 'L2', 'cap-4'),
  ];

  const renderWithDnd = (component: React.ReactNode) => {
    return render(<DndContext>{component}</DndContext>);
  };

  it('should render all L1 capabilities as top-level items', () => {
    renderWithDnd(
      <CapabilityExplorer
        capabilities={mockCapabilities}
        assignedCapabilityIds={new Set()}
        isLoading={false}
      />
    );

    expect(screen.getByText('Customer Management')).toBeInTheDocument();
    expect(screen.getByText('Order Processing')).toBeInTheDocument();
  });

  it('should render L2 capabilities nested under their L1 parent', () => {
    renderWithDnd(
      <CapabilityExplorer
        capabilities={mockCapabilities}
        assignedCapabilityIds={new Set()}
        isLoading={false}
      />
    );

    expect(screen.getByText('Customer Onboarding')).toBeInTheDocument();
    expect(screen.getByText('Order Validation')).toBeInTheDocument();
  });

  it('should render L3 capabilities nested under their L2 parent', () => {
    renderWithDnd(
      <CapabilityExplorer
        capabilities={mockCapabilities}
        assignedCapabilityIds={new Set()}
        isLoading={false}
      />
    );

    expect(screen.getByText('Customer Verification')).toBeInTheDocument();
  });

  it('should mark L1 capabilities that are assigned to other domains', () => {
    const assignedIds = new Set<CapabilityId>(['cap-1' as CapabilityId]);

    renderWithDnd(
      <CapabilityExplorer
        capabilities={mockCapabilities}
        assignedCapabilityIds={assignedIds}
        isLoading={false}
      />
    );

    const assignedIndicator = screen.getByTestId('assigned-indicator-cap-1');
    expect(assignedIndicator).toBeInTheDocument();
  });

  it('should not show assigned indicator for unassigned L1 capabilities', () => {
    renderWithDnd(
      <CapabilityExplorer
        capabilities={mockCapabilities}
        assignedCapabilityIds={new Set()}
        isLoading={false}
      />
    );

    expect(screen.queryByTestId('assigned-indicator-cap-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('assigned-indicator-cap-4')).not.toBeInTheDocument();
  });

  it('should make L1 items draggable', () => {
    renderWithDnd(
      <CapabilityExplorer
        capabilities={mockCapabilities}
        assignedCapabilityIds={new Set()}
        isLoading={false}
      />
    );

    const l1Item = screen.getByTestId('draggable-cap-1');
    expect(l1Item).toBeInTheDocument();
    expect(l1Item).toHaveAttribute('data-draggable', 'true');
  });

  it('should not make non-L1 items draggable', () => {
    renderWithDnd(
      <CapabilityExplorer
        capabilities={mockCapabilities}
        assignedCapabilityIds={new Set()}
        isLoading={false}
      />
    );

    expect(screen.queryByTestId('draggable-cap-2')).not.toBeInTheDocument();
    expect(screen.queryByTestId('draggable-cap-3')).not.toBeInTheDocument();
  });

  it('should display loading state', () => {
    renderWithDnd(
      <CapabilityExplorer
        capabilities={[]}
        assignedCapabilityIds={new Set()}
        isLoading={true}
      />
    );

    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('should display empty state when no capabilities', () => {
    renderWithDnd(
      <CapabilityExplorer
        capabilities={[]}
        assignedCapabilityIds={new Set()}
        isLoading={false}
      />
    );

    expect(screen.getByText(/no capabilities/i)).toBeInTheDocument();
  });

  it('should sort L1 capabilities alphabetically', () => {
    const unsortedCapabilities: Capability[] = [
      createCapability('cap-z', 'Zebra', 'L1'),
      createCapability('cap-a', 'Alpha', 'L1'),
      createCapability('cap-m', 'Middle', 'L1'),
    ];

    renderWithDnd(
      <CapabilityExplorer
        capabilities={unsortedCapabilities}
        assignedCapabilityIds={new Set()}
        isLoading={false}
      />
    );

    const items = screen.getAllByTestId(/^draggable-cap/);
    expect(items[0]).toHaveTextContent('Alpha');
    expect(items[1]).toHaveTextContent('Middle');
    expect(items[2]).toHaveTextContent('Zebra');
  });
});
