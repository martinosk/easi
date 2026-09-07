import { MantineProvider } from '@mantine/core';
import { QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../../../api/types';
import { createTestQueryClient } from '../../../test/helpers';
import { theme } from '../../../theme/mantine';
import type { EditGrant } from '../../edit-grants/types';
import type { OnePagerQualityResponse, OnePagerQualityRow } from '../types';
import { OnePagerQualityPage } from './OnePagerQualityPage';

vi.mock('../api/onePagerQualityApi', () => ({
  onePagerQualityApi: {
    getList: vi.fn(),
  },
}));

vi.mock('../../edit-grants/api/editGrantApi', () => ({
  editGrantApi: {
    create: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import toast from 'react-hot-toast';
import { editGrantApi } from '../../edit-grants/api/editGrantApi';
import { onePagerQualityApi } from '../api/onePagerQualityApi';

function buildRow(overrides: Partial<OnePagerQualityRow> = {}): OnePagerQualityRow {
  return {
    subjectType: 'application',
    subjectId: 'app-1',
    name: 'Billing',
    completeness: 'incomplete',
    requiredCount: 3,
    filledCount: 1,
    missingCount: 2,
    creatorId: 'user-1',
    creatorEmail: 'alice@dfds.com',
    createdAt: '2026-01-01T10:00:00Z',
    lastUpdatedAt: '2026-06-01T12:00:00Z',
    ...overrides,
  };
}

function buildResponse(overrides: Partial<OnePagerQualityResponse> = {}): OnePagerQualityResponse {
  return {
    data: [buildRow()],
    pagination: { hasMore: false, limit: 50 },
    _links: { self: { href: '/api/v1/one-pager-quality', method: 'GET' } },
    ...overrides,
  };
}

function buildGrant(overrides: Partial<EditGrant> = {}): EditGrant {
  return {
    id: 'grant-1',
    grantorId: 'grantor-1',
    grantorEmail: 'grantor@dfds.com',
    granteeEmail: 'colleague@dfds.com',
    artifactType: 'component',
    artifactId: 'app-1',
    scope: 'write',
    status: 'active',
    createdAt: '2026-01-01T00:00:00Z',
    expiresAt: '2026-01-31T00:00:00Z',
    _links: {},
    ...overrides,
  };
}

const EDIT_GRANTS_LINK = { 'x-edit-grants': { href: '/api/v1/edit-grants', method: 'POST' as const } };

function renderPage() {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MantineProvider theme={theme}>
        <MemoryRouter>
          <OnePagerQualityPage />
        </MemoryRouter>
      </MantineProvider>
    </QueryClientProvider>,
  );
}

describe('OnePagerQualityPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('shows a loading state while fetching', () => {
    vi.mocked(onePagerQualityApi.getList).mockReturnValue(new Promise(() => {}));

    renderPage();

    expect(screen.getByText('Loading one-pager quality...')).toBeInTheDocument();
  });

  it('renders rows once loaded', async () => {
    vi.mocked(onePagerQualityApi.getList).mockResolvedValue(
      buildResponse({ data: [buildRow({ subjectId: 'app-1', name: 'Billing' })] }),
    );

    renderPage();

    await waitFor(() => expect(screen.getByTestId('quality-row-app-1')).toBeInTheDocument());
    expect(screen.getByText('Billing')).toBeInTheDocument();
  });

  it('shows a visibly distinct completeness indicator for incomplete vs complete rows', async () => {
    vi.mocked(onePagerQualityApi.getList).mockResolvedValue(
      buildResponse({
        data: [
          buildRow({ subjectId: 'app-1', name: 'Billing', completeness: 'incomplete' }),
          buildRow({ subjectId: 'app-2', name: 'Payments', completeness: 'complete' }),
        ],
      }),
    );

    renderPage();

    await waitFor(() => expect(screen.getByTestId('quality-row-app-1')).toBeInTheDocument());

    const incompleteBadge = within(screen.getByTestId('quality-row-app-1')).getByTestId('completeness-badge');
    const completeBadge = within(screen.getByTestId('quality-row-app-2')).getByTestId('completeness-badge');

    expect(incompleteBadge).toHaveTextContent('Incomplete');
    expect(completeBadge).toHaveTextContent('Complete');
    expect(incompleteBadge.getAttribute('style')).not.toBe(completeBadge.getAttribute('style'));
  });

  it('shows an empty state when there are no rows', async () => {
    vi.mocked(onePagerQualityApi.getList).mockResolvedValue(buildResponse({ data: [] }));

    renderPage();

    await waitFor(() => expect(screen.getByText(/no one-pagers found/i)).toBeInTheDocument());
  });

  it('shows a graceful message on 403', async () => {
    vi.mocked(onePagerQualityApi.getList).mockRejectedValue(new ApiError('Forbidden', 403));

    renderPage();

    await waitFor(() => expect(screen.getByTestId('one-pager-quality-error')).toBeInTheDocument());
    expect(screen.getByText(/do not have permission/i)).toBeInTheDocument();
  });

  it('shows a generic error message on other failures', async () => {
    vi.mocked(onePagerQualityApi.getList).mockRejectedValue(new ApiError('Server error', 500));

    renderPage();

    await waitFor(() => expect(screen.getByTestId('one-pager-quality-error')).toBeInTheDocument());
    expect(screen.getByText(/something went wrong/i)).toBeInTheDocument();
  });

  it('clicking a sort header changes sort and clicking again toggles order', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagerQualityApi.getList).mockResolvedValue(buildResponse());

    renderPage();

    await waitFor(() => expect(screen.getByTestId('one-pager-quality-table')).toBeInTheDocument());
    expect(onePagerQualityApi.getList).toHaveBeenCalledWith(
      expect.objectContaining({ sort: 'completeness', order: 'asc' }),
    );

    await user.click(screen.getByTestId('sort-header-name'));

    await waitFor(() =>
      expect(onePagerQualityApi.getList).toHaveBeenLastCalledWith(
        expect.objectContaining({ sort: 'name', order: 'asc' }),
      ),
    );

    await user.click(screen.getByTestId('sort-header-name'));

    await waitFor(() =>
      expect(onePagerQualityApi.getList).toHaveBeenLastCalledWith(
        expect.objectContaining({ sort: 'name', order: 'desc' }),
      ),
    );
  });

  it('pagination Next uses the cursor returned by the previous page', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagerQualityApi.getList).mockResolvedValueOnce(
      buildResponse({ pagination: { hasMore: true, limit: 50, cursor: 'cursor-1' } }),
    );

    renderPage();

    await waitFor(() => expect(screen.getByTestId('pagination-next')).not.toBeDisabled());

    vi.mocked(onePagerQualityApi.getList).mockResolvedValueOnce(
      buildResponse({
        data: [buildRow({ subjectId: 'app-2', name: 'Payments' })],
        pagination: { hasMore: false, limit: 50 },
      }),
    );

    await user.click(screen.getByTestId('pagination-next'));

    await waitFor(() =>
      expect(onePagerQualityApi.getList).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 'cursor-1' })),
    );
  });

  it('pagination Prev returns to the previous cursor', async () => {
    const user = userEvent.setup();
    vi.mocked(onePagerQualityApi.getList).mockResolvedValueOnce(
      buildResponse({ pagination: { hasMore: true, limit: 50, cursor: 'cursor-1' } }),
    );

    renderPage();

    await waitFor(() => expect(screen.getByTestId('pagination-next')).not.toBeDisabled());

    vi.mocked(onePagerQualityApi.getList).mockResolvedValueOnce(
      buildResponse({ pagination: { hasMore: false, limit: 50 } }),
    );

    await user.click(screen.getByTestId('pagination-next'));

    await waitFor(() => expect(screen.getByTestId('pagination-prev')).not.toBeDisabled());

    vi.mocked(onePagerQualityApi.getList).mockResolvedValueOnce(
      buildResponse({ pagination: { hasMore: true, limit: 50, cursor: 'cursor-1' } }),
    );

    await user.click(screen.getByTestId('pagination-prev'));

    await waitFor(() =>
      expect(onePagerQualityApi.getList).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: undefined })),
    );
  });

  describe('Invite to edit', () => {
    it('shows the Invite to Edit action when the row carries an x-edit-grants link', async () => {
      vi.mocked(onePagerQualityApi.getList).mockResolvedValue(
        buildResponse({ data: [buildRow({ subjectId: 'app-1', _links: EDIT_GRANTS_LINK })] }),
      );

      renderPage();

      await waitFor(() => expect(screen.getByTestId('quality-row-app-1')).toBeInTheDocument());
      expect(within(screen.getByTestId('quality-row-app-1')).getByTestId('invite-to-edit-btn')).toBeInTheDocument();
    });

    it('shows no Invite to Edit action when the row carries no x-edit-grants link', async () => {
      vi.mocked(onePagerQualityApi.getList).mockResolvedValue(
        buildResponse({ data: [buildRow({ subjectId: 'app-1' })] }),
      );

      renderPage();

      await waitFor(() => expect(screen.getByTestId('quality-row-app-1')).toBeInTheDocument());
      expect(
        within(screen.getByTestId('quality-row-app-1')).queryByTestId('invite-to-edit-btn'),
      ).not.toBeInTheDocument();
    });

    it('opens the existing invite dialog prefilled with the mapped artifact type and creates the grant through the existing mechanism', async () => {
      const user = userEvent.setup();
      vi.mocked(onePagerQualityApi.getList).mockResolvedValue(
        buildResponse({
          data: [buildRow({ subjectType: 'application', subjectId: 'app-1', name: 'Billing', _links: EDIT_GRANTS_LINK })],
        }),
      );
      vi.mocked(editGrantApi.create).mockResolvedValue(buildGrant());

      renderPage();

      await waitFor(() => expect(screen.getByTestId('quality-row-app-1')).toBeInTheDocument());
      await user.click(within(screen.getByTestId('quality-row-app-1')).getByTestId('invite-to-edit-btn'));

      expect(screen.getByTestId('invite-to-edit-dialog')).toBeInTheDocument();

      await user.type(screen.getByTestId('grantee-email-input'), 'colleague@dfds.com');
      await user.click(screen.getByTestId('grant-submit-btn'));

      await waitFor(() =>
        expect(editGrantApi.create).toHaveBeenCalledWith({
          granteeEmail: 'colleague@dfds.com',
          artifactType: 'component',
          artifactId: 'app-1',
          reason: undefined,
        }),
      );
    });

    it('surfaces the existing onboarding-invitation toast when the invited email belongs to no user', async () => {
      const user = userEvent.setup();
      vi.mocked(onePagerQualityApi.getList).mockResolvedValue(
        buildResponse({
          data: [buildRow({ subjectType: 'application', subjectId: 'app-1', _links: EDIT_GRANTS_LINK })],
        }),
      );
      vi.mocked(editGrantApi.create).mockResolvedValue(
        buildGrant({ granteeEmail: 'newperson@dfds.com', invitationCreated: true }),
      );

      renderPage();

      await waitFor(() => expect(screen.getByTestId('quality-row-app-1')).toBeInTheDocument());
      await user.click(within(screen.getByTestId('quality-row-app-1')).getByTestId('invite-to-edit-btn'));
      await user.type(screen.getByTestId('grantee-email-input'), 'newperson@dfds.com');
      await user.click(screen.getByTestId('grant-submit-btn'));

      await waitFor(() =>
        expect(toast.success).toHaveBeenCalledWith(
          'Edit access granted. An invitation to join EASI was also created for newperson@dfds.com.',
        ),
      );
    });

    it.each([
      ['capability', 'capability'],
      ['application', 'component'],
      ['acquired-entity', 'acquired_entity'],
      ['vendor', 'vendor'],
      ['internal-team', 'internal_team'],
    ] as const)('creates the edit grant with artifact type %s for a %s row', async (subjectType, artifactType) => {
      const user = userEvent.setup();
      vi.mocked(onePagerQualityApi.getList).mockResolvedValue(
        buildResponse({
          data: [buildRow({ subjectType, subjectId: 'subject-1', name: 'Subject', _links: EDIT_GRANTS_LINK })],
        }),
      );
      vi.mocked(editGrantApi.create).mockResolvedValue(buildGrant());

      renderPage();

      await waitFor(() => expect(screen.getByTestId('quality-row-subject-1')).toBeInTheDocument());
      await user.click(within(screen.getByTestId('quality-row-subject-1')).getByTestId('invite-to-edit-btn'));
      await user.type(screen.getByTestId('grantee-email-input'), 'colleague@dfds.com');
      await user.click(screen.getByTestId('grant-submit-btn'));

      await waitFor(() =>
        expect(editGrantApi.create).toHaveBeenCalledWith(
          expect.objectContaining({ artifactType, artifactId: 'subject-1' }),
        ),
      );
    });
  });
});
