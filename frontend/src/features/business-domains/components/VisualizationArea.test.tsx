import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { BusinessDomain, BusinessDomainId, Capability, CapabilityId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers/renderWithProviders';
import { VisualizationArea } from './VisualizationArea';

function render(ui: React.ReactElement) {
  return renderWithProviders(ui, { withRouter: false });
}

describe('VisualizationArea', () => {
  const mockDomain: BusinessDomain = {
    id: 'domain-1' as BusinessDomainId,
    name: 'Finance',
    description: '',
    capabilityCount: 2,
    createdAt: '2024-01-01',
    _links: {
      self: { href: '/api/v1/business-domains/domain-1', method: 'GET' },
      capabilities: { href: '/api/v1/business-domains/domain-1/capabilities', method: 'GET' },
    },
  };

  const mockCapabilities: Capability[] = [
    {
      id: 'cap-1' as CapabilityId,
      name: 'Financial Management',
      level: 'L1',
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/capabilities/cap-1', method: 'GET' } },
    },
    {
      id: 'cap-2' as CapabilityId,
      name: 'Accounting',
      level: 'L2',
      parentId: 'cap-1' as CapabilityId,
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/capabilities/cap-2', method: 'GET' } },
    },
  ];

  const defaultProps = {
    visualizedDomain: mockDomain,
    capabilities: mockCapabilities,
    capabilitiesLoading: false,
    depth: 4 as const,
    positions: {},
    onDepthChange: vi.fn(),
    onCapabilityClick: vi.fn(),
    onContextMenu: vi.fn(),
    selectedCapabilities: new Set<CapabilityId>(),
    showApplications: false,
    onShowApplicationsChange: vi.fn(),
    getRealizationsForCapability: vi.fn().mockReturnValue([]),
    onApplicationClick: vi.fn(),
  };

  it('renders capabilities when domain is selected', () => {
    render(<VisualizationArea {...defaultProps} />);

    expect(screen.getByText('Financial Management')).toBeInTheDocument();
    expect(screen.getByText('Accounting')).toBeInTheDocument();
  });

  it.each([
    {
      label: 'onCapabilityClick',
      propKey: 'onCapabilityClick' as const,
      fire: (el: HTMLElement) => fireEvent.click(el),
    },
    {
      label: 'onContextMenu',
      propKey: 'onContextMenu' as const,
      fire: (el: HTMLElement) => fireEvent.contextMenu(el),
    },
  ])('forwards $label with the matching capability and event', ({ propKey, fire }) => {
    const handler = vi.fn();
    render(<VisualizationArea {...defaultProps} {...{ [propKey]: handler }} />);

    fire(screen.getByText('Financial Management'));

    expect(handler).toHaveBeenCalledTimes(1);
    expect(handler).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'cap-1', name: 'Financial Management' }),
      expect.any(Object),
    );
  });

  it('marks selected capabilities so the grid can highlight them', () => {
    const selectedCapabilities = new Set(['cap-1' as CapabilityId]);
    render(<VisualizationArea {...defaultProps} selectedCapabilities={selectedCapabilities} />);

    expect(screen.getByTestId('capability-cap-1')).toHaveAttribute('data-selected', 'true');
  });

  it('shows placeholder when no domain is selected', () => {
    render(<VisualizationArea {...defaultProps} visualizedDomain={null} />);

    expect(screen.getByText('Click a domain to see its capabilities')).toBeInTheDocument();
  });

  it('shows loading state when capabilities are loading', () => {
    render(<VisualizationArea {...defaultProps} capabilitiesLoading={true} />);

    expect(screen.getByText('Loading capabilities...')).toBeInTheDocument();
  });
});
