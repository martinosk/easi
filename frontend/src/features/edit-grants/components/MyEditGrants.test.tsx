import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MyEditGrants } from './MyEditGrants';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import type { EditGrant } from '../types';

vi.mock('../hooks/useEditGrants', () => ({
  useMyEditGrants: vi.fn(),
}));

import { useMyEditGrants } from '../hooks/useEditGrants';

function createGrant(overrides: Partial<EditGrant> = {}): EditGrant {
  return {
    id: 'grant-1',
    grantorId: 'grantor-id',
    grantorEmail: 'grantor@example.com',
    granteeEmail: 'grantee@example.com',
    artifactType: 'capability',
    artifactId: 'cap-123',
    scope: 'write',
    status: 'active',
    createdAt: '2025-01-01T00:00:00Z',
    expiresAt: '2025-01-31T00:00:00Z',
    _links: {},
    ...overrides,
  };
}

function mockHook(data: EditGrant[] | undefined, isLoading = false) {
  vi.mocked(useMyEditGrants).mockReturnValue({
    data,
    isLoading,
    error: null,
    isError: false,
    isPending: isLoading,
    isSuccess: !isLoading && data !== undefined,
    status: isLoading ? 'pending' : 'success',
  } as unknown as ReturnType<typeof useMyEditGrants>);
}

describe('MyEditGrants', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function renderComponent() {
    return render(
      <MantineTestWrapper>
        <MyEditGrants />
      </MantineTestWrapper>
    );
  }

  it('should show loading spinner while data is loading', () => {
    mockHook(undefined, true);
    renderComponent();
    expect(screen.getByTestId('my-edit-grants-loading')).toBeInTheDocument();
  });

  it('should render nothing when there are no active grants', () => {
    mockHook([]);
    renderComponent();
    expect(screen.queryByTestId('my-edit-grants')).not.toBeInTheDocument();
  });

  it('should render nothing when all grants are revoked', () => {
    mockHook([createGrant({ status: 'revoked' })]);
    renderComponent();
    expect(screen.queryByTestId('my-edit-grants')).not.toBeInTheDocument();
  });

  it('should render active grants', () => {
    mockHook([
      createGrant({ id: 'g1', artifactName: 'Customer Onboarding' }),
      createGrant({ id: 'g2', artifactName: 'Payment Service' }),
    ]);
    renderComponent();

    expect(screen.getByTestId('my-edit-grants')).toBeInTheDocument();
    expect(screen.getByTestId('my-grant-g1')).toBeInTheDocument();
    expect(screen.getByTestId('my-grant-g2')).toBeInTheDocument();
  });

  it('should display artifact name instead of raw ID', () => {
    mockHook([createGrant({ artifactName: 'Customer Onboarding' })]);
    renderComponent();
    expect(screen.getByText('Customer Onboarding')).toBeInTheDocument();
  });

  it('should show "Deleted artifact" when artifactName is missing', () => {
    mockHook([createGrant({ artifactName: undefined })]);
    renderComponent();
    expect(screen.getByText('Deleted artifact')).toBeInTheDocument();
  });

  it('should render artifact name as a link when artifact HATEOAS link exists', () => {
    mockHook([
      createGrant({
        artifactName: 'Customer Onboarding',
        _links: {
          artifact: { href: '/business-domains?capability=cap-123', method: 'GET' },
        },
      }),
    ]);
    renderComponent();

    const link = screen.getByRole('link', { name: 'Customer Onboarding' });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute('href', '/business-domains?capability=cap-123');
  });

  it('should render artifact name without link when artifact HATEOAS link is absent', () => {
    mockHook([createGrant({ artifactName: 'Customer Onboarding', _links: {} })]);
    renderComponent();

    expect(screen.getByText('Customer Onboarding')).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Customer Onboarding' })).not.toBeInTheDocument();
  });

  it('should display grantor email', () => {
    mockHook([createGrant({ grantorEmail: 'alice@example.com' })]);
    renderComponent();
    expect(screen.getByText('Granted by alice@example.com')).toBeInTheDocument();
  });

  it('should display expiration date', () => {
    mockHook([createGrant({ expiresAt: '2025-01-31T00:00:00Z' })]);
    renderComponent();
    expect(screen.getByText(/Expires/)).toBeInTheDocument();
  });

  it('should display reason when provided', () => {
    mockHook([createGrant({ reason: 'For Q1 review' })]);
    renderComponent();
    expect(screen.getByText('For Q1 review')).toBeInTheDocument();
  });

  it('should not display reason when not provided', () => {
    mockHook([createGrant({ reason: undefined })]);
    renderComponent();
    expect(screen.queryByText('For Q1 review')).not.toBeInTheDocument();
  });

  it('should filter out non-active grants', () => {
    mockHook([
      createGrant({ id: 'active-1', status: 'active', artifactName: 'Active Cap' }),
      createGrant({ id: 'revoked-1', status: 'revoked', artifactName: 'Revoked Cap' }),
      createGrant({ id: 'expired-1', status: 'expired', artifactName: 'Expired Cap' }),
    ]);
    renderComponent();

    expect(screen.getByTestId('my-grant-active-1')).toBeInTheDocument();
    expect(screen.queryByTestId('my-grant-revoked-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('my-grant-expired-1')).not.toBeInTheDocument();
  });
});
