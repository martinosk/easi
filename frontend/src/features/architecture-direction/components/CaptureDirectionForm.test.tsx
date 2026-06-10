import { fireEvent, screen, waitFor } from '@testing-library/react';
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

function search(term: string) {
  fireEvent.change(screen.getByTestId('source-search-input'), { target: { value: term } });
}

describe('CaptureDirectionForm — search-driven source picker', () => {
  it('searches capabilities and lists eligible candidates with an Add control', async () => {
    seed();
    renderForm();
    search('customer');
    await waitFor(() => expect(screen.getByTestId('candidate-cap-cim')).toBeInTheDocument());
    expect(screen.getByTestId('add-candidate-cap-cim')).toBeEnabled();
  });

  it('marks a candidate already sourced elsewhere as ineligible and disables Add (R1)', async () => {
    seed();
    renderForm();
    search('customer');
    await waitFor(() => expect(screen.getByTestId('candidate-cap-fraud')).toBeInTheDocument());
    expect(screen.getByTestId('candidate-cap-fraud')).toHaveTextContent(/Take Payment/);
    expect(screen.getByTestId('add-candidate-cap-fraud')).toBeDisabled();
  });

  it('adds an eligible candidate as a selected chip and reflects the count', async () => {
    seed();
    renderForm();
    search('customer');
    await waitFor(() => expect(screen.getByTestId('add-candidate-cap-cim')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('add-candidate-cap-cim'));
    expect(await screen.findByTestId('selected-chip-cap-cim')).toBeInTheDocument();
    expect(screen.getByTestId('selected-count')).toHaveTextContent('1');
  });

  it('shows a draft cardinality hint explaining proposed needs 2 sources for consolidate (R8)', async () => {
    seed();
    renderForm();
    search('customer');
    await waitFor(() => expect(screen.getByTestId('add-candidate-cap-cim')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('add-candidate-cap-cim'));
    expect(await screen.findByTestId('draft-cardinality-hint')).toHaveTextContent(/2 sources/i);
  });

  it('previews carve-outs for the selected source set (R2)', async () => {
    seed();
    renderForm();
    search('customer');
    await waitFor(() => expect(screen.getByTestId('add-candidate-cap-cim')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('add-candidate-cap-cim'));
    await waitFor(() => expect(screen.getByTestId('composition-preview')).toHaveTextContent(/Customer Consent/));
    expect(screen.getByTestId('composition-preview')).toHaveTextContent(/Take Payment/);
  });

  it('captures a draft direction with the selected source and calls onCaptured', async () => {
    seed();
    const onCaptured = vi.fn();
    renderForm({ onCaptured });
    search('customer');
    await waitFor(() => expect(screen.getByTestId('add-candidate-cap-cim')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('add-candidate-cap-cim'));
    await screen.findByTestId('selected-chip-cap-cim');

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
