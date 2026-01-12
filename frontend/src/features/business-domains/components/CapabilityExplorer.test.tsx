import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
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
  _links: { self: { href: `/api/v1/capabilities/${id}`, method: 'GET' } },
});

describe('CapabilityExplorer', () => {
  const mockCapabilities: Capability[] = [
    createCapability('cap-1', 'Customer Management', 'L1'),
    createCapability('cap-2', 'Customer Onboarding', 'L2', 'cap-1'),
    createCapability('cap-3', 'Customer Verification', 'L3', 'cap-2'),
    createCapability('cap-4', 'Order Processing', 'L1'),
    createCapability('cap-5', 'Order Validation', 'L2', 'cap-4'),
  ];

  describe('Hierarchical Nesting', () => {
    it('should render L2 capabilities nested under their L1 parent', () => {
      render(
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
      render(
        <CapabilityExplorer
          capabilities={mockCapabilities}
          assignedCapabilityIds={new Set()}
          isLoading={false}
        />
      );

      expect(screen.getByText('Customer Verification')).toBeInTheDocument();
    });
  });

  describe('L1-Only Draggable Rule', () => {
    it('should make L1 items draggable', () => {
      render(
        <CapabilityExplorer
          capabilities={mockCapabilities}
          assignedCapabilityIds={new Set()}
          isLoading={false}
        />
      );

      const l1Item = screen.getByTestId('draggable-cap-1');
      expect(l1Item).toBeInTheDocument();
      expect(l1Item).toHaveAttribute('data-draggable', 'true');
      expect(l1Item).toHaveAttribute('draggable', 'true');
    });

    it('should not make non-L1 items draggable', () => {
      render(
        <CapabilityExplorer
          capabilities={mockCapabilities}
          assignedCapabilityIds={new Set()}
          isLoading={false}
        />
      );

      expect(screen.queryByTestId('draggable-cap-2')).not.toBeInTheDocument();
      expect(screen.queryByTestId('draggable-cap-3')).not.toBeInTheDocument();
    });
  });

  describe('Alphabetical Sorting', () => {
    it('should sort L1 capabilities alphabetically', () => {
      const unsortedCapabilities: Capability[] = [
        createCapability('cap-z', 'Zebra', 'L1'),
        createCapability('cap-a', 'Alpha', 'L1'),
        createCapability('cap-m', 'Middle', 'L1'),
      ];

      render(
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
});
