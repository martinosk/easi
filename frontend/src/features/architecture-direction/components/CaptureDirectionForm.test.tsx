import { fireEvent, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { toEnterpriseCapabilityId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { seedSpec172Db } from '../../../test/mocks/spec172/store';
import { directionApi } from '../api/directionApi';
import { CaptureDirectionForm } from './CaptureDirectionForm';

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

function seed() {
  seedSpec172Db({
    enterpriseCapabilities: [
      { id: 'ec-crm', name: 'CRM', active: true, createdAt: '2026-01-01T00:00:00Z' },
      { id: 'ec-tp', name: 'Take Payment', active: true, createdAt: '2026-01-01T00:00:00Z' },
    ],
    directions: [
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
        businessDomainName: 'Customer',
      },
      {
        id: 'cap-consent',
        name: 'Customer Consent',
        level: 'L2',
        parentId: 'cap-cim',
        businessDomainId: 'bd-c',
        businessDomainName: 'Customer',
      },
      {
        id: 'cap-fraud',
        name: 'Customer Fraud Prevention',
        level: 'L2',
        parentId: 'cap-cim',
        businessDomainId: 'bd-c',
        businessDomainName: 'Customer',
      },
    ],
  });
}

function renderForm(overrides: Partial<Parameters<typeof CaptureDirectionForm>[0]> = {}) {
  return renderWithProviders(
    <CaptureDirectionForm
      enterpriseCapabilityId={toEnterpriseCapabilityId('ec-crm')}
      onCaptured={vi.fn()}
      onCancel={vi.fn()}
      {...overrides}
    />,
    { withRouter: false },
  );
}

async function openDropdown(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByTestId('source-search-input'));
}

async function selectCandidate(user: ReturnType<typeof userEvent.setup>, name: string) {
  await openDropdown(user);
  await user.click(await screen.findByText(name));
}

describe('CaptureDirectionForm — source picker', () => {
  it('shows candidates immediately without searching', async () => {
    seed();
    const user = userEvent.setup();
    renderForm();
    await openDropdown(user);
    expect(await screen.findByText('Customer Identity Mgmt')).toBeInTheDocument();
    expect(screen.getByText('Customer Consent')).toBeInTheDocument();
  });

  it('marks a candidate already sourced elsewhere as ineligible — clicking it does not add it (R1)', async () => {
    seed();
    const user = userEvent.setup();
    renderForm();
    await openDropdown(user);
    await screen.findByText('Customer Fraud Prevention');
    await user.click(screen.getByText('Customer Fraud Prevention'));
    expect(screen.getByTestId('selected-count')).toHaveTextContent('0');
  });

  it('adds an eligible candidate and reflects the count', async () => {
    seed();
    const user = userEvent.setup();
    renderForm();
    await selectCandidate(user, 'Customer Identity Mgmt');
    await waitFor(() => expect(screen.getByTestId('selected-count')).toHaveTextContent('1'));
  });

  it('shows a draft cardinality hint explaining proposed needs 2 sources for consolidate (R8)', async () => {
    seed();
    const user = userEvent.setup();
    renderForm();
    await selectCandidate(user, 'Customer Identity Mgmt');
    await waitFor(() => expect(screen.getByTestId('draft-cardinality-hint')).toHaveTextContent(/2 sources/i));
  });

  it('filters candidates by search term', async () => {
    seed();
    const user = userEvent.setup();
    renderForm();
    await waitFor(() => expect(screen.getByTestId('source-search-input')).not.toBeDisabled());
    await user.type(screen.getByTestId('source-search-input'), 'Consent');
    expect(await screen.findByText('Customer Consent')).toBeInTheDocument();
    expect(screen.queryByText('Customer Identity Mgmt')).toBeNull();
  });

  it('previews carve-outs for the selected source set (R2)', async () => {
    seed();
    const user = userEvent.setup();
    renderForm();
    await selectCandidate(user, 'Customer Identity Mgmt');
    await waitFor(() => expect(screen.getByTestId('composition-preview')).toHaveTextContent(/Customer Consent/));
    expect(screen.getByTestId('composition-preview')).toHaveTextContent(/Take Payment/);
  });

  it('captures a draft direction with the selected source and calls onCaptured', async () => {
    seed();
    const user = userEvent.setup();
    const onCaptured = vi.fn();
    renderForm({ onCaptured });
    await selectCandidate(user, 'Customer Identity Mgmt');
    await waitFor(() => expect(screen.getByTestId('selected-count')).toHaveTextContent('1'));

    fireEvent.click(screen.getByTestId('capture-submit'));

    await waitFor(() => expect(onCaptured).toHaveBeenCalled());
    const result = await directionApi.getForEnterpriseCapability(toEnterpriseCapabilityId('ec-crm'));
    expect(result.direction?.status).toBe('draft');
    expect(result.direction?.sourceCapabilities.map((s) => s.id)).toEqual(['cap-cim']);
  });

  it('disables capture when no source is selected', async () => {
    seed();
    renderForm();
    await waitFor(() => expect(screen.getByTestId('capture-submit')).toBeDisabled());
  });
});
