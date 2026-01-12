import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DomainForm } from './DomainForm';
import type { BusinessDomain, BusinessDomainId } from '../../../api/types';

describe('DomainForm', () => {
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

  describe('Create mode', () => {
    it('renders empty form in create mode', () => {
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const nameInput = screen.getByTestId('domain-name-input') as HTMLInputElement;
      const descriptionInput = screen.getByTestId('domain-description-input') as HTMLTextAreaElement;

      expect(nameInput.value).toBe('');
      expect(descriptionInput.value).toBe('');
      expect(screen.getByTestId('domain-form-submit')).toHaveTextContent('Create');
    });

    it('validates required name field', async () => {
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const form = screen.getByTestId('domain-form');
      fireEvent.submit(form);

      await waitFor(() => {
        expect(screen.getByTestId('domain-name-error')).toHaveTextContent('Name is required');
      });

      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('validates name length', async () => {
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const nameInput = screen.getByTestId('domain-name-input');
      fireEvent.change(nameInput, { target: { value: 'a'.repeat(101) } });

      const submitButton = screen.getByTestId('domain-form-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByTestId('domain-name-error')).toHaveTextContent('Name must be 100 characters or less');
      });

      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('validates description length', async () => {
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const nameInput = screen.getByTestId('domain-name-input');
      const descriptionInput = screen.getByTestId('domain-description-input');

      fireEvent.change(nameInput, { target: { value: 'Valid Name' } });
      fireEvent.change(descriptionInput, { target: { value: 'a'.repeat(501) } });

      const submitButton = screen.getByTestId('domain-form-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByTestId('domain-description-error')).toHaveTextContent('Description must be 500 characters or less');
      });

      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('submits valid form data', async () => {
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const nameInput = screen.getByTestId('domain-name-input');
      const descriptionInput = screen.getByTestId('domain-description-input');

      fireEvent.change(nameInput, { target: { value: 'Customer Experience' } });
      fireEvent.change(descriptionInput, { target: { value: 'All customer-facing capabilities' } });

      const submitButton = screen.getByTestId('domain-form-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith('Customer Experience', 'All customer-facing capabilities');
      });
    });

    it('trims whitespace from inputs', async () => {
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const nameInput = screen.getByTestId('domain-name-input');
      const descriptionInput = screen.getByTestId('domain-description-input');

      fireEvent.change(nameInput, { target: { value: '  Customer Experience  ' } });
      fireEvent.change(descriptionInput, { target: { value: '  Description  ' } });

      const submitButton = screen.getByTestId('domain-form-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith('Customer Experience', 'Description');
      });
    });
  });

  describe('Edit mode', () => {
    it('renders form with domain data in edit mode', () => {
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="edit" domain={mockDomain} onSubmit={onSubmit} onCancel={onCancel} />);

      const nameInput = screen.getByTestId('domain-name-input') as HTMLInputElement;
      const descriptionInput = screen.getByTestId('domain-description-input') as HTMLTextAreaElement;

      expect(nameInput.value).toBe('Customer Experience');
      expect(descriptionInput.value).toBe('All customer-facing capabilities');
      expect(screen.getByTestId('domain-form-submit')).toHaveTextContent('Save');
    });

    it('handles empty description in edit mode', () => {
      const domain: BusinessDomain = { ...mockDomain, description: '' };
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="edit" domain={domain} onSubmit={onSubmit} onCancel={onCancel} />);

      const descriptionInput = screen.getByTestId('domain-description-input') as HTMLTextAreaElement;
      expect(descriptionInput.value).toBe('');
    });
  });

  describe('Common behavior', () => {
    it('calls onCancel when cancel button is clicked', () => {
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const cancelButton = screen.getByTestId('domain-form-cancel');
      fireEvent.click(cancelButton);

      expect(onCancel).toHaveBeenCalled();
      expect(onSubmit).not.toHaveBeenCalled();
    });

    it('disables submit button when name is empty', () => {
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const submitButton = screen.getByTestId('domain-form-submit') as HTMLButtonElement;
      expect(submitButton.disabled).toBe(true);
    });

    it('enables submit button when name is provided', () => {
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const nameInput = screen.getByTestId('domain-name-input');
      fireEvent.change(nameInput, { target: { value: 'Customer Experience' } });

      const submitButton = screen.getByTestId('domain-form-submit') as HTMLButtonElement;
      expect(submitButton.disabled).toBe(false);
    });

    it('displays backend error message', async () => {
      const onSubmit = vi.fn().mockRejectedValue(new Error('Domain name already exists'));
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const nameInput = screen.getByTestId('domain-name-input');
      fireEvent.change(nameInput, { target: { value: 'Customer Experience' } });

      const submitButton = screen.getByTestId('domain-form-submit');
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByTestId('domain-form-error')).toHaveTextContent('Domain name already exists');
      });
    });

    it('clears field error when user types', async () => {
      const onSubmit = vi.fn();
      const onCancel = vi.fn();

      render(<DomainForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />);

      const form = screen.getByTestId('domain-form');
      fireEvent.submit(form);

      await waitFor(() => {
        expect(screen.getByTestId('domain-name-error')).toBeInTheDocument();
      });

      const nameInput = screen.getByTestId('domain-name-input');
      fireEvent.change(nameInput, { target: { value: 'Valid Name' } });

      expect(screen.queryByTestId('domain-name-error')).not.toBeInTheDocument();
    });
  });
});
