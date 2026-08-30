import { screen, waitFor, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { EnterpriseCapabilityId, HATEOASLink } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { seedOnePagerCompleteness } from '../../../test/mocks/onePagerCompleteness';
import { seedSpec172Db } from '../../../test/mocks/spec172/store';
import type { EnterpriseCapability } from '../types';
import { EnterpriseCapabilitiesTable } from './EnterpriseCapabilitiesTable';

function render(ui: React.ReactElement) {
  return renderWithProviders(ui, { withRouter: false });
}

function createCapability(overrides: Partial<EnterpriseCapability> = {}): EnterpriseCapability {
  const self: HATEOASLink = { href: '/api/v1/enterprise-capabilities/ec-1', method: 'GET' };
  return {
    id: 'ec-1' as EnterpriseCapabilityId,
    name: 'Order Management',
    description: 'Manages orders',
    category: 'Core',
    active: true,
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self },
    ...overrides,
  };
}

describe('EnterpriseCapabilitiesTable', () => {
  describe('composition counts', () => {
    it('shows the included-capability and domain counts from the composition summaries', async () => {
      seedSpec172Db({
        enterpriseCapabilities: [{ id: 'ec-1', name: 'Order Management', active: true, createdAt: '2024-01-01T00:00:00Z' }],
        directions: [
          {
            id: 'dir-1',
            enterpriseCapabilityId: 'ec-1',
            type: 'consolidate',
            status: 'proposed',
            horizon: 'next',
            sourceCapabilityIds: ['cap-a'],
            createdAt: '2024-01-01T00:00:00Z',
          },
        ],
        capabilities: [
          {
            id: 'cap-a',
            name: 'A',
            level: 'L1',
            parentId: null,
            businessDomainId: 'bd-1',
            businessDomainName: 'Sales',
          },
          {
            id: 'cap-b',
            name: 'B',
            level: 'L2',
            parentId: 'cap-a',
            businessDomainId: 'bd-1',
            businessDomainName: 'Sales',
          },
        ],
      });

      render(<EnterpriseCapabilitiesTable capabilities={[createCapability()]} onSelect={vi.fn()} onDelete={vi.fn()} />);

      const row = screen.getByTestId('capability-row-ec-1');
      await waitFor(() => expect(within(row).getByTestId('included-count-ec-1')).toHaveTextContent('2'));
      expect(within(row).getByTestId('domain-count-ec-1')).toHaveTextContent('1');
    });

    it('shows zero counts for an enterprise capability without an active direction', async () => {
      seedSpec172Db({
        enterpriseCapabilities: [{ id: 'ec-1', name: 'Order Management', active: true, createdAt: '2024-01-01T00:00:00Z' }],
      });

      render(<EnterpriseCapabilitiesTable capabilities={[createCapability()]} onSelect={vi.fn()} onDelete={vi.fn()} />);

      await waitFor(() => expect(screen.getByTestId('included-count-ec-1')).toHaveTextContent('0'));
      expect(screen.getByTestId('domain-count-ec-1')).toHaveTextContent('0');
    });

    it('shows a dash while no summary is available for the enterprise capability', async () => {
      seedSpec172Db({ enterpriseCapabilities: [] });

      render(<EnterpriseCapabilitiesTable capabilities={[createCapability()]} onSelect={vi.fn()} onDelete={vi.fn()} />);

      await waitFor(() => expect(screen.getByTestId('included-count-ec-1')).toHaveTextContent('—'));
      expect(screen.getByTestId('domain-count-ec-1')).toHaveTextContent('—');
    });
  });

  describe('one-pager completeness indicator', () => {
    it('shows the indicator only for enterprise capabilities reported as incomplete', async () => {
      seedOnePagerCompleteness('enterprise-capability', [
        { subjectId: 'ec-1', complete: false },
        { subjectId: 'ec-2', complete: true },
      ]);
      const capabilities = [
        createCapability(),
        createCapability({ id: 'ec-2' as EnterpriseCapabilityId, name: 'Billing' }),
      ];
      render(<EnterpriseCapabilitiesTable capabilities={capabilities} onSelect={vi.fn()} onDelete={vi.fn()} />);

      expect(await screen.findByTestId('one-pager-incomplete-ec-1')).toBeInTheDocument();
      expect(screen.queryByTestId('one-pager-incomplete-ec-2')).not.toBeInTheDocument();
    });

    it('shows no indicator when the subject type has no required field', async () => {
      seedOnePagerCompleteness('enterprise-capability', []);
      render(<EnterpriseCapabilitiesTable capabilities={[createCapability()]} onSelect={vi.fn()} onDelete={vi.fn()} />);

      await waitFor(() => expect(screen.queryByTestId('one-pager-incomplete-ec-1')).not.toBeInTheDocument());
    });
  });
});
