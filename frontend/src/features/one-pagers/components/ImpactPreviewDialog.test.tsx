import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { OnePagerConfiguration } from '../types';
import { ImpactPreviewDialog } from './ImpactPreviewDialog';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    getImpactPreview: vi.fn(),
  },
}));

import { screen, waitFor } from '@testing-library/react';
import { onePagersApi } from '../api/onePagersApi';

function buildConfiguration(overrides: Partial<OnePagerConfiguration> = {}): OnePagerConfiguration {
  return {
    id: 'config-1',
    subjectType: 'application',
    builtInFields: [],
    customFields: [],
    displayOrder: [],
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'admin@example.com',
    _links: {
      self: { href: '/api/v1/one-pagers/configurations/application', method: 'GET' },
      'x-impact-preview': { href: '/api/v1/one-pagers/configurations/application/impact-preview', method: 'GET' },
    },
    ...overrides,
  };
}

describe('ImpactPreviewDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('shows the affected subject count once the preview loads', async () => {
    vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({
      subjectType: 'application',
      fieldId: 'field-1',
      affectedSubjectCount: 37,
    });

    renderWithProviders(
      <ImpactPreviewDialog
        configuration={buildConfiguration()}
        fieldName="Contract link"
        fieldId="field-1"
        isConfirming={false}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />,
      { withRouter: false },
    );

    await waitFor(() =>
      expect(screen.getByTestId('one-pager-impact-preview-message')).toHaveTextContent(
        'Making Contract link required will mark 37 Applications incomplete',
      ),
    );
  });

  it('shows a count-unavailable message but still allows confirm when the preview fetch fails', async () => {
    vi.mocked(onePagersApi.getImpactPreview).mockRejectedValue(new Error('boom'));
    const onConfirm = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(
      <ImpactPreviewDialog
        configuration={buildConfiguration()}
        fieldName="Contract link"
        fieldId="field-1"
        isConfirming={false}
        onConfirm={onConfirm}
        onCancel={vi.fn()}
      />,
      { withRouter: false },
    );

    await waitFor(() => expect(screen.getByTestId('one-pager-impact-preview-message')).toBeInTheDocument());
    expect(screen.getByTestId('one-pager-impact-preview-message')).not.toHaveTextContent('undefined');

    await user.click(screen.getByTestId('one-pager-impact-preview-confirm'));
    expect(onConfirm).toHaveBeenCalled();
  });

  it('calls onCancel when Cancel is clicked', async () => {
    vi.mocked(onePagersApi.getImpactPreview).mockResolvedValue({
      subjectType: 'vendor',
      affectedSubjectCount: 120,
    });
    const onCancel = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(
      <ImpactPreviewDialog
        configuration={buildConfiguration({ subjectType: 'vendor' })}
        fieldName="New field"
        isConfirming={false}
        onConfirm={vi.fn()}
        onCancel={onCancel}
      />,
      { withRouter: false },
    );

    await waitFor(() => expect(screen.getByTestId('one-pager-impact-preview-message')).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onCancel).toHaveBeenCalled();
  });
});
