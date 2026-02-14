import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import { DomainForm } from './DomainForm';
import type { BusinessDomain, BusinessDomainId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers/renderWithProviders';
import * as useUsersModule from '../../users/hooks/useUsers';

vi.mock('../../users/hooks/useUsers', () => ({
  useEAOwnerCandidates: vi.fn(),
}));

describe('DomainForm', () => {
  type EAOwnerCandidatesResult = ReturnType<typeof useUsersModule.useEAOwnerCandidates>;

  function createEAOwnerCandidatesResult(result: {
    data: unknown[];
    isLoading: boolean;
    error: null;
  }): EAOwnerCandidatesResult {
    return result as EAOwnerCandidatesResult;
  }

  const mockUsers = [
    { id: 'user-1', name: 'Alice Smith', email: 'alice@example.com', role: 'architect' },
    { id: 'user-2', name: 'Bob Johnson', email: 'bob@example.com', role: 'admin' },
    { id: 'user-3', name: null, email: 'carol@example.com', role: 'architect' },
  ];

  const mockDomain: BusinessDomain = {
    id: 'domain-1' as BusinessDomainId,
    name: 'Customer Experience',
    description: 'All customer-facing capabilities',
    capabilityCount: 3,
    createdAt: '2025-01-15T10:00:00Z',
    _links: {
      self: { href: '/api/v1/business-domains/domain-1', method: 'GET' },
    },
  };

  let onSubmit: ReturnType<typeof vi.fn<(name: string, description: string, domainArchitectId?: string) => Promise<void>>>;
  let onCancel: ReturnType<typeof vi.fn<() => void>>;

  beforeEach(() => {
    onSubmit = vi.fn<(name: string, description: string, domainArchitectId?: string) => Promise<void>>();
    onCancel = vi.fn<() => void>();
    vi.mocked(useUsersModule.useEAOwnerCandidates).mockReturnValue({
      ...createEAOwnerCandidatesResult({
        data: mockUsers,
        isLoading: false,
        error: null,
      }),
    });
  });

  function renderCreateForm(submitOverride?: typeof onSubmit) {
    renderWithProviders(
      <DomainForm mode="create" onSubmit={submitOverride ?? onSubmit} onCancel={onCancel} />,
    );
  }

  function renderEditForm(domain: BusinessDomain, submitOverride?: typeof onSubmit) {
    renderWithProviders(
      <DomainForm mode="edit" domain={domain} onSubmit={submitOverride ?? onSubmit} onCancel={onCancel} />,
    );
  }

  function fillAndSubmit(fields: { name?: string; description?: string; architect?: string }) {
    if (fields.name !== undefined) {
      fireEvent.change(screen.getByTestId('domain-name-input'), { target: { value: fields.name } });
    }
    if (fields.description !== undefined) {
      fireEvent.change(screen.getByTestId('domain-description-input'), { target: { value: fields.description } });
    }
    if (fields.architect !== undefined) {
      fireEvent.change(screen.getByTestId('domain-architect-select'), { target: { value: fields.architect } });
    }
    fireEvent.click(screen.getByTestId('domain-form-submit'));
  }

  function mockLoadingUsers() {
    vi.mocked(useUsersModule.useEAOwnerCandidates).mockReturnValue(createEAOwnerCandidatesResult({
      data: [],
      isLoading: true,
      error: null,
    }));
  }

  describe('Create mode', () => {
    it('renders empty form in create mode', () => {
      renderCreateForm();

      const nameInput = screen.getByTestId('domain-name-input') as HTMLInputElement;
      const descriptionInput = screen.getByTestId('domain-description-input') as HTMLTextAreaElement;

      expect(nameInput.value).toBe('');
      expect(descriptionInput.value).toBe('');
      expect(screen.getByTestId('domain-form-submit')).toHaveTextContent('Create');
    });

    it('validates required name field', async () => {
      renderCreateForm();

      fireEvent.submit(screen.getByTestId('domain-form'));

      await waitFor(() => {
        expect(screen.getByTestId('domain-name-error')).toHaveTextContent('Name is required');
      });

      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('validates name length', async () => {
      renderCreateForm();
      fillAndSubmit({ name: 'a'.repeat(101) });

      await waitFor(() => {
        expect(screen.getByTestId('domain-name-error')).toHaveTextContent('Name must be 100 characters or less');
      });

      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('validates description length', async () => {
      renderCreateForm();
      fillAndSubmit({ name: 'Valid Name', description: 'a'.repeat(501) });

      await waitFor(() => {
        expect(screen.getByTestId('domain-description-error')).toHaveTextContent('Description must be 500 characters or less');
      });

      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('submits valid form data', async () => {
      onSubmit.mockResolvedValue(undefined);
      renderCreateForm();
      fillAndSubmit({ name: 'Customer Experience', description: 'All customer-facing capabilities' });

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith('Customer Experience', 'All customer-facing capabilities', undefined);
      });
    });

    it('trims whitespace from inputs', async () => {
      onSubmit.mockResolvedValue(undefined);
      renderCreateForm();
      fillAndSubmit({ name: '  Customer Experience  ', description: '  Description  ' });

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith('Customer Experience', 'Description', undefined);
      });
    });
  });

  describe('Edit mode', () => {
    it('renders form with domain data in edit mode', () => {
      renderEditForm(mockDomain);

      const nameInput = screen.getByTestId('domain-name-input') as HTMLInputElement;
      const descriptionInput = screen.getByTestId('domain-description-input') as HTMLTextAreaElement;

      expect(nameInput.value).toBe('Customer Experience');
      expect(descriptionInput.value).toBe('All customer-facing capabilities');
      expect(screen.getByTestId('domain-form-submit')).toHaveTextContent('Save');
    });

    it('handles empty description in edit mode', () => {
      renderEditForm({ ...mockDomain, description: '' });

      const descriptionInput = screen.getByTestId('domain-description-input') as HTMLTextAreaElement;
      expect(descriptionInput.value).toBe('');
    });
  });

  describe('Common behavior', () => {
    it('calls onCancel when cancel button is clicked', () => {
      renderCreateForm();

      fireEvent.click(screen.getByTestId('domain-form-cancel'));

      expect(onCancel).toHaveBeenCalled();
      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('disables submit button when name is empty', () => {
      renderCreateForm();

      const submitButton = screen.getByTestId('domain-form-submit') as HTMLButtonElement;
      expect(submitButton.disabled).toBe(true);
    });

    it('enables submit button when name is provided', () => {
      renderCreateForm();

      fireEvent.change(screen.getByTestId('domain-name-input'), { target: { value: 'Customer Experience' } });

      const submitButton = screen.getByTestId('domain-form-submit') as HTMLButtonElement;
      expect(submitButton.disabled).toBe(false);
    });

    it('displays backend error message', async () => {
      onSubmit.mockRejectedValue(new Error('Domain name already exists'));
      renderCreateForm();
      fillAndSubmit({ name: 'Customer Experience' });

      await waitFor(() => {
        expect(screen.getByTestId('domain-form-error')).toHaveTextContent('Domain name already exists');
      });
    });

    it('clears field error when user types', async () => {
      renderCreateForm();

      fireEvent.submit(screen.getByTestId('domain-form'));

      await waitFor(() => {
        expect(screen.getByTestId('domain-name-error')).toBeInTheDocument();
      });

      fireEvent.change(screen.getByTestId('domain-name-input'), { target: { value: 'Valid Name' } });

      expect(screen.queryByTestId('domain-name-error')).not.toBeInTheDocument();
    });
  });

  describe('Domain Architect', () => {
    it('should render domain architect select field', () => {
      renderCreateForm();

      expect(screen.getByTestId('domain-architect-select')).toBeInTheDocument();
      expect(screen.getByText('Domain Architect')).toBeInTheDocument();
    });

    it('should populate dropdown with eligible users (Architect and Admin roles)', () => {
      renderCreateForm();

      expect(screen.getByTestId('domain-architect-select')).toBeInTheDocument();
      expect(screen.getByText('Alice Smith (architect)')).toBeInTheDocument();
      expect(screen.getByText('Bob Johnson (admin)')).toBeInTheDocument();
    });

    it('should display email when user name is null', () => {
      renderCreateForm();

      expect(screen.getByText('carol@example.com (architect)')).toBeInTheDocument();
    });

    it('should have empty selection option as default', () => {
      renderCreateForm();

      const select = screen.getByTestId('domain-architect-select') as HTMLSelectElement;
      expect(select.value).toBe('');
      expect(screen.getByText('-- Select Domain Architect (optional) --')).toBeInTheDocument();
    });

    it('should submit with selected domain architect ID', async () => {
      onSubmit.mockResolvedValue(undefined);
      renderCreateForm();
      fillAndSubmit({ name: 'Finance', architect: 'user-1' });

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith('Finance', '', 'user-1');
      });
    });

    it('should submit without domain architect when none selected', async () => {
      onSubmit.mockResolvedValue(undefined);
      renderCreateForm();
      fillAndSubmit({ name: 'Finance' });

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith('Finance', '', undefined);
      });
    });

    it('should show loading state while fetching users', () => {
      mockLoadingUsers();
      renderCreateForm();

      expect(screen.getByText('Loading eligible users...')).toBeInTheDocument();
    });

    it('should disable select while loading users', () => {
      mockLoadingUsers();
      renderCreateForm();

      const select = screen.getByTestId('domain-architect-select') as HTMLSelectElement;
      expect(select.disabled).toBe(true);
    });

    it('should pre-select domain architect in edit mode', () => {
      renderEditForm({ ...mockDomain, domainArchitectId: 'user-2' });

      const select = screen.getByTestId('domain-architect-select') as HTMLSelectElement;
      expect(select.value).toBe('user-2');
    });

    it('should allow clearing domain architect selection', async () => {
      onSubmit.mockResolvedValue(undefined);
      renderEditForm({ ...mockDomain, domainArchitectId: 'user-2' });

      fireEvent.change(screen.getByTestId('domain-architect-select'), { target: { value: '' } });
      fireEvent.click(screen.getByTestId('domain-form-submit'));

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith('Customer Experience', 'All customer-facing capabilities', undefined);
      });
    });
  });
});
