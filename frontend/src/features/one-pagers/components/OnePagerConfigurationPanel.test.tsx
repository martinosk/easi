import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { BuiltInField, CustomField, OnePagerConfiguration } from '../types';
import { OnePagerConfigurationPanel } from './OnePagerConfigurationPanel';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getConfiguration: vi.fn(),
    changeFieldRequirement: vi.fn(),
    changeBuiltInFieldRequirement: vi.fn(),
    defineCustomField: vi.fn(),
    getImpactPreview: vi.fn(),
    setNumberFieldBounds: vi.fn(),
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

const expertsBuiltIn: BuiltInField = {
  id: 'experts',
  label: 'Experts',
  included: true,
  required: false,
  _links: {
    'x-exclude': {
      href: '/api/v1/one-pagers/configurations/application/built-in-fields/experts/exclude',
      method: 'POST',
    },
    'x-set-requirement': {
      href: '/api/v1/one-pagers/configurations/application/built-in-fields/experts/requirement',
      method: 'PUT',
    },
  },
};

function buildConfigurationWithBuiltIn(overrides: Partial<OnePagerConfiguration> = {}): OnePagerConfiguration {
  return buildConfiguration({
    builtInFields: [expertsBuiltIn],
    displayOrder: [{ kind: 'builtIn', id: 'experts' }],
    ...overrides,
  });
}

function renderPanel() {
  return renderWithProviders(<OnePagerConfigurationPanel subjectType="application" />, { withRouter: false });
}

function numberFieldWithBounds(id: string, name: string, min: number, max: number): CustomField {
  return {
    id,
    name,
    type: 'number',
    required: false,
    helpText: '',
    active: true,
    min,
    max,
    _links: {
      'x-set-bounds': {
        href: `/api/v1/one-pagers/configurations/application/custom-fields/${id}/bounds`,
        method: 'PUT',
      },
    },
  };
}

describe('OnePagerConfigurationPanel — required field impact preview', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it.each([
    {
      label: 'a custom field',
      configuration: buildConfiguration(),
      updatedConfiguration: buildConfiguration({ version: 2 }),
      checkboxTestId: 'one-pager-required-field-1',
      fieldId: 'field-1',
      fieldKind: 'custom' as const,
      affectedSubjectCount: 37,
      message: 'Making Contract link required will mark 37 Applications incomplete',
      changeApi: onePagersApi.changeFieldRequirement,
      expectedField: contractLinkField as unknown,
    },
    {
      label: 'a built-in field',
      configuration: buildConfigurationWithBuiltIn(),
      updatedConfiguration: buildConfigurationWithBuiltIn({ version: 2 }),
      checkboxTestId: 'one-pager-builtin-required-experts',
      fieldId: 'experts',
      fieldKind: 'builtIn' as const,
      affectedSubjectCount: 40,
      message: 'Making Experts required will mark 40 Applications incomplete',
      changeApi: onePagersApi.changeBuiltInFieldRequirement,
      expectedField: expertsBuiltIn as unknown,
    },
  ])(
    'opens a confirmation dialog with the fetched impact count for $label, then confirms the requirement change',
    async ({
      configuration,
      updatedConfiguration,
      checkboxTestId,
      fieldId,
      fieldKind,
      affectedSubjectCount,
      message,
      changeApi,
      expectedField,
    }) => {
      const user = userEvent.setup();
      vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
      vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({
        subjectType: 'application',
        fieldId,
        affectedSubjectCount,
      });
      vi.mocked(changeApi).mockResolvedValue(updatedConfiguration);

      renderPanel();

      const checkbox = await screen.findByTestId(checkboxTestId);
      await user.click(checkbox);

      expect(await screen.findByText(message)).toBeInTheDocument();
      expect(onePagersApi.getImpactPreview).toHaveBeenCalledWith(expect.anything(), fieldId, fieldKind);
      expect(changeApi).not.toHaveBeenCalled();

      await user.click(screen.getByTestId('one-pager-impact-preview-confirm'));

      await waitFor(() =>
        expect(changeApi).toHaveBeenCalledWith(expectedField, { required: true, version: 1 }),
      );
      await waitFor(() => expect(screen.queryByTestId('one-pager-impact-preview-dialog')).not.toBeInTheDocument());
    },
  );

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

describe('OnePagerConfigurationPanel — Number field bounds', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('shows minimum and maximum inputs only when the field type is Number', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration({ customFields: [], displayOrder: [] }));

    renderPanel();

    expect(screen.queryByTestId('one-pager-new-field-min')).not.toBeInTheDocument();

    await user.click(await screen.findByTestId('one-pager-new-field-type'));
    await user.click(await screen.findByRole('option', { name: 'Number', hidden: true }));

    expect(await screen.findByTestId('one-pager-new-field-min')).toBeInTheDocument();
    expect(screen.getByTestId('one-pager-new-field-max')).toBeInTheDocument();
  });

  it('defines the field then composes a set-bounds call using the returned field and version', async () => {
    const user = userEvent.setup();
    const configuration = buildConfiguration({ customFields: [], displayOrder: [] });
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
    const maturityField = {
      id: 'field-2',
      name: 'Maturity score',
      type: 'number' as const,
      required: false,
      helpText: '',
      active: true,
      _links: {
        'x-set-bounds': {
          href: '/api/v1/one-pagers/configurations/application/custom-fields/field-2/bounds',
          method: 'PUT' as const,
        },
      },
    };
    vi.mocked(onePagersApi.defineCustomField).mockResolvedValue(
      buildConfiguration({ version: 2, customFields: [maturityField], displayOrder: [{ kind: 'custom', id: 'field-2' }] }),
    );
    vi.mocked(onePagersApi.setNumberFieldBounds).mockResolvedValue(buildConfiguration({ version: 3 }));

    renderPanel();

    await user.type(await screen.findByTestId('one-pager-new-field-name'), 'Maturity score');
    await user.click(screen.getByTestId('one-pager-new-field-type'));
    await user.click(await screen.findByRole('option', { name: 'Number', hidden: true }));
    await user.type(await screen.findByTestId('one-pager-new-field-min'), '0');
    await user.type(screen.getByTestId('one-pager-new-field-max'), '5');
    await user.click(screen.getByTestId('one-pager-new-field-submit'));

    await waitFor(() => expect(onePagersApi.defineCustomField).toHaveBeenCalled());
    await waitFor(() =>
      expect(onePagersApi.setNumberFieldBounds).toHaveBeenCalledWith(maturityField, { min: 0, max: 5, version: 2 }),
    );
  });

  it.each([
    {
      name: 'tightens the maximum',
      field: numberFieldWithBounds('field-3', 'Maturity score', 0, 5),
      typeIntoMax: '3',
      expectedRequest: { min: 0, max: 3, version: 1 },
    },
    {
      name: 'clears a bound while keeping the other',
      field: numberFieldWithBounds('field-4', 'Headcount', 0, 500),
      typeIntoMax: '',
      expectedRequest: { min: 0, max: undefined, version: 1 },
    },
  ])('edits the bounds of an existing Number field: $name', async ({ field, typeIntoMax, expectedRequest }) => {
    const user = userEvent.setup();
    const configuration = buildConfiguration({
      customFields: [field],
      displayOrder: [{ kind: 'custom', id: field.id }],
    });
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
    vi.mocked(onePagersApi.setNumberFieldBounds).mockResolvedValue(buildConfiguration({ version: 2 }));

    renderPanel();

    const maxInput = await screen.findByTestId(`one-pager-bounds-max-${field.id}`);
    await user.clear(maxInput);
    if (typeIntoMax) await user.type(maxInput, typeIntoMax);
    await user.click(screen.getByTestId(`one-pager-bounds-save-${field.id}`));

    await waitFor(() => expect(onePagersApi.setNumberFieldBounds).toHaveBeenCalledWith(field, expectedRequest));
  });

  it('shows a client-side hint and blocks submit when minimum exceeds maximum', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration({ customFields: [], displayOrder: [] }));

    renderPanel();

    await user.type(await screen.findByTestId('one-pager-new-field-name'), 'Maturity score');
    await user.click(screen.getByTestId('one-pager-new-field-type'));
    await user.click(await screen.findByRole('option', { name: 'Number', hidden: true }));
    await user.type(await screen.findByTestId('one-pager-new-field-min'), '10');
    await user.type(screen.getByTestId('one-pager-new-field-max'), '5');

    expect(await screen.findByText('Minimum must not exceed maximum')).toBeInTheDocument();
    expect(screen.getByTestId('one-pager-new-field-submit')).toBeDisabled();
  });
});
