import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { InviteToEditDialog } from './InviteToEditDialog';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';

describe('InviteToEditDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnSubmit = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockOnSubmit.mockResolvedValue(undefined);

    HTMLDialogElement.prototype.showModal =
      HTMLDialogElement.prototype.showModal || vi.fn();
    HTMLDialogElement.prototype.close =
      HTMLDialogElement.prototype.close || vi.fn();
  });

  function renderDialog(isOpen = true) {
    return render(
      <MantineTestWrapper>
        <InviteToEditDialog
          isOpen={isOpen}
          onClose={mockOnClose}
          onSubmit={mockOnSubmit}
          artifactType="capability"
          artifactId="cap-123"
        />
      </MantineTestWrapper>
    );
  }

  describe('Dialog rendering', () => {
    it('should render dialog with title and description', () => {
      renderDialog();

      expect(screen.getByText('Invite to Edit')).toBeInTheDocument();
      expect(
        screen.getByText('Grant temporary edit access for this capability to a stakeholder.')
      ).toBeInTheDocument();
    });

    it('should render email and reason form fields', () => {
      renderDialog();

      expect(screen.getByTestId('grantee-email-input')).toBeInTheDocument();
      expect(screen.getByTestId('grant-reason-input')).toBeInTheDocument();
    });

    it('should render submit and cancel buttons', () => {
      renderDialog();

      expect(screen.getByTestId('grant-submit-btn')).toBeInTheDocument();
      expect(screen.getByTestId('grant-cancel-btn')).toBeInTheDocument();
      expect(screen.getByText('Grant Edit Access')).toBeInTheDocument();
      expect(screen.getByText('Cancel')).toBeInTheDocument();
    });
  });

  describe('Form submission', () => {
    it('should call onSubmit with email and artifact info', async () => {
      renderDialog();

      fireEvent.change(screen.getByTestId('grantee-email-input'), {
        target: { value: 'user@example.com' },
      });

      fireEvent.submit(screen.getByTestId('grant-submit-btn').closest('form')!);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith({
          granteeEmail: 'user@example.com',
          artifactType: 'capability',
          artifactId: 'cap-123',
          reason: undefined,
        });
      });
    });

    it('should include reason when provided', async () => {
      renderDialog();

      fireEvent.change(screen.getByTestId('grantee-email-input'), {
        target: { value: 'user@example.com' },
      });
      fireEvent.change(screen.getByTestId('grant-reason-input'), {
        target: { value: 'Quarterly review collaboration' },
      });

      fireEvent.submit(screen.getByTestId('grant-submit-btn').closest('form')!);

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith({
          granteeEmail: 'user@example.com',
          artifactType: 'capability',
          artifactId: 'cap-123',
          reason: 'Quarterly review collaboration',
        });
      });
    });

    it('should close dialog and reset form on successful submission', async () => {
      renderDialog();

      fireEvent.change(screen.getByTestId('grantee-email-input'), {
        target: { value: 'user@example.com' },
      });

      fireEvent.submit(screen.getByTestId('grant-submit-btn').closest('form')!);

      await waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled();
      });

      expect(
        (screen.getByTestId('grantee-email-input') as HTMLInputElement).value
      ).toBe('');
    });
  });

  describe('Error handling', () => {
    it('should display error message when submission fails', async () => {
      mockOnSubmit.mockRejectedValueOnce(new Error('Cannot grant edit access to yourself'));

      renderDialog();

      fireEvent.change(screen.getByTestId('grantee-email-input'), {
        target: { value: 'user@example.com' },
      });

      fireEvent.submit(screen.getByTestId('grant-submit-btn').closest('form')!);

      await waitFor(() => {
        expect(screen.getByTestId('grant-error-message')).toHaveTextContent(
          'Cannot grant edit access to yourself'
        );
      });

      expect(mockOnClose).not.toHaveBeenCalled();
    });

    it('should display generic error for non-Error exceptions', async () => {
      mockOnSubmit.mockRejectedValueOnce('unknown');

      renderDialog();

      fireEvent.change(screen.getByTestId('grantee-email-input'), {
        target: { value: 'user@example.com' },
      });

      fireEvent.submit(screen.getByTestId('grant-submit-btn').closest('form')!);

      await waitFor(() => {
        expect(screen.getByTestId('grant-error-message')).toHaveTextContent(
          'Failed to grant edit access'
        );
      });
    });
  });

  describe('Cancel behavior', () => {
    it('should call onClose and reset form when cancel is clicked', () => {
      renderDialog();

      fireEvent.change(screen.getByTestId('grantee-email-input'), {
        target: { value: 'partial@example.com' },
      });
      fireEvent.change(screen.getByTestId('grant-reason-input'), {
        target: { value: 'some reason' },
      });

      fireEvent.click(screen.getByTestId('grant-cancel-btn'));

      expect(mockOnClose).toHaveBeenCalled();
      expect(
        (screen.getByTestId('grantee-email-input') as HTMLInputElement).value
      ).toBe('');
      expect(
        (screen.getByTestId('grant-reason-input') as HTMLInputElement).value
      ).toBe('');
    });
  });
});
