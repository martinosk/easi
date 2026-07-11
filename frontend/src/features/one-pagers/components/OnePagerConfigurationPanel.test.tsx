import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { CustomField, OnePagerConfiguration } from '../types';
import { OnePagerConfigurationPanel } from './OnePagerConfigurationPanel';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getConfiguration: vi.fn(),
    changeFieldRequirement: vi.fn(),
    defineCustomField: vi.fn(),
    getImpactPreview: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import { onePagersApi } from '../api/onePagersApi';

const contractLinkField: CustomField = {
  id: 'field-1',
  name: 'Contract link',
  type: 'link',
  required: false,
  helpText: '',
  active: true,
  _links: {
    'x-set-requirement': {
      href: '/api/v1/one-pagers/configurations/application/custom-fields/field-1/requirement',
      method: 'PUT',
    },
  },
};

function buildConfiguration(overrides: Partial<OnePagerConfiguration> = {}): OnePagerConfiguration {
  return {
    id: 'config-1',
    subjectType: 'application',
    builtInFields: [],
    customFields: [contractLinkField],
    displayOrder: [{ kind: 'custom', id: 'field-1' }],
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'admin@example.com',
    _links: {
      self: { href: '/api/v1/one-pagers/configurations/application', method: 'GET' },
      'x-define-custom-field': {
        href: '/api/v1/one-pagers/configurations/application/custom-fields',
        method: 'POST',
      },
      'x-impact-preview': { href: '/api/v1/one-pagers/configurations/application/impact-preview', method: 'GET' },
    },
    ...overrides,
  };
}

function renderPanel() {
  return renderWithProviders(<OnePagerConfigurationPanel subjectType="application" />, { withRouter: false });
}

describe('OnePagerConfigurationPanel — required field impact preview', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('opens a confirmation dialog with the fetched impact count, then confirms the requirement change', async () => {
    const user = userEvent.setup();
    const configuration = buildConfiguration();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
    vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({
      subjectType: 'application',
      fieldId: 'field-1',
      affectedSubjectCount: 37,
    });
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(buildConfiguration({ version: 2 }));

    renderPanel();

    const checkbox = await screen.findByTestId('one-pager-required-field-1');
    await user.click(checkbox);

    expect(
      await screen.findByText('Making Contract link required will mark 37 Applications incomplete'),
    ).toBeInTheDocument();
    expect(onePagersApi.changeFieldRequirement).not.toHaveBeenCalled();

    await user.click(screen.getByTestId('one-pager-impact-preview-confirm'));

    await waitFor(() =>
      expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(contractLinkField, {
        required: true,
        version: 1,
      }),
    );
    await waitFor(() => expect(screen.queryByTestId('one-pager-impact-preview-dialog')).not.toBeInTheDocument());
  });

  it('fires no mutation when the confirmation dialog is cancelled', async () => {
    const user = userEvent.setup();
    const configuration = buildConfiguration();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
    vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({
      subjectType: 'application',
      fieldId: 'field-1',
      affectedSubjectCount: 37,
    });

    renderPanel();

    const checkbox = await screen.findByTestId('one-pager-required-field-1');
    await user.click(checkbox);

    await screen.findByTestId('one-pager-impact-preview-dialog');
    await user.click(screen.getByRole('button', { name: 'Cancel' }));

    expect(onePagersApi.changeFieldRequirement).not.toHaveBeenCalled();
    expect(screen.queryByTestId('one-pager-impact-preview-dialog')).not.toBeInTheDocument();
  });

  it('mutates directly with no dialog when flipping a field to optional', async () => {
    const user = userEvent.setup();
    const requiredField = { ...contractLinkField, required: true };
    const configuration = buildConfiguration({ customFields: [requiredField] });
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(buildConfiguration({ version: 2 }));

    renderPanel();

    const checkbox = await screen.findByTestId('one-pager-required-field-1');
    expect(checkbox).toBeChecked();
    await user.click(checkbox);

    await waitFor(() =>
      expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(requiredField, {
        required: false,
        version: 1,
      }),
    );
    expect(onePagersApi.getImpactPreview).not.toHaveBeenCalled();
    expect(screen.queryByTestId('one-pager-impact-preview-dialog')).not.toBeInTheDocument();
  });

  it('mutates directly with no dialog when the x-impact-preview link is absent', async () => {
    const user = userEvent.setup();
    const configuration = buildConfiguration({
      _links: {
        self: { href: '/api/v1/one-pagers/configurations/application', method: 'GET' },
        'x-define-custom-field': {
          href: '/api/v1/one-pagers/configurations/application/custom-fields',
          method: 'POST',
        },
      },
    });
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(buildConfiguration({ version: 2 }));

    renderPanel();

    const checkbox = await screen.findByTestId('one-pager-required-field-1');
    await user.click(checkbox);

    await waitFor(() =>
      expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(contractLinkField, {
        required: true,
        version: 1,
      }),
    );
    expect(onePagersApi.getImpactPreview).not.toHaveBeenCalled();
    expect(screen.queryByTestId('one-pager-impact-preview-dialog')).not.toBeInTheDocument();
  });

  it('shows the population count and submits the new field on confirm when Required is checked', async () => {
    const user = userEvent.setup();
    const configuration = buildConfiguration({ customFields: [], displayOrder: [] });
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
    vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({
      subjectType: 'application',
      affectedSubjectCount: 120,
    });
    vi.mocked(onePagersApi.defineCustomField).mockResolvedValue(buildConfiguration({ version: 2 }));

    renderPanel();

    await user.type(await screen.findByTestId('one-pager-new-field-name'), 'Data classification');
    await user.click(screen.getByTestId('one-pager-new-field-required'));
    await user.click(screen.getByTestId('one-pager-new-field-submit'));

    expect(
      await screen.findByText('Making Data classification required will mark 120 Applications incomplete'),
    ).toBeInTheDocument();
    expect(onePagersApi.defineCustomField).not.toHaveBeenCalled();

    await user.click(screen.getByTestId('one-pager-impact-preview-confirm'));

    await waitFor(() =>
      expect(onePagersApi.defineCustomField).toHaveBeenCalledWith(configuration, {
        name: 'Data classification',
        fieldType: 'text',
        required: true,
        helpText: '',
        options: undefined,
        version: 1,
      }),
    );
  });

  it('submits the new field directly with no dialog when Required is unchecked', async () => {
    const user = userEvent.setup();
    const configuration = buildConfiguration({ customFields: [], displayOrder: [] });
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
    vi.mocked(onePagersApi.defineCustomField).mockResolvedValue(buildConfiguration({ version: 2 }));

    renderPanel();

    await user.type(await screen.findByTestId('one-pager-new-field-name'), 'Notes');
    await user.click(screen.getByTestId('one-pager-new-field-submit'));

    await waitFor(() =>
      expect(onePagersApi.defineCustomField).toHaveBeenCalledWith(configuration, {
        name: 'Notes',
        fieldType: 'text',
        required: false,
        helpText: '',
        options: undefined,
        version: 1,
      }),
    );
    expect(onePagersApi.getImpactPreview).not.toHaveBeenCalled();
    expect(screen.queryByTestId('one-pager-impact-preview-dialog')).not.toBeInTheDocument();
  });
});
