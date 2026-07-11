import { MantineProvider } from '@mantine/core';
import { QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../../../api/types';
import { createTestQueryClient } from '../../../test/helpers';
import { theme } from '../../../theme/mantine';
import type { OnePagerView, OnePagerViewField } from '../types';
import { OnePagerPage } from './OnePagerPage';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getOnePager: vi.fn(),
  },
}));

vi.mock('../../../utils/clipboard', () => ({
  copyToClipboard: vi.fn(),
  generateOnePagerShareUrl: vi.fn(
    (subjectType: string, subjectId: string) => `https://app.example.com/one-pagers/${subjectType}/${subjectId}`,
  ),
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
});
