import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { VisualizationArea } from './VisualizationArea';
import type { BusinessDomain, Capability, CapabilityId } from '../../../api/types';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';

const renderWithMantine = (ui: React.ReactElement) => render(<MantineTestWrapper>{ui}</MantineTestWrapper>);

describe('VisualizationArea', () => {
  const mockDomain: BusinessDomain = {
    id: 'domain-1' as any,
    name: 'Finance',
    description: '',
    createdAt: '2024-01-01',
    _links: {
      self: { href: '/api/v1/business-domains/domain-1' },
      capabilities: '/api/v1/business-domains/domain-1/capabilities',
    },
  };

  const mockCapabilities: Capability[] = [
    {
      id: 'cap-1' as CapabilityId,
      name: 'Financial Management',
      level: 'L1',
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/capabilities/cap-1' } },
    },
    {
      id: 'cap-2' as CapabilityId,
      name: 'Accounting',
      level: 'L2',
      parentId: 'cap-1' as CapabilityId,
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/capabilities/cap-2' } },
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
    renderWithMantine(<VisualizationArea {...defaultProps} />);

    expect(screen.getByText('Financial Management')).toBeInTheDocument();
    expect(screen.getByText('Accounting')).toBeInTheDocument();
  });

  it('calls onCapabilityClick with capability and event when clicked', () => {
    const onCapabilityClick = vi.fn();
    renderWithMantine(<VisualizationArea {...defaultProps} onCapabilityClick={onCapabilityClick} />);

    fireEvent.click(screen.getByText('Financial Management'));

    expect(onCapabilityClick).toHaveBeenCalledTimes(1);
    expect(onCapabilityClick).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'cap-1', name: 'Financial Management' }),
      expect.any(Object)
    );
  });

  it('calls onContextMenu with capability and event on right-click', () => {
    const onContextMenu = vi.fn();
    renderWithMantine(<VisualizationArea {...defaultProps} onContextMenu={onContextMenu} />);

    fireEvent.contextMenu(screen.getByText('Financial Management'));

    expect(onContextMenu).toHaveBeenCalledTimes(1);
    expect(onContextMenu).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'cap-1', name: 'Financial Management' }),
      expect.any(Object)
    );
  });

  it('passes selectedCapabilities to grid for visual highlighting', () => {
    const selectedCapabilities = new Set(['cap-1' as CapabilityId]);
    renderWithMantine(<VisualizationArea {...defaultProps} selectedCapabilities={selectedCapabilities} />);

    const capability = screen.getByTestId('capability-cap-1');
    expect(capability).toHaveClass('selected');
  });

  it('shows placeholder when no domain is selected', () => {
    renderWithMantine(<VisualizationArea {...defaultProps} visualizedDomain={null} />);

    expect(screen.getByText('Click a domain to see its capabilities')).toBeInTheDocument();
  });

  it('shows loading state when capabilities are loading', () => {
    renderWithMantine(<VisualizationArea {...defaultProps} capabilitiesLoading={true} />);

    expect(screen.getByText('Loading capabilities...')).toBeInTheDocument();
  });
});
