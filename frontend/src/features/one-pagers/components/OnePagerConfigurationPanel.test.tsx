import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { SubjectAttribute, SubjectAttributeSchema } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import type { BuiltInField, CustomField, OnePagerConfiguration } from '../types';
import { OnePagerConfigurationPanel } from './OnePagerConfigurationPanel';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getConfiguration: vi.fn(),
    changeFieldRequirement: vi.fn(),
    changeBuiltInFieldRequirement: vi.fn(),
    getImpactPreview: vi.fn(),
  },
}));

vi.mock('../../../api/metadata', () => ({
  subjectAttributeSchemaApi: {
    getSchema: vi.fn(),
    defineAttribute: vi.fn(),
    renameAttribute: vi.fn(),
    retireAttribute: vi.fn(),
    reactivateAttribute: vi.fn(),
    addOption: vi.fn(),
    retireOption: vi.fn(),
    setBounds: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import { subjectAttributeSchemaApi } from '../../../api/metadata';
import { onePagersApi } from '../api/onePagersApi';

const contractLinkField: CustomField = {
  id: 'field-1',
  name: 'Contract link',
  type: 'link',
  required: false,
  helpText: '',
  active: true,
  included: true,
  _links: {
    'x-set-requirement': {
      href: '/api/v1/one-pagers/configurations/application/custom-fields/field-1/requirement',
      method: 'PUT',
    },
  },
};

const contractLinkAttribute: SubjectAttribute = {
  id: 'field-1',
  name: 'Contract link',
  type: 'link',
  helpText: '',
  active: true,
  _links: {
    'x-rename': { href: '/api/v1/meta-model/subject-types/application/attributes/field-1', method: 'PUT' },
    'x-retire': { href: '/api/v1/meta-model/subject-types/application/attributes/field-1/retire', method: 'POST' },
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
      'x-attribute-schema': { href: '/api/v1/meta-model/subject-types/application/attributes', method: 'GET' },
      'x-impact-preview': { href: '/api/v1/one-pagers/configurations/application/impact-preview', method: 'GET' },
    },
    ...overrides,
  };
}

function buildSchema(overrides: Partial<SubjectAttributeSchema> = {}): SubjectAttributeSchema {
  return {
    id: 'schema-1',
    subjectType: 'application',
    attributes: [contractLinkAttribute],
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'steward@example.com',
    _links: {
      self: { href: '/api/v1/meta-model/subject-types/application/attributes', method: 'GET' },
      'x-define': { href: '/api/v1/meta-model/subject-types/application/attributes', method: 'POST' },
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

function numberAttributeWithBounds(id: string, name: string, min: number, max: number): SubjectAttribute {
  return {
    id,
    name,
    type: 'number',
    helpText: '',
    active: true,
    min,
    max,
    _links: {
      'x-set-bounds': { href: `/api/v1/meta-model/subject-types/application/attributes/${id}/bounds`, method: 'PUT' },
    },
  };
}

function numberField(id: string, name: string): CustomField {
  return { id, name, type: 'number', required: false, helpText: '', active: true, included: true };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(buildSchema());
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('OnePagerConfigurationPanel — required field impact preview', () => {
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
    async ({ configuration, updatedConfiguration, checkboxTestId, fieldId, fieldKind, affectedSubjectCount, message, changeApi, expectedField }) => {
      const user = userEvent.setup();
      vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
      vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({ subjectType: 'application', fieldId, affectedSubjectCount });
      vi.mocked(changeApi).mockResolvedValue(updatedConfiguration);

      renderPanel();

      await user.click(await screen.findByTestId(checkboxTestId));

      expect(await screen.findByText(message)).toBeInTheDocument();
      expect(onePagersApi.getImpactPreview).toHaveBeenCalledWith(expect.anything(), fieldId, fieldKind);
      expect(changeApi).not.toHaveBeenCalled();

      await user.click(screen.getByTestId('one-pager-impact-preview-confirm'));

      await waitFor(() => expect(changeApi).toHaveBeenCalledWith(expectedField, { required: true, version: 1 }));
      await waitFor(() => expect(screen.queryByTestId('one-pager-impact-preview-dialog')).not.toBeInTheDocument());
    },
  );

  it('fires no mutation when the confirmation dialog is cancelled', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration());
    vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({ subjectType: 'application', fieldId: 'field-1', affectedSubjectCount: 37 });

    renderPanel();

    await user.click(await screen.findByTestId('one-pager-required-field-1'));
    await screen.findByTestId('one-pager-impact-preview-dialog');
    await user.click(screen.getByRole('button', { name: 'Cancel' }));

    expect(onePagersApi.changeFieldRequirement).not.toHaveBeenCalled();
    expect(screen.queryByTestId('one-pager-impact-preview-dialog')).not.toBeInTheDocument();
  });

  it('mutates directly with no dialog when flipping a field to optional', async () => {
    const user = userEvent.setup();
    const requiredField = { ...contractLinkField, required: true };
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration({ customFields: [requiredField] }));
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(buildConfiguration({ version: 2 }));

    renderPanel();

    const checkbox = await screen.findByTestId('one-pager-required-field-1');
    expect(checkbox).toBeChecked();
    await user.click(checkbox);

    await waitFor(() =>
      expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(requiredField, { required: false, version: 1 }),
    );
    expect(onePagersApi.getImpactPreview).not.toHaveBeenCalled();
  });

  it('mutates directly with no dialog when the x-impact-preview link is absent', async () => {
    const user = userEvent.setup();
    const configuration = buildConfiguration({
      _links: {
        self: { href: '/api/v1/one-pagers/configurations/application', method: 'GET' },
        'x-attribute-schema': { href: '/api/v1/meta-model/subject-types/application/attributes', method: 'GET' },
      },
    });
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(configuration);
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(buildConfiguration({ version: 2 }));

    renderPanel();

    await user.click(await screen.findByTestId('one-pager-required-field-1'));

    await waitFor(() =>
      expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(contractLinkField, { required: true, version: 1 }),
    );
    expect(onePagersApi.getImpactPreview).not.toHaveBeenCalled();
  });
});

describe('OnePagerConfigurationPanel — defining custom fields in MetaModel', () => {
  it('shows the population count and defines the attribute on confirm when Required is checked, then requires it on the one-pager', async () => {
    const user = userEvent.setup();
    const configuration = buildConfiguration({ customFields: [], displayOrder: [] });
    const definedField: CustomField = {
      ...contractLinkField,
      id: 'attr-new',
      name: 'Data classification',
      type: 'text',
      _links: {
        'x-set-requirement': {
          href: '/api/v1/one-pagers/configurations/application/custom-fields/attr-new/requirement',
          method: 'PUT',
        },
      },
    };
    const refreshed = buildConfiguration({ version: 2, customFields: [definedField], displayOrder: [{ kind: 'custom', id: 'attr-new' }] });
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValueOnce(configuration).mockResolvedValue(refreshed);
    vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({ subjectType: 'application', affectedSubjectCount: 120 });
    const schema = buildSchema({ attributes: [] });
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(schema);
    vi.mocked(subjectAttributeSchemaApi.defineAttribute).mockResolvedValue(
      buildSchema({ version: 2, attributes: [{ id: 'attr-new', name: 'Data classification', type: 'text', helpText: '', active: true }] }),
    );
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(buildConfiguration({ version: 3 }));

    renderPanel();

    await user.type(await screen.findByTestId('one-pager-new-field-name'), 'Data classification');
    await user.click(screen.getByTestId('one-pager-new-field-required'));
    await user.click(screen.getByTestId('one-pager-new-field-submit'));

    expect(await screen.findByText('Making Data classification required will mark 120 Applications incomplete')).toBeInTheDocument();
    expect(subjectAttributeSchemaApi.defineAttribute).not.toHaveBeenCalled();

    await user.click(screen.getByTestId('one-pager-impact-preview-confirm'));

    await waitFor(() =>
      expect(subjectAttributeSchemaApi.defineAttribute).toHaveBeenCalledWith(schema, {
        name: 'Data classification',
        type: 'text',
        helpText: '',
        options: undefined,
        min: undefined,
        max: undefined,
        version: 1,
      }),
    );
    await waitFor(() =>
      expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(definedField, { required: true, version: 2 }),
    );
  });

  it('defines the attribute directly with no dialog when Required is unchecked', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration({ customFields: [], displayOrder: [] }));
    const schema = buildSchema({ attributes: [] });
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(schema);
    vi.mocked(subjectAttributeSchemaApi.defineAttribute).mockResolvedValue(buildSchema({ version: 2 }));

    renderPanel();

    await user.type(await screen.findByTestId('one-pager-new-field-name'), 'Notes');
    await user.click(screen.getByTestId('one-pager-new-field-submit'));

    await waitFor(() =>
      expect(subjectAttributeSchemaApi.defineAttribute).toHaveBeenCalledWith(schema, {
        name: 'Notes',
        type: 'text',
        helpText: '',
        options: undefined,
        min: undefined,
        max: undefined,
        version: 1,
      }),
    );
    expect(onePagersApi.getImpactPreview).not.toHaveBeenCalled();
    expect(onePagersApi.changeFieldRequirement).not.toHaveBeenCalled();
  });

  it('hides the define form when the schema offers no x-define affordance', async () => {
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration());
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(
      buildSchema({ _links: { self: { href: '/api/v1/meta-model/subject-types/application/attributes', method: 'GET' } } }),
    );

    renderPanel();

    expect(await screen.findByTestId('one-pager-field-list')).toBeInTheDocument();
    expect(screen.queryByTestId('one-pager-add-field-form')).not.toBeInTheDocument();
  });
});

describe('OnePagerConfigurationPanel — schema affordances on custom fields', () => {
  it('retires an attribute through its MetaModel link with the schema version', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration());
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(buildSchema({ version: 5 }));
    vi.mocked(subjectAttributeSchemaApi.retireAttribute).mockResolvedValue(buildSchema({ version: 6 }));

    renderPanel();

    await user.click(await screen.findByTestId('one-pager-retire-field-1'));

    await waitFor(() =>
      expect(subjectAttributeSchemaApi.retireAttribute).toHaveBeenCalledWith(contractLinkAttribute, { version: 5 }),
    );
  });

  it('offers no schema controls when the attribute carries no links', async () => {
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration());
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(
      buildSchema({ attributes: [{ ...contractLinkAttribute, _links: {} }] }),
    );

    renderPanel();

    expect(await screen.findByTestId('one-pager-required-field-1')).toBeInTheDocument();
    expect(screen.queryByTestId('one-pager-retire-field-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('one-pager-rename-field-1')).not.toBeInTheDocument();
  });

  it('lists retired attributes from the schema and reactivates them', async () => {
    const user = userEvent.setup();
    const retired: SubjectAttribute = {
      id: 'old',
      name: 'Old field',
      type: 'text',
      helpText: '',
      active: false,
      _links: { 'x-reactivate': { href: '/api/v1/meta-model/subject-types/application/attributes/old/reactivate', method: 'POST' } },
    };
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration());
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(buildSchema({ version: 2, attributes: [contractLinkAttribute, retired] }));
    vi.mocked(subjectAttributeSchemaApi.reactivateAttribute).mockResolvedValue(buildSchema({ version: 3 }));

    renderPanel();

    await user.click(await screen.findByTestId('one-pager-reactivate-old'));

    await waitFor(() => expect(subjectAttributeSchemaApi.reactivateAttribute).toHaveBeenCalledWith(retired, { version: 2 }));
  });

  it('renames an attribute through the dialog with the schema version', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration());
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(buildSchema({ version: 4 }));
    vi.mocked(subjectAttributeSchemaApi.renameAttribute).mockResolvedValue(buildSchema({ version: 5 }));

    renderPanel();

    await user.click(await screen.findByTestId('one-pager-rename-field-1'));
    const nameInput = await screen.findByTestId('one-pager-rename-name-input');
    await user.clear(nameInput);
    await user.type(nameInput, 'Contract URL');
    await user.click(screen.getByTestId('one-pager-rename-submit'));

    await waitFor(() =>
      expect(subjectAttributeSchemaApi.renameAttribute).toHaveBeenCalledWith(contractLinkAttribute, {
        name: 'Contract URL',
        helpText: '',
        type: 'link',
        version: 4,
      }),
    );
  });
});

describe('OnePagerConfigurationPanel — Number field bounds', () => {
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

  it('defines a number field with its bounds in a single MetaModel request', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration({ customFields: [], displayOrder: [] }));
    const schema = buildSchema({ attributes: [] });
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(schema);
    vi.mocked(subjectAttributeSchemaApi.defineAttribute).mockResolvedValue(buildSchema({ version: 2 }));

    renderPanel();

    await user.type(await screen.findByTestId('one-pager-new-field-name'), 'Maturity score');
    await user.click(screen.getByTestId('one-pager-new-field-type'));
    await user.click(await screen.findByRole('option', { name: 'Number', hidden: true }));
    await user.type(await screen.findByTestId('one-pager-new-field-min'), '0');
    await user.type(screen.getByTestId('one-pager-new-field-max'), '5');
    await user.click(screen.getByTestId('one-pager-new-field-submit'));

    await waitFor(() =>
      expect(subjectAttributeSchemaApi.defineAttribute).toHaveBeenCalledWith(schema, {
        name: 'Maturity score',
        type: 'number',
        helpText: '',
        options: undefined,
        min: 0,
        max: 5,
        version: 1,
      }),
    );
  });

  it.each([
    {
      name: 'tightens the maximum',
      attribute: numberAttributeWithBounds('field-3', 'Maturity score', 0, 5),
      typeIntoMax: '3',
      expectedRequest: { min: 0, max: 3, version: 1 },
    },
    {
      name: 'clears a bound while keeping the other',
      attribute: numberAttributeWithBounds('field-4', 'Headcount', 0, 500),
      typeIntoMax: '',
      expectedRequest: { min: 0, max: undefined, version: 1 },
    },
  ])('edits the bounds of an existing Number field: $name', async ({ attribute, typeIntoMax, expectedRequest }) => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(
      buildConfiguration({ customFields: [numberField(attribute.id, attribute.name)], displayOrder: [{ kind: 'custom', id: attribute.id }] }),
    );
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(buildSchema({ attributes: [attribute] }));
    vi.mocked(subjectAttributeSchemaApi.setBounds).mockResolvedValue(buildSchema({ version: 2 }));

    renderPanel();

    const maxInput = await screen.findByTestId(`one-pager-bounds-max-${attribute.id}`);
    await user.clear(maxInput);
    if (typeIntoMax) await user.type(maxInput, typeIntoMax);
    await user.click(screen.getByTestId(`one-pager-bounds-save-${attribute.id}`));

    await waitFor(() => expect(subjectAttributeSchemaApi.setBounds).toHaveBeenCalledWith(attribute, expectedRequest));
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
