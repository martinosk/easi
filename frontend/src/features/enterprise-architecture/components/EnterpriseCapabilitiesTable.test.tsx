import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { EnterpriseCapabilityId, HATEOASLink } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
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
    includedCapabilityCount: 3,
    domainCount: 1,
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self },
    ...overrides,
  };
}

describe('EnterpriseCapabilitiesTable', () => {
  describe('one-pager completeness indicator', () => {
    it('should show the indicator when onePagerComplete is false', () => {
      const capability = createCapability({ id: 'ec-1' as EnterpriseCapabilityId, onePagerComplete: false });
      render(<EnterpriseCapabilitiesTable capabilities={[capability]} onSelect={vi.fn()} onDelete={vi.fn()} />);

      expect(screen.getByTestId('one-pager-incomplete-ec-1')).toBeInTheDocument();
    });

    it('should not show the indicator when onePagerComplete is true', () => {
      const capability = createCapability({ id: 'ec-1' as EnterpriseCapabilityId, onePagerComplete: true });
      render(<EnterpriseCapabilitiesTable capabilities={[capability]} onSelect={vi.fn()} onDelete={vi.fn()} />);

      expect(screen.queryByTestId('one-pager-incomplete-ec-1')).not.toBeInTheDocument();
    });

    it('should not show the indicator when onePagerComplete is absent', () => {
      const capability = createCapability({ id: 'ec-1' as EnterpriseCapabilityId });
      render(<EnterpriseCapabilitiesTable capabilities={[capability]} onSelect={vi.fn()} onDelete={vi.fn()} />);

      expect(screen.queryByTestId('one-pager-incomplete-ec-1')).not.toBeInTheDocument();
    });
  });
});
