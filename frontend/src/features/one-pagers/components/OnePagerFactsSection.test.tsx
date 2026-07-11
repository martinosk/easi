import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { CustomField, OnePagerConfiguration, OnePagerFacts } from '../types';
import { OnePagerFactsSection } from './OnePagerFactsSection';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getConfiguration: vi.fn(),
    getFacts: vi.fn(),
    recordFieldValue: vi.fn(),
    clearFieldValue: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import { onePagersApi } from '../api/onePagersApi';

const textField: CustomField = {
  id: 'f-text',
  name: 'Business summary',
  type: 'text',
  required: true,
  helpText: 'What stakeholders should know',
  active: true,
};

const selectionField: CustomField = {
  id: 'f-select',
  name: 'Tier',
  type: 'selection',
  required: false,
  helpText: '',
  active: true,
  options: [
    { id: 'opt-1', label: 'Tier 1', active: true },
    { id: 'opt-2', label: 'Tier 2', active: false },
  ],
};

const retiredField: CustomField = {
  id: 'f-retired',
  name: 'Legacy field',
  type: 'text',
  required: false,
  helpText: '',
  active: false,
};

function buildConfiguration(customFields: CustomField[]): OnePagerConfiguration {
  return {
    id: 'config-1',
    subjectType: 'vendor',
    builtInFields: [],
    customFields,
    displayOrder: customFields.map((field) => ({ kind: 'custom', id: field.id })),
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'admin@example.com',
    _links: { self: { href: '/api/v1/one-pagers/configurations/vendor', method: 'GET' } },
  };
}

function buildFacts(overrides: Partial<OnePagerFacts> = {}): OnePagerFacts {
  return {
    subjectType: 'vendor',
    subjectId: 'vendor-1',
    values: [],
    _links: {
      self: { href: '/api/v1/one-pagers/vendor/vendor-1/facts', method: 'GET' },
      'x-record': { href: '/api/v1/one-pagers/vendor/vendor-1/facts', method: 'PUT' },
    },
    ...overrides,
  };
}

function readOnlyFacts(overrides: Partial<OnePagerFacts> = {}): OnePagerFacts {
  return buildFacts({
    _links: { self: { href: '/api/v1/one-pagers/vendor/vendor-1/facts', method: 'GET' } },
    ...overrides,
  });
}

function recordedTextValue() {
  return {
    fieldId: 'f-text',
    value: { type: 'text', version: 1, value: 'hello' },
    displayText: 'hello',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'ea@example.com',
    _links: {
      'x-record': { href: '/api/v1/one-pagers/vendor/vendor-1/facts/f-text', method: 'PUT' as const },
      'x-clear': { href: '/api/v1/one-pagers/vendor/vendor-1/facts/f-text', method: 'DELETE' as const },
    },
  };
}

function recordedRetiredSelection() {
  return {
    fieldId: 'f-select',
    value: { type: 'selection', version: 1, value: { optionId: 'opt-2' } },
    displayText: 'opt-2',
    retiredOption: true,
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'ea@example.com',
  };
}

function renderSection() {
  return renderWithProviders(<OnePagerFactsSection subjectType="vendor" subjectId="vendor-1" />, {
    withRouter: false,
  });
}

describe('OnePagerFactsSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders nothing when the configuration has no active custom fields', async () => {
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([retiredField]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts());

    renderSection();

    await waitFor(() => expect(onePagersApi.getFacts).toHaveBeenCalled());
    expect(screen.queryByTestId('one-pager-facts-section')).not.toBeInTheDocument();
  });

  it('renders inputs for active fields and never renders retired fields', async () => {
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(
      buildConfiguration([textField, selectionField, retiredField]),
    );
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts());

    renderSection();

    expect(await screen.findByLabelText('Business summary')).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: 'Tier' })).toBeInTheDocument();
    expect(screen.queryByText('Legacy field')).not.toBeInTheDocument();
  });

  it('highlights empty required fields without blocking anything', async () => {
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([textField]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts());

    renderSection();

    expect(await screen.findByText('Required')).toBeInTheDocument();
  });

  it('renders values read-only when the facts carry no x-record link', async () => {
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([textField, selectionField]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(
      readOnlyFacts({ values: [recordedTextValue(), recordedRetiredSelection()] }),
    );

    renderSection();

    expect(await screen.findByText('hello')).toBeInTheDocument();
    expect(screen.getByText('Tier 2 (retired)')).toBeInTheDocument();
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Save one-pager' })).not.toBeInTheDocument();
  });

  it('flags a retired selection option in the edit form until changed', async () => {
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([selectionField]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts({ values: [recordedRetiredSelection()] }));

    renderSection();

    expect(await screen.findByRole('textbox', { name: 'Tier' })).toHaveValue('Tier 2 (retired)');
  });

  it('submits only dirty fields on save', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([textField, selectionField]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts());
    vi.mocked(onePagersApi.recordFieldValue).mockResolvedValue(buildFacts());

    renderSection();

    const input = await screen.findByLabelText('Business summary');
    await user.type(input, 'Runs on shared Kubernetes cluster');
    await user.click(screen.getByRole('button', { name: 'Save one-pager' }));

    await waitFor(() => expect(onePagersApi.recordFieldValue).toHaveBeenCalledTimes(1));
    expect(onePagersApi.recordFieldValue).toHaveBeenCalledWith(expect.objectContaining({ subjectId: 'vendor-1' }), 'f-text', {
      value: { type: 'text', version: 1, value: 'Runs on shared Kubernetes cluster' },
    });
    expect(onePagersApi.clearFieldValue).not.toHaveBeenCalled();
  });

  it('records every dirty field when several fields are edited before one save', async () => {
    const user = userEvent.setup();
    const secondTextField: CustomField = {
      id: 'f-text-2',
      name: 'Contract notes',
      type: 'text',
      required: false,
      helpText: '',
      active: true,
    };
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(
      buildConfiguration([textField, secondTextField]),
    );
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts());
    vi.mocked(onePagersApi.recordFieldValue).mockResolvedValue(buildFacts());

    renderSection();

    await user.type(await screen.findByLabelText('Business summary'), 'hello');
    await user.type(screen.getByLabelText('Contract notes'), 'world');
    await user.click(screen.getByRole('button', { name: 'Save one-pager' }));

    await waitFor(() => expect(onePagersApi.recordFieldValue).toHaveBeenCalledTimes(2));
    expect(onePagersApi.recordFieldValue).toHaveBeenCalledWith(expect.anything(), 'f-text', {
      value: { type: 'text', version: 1, value: 'hello' },
    });
    expect(onePagersApi.recordFieldValue).toHaveBeenCalledWith(expect.anything(), 'f-text-2', {
      value: { type: 'text', version: 1, value: 'world' },
    });
  });

  it('clears a field whose value was removed', async () => {
    const user = userEvent.setup();
    const recorded = recordedTextValue();
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([textField]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts({ values: [recorded] }));
    vi.mocked(onePagersApi.clearFieldValue).mockResolvedValue(buildFacts());

    renderSection();

    const input = await screen.findByLabelText('Business summary');
    await user.clear(input);
    await user.click(screen.getByRole('button', { name: 'Save one-pager' }));

    await waitFor(() => expect(onePagersApi.clearFieldValue).toHaveBeenCalledTimes(1));
    expect(onePagersApi.clearFieldValue).toHaveBeenCalledWith(expect.objectContaining({ fieldId: 'f-text' }));
    expect(onePagersApi.recordFieldValue).not.toHaveBeenCalled();
  });

  it('disables save while the form is pristine', async () => {
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([textField]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts());

    renderSection();

    expect(await screen.findByRole('button', { name: 'Save one-pager' })).toBeDisabled();
  });
});
