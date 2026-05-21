import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Capability } from '../../../api/types';
import { toCapabilityId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import type { CapabilityLinkStatus, CapabilityLinkStatusResponse } from '../types';
import { DomainCapabilityPanel } from './DomainCapabilityPanel';

function render(ui: React.ReactElement) {
  return renderWithProviders(ui, { withRouter: false });
}

const mockCapabilities: Capability[] = [
  {
    id: toCapabilityId('cap-1'),
    name: 'Payment Processing',
    level: 'L1',
    status: 'active',
    createdAt: '2025-01-01T00:00:00Z',
    _links: { self: { href: '/api/v1/capabilities/cap-1', method: 'GET' } },
  },
  {
    id: toCapabilityId('cap-2'),
    name: 'Order Management',
    level: 'L1',
    status: 'active',
    createdAt: '2025-01-01T00:00:00Z',
    _links: { self: { href: '/api/v1/capabilities/cap-2', method: 'GET' } },
  },
];

function statusMap(entries: Array<[string, CapabilityLinkStatusResponse]>) {
  return new Map(entries);
}

function renderPanel(linkStatuses: Map<string, CapabilityLinkStatusResponse>, capabilities = mockCapabilities) {
  return render(<DomainCapabilityPanel capabilities={capabilities} linkStatuses={linkStatuses} isLoading={false} />);
}

describe('DomainCapabilityPanel', () => {
  describe('Visual State Rendering', () => {
    type VisualCase = {
      name: string;
      response: CapabilityLinkStatusResponse;
      expectedStatusText: RegExp | null;
    };

    const cases: VisualCase[] = [
      {
        name: 'available capabilities omit status indicators',
        response: { capabilityId: 'cap-1', status: 'available' },
        expectedStatusText: null,
      },
      {
        name: 'linked capabilities show arrow indicator',
        response: {
          capabilityId: 'cap-1',
          status: 'linked',
          linkedTo: { id: 'ec-1', name: 'Customer Management' },
        },
        expectedStatusText: /──► Customer Management/,
      },
      {
        name: 'blocked_by_parent shows parent name',
        response: {
          capabilityId: 'cap-1',
          status: 'blocked_by_parent',
          blockingCapability: { id: 'cap-parent', name: 'Parent Capability' },
        },
        expectedStatusText: /Parent linked to Parent Capability/i,
      },
      {
        name: 'blocked_by_child shows child name',
        response: {
          capabilityId: 'cap-1',
          status: 'blocked_by_child',
          blockingCapability: { id: 'cap-child', name: 'Child Capability' },
        },
        expectedStatusText: /Child linked to Child Capability/i,
      },
    ];

    it.each(cases)('$name', ({ response, expectedStatusText }) => {
      renderPanel(statusMap([['cap-1', response]]));

      expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      if (expectedStatusText) {
        expect(screen.getByText(expectedStatusText)).toBeInTheDocument();
      } else {
        expect(screen.queryByText(/Linked/)).not.toBeInTheDocument();
        expect(screen.queryByText(/blocked/i)).not.toBeInTheDocument();
      }
    });
  });

  describe('Drag and Drop', () => {
    type DraggabilityCase = {
      name: string;
      status: CapabilityLinkStatus;
      response: CapabilityLinkStatusResponse;
      expectedDraggable: 'true' | 'false';
    };

    const cases: DraggabilityCase[] = [
      {
        name: 'available is draggable',
        status: 'available',
        response: { capabilityId: 'cap-1', status: 'available' },
        expectedDraggable: 'true',
      },
      {
        name: 'linked is not draggable',
        status: 'linked',
        response: {
          capabilityId: 'cap-1',
          status: 'linked',
          linkedTo: { id: 'ec-1', name: 'Customer Management' },
        },
        expectedDraggable: 'false',
      },
      {
        name: 'blocked_by_parent is not draggable',
        status: 'blocked_by_parent',
        response: {
          capabilityId: 'cap-1',
          status: 'blocked_by_parent',
          blockingCapability: { id: 'cap-parent', name: 'Parent Capability' },
        },
        expectedDraggable: 'false',
      },
    ];

    it.each(cases)('$name', ({ response, expectedDraggable }) => {
      renderPanel(statusMap([['cap-1', response]]));

      const element = screen.getByText('Payment Processing').closest(`div[draggable="${expectedDraggable}"]`);
      expect(element).toBeInTheDocument();
      expect(element).toHaveAttribute('draggable', expectedDraggable);
    });

    it('calls onDragStart when dragging begins', () => {
      const onDragStart = vi.fn();
      const linkStatuses = statusMap([['cap-1', { capabilityId: 'cap-1', status: 'available' }]]);

      render(
        <DomainCapabilityPanel
          capabilities={mockCapabilities}
          linkStatuses={linkStatuses}
          isLoading={false}
          onDragStart={onDragStart}
          onDragEnd={vi.fn()}
        />,
      );

      const draggable = screen.getByText('Payment Processing').closest('div[draggable="true"]')!;
      fireEvent.dragStart(draggable, {
        dataTransfer: { setData: vi.fn(), effectAllowed: 'move' },
      });

      expect(onDragStart).toHaveBeenCalledWith(mockCapabilities[0]);
    });
  });

  describe('Data Loading', () => {
    it('shows loading message when isLoading is true', () => {
      render(<DomainCapabilityPanel capabilities={[]} linkStatuses={new Map()} isLoading={true} />);
      expect(screen.getByText('Loading domain capabilities...')).toBeInTheDocument();
    });

    it('shows empty message when no capabilities are available', () => {
      render(<DomainCapabilityPanel capabilities={[]} linkStatuses={new Map()} isLoading={false} />);
      expect(screen.getByText('No domain capabilities available')).toBeInTheDocument();
    });

    it('renders capability list when data is loaded', () => {
      renderPanel(
        statusMap([
          ['cap-1', { capabilityId: 'cap-1', status: 'available' }],
          ['cap-2', { capabilityId: 'cap-2', status: 'available' }],
        ]),
      );

      expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      expect(screen.getByText('Order Management')).toBeInTheDocument();
    });
  });

  describe('Hierarchy Tree Structure', () => {
    it('displays nested capabilities in tree structure', () => {
      const hierarchicalCapabilities: Capability[] = [
        {
          id: toCapabilityId('parent-1'),
          name: 'Parent Capability',
          level: 'L1',
          status: 'active',
          createdAt: '2025-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/parent-1', method: 'GET' } },
        },
        {
          id: toCapabilityId('child-1'),
          name: 'Child Capability',
          level: 'L2',
          parentId: toCapabilityId('parent-1'),
          status: 'active',
          createdAt: '2025-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/child-1', method: 'GET' } },
        },
      ];
      const linkStatuses = statusMap([
        ['parent-1', { capabilityId: 'parent-1', status: 'available' }],
        ['child-1', { capabilityId: 'child-1', status: 'available' }],
      ]);

      renderPanel(linkStatuses, hierarchicalCapabilities);

      expect(screen.getByText('Parent Capability')).toBeInTheDocument();
      expect(screen.getByText('Child Capability')).toBeInTheDocument();
    });
  });
});
