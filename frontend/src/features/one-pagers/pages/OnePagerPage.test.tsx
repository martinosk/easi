import { MantineProvider } from '@mantine/core';
import { QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../../../api/types';
import { createTestQueryClient } from '../../../test/helpers';
import { theme } from '../../../theme/mantine';
import type {
  CustomField,
  FieldValue,
  OnePagerConfiguration,
  OnePagerFacts,
  OnePagerSubjectType,
  OnePagerView,
  OnePagerViewField,
} from '../types';
import { OnePagerPage } from './OnePagerPage';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getOnePager: vi.fn(),
    getConfiguration: vi.fn(),
    getFacts: vi.fn(),
    recordFieldValue: vi.fn(),
    clearFieldValue: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

vi.mock('../../../utils/clipboard', () => ({
  copyToClipboard: vi.fn(),
  generateOnePagerShareUrl: vi.fn(
    (subjectType: string, subjectId: string) => `https://app.example.com/one-pagers/${subjectType}/${subjectId}`,
  ),
}));

vi.mock('../components/SubjectDrawer', () => ({
  SubjectDrawer: ({ opened, subjectType, subjectId }: { opened: boolean; subjectType: string; subjectId: string }) =>
    opened ? <div data-testid="one-pager-subject-drawer">{`${subjectType}:${subjectId}`}</div> : null,
}));

import { onePagersApi } from '../api/onePagersApi';
import { copyToClipboard, generateOnePagerShareUrl } from '../../../utils/clipboard';

function builtInField(
  id: string,
  label: string,
  value: Extract<OnePagerViewField, { kind: 'builtIn' }>['builtIn']['value'],
): OnePagerViewField {
  return { kind: 'builtIn', builtIn: { id, label, value } };
}

function customField(
  fieldId: string,
  name: string,
  custom: Partial<Extract<OnePagerViewField, { kind: 'custom' }>['custom']> = {},
): OnePagerViewField {
  return {
    kind: 'custom',
    custom: {
      fieldId,
      name,
      type: 'text',
      value: null,
      displayText: '',
      ...custom,
    },
  };
}

function buildView(overrides: Partial<OnePagerView> = {}): OnePagerView {
  return {
    subjectType: 'capability',
    subjectId: 'cap-1',
    subjectName: 'Payments',
    fields: [],
    completeness: { requiredCount: 0, filledCount: 0, missingFields: [] },
    _links: {
      self: { href: '/api/v1/one-pagers/capability/cap-1', method: 'GET' },
      'x-subject': { href: '/api/v1/capabilities/cap-1', method: 'GET' },
    },
    ...overrides,
  };
}

function renderPage(path = '/one-pagers/capability/cap-1') {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MantineProvider theme={theme}>
        <MemoryRouter initialEntries={[path]}>
          <Routes>
            <Route path="/one-pagers/:subjectType/:subjectId" element={<OnePagerPage />} />
          </Routes>
        </MemoryRouter>
      </MantineProvider>
    </QueryClientProvider>,
  );
}

function fieldLabels(container: HTMLElement): string[] {
  return Array.from(container.querySelectorAll('label')).map((el) => el.textContent ?? '');
}

function withRecordLink() {
  return {
    self: { href: '/api/v1/one-pagers/capability/cap-1', method: 'GET' as const },
    'x-subject': { href: '/api/v1/capabilities/cap-1', method: 'GET' as const },
    'x-record': { href: '/api/v1/one-pagers/capability/cap-1/facts', method: 'PUT' as const },
  };
}

function customFieldDef(id: string, name: string, overrides: Partial<CustomField> = {}): CustomField {
  return { id, name, type: 'text', required: false, helpText: '', active: true, ...overrides };
}

function buildConfiguration(customFields: CustomField[]): OnePagerConfiguration {
  return {
    id: 'config-1',
    subjectType: 'capability',
    builtInFields: [],
    customFields,
    displayOrder: customFields.map((field) => ({ kind: 'custom', id: field.id })),
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'admin@example.com',
    _links: { self: { href: '/api/v1/one-pagers/configurations/capability', method: 'GET' } },
  };
}

function textFieldValue(fieldId: string, value: string): FieldValue {
  return {
    fieldId,
    value: { type: 'text', version: 1, value },
    displayText: value,
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'ea@example.com',
  };
}

function buildFacts(overrides: Partial<OnePagerFacts> = {}): OnePagerFacts {
  return {
    subjectType: 'capability',
    subjectId: 'cap-1',
    values: [],
    _links: {
      self: { href: '/api/v1/one-pagers/capability/cap-1/facts', method: 'GET' },
      'x-record': { href: '/api/v1/one-pagers/capability/cap-1/facts', method: 'PUT' },
    },
    ...overrides,
  };
}

describe('OnePagerPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders the subject header with name and subject type', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(buildView());

    renderPage();

    expect(await screen.findByRole('heading', { name: 'Payments' })).toBeInTheDocument();
    expect(screen.getByText('Capability')).toBeInTheDocument();
  });

  it('renders fields in exactly the server-provided interleaved order', async () => {
    const view = buildView({
      fields: [
        builtInField('description', 'Description', { type: 'text', text: 'Handles all payments' }),
        customField('f1', 'Contract link', {
          type: 'link',
          value: { type: 'link', version: 1, value: { label: 'MSA', url: 'https://example.com' } },
          displayText: 'MSA',
        }),
        builtInField('maturity', 'Maturity', { type: 'maturity', maturity: { value: 85, section: 'Optimizing' } }),
      ],
    });
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(view);

    const { container } = renderPage();

    await waitFor(() => expect(screen.getByText('Handles all payments')).toBeInTheDocument());
    expect(fieldLabels(container)).toEqual(['Description', 'Contract link', 'Maturity']);
  });

  it('renders a link custom field as an anchor with its label and href', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          customField('f1', 'Contract link', {
            type: 'link',
            value: { type: 'link', version: 1, value: { label: 'MSA', url: 'https://example.com' } },
            displayText: 'MSA',
          }),
        ],
      }),
    );

    renderPage();

    const anchor = await screen.findByRole('link', { name: 'MSA' });
    expect(anchor).toHaveAttribute('href', 'https://example.com');
  });

  it('renders a contact-person custom field as a name, email, and company block', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          customField('f2', 'Primary Contact', {
            type: 'contact-person',
            value: {
              type: 'contact-person',
              version: 1,
              value: { name: 'Jane', email: 'jane@example.com', company: 'Acme' },
            },
            displayText: 'Jane',
          }),
        ],
      }),
    );

    renderPage();

    expect(await screen.findByText('Jane (jane@example.com), Acme')).toBeInTheDocument();
  });

  it('renders a selection custom field value flagged as retired', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          customField('f3', 'Tier', {
            type: 'selection',
            value: { type: 'selection', version: 1, value: { optionId: 'opt-2' } },
            displayText: 'Tier 2',
            retiredOption: true,
          }),
        ],
      }),
    );

    renderPage();

    expect(await screen.findByText('Tier 2 (retired)')).toBeInTheDocument();
  });

  it('renders a Number custom field value outside its current bounds flagged', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          customField('f4', 'Maturity score', {
            type: 'number',
            value: { type: 'number', version: 1, value: 5 },
            displayText: '5',
            outOfBounds: true,
          }),
        ],
      }),
    );

    renderPage();

    expect(await screen.findByText('5 (outside range)')).toBeInTheDocument();
  });

  it('renders a built-in date field as the local calendar date, not UTC-shifted', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [builtInField('acquisition-date', 'Acquisition Date', { type: 'date', date: '2023-05-01' })],
      }),
    );

    renderPage();

    const expected = new Date(2023, 4, 1).toLocaleDateString();
    expect(await screen.findByText(expected)).toBeInTheDocument();
  });

  it('renders a custom date field as the local calendar date, not the raw ISO string', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          customField('f5', 'Renewal', {
            type: 'date',
            value: { type: 'date', version: 1, value: '2023-05-01' },
            displayText: '2023-05-01',
          }),
        ],
      }),
    );

    renderPage();

    const expected = new Date(2023, 4, 1).toLocaleDateString();
    expect(await screen.findByText(expected)).toBeInTheDocument();
  });

  it('renders a built-in maturity field with the tenant section name', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          builtInField('maturity', 'Maturity', { type: 'maturity', maturity: { value: 85, section: 'Optimizing' } }),
        ],
      }),
    );

    renderPage();

    expect(await screen.findByText('85')).toBeInTheDocument();
    expect(screen.getByText('Optimizing')).toBeInTheDocument();
  });

  it('renders a built-in maturity field without a section when none is configured', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [builtInField('maturity', 'Maturity', { type: 'maturity', maturity: { value: 40 } })],
      }),
    );

    renderPage();

    expect(await screen.findByText('40')).toBeInTheDocument();
  });

  it('renders a built-in experts field as one line per expert', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          builtInField('experts', 'Experts', {
            type: 'experts',
            experts: [{ name: 'Jane', role: 'CTO', contact: 'jane@x.com' }],
          }),
        ],
      }),
    );

    renderPage();

    expect(await screen.findByText('Jane (CTO), jane@x.com')).toBeInTheDocument();
  });

  it('renders an em-dash for both a null built-in value and a null custom value', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [builtInField('notes', 'Notes', null), customField('f4', 'Renewal', { type: 'date', value: null })],
      }),
    );

    const { container } = renderPage();

    await waitFor(() => expect(fieldLabels(container)).toEqual(['Notes', 'Renewal']));
    expect(screen.getAllByText('—')).toHaveLength(2);
  });

  it('copies the share URL when Share (copy URL) is clicked', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(buildView());

    renderPage();

    await user.click(await screen.findByRole('button', { name: 'Share (copy URL)' }));

    expect(generateOnePagerShareUrl).toHaveBeenCalledWith('capability', 'cap-1');
    expect(copyToClipboard).toHaveBeenCalledWith('https://app.example.com/one-pagers/capability/cap-1');
  });

  it('shows a not-found state when the subject does not exist (404)', async () => {
    vi.mocked(onePagersApi.getOnePager).mockRejectedValue(new ApiError('Not found', 404));

    renderPage();

    expect(await screen.findByText('One-pager not found')).toBeInTheDocument();
  });

  it('shows a not-found state when the subjectType route param is invalid', async () => {
    renderPage('/one-pagers/not-a-real-type/abc123');

    expect(await screen.findByText('One-pager not found')).toBeInTheDocument();
    expect(onePagersApi.getOnePager).not.toHaveBeenCalled();
  });

  it('shows a complete-subject summary and no missing markers when every required field is filled', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          customField('f1', 'Contract link', {
            type: 'link',
            value: { type: 'link', version: 1, value: { label: 'MSA', url: 'https://example.com' } },
            displayText: 'MSA',
          }),
          customField('f2', 'Contact person', {
            type: 'contact-person',
            value: { type: 'contact-person', version: 1, value: { name: 'Jane', email: 'jane@example.com' } },
            displayText: 'Jane',
          }),
        ],
        completeness: { requiredCount: 2, filledCount: 2, missingFields: [] },
      }),
    );

    renderPage();

    expect(await screen.findByTestId('one-pager-completeness-summary')).toHaveTextContent(
      '2 of 2 required fields filled',
    );
    expect(screen.queryByText('missing — required')).not.toBeInTheDocument();
  });

  it('flags a missing required field and shows a partial completeness summary', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          customField('f1', 'Contract link', { type: 'link', value: null }),
          customField('f2', 'Contact person', {
            type: 'contact-person',
            value: { type: 'contact-person', version: 1, value: { name: 'Jane', email: 'jane@example.com' } },
            displayText: 'Jane',
          }),
        ],
        completeness: {
          requiredCount: 2,
          filledCount: 1,
          missingFields: [{ fieldId: 'f1', name: 'Contract link' }],
        },
      }),
    );

    renderPage();

    expect(await screen.findByTestId('one-pager-missing-required-f1')).toHaveTextContent('missing — required');
    expect(screen.getByTestId('one-pager-completeness-summary')).toHaveTextContent('1 of 2 required fields filled');
    expect(screen.queryByTestId('one-pager-missing-required-f2')).not.toBeInTheDocument();
  });

  it('flags a missing required built-in field as missing — required', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [builtInField('experts', 'Experts', null)],
        completeness: {
          requiredCount: 1,
          filledCount: 0,
          missingFields: [{ fieldId: 'experts', name: 'Experts' }],
        },
      }),
    );

    renderPage();

    expect(await screen.findByTestId('one-pager-missing-required-experts')).toHaveTextContent('missing — required');
    expect(screen.getByTestId('one-pager-completeness-summary')).toHaveTextContent('0 of 1 required fields filled');
  });

  it('does not flag an optional field with no value as missing', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [
          customField('f1', 'Contract link', {
            type: 'link',
            value: { type: 'link', version: 1, value: { label: 'MSA', url: 'https://example.com' } },
            displayText: 'MSA',
          }),
          customField('f2', 'Notes', { type: 'text', value: null }),
        ],
        completeness: { requiredCount: 1, filledCount: 1, missingFields: [] },
      }),
    );

    renderPage();

    expect(await screen.findByTestId('one-pager-completeness-summary')).toHaveTextContent(
      '1 of 1 required fields filled',
    );
    expect(screen.queryByTestId('one-pager-missing-required-f2')).not.toBeInTheDocument();
  });

  it('renders no completeness summary when the configuration has no required fields', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        fields: [customField('f1', 'Notes', { type: 'text', value: null })],
        completeness: { requiredCount: 0, filledCount: 0, missingFields: [] },
      }),
    );

    renderPage();

    await waitFor(() => expect(screen.getByText('Payments')).toBeInTheDocument());
    expect(screen.queryByTestId('one-pager-completeness-summary')).not.toBeInTheDocument();
  });

  it('shows an Edit action when the loaded view carries x-record', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(buildView({ _links: withRecordLink() }));

    renderPage();

    expect(await screen.findByRole('button', { name: 'Edit' })).toBeInTheDocument();
  });

  it('does not show an Edit action when the actor lacks the subject write permission', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(buildView());

    renderPage();

    await screen.findByRole('button', { name: 'Share (copy URL)' });
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
  });

  it('does not show an Edit action for a shared read-only link without x-record', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({ _links: { self: { href: '/one-pagers/capability/cap-1', method: 'GET' } } }),
    );

    renderPage();

    await screen.findByRole('button', { name: 'Share (copy URL)' });
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
  });

  it('shows Share in read mode and swaps to Save/Cancel while editing', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(buildView({ _links: withRecordLink() }));
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts());

    renderPage();
    expect(await screen.findByRole('button', { name: 'Share (copy URL)' })).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Edit' }));

    expect(screen.queryByRole('button', { name: 'Share (copy URL)' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
  });

  it('renders active custom fields as editable inputs in their interleaved position, built-ins stay read-only', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        _links: withRecordLink(),
        fields: [
          builtInField('description', 'Description', { type: 'text', text: 'Handles all payments' }),
          customField('f1', 'Contract link', {
            type: 'link',
            value: { type: 'link', version: 1, value: { label: 'MSA', url: 'https://example.com' } },
            displayText: 'MSA',
          }),
          builtInField('maturity', 'Maturity', { type: 'maturity', maturity: { value: 85, section: 'Optimizing' } }),
        ],
      }),
    );
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(
      buildConfiguration([customFieldDef('f1', 'Contract link', { type: 'link' })]),
    );
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(
      buildFacts({
        values: [
          {
            fieldId: 'f1',
            value: { type: 'link', version: 1, value: { label: 'MSA', url: 'https://example.com' } },
            displayText: 'MSA',
            modifiedAt: '2026-01-01T00:00:00Z',
            modifiedBy: 'ea@example.com',
          },
        ],
      }),
    );

    renderPage();
    await user.click(await screen.findByRole('button', { name: 'Edit' }));

    expect(await screen.findByLabelText('Contract link label')).toHaveValue('MSA');
    expect(screen.getByLabelText('Contract link URL')).toHaveValue('https://example.com');
    expect(screen.getByText('Handles all payments')).toBeInTheDocument();
    expect(screen.getByText('85')).toBeInTheDocument();
    expect(screen.getByText('Optimizing')).toBeInTheDocument();

    const order = document.body.textContent ?? '';
    expect(order.indexOf('Description')).toBeLessThan(order.indexOf('Contract link'));
    expect(order.indexOf('Contract link')).toBeLessThan(order.indexOf('Maturity'));
  });

  it('records exactly the dirty custom fields, writes nothing for the unchanged one, and returns to read mode with fresh data', async () => {
    const user = userEvent.setup();
    const initialView = buildView({
      _links: withRecordLink(),
      fields: [
        customField('f1', 'Contract link', {
          value: { type: 'text', version: 1, value: 'old-1' },
          displayText: 'old-1',
        }),
        customField('f2', 'Notes', { value: { type: 'text', version: 1, value: 'old-2' }, displayText: 'old-2' }),
        customField('f3', 'Owner', { value: null }),
      ],
      completeness: { requiredCount: 3, filledCount: 2, missingFields: [{ fieldId: 'f3', name: 'Owner' }] },
    });
    const updatedView = buildView({
      _links: withRecordLink(),
      fields: [
        customField('f1', 'Contract link', {
          value: { type: 'text', version: 1, value: 'new-1' },
          displayText: 'new-1',
        }),
        customField('f2', 'Notes', { value: { type: 'text', version: 1, value: 'old-2' }, displayText: 'old-2' }),
        customField('f3', 'Owner', { value: { type: 'text', version: 1, value: 'new-3' }, displayText: 'new-3' }),
      ],
      completeness: { requiredCount: 3, filledCount: 3, missingFields: [] },
    });
    vi.mocked(onePagersApi.getOnePager).mockResolvedValueOnce(initialView).mockResolvedValue(updatedView);
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(
      buildConfiguration([customFieldDef('f1', 'Contract link'), customFieldDef('f2', 'Notes'), customFieldDef('f3', 'Owner')]),
    );
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(
      buildFacts({ values: [textFieldValue('f1', 'old-1'), textFieldValue('f2', 'old-2')] }),
    );
    vi.mocked(onePagersApi.recordFieldValue).mockResolvedValue(buildFacts());

    renderPage();
    await user.click(await screen.findByRole('button', { name: 'Edit' }));

    const f1Input = await screen.findByLabelText('Contract link');
    await user.clear(f1Input);
    await user.type(f1Input, 'new-1');

    const f3Input = screen.getByLabelText('Owner');
    await user.type(f3Input, 'new-3');

    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(onePagersApi.recordFieldValue).toHaveBeenCalledTimes(2));
    expect(onePagersApi.recordFieldValue).toHaveBeenCalledWith(expect.anything(), 'f1', {
      value: { type: 'text', version: 1, value: 'new-1' },
    });
    expect(onePagersApi.recordFieldValue).toHaveBeenCalledWith(expect.anything(), 'f3', {
      value: { type: 'text', version: 1, value: 'new-3' },
    });
    expect(onePagersApi.clearFieldValue).not.toHaveBeenCalled();

    expect(await screen.findByRole('button', { name: 'Share (copy URL)' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Save' })).not.toBeInTheDocument();
    expect(await screen.findByText('new-1')).toBeInTheDocument();
    expect(await screen.findByTestId('one-pager-completeness-summary')).toHaveTextContent(
      '3 of 3 required fields filled',
    );
  });

  it('clears a field emptied during editing instead of recording it', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        _links: withRecordLink(),
        fields: [customField('f1', 'Notes', { value: { type: 'text', version: 1, value: 'old' }, displayText: 'old' })],
      }),
    );
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([customFieldDef('f1', 'Notes')]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts({ values: [textFieldValue('f1', 'old')] }));
    vi.mocked(onePagersApi.clearFieldValue).mockResolvedValue(buildFacts());

    renderPage();
    await user.click(await screen.findByRole('button', { name: 'Edit' }));

    const input = await screen.findByLabelText('Notes');
    await user.clear(input);
    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(onePagersApi.clearFieldValue).toHaveBeenCalledTimes(1));
    expect(onePagersApi.clearFieldValue).toHaveBeenCalledWith(expect.objectContaining({ fieldId: 'f1' }));
    expect(onePagersApi.recordFieldValue).not.toHaveBeenCalled();
  });

  it('discards in-progress edits on Cancel and returns to read mode with the pre-edit value', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        _links: withRecordLink(),
        fields: [
          customField('f1', 'Notes', { value: { type: 'text', version: 1, value: 'original' }, displayText: 'original' }),
        ],
      }),
    );
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([customFieldDef('f1', 'Notes')]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts({ values: [textFieldValue('f1', 'original')] }));

    renderPage();
    await user.click(await screen.findByRole('button', { name: 'Edit' }));

    const input = await screen.findByLabelText('Notes');
    await user.clear(input);
    await user.type(input, 'changed');

    await user.click(screen.getByRole('button', { name: 'Cancel' }));

    expect(await screen.findByText('original')).toBeInTheDocument();
    expect(onePagersApi.recordFieldValue).not.toHaveBeenCalled();
    expect(onePagersApi.clearFieldValue).not.toHaveBeenCalled();
    expect(screen.getByRole('button', { name: 'Share (copy URL)' })).toBeInTheDocument();
  });
});

describe('OnePagerPage subject drawer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const labelCases: [OnePagerSubjectType, string][] = [
    ['capability', 'Open capability'],
    ['application', 'Open application'],
    ['acquired-entity', 'Open acquired entity'],
    ['vendor', 'Open vendor'],
    ['internal-team', 'Open internal team'],
  ];

  it.each(labelCases)('offers an Open button labeled for a %s subject when x-subject is present', async (subjectType, label) => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({
        subjectType,
        subjectId: 'subj-1',
        _links: {
          self: { href: `/api/v1/one-pagers/${subjectType}/subj-1`, method: 'GET' },
          'x-subject': { href: '/api/v1/subjects/subj-1', method: 'GET' },
        },
      }),
    );

    renderPage(`/one-pagers/${subjectType}/subj-1`);

    expect(await screen.findByRole('button', { name: label })).toBeInTheDocument();
  });

  it('hides the Open button when the view carries no x-subject link', async () => {
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(
      buildView({ _links: { self: { href: '/api/v1/one-pagers/capability/cap-1', method: 'GET' } } }),
    );

    renderPage();

    await screen.findByRole('button', { name: 'Share (copy URL)' });
    expect(screen.queryByRole('button', { name: 'Open capability' })).not.toBeInTheDocument();
  });

  it('opens the drawer for the viewed subject when the Open button is clicked', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(buildView());

    renderPage();

    expect(screen.queryByTestId('one-pager-subject-drawer')).not.toBeInTheDocument();
    await user.click(await screen.findByRole('button', { name: 'Open capability' }));

    expect(await screen.findByTestId('one-pager-subject-drawer')).toHaveTextContent('capability:cap-1');
  });

  it('does not offer the Open button while facts edit mode is active', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagersApi.getOnePager).mockResolvedValue(buildView({ _links: withRecordLink() }));
    vi.mocked(onePagersApi.getConfiguration).mockResolvedValue(buildConfiguration([]));
    vi.mocked(onePagersApi.getFacts).mockResolvedValue(buildFacts());

    renderPage();
    expect(await screen.findByRole('button', { name: 'Open capability' })).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Edit' }));

    expect(screen.queryByRole('button', { name: 'Open capability' })).not.toBeInTheDocument();
  });
});
