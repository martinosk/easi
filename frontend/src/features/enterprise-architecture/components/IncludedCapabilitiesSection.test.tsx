import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { CompositionResponse } from '../types';
import { IncludedCapabilitiesSection } from './IncludedCapabilitiesSection';

const selfLink = (id: string) => ({ self: { href: `/api/v1/capabilities/${id}`, method: 'GET' as const } });

function composition(overrides: Partial<CompositionResponse> = {}): CompositionResponse {
  return {
    data: [
      {
        businessDomainId: 'bd-c',
        businessDomainName: 'Customer Domain',
        items: [
          {
            capabilityId: 'cap-cim',
            name: 'Customer Account Management',
            level: 'L1',
            businessDomainId: 'bd-c',
            businessDomainName: 'Customer Domain',
            role: 'source',
            carvedOutBy: null,
            _links: {
              ...selfLink('cap-cim'),
              'x-exclude': { href: '/api/v1/enterprise-capabilities/ec-1/direction/sources/cap-cim', method: 'DELETE' },
            },
          },
          {
            capabilityId: 'cap-consent',
            name: 'Customer Account Creation',
            level: 'L2',
            businessDomainId: 'bd-c',
            businessDomainName: 'Customer Domain',
            role: 'implicit',
            carvedOutBy: null,
            _links: selfLink('cap-consent'),
          },
          {
            capabilityId: 'cap-recovery',
            name: 'Account Recovery',
            level: 'L3',
            businessDomainId: 'bd-c',
            businessDomainName: 'Customer Domain',
            role: 'carved-out',
            carvedOutBy: { enterpriseCapabilityId: 'ec-tp', enterpriseCapabilityName: 'Take Payment' },
            _links: {
              ...selfLink('cap-recovery'),
              'x-owning-ec': { href: '/api/v1/enterprise-capabilities/ec-tp', method: 'GET' },
            },
          },
        ],
      },
    ],
    meta: { sourceCount: 1, includedCount: 2, carvedOutCount: 1, domainCount: 1 },
    _links: {},
    ...overrides,
  };
}

const noop = vi.fn();

describe('IncludedCapabilitiesSection', () => {
  it('shows a loader while loading', () => {
    renderWithProviders(
      <IncludedCapabilitiesSection composition={undefined} isLoading onExclude={noop} isExcluding={false} />,
      { withRouter: false },
    );
    expect(screen.getByTestId('included-capabilities-loading')).toBeInTheDocument();
  });

  it('shows an empty state when there is no composition', () => {
    const empty = composition({
      data: [],
      meta: { sourceCount: 0, includedCount: 0, carvedOutCount: 0, domainCount: 0 },
    });
    renderWithProviders(
      <IncludedCapabilitiesSection composition={empty} isLoading={false} onExclude={noop} isExcluding={false} />,
      { withRouter: false },
    );
    expect(screen.getByTestId('included-empty-state')).toHaveTextContent(/no capabilities included yet/i);
  });

  it('renders the source/included counts and domain-grouped rows with roles', () => {
    renderWithProviders(
      <IncludedCapabilitiesSection
        composition={composition()}
        isLoading={false}
        onExclude={noop}
        isExcluding={false}
      />,
      { withRouter: false },
    );
    expect(screen.getByTestId('composition-counts')).toHaveTextContent('1 sources · 2 included');
    expect(screen.getByText('Customer Domain')).toBeInTheDocument();
    expect(screen.getByText('Customer Account Management')).toBeInTheDocument();
    expect(screen.getByTestId('included-row-cap-consent')).toHaveTextContent(/via parent/i);
    const carved = screen.getByTestId('included-row-cap-recovery');
    expect(carved).toHaveTextContent(/carved out/i);
    expect(carved).toHaveTextContent(/Take Payment/);
  });

  it('renders an Exclude button only for source rows that carry an x-exclude link', () => {
    renderWithProviders(
      <IncludedCapabilitiesSection
        composition={composition()}
        isLoading={false}
        onExclude={noop}
        isExcluding={false}
      />,
      { withRouter: false },
    );
    expect(screen.getByTestId('exclude-cap-cim')).toBeInTheDocument();
    expect(screen.queryByTestId('exclude-cap-consent')).not.toBeInTheDocument();
    expect(screen.queryByTestId('exclude-cap-recovery')).not.toBeInTheDocument();
  });

  it('omits the Exclude button on a source without an x-exclude link (agreed direction)', () => {
    const agreed = composition();
    agreed.data[0].items[0]._links = selfLink('cap-cim'); // strip x-exclude
    renderWithProviders(
      <IncludedCapabilitiesSection composition={agreed} isLoading={false} onExclude={noop} isExcluding={false} />,
      { withRouter: false },
    );
    expect(screen.queryByTestId('exclude-cap-cim')).not.toBeInTheDocument();
  });

  it('calls onExclude with the capability id when Exclude is clicked', () => {
    const onExclude = vi.fn();
    renderWithProviders(
      <IncludedCapabilitiesSection
        composition={composition()}
        isLoading={false}
        onExclude={onExclude}
        isExcluding={false}
      />,
      { withRouter: false },
    );
    fireEvent.click(screen.getByTestId('exclude-cap-cim'));
    expect(onExclude).toHaveBeenCalledWith('cap-cim');
  });
});
