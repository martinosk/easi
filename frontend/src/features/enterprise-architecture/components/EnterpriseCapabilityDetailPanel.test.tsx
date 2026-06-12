import { screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { toEnterpriseCapabilityId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { seedSpec172Db } from '../../../test/mocks/spec172/store';
import type { EnterpriseCapability } from '../types';
import { EnterpriseCapabilityDetailPanel } from './EnterpriseCapabilityDetailPanel';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function ec(overrides: Partial<EnterpriseCapability> = {}): EnterpriseCapability {
  return {
    id: toEnterpriseCapabilityId('ec-crm'),
    name: 'CRM',
    description: 'Customer relationship management',
    category: 'Customer Domain',
    active: true,
    includedCapabilityCount: 2,
    domainCount: 1,
    createdAt: '2026-01-01T00:00:00Z',
    _links: { self: { href: '/api/v1/enterprise-capabilities/ec-crm', method: 'GET' } },
    ...overrides,
  };
}

function seedComposed() {
  seedSpec172Db({
    enterpriseCapabilities: [
      { id: 'ec-crm', name: 'CRM', active: true, createdAt: '2026-01-01T00:00:00Z' },
      { id: 'ec-tp', name: 'Take Payment', active: true, createdAt: '2026-01-01T00:00:00Z' },
    ],
    directions: [
      {
        id: 'dir-crm',
        enterpriseCapabilityId: 'ec-crm',
        type: 'consolidate',
        status: 'proposed',
        horizon: 'next',
        sourceCapabilityIds: ['cap-cim'],
        createdAt: '2026-01-01T00:00:00Z',
      },
      {
        id: 'dir-tp',
        enterpriseCapabilityId: 'ec-tp',
        type: 'consolidate',
        status: 'proposed',
        horizon: 'next',
        sourceCapabilityIds: ['cap-fraud'],
        createdAt: '2026-01-01T00:00:00Z',
      },
    ],
    capabilities: [
      {
        id: 'cap-cim',
        name: 'Customer Identity Mgmt',
        level: 'L1',
        parentId: null,
        businessDomainId: 'bd-c',
        businessDomainName: 'Customer Domain',
      },
      {
        id: 'cap-consent',
        name: 'Customer Consent',
        level: 'L2',
        parentId: 'cap-cim',
        businessDomainId: 'bd-c',
        businessDomainName: 'Customer Domain',
      },
      {
        id: 'cap-fraud',
        name: 'Customer Fraud Prevention',
        level: 'L2',
        parentId: 'cap-cim',
        businessDomainId: 'bd-c',
        businessDomainName: 'Customer Domain',
      },
    ],
  });
}

describe('EnterpriseCapabilityDetailPanel', () => {
  it('renders the EC name and the included-capabilities composition', async () => {
    seedComposed();
    renderWithProviders(<EnterpriseCapabilityDetailPanel capability={ec()} onClose={vi.fn()} />, {
      withRouter: false,
    });

    expect(screen.getByRole('heading', { name: 'CRM' })).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText('Customer Identity Mgmt')).toBeInTheDocument());
    expect(screen.getByTestId('included-row-cap-consent')).toHaveTextContent(/via parent/i);
    expect(screen.getByTestId('included-row-cap-fraud')).toHaveTextContent(/Take Payment/);
  });

  it('shows an empty included-capabilities state when there is no direction', async () => {
    seedSpec172Db({
      enterpriseCapabilities: [{ id: 'ec-crm', name: 'CRM', active: true, createdAt: '2026-01-01T00:00:00Z' }],
    });
    renderWithProviders(
      <EnterpriseCapabilityDetailPanel capability={ec({ includedCapabilityCount: 0 })} onClose={vi.fn()} />,
      {
        withRouter: false,
      },
    );

    await waitFor(() => expect(screen.getByTestId('included-empty-state')).toBeInTheDocument());
  });

  it('does not render any "Linked Capabilities" linking section', async () => {
    seedComposed();
    renderWithProviders(<EnterpriseCapabilityDetailPanel capability={ec()} onClose={vi.fn()} />, {
      withRouter: false,
    });
    await waitFor(() => expect(screen.getByText('Customer Identity Mgmt')).toBeInTheDocument());
    expect(screen.queryByText(/Linked Capabilities/i)).not.toBeInTheDocument();
  });
});
