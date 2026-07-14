import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { CustomField, OnePagerFacts } from '../types';
import { OnePagerFactsForm } from './OnePagerFactsForm';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
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

const secondTextField: CustomField = {
  id: 'f-text-2',
  name: 'Contract notes',
  type: 'text',
  required: false,
  helpText: '',
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

const boundedNumberField: CustomField = {
  id: 'f-number',
  name: 'Maturity score',
  type: 'number',
  required: false,
  helpText: '',
  active: true,
  min: 0,
  max: 5,
};

const linkField: CustomField = {
  id: 'f-link',
  name: 'Contract link',
  type: 'link',
  required: false,
  helpText: '',
  active: true,
};

const contactPersonField: CustomField = {
  id: 'f-contact',
  name: 'Primary contact',
  type: 'contact-person',
  required: false,
  helpText: '',
  active: true,
};

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

function renderForm(fields: CustomField[], facts: OnePagerFacts) {
  return renderWithProviders(<OnePagerFactsForm fields={fields} facts={facts} />, { withRouter: false });
}

describe('OnePagerFactsForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders an input per field and highlights an empty required field', async () => {
    renderForm([textField, selectionField], buildFacts());

    expect(await screen.findByLabelText('Business summary')).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: 'Tier' })).toBeInTheDocument();
    expect(screen.getByText('Required')).toBeInTheDocument();
  });

  it('flags a retired selection option until changed', async () => {
    renderForm([selectionField], buildFacts({ values: [recordedRetiredSelection()] }));

    expect(await screen.findByRole('textbox', { name: 'Tier' })).toHaveValue('Tier 2 (retired)');
  });

  it('disables save while the form is pristine', async () => {
    renderForm([textField], buildFacts());

    expect(await screen.findByRole('button', { name: 'Save one-pager' })).toBeDisabled();
  });

  it('submits only dirty fields on save', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.recordFieldValue).mockResolvedValue(buildFacts());
    renderForm([textField, selectionField], buildFacts());

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
    vi.mocked(onePagersApi.recordFieldValue).mockResolvedValue(buildFacts());
    renderForm([textField, secondTextField], buildFacts());

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
    vi.mocked(onePagersApi.clearFieldValue).mockResolvedValue(buildFacts());
    renderForm([textField], buildFacts({ values: [recorded] }));

    const input = await screen.findByLabelText('Business summary');
    await user.clear(input);
    await user.click(screen.getByRole('button', { name: 'Save one-pager' }));

    await waitFor(() => expect(onePagersApi.clearFieldValue).toHaveBeenCalledTimes(1));
    expect(onePagersApi.clearFieldValue).toHaveBeenCalledWith(expect.objectContaining({ fieldId: 'f-text' }));
    expect(onePagersApi.recordFieldValue).not.toHaveBeenCalled();
  });

  it('clamps a number input to the field bounds on blur, using them as soft hints', async () => {
    const user = userEvent.setup();
    renderForm([boundedNumberField], buildFacts());

    const input = await screen.findByLabelText('Maturity score');
    await user.type(input, '9');
    await user.tab();

    expect(input).toHaveValue('5');
  });

  it('renders a link field with empty label and url inputs, not crashing on first paint', async () => {
    renderForm([linkField], buildFacts());

    expect(await screen.findByLabelText('Contract link label')).toHaveValue('');
    expect(screen.getByLabelText('Contract link URL')).toHaveValue('');
  });

  it('renders a contact-person field with empty name, email, and company inputs', async () => {
    renderForm([contactPersonField], buildFacts());

    expect(await screen.findByLabelText('Primary contact name')).toHaveValue('');
    expect(screen.getByLabelText('Primary contact email')).toHaveValue('');
    expect(screen.getByLabelText('Primary contact company')).toHaveValue('');
  });
});
