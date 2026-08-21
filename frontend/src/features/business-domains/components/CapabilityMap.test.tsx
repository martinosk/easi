import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap, buildCapabilityRealization } from '../../../test/helpers/entityBuilders';
import type { MapDepth } from '../hooks/useMapViewState';
import { buildViewModel } from '../testkit/hookData';
import { CapabilityMap } from './CapabilityMap';

const capabilities = [
  cap('l1-b', 'Billing', 'L1'),
  cap('l2-inv', 'Invoicing', 'L2', 'l1-b'),
  cap('l3-du', 'Dunning', 'L3', 'l2-inv'),
  cap('l1-a', 'Analytics', 'L1'),
];

const billingRealizations = [
  {
    capabilityId: toCapabilityId('l1-b'),
    realizations: [
      buildCapabilityRealization({
        capabilityId: toCapabilityId('l1-b'),
        componentId: toComponentId('comp-erp'),
        componentName: 'ERP System',
      }),
    ],
  },
];

interface RenderMapOptions {
  depth?: MapDepth;
  searchQuery?: string;
  assigned?: typeof capabilities;
  showApps?: boolean;
  withRealizations?: boolean;
  onCapabilityClick?: ReturnType<typeof vi.fn>;
  onChipClick?: ReturnType<typeof vi.fn>;
}

function renderMap({
  depth = 4,
  searchQuery = '',
  assigned = capabilities,
  showApps = false,
  withRealizations = false,
  onCapabilityClick = vi.fn(),
  onChipClick = vi.fn(),
}: RenderMapOptions = {}) {
  return renderWithProviders(
    <CapabilityMap
      viewModel={buildViewModel('Finance', 'dom-finance', assigned, withRealizations ? billingRealizations : [])}
      depth={depth}
      searchQuery={searchQuery}
      showApps={showApps}
      getColorForValue={() => '#123456'}
      onCapabilityClick={onCapabilityClick}
      onChipClick={onChipClick}
    />,
    { withRouter: false },
  );
}

describe('CapabilityMap', () => {
  it('renders L1 blocks with nested children down to the selected depth', () => {
    renderMap({ depth: 2 });
    expect(screen.getByTestId('map-cell-l1-b')).toBeInTheDocument();
    expect(screen.getByTestId('map-cell-l2-inv')).toBeInTheDocument();
    expect(screen.queryByTestId('map-cell-l3-du')).toBeNull();
  });

  it('shows deeper levels when depth allows', () => {
    renderMap({ depth: 3 });
    expect(screen.getByTestId('map-cell-l3-du')).toBeInTheDocument();
  });

  it('orders L1 blocks alphabetically', () => {
    renderMap({ depth: 1 });
    const cells = screen.getAllByTestId(/^map-cell-l1/);
    expect(cells.map((el) => el.dataset.testid)).toEqual(['map-cell-l1-a', 'map-cell-l1-b']);
  });

  it('invokes the click handler with the clicked capability', () => {
    const onCapabilityClick = vi.fn();
    renderMap({ depth: 2, onCapabilityClick });
    fireEvent.click(screen.getByTestId('map-cell-l2-inv'));
    expect(onCapabilityClick).toHaveBeenCalledWith(expect.objectContaining({ name: 'Invoicing' }), expect.anything());
  });

  it('renders children as boxes inside their parent box', () => {
    renderMap({ depth: 3 });
    expect(screen.getByTestId('map-cell-l1-b')).toContainElement(screen.getByTestId('map-cell-l2-inv'));
    expect(screen.getByTestId('map-cell-l2-inv')).toContainElement(screen.getByTestId('map-cell-l3-du'));
  });

  it('opens only the clicked box, not its ancestors', () => {
    const onCapabilityClick = vi.fn();
    renderMap({ depth: 3, onCapabilityClick });
    fireEvent.click(screen.getByTestId('map-cell-l3-du'));
    expect(onCapabilityClick).toHaveBeenCalledTimes(1);
    expect(onCapabilityClick).toHaveBeenCalledWith(expect.objectContaining({ name: 'Dunning' }), expect.anything());
  });

  it('de-emphasizes cells that do not match the search but keeps ancestors of matches visible', () => {
    renderMap({ depth: 3, searchQuery: 'dunning' });
    expect(screen.getByTestId('map-cell-l1-a').dataset.dimmed).toBe('true');
    expect(screen.getByTestId('map-cell-l1-b').dataset.dimmed).toBeUndefined();
    expect(screen.getByTestId('map-cell-l3-du').dataset.dimmed).toBeUndefined();
  });

  it('never renders journey status pills — the map is Now-only', () => {
    renderMap({ depth: 1 });
    expect(document.querySelector('[data-journey-status]')).toBeNull();
    expect(screen.queryByTestId('status-pill-in-flight')).toBeNull();
  });

  it('shows an empty state for a domain without capabilities', () => {
    renderMap({ assigned: [] });
    expect(screen.queryByTestId('capability-map')).toBeNull();
    expect(screen.getByTestId('capability-map-empty')).toHaveTextContent(/no capabilities/i);
  });

  it('hides application chips while the apps toggle is off', () => {
    renderMap({ depth: 1, withRealizations: true, showApps: false });
    expect(screen.queryByTestId('app-chip-comp-erp')).toBeNull();
  });

  it('shows application chips on cells when the apps toggle is on', () => {
    renderMap({ depth: 1, withRealizations: true, showApps: true });
    expect(screen.getByTestId('app-chip-comp-erp')).toHaveTextContent('ERP System');
  });

  it('opens the application, not the capability, when a chip is clicked', () => {
    const onCapabilityClick = vi.fn();
    const onChipClick = vi.fn();
    renderMap({ depth: 1, withRealizations: true, showApps: true, onCapabilityClick, onChipClick });

    fireEvent.click(screen.getByTestId('app-chip-comp-erp'));

    expect(onChipClick).toHaveBeenCalledWith('comp-erp');
    expect(onCapabilityClick).not.toHaveBeenCalled();
  });

});
