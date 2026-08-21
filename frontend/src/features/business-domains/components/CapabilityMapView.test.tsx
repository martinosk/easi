import { fireEvent, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap, buildCapabilityRealization } from '../../../test/helpers/entityBuilders';
import { buildHookData, buildViewModel } from '../testkit/hookData';
import { CapabilityMapView } from './CapabilityMapView';

const financeRealizations = [
  {
    capabilityId: toCapabilityId('d1-l1'),
    realizations: [
      buildCapabilityRealization({
        capabilityId: toCapabilityId('d1-l1'),
        componentId: toComponentId('comp-erp'),
        componentName: 'ERP System',
      }),
    ],
  },
];

function nestedHookData(overrides = {}) {
  return buildHookData({
    boardDomains: [
      buildViewModel(
        'Finance',
        'domain-1',
        [cap('d1-l1', 'Billing', 'L1'), cap('d1-l2', 'Invoicing', 'L2', 'd1-l1'), cap('d1-l3', 'Dunning', 'L3', 'd1-l2')],
        financeRealizations,
      ),
      buildViewModel('Logistics', 'domain-2'),
    ],
    ...overrides,
  });
}

function renderMapView(hookData = nestedHookData(), onViewModeChange = vi.fn()) {
  renderWithProviders(<CapabilityMapView hookData={hookData} viewMode="map" onViewModeChange={onViewModeChange} />);
  return { hookData, onViewModeChange };
}

beforeEach(() => {
  localStorage.clear();
});

describe('CapabilityMapView', () => {
  it('renders the map for the first domain by default', () => {
    renderMapView();
    expect(screen.getByTestId('capability-map')).toBeInTheDocument();
    expect(screen.getByTestId('map-cell-d1-l1')).toBeInTheDocument();
    expect(screen.queryByTestId('map-cell-domain-2-l1')).toBeNull();
  });

  it('switches the mapped domain via the domain selector', async () => {
    const user = userEvent.setup();
    renderMapView();

    await user.click(screen.getByTestId('map-domain-select'));
    const option = (await screen.findAllByText('Logistics')).find((el) => el.closest('[data-combobox-option]'));
    if (!option) throw new Error('Logistics option not found');
    await user.click(option);

    expect(screen.getByTestId('map-cell-domain-2-l1')).toBeInTheDocument();
    expect(screen.queryByTestId('map-cell-d1-l1')).toBeNull();
  });

  it('defaults to depth L1–L2 and reveals deeper levels when the depth changes', async () => {
    const user = userEvent.setup();
    renderMapView();

    expect(screen.getByTestId('map-cell-d1-l2')).toBeInTheDocument();
    expect(screen.queryByTestId('map-cell-d1-l3')).toBeNull();

    await user.click(screen.getByRole('radio', { name: 'L3' }));

    expect(screen.getByTestId('map-cell-d1-l3')).toBeInTheDocument();
  });

  it('opens the capability drawer flow when a cell is clicked', () => {
    const { hookData } = renderMapView();

    fireEvent.click(screen.getByTestId('map-cell-d1-l2'));

    expect(hookData.handleCapabilityClick).toHaveBeenCalledWith(
      'domain-1',
      expect.objectContaining({ name: 'Invoicing' }),
      expect.anything(),
    );
  });

  it('switches back to the board through the view toggle', async () => {
    const user = userEvent.setup();
    const { onViewModeChange } = renderMapView();

    await user.click(screen.getByRole('radio', { name: 'Board' }));

    expect(onViewModeChange).toHaveBeenCalledWith('board');
  });

  it('keeps search available but offers no lens switcher in map view', () => {
    renderMapView();
    expect(screen.queryByTestId('lens-switcher')).toBeNull();
    expect(screen.getByTestId('board-search-input')).toBeInTheDocument();
  });

  it('renders the Now legend even when the board lens is journey', () => {
    renderMapView(nestedHookData({ lens: 'journey' }));
    expect(screen.getByTestId('board-legend')).toHaveTextContent('Full');
    expect(screen.queryByText('in transition')).toBeNull();
    expect(screen.queryByTestId('changes-only-toggle')).toBeNull();
  });

  it('reveals application chips through the persisted Show apps toggle', async () => {
    const user = userEvent.setup();
    renderMapView();

    expect(screen.queryByTestId('app-chip-comp-erp')).toBeNull();

    await user.click(screen.getByTestId('map-show-apps-toggle'));
    expect(screen.getByTestId('app-chip-comp-erp')).toBeInTheDocument();
    expect(localStorage.getItem('business-domains-map-apps')).toBe('true');
  });

  it('opens the application drawer flow when a chip is clicked', async () => {
    const user = userEvent.setup();
    const { hookData } = renderMapView();

    await user.click(screen.getByTestId('map-show-apps-toggle'));
    await user.click(screen.getByTestId('app-chip-comp-erp'));

    expect(hookData.handleApplicationClick).toHaveBeenCalledWith('comp-erp');
    expect(hookData.handleCapabilityClick).not.toHaveBeenCalled();
  });

  it('offers no board mutation affordances', () => {
    renderMapView();
    expect(screen.queryByTestId('assign-rail-toggle')).toBeNull();
    expect(screen.queryByTestId('create-domain-button')).toBeNull();
  });
});
