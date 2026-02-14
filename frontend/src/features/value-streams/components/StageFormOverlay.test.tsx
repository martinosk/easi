import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { StageFormOverlay } from './StageFormOverlay';

const defaultProps = {
  isEditing: false,
  formData: { name: '', description: '' },
  onFormDataChange: vi.fn(),
  onSubmit: vi.fn(),
  onCancel: vi.fn(),
};

describe('StageFormOverlay', () => {
  describe('Add mode', () => {
    it('should display "Add Stage" heading when not editing', () => {
      render(<StageFormOverlay {...defaultProps} />);

      expect(screen.getByText('Add Stage')).toBeInTheDocument();
    });

    it('should display "Add" on the submit button when not editing', () => {
      render(<StageFormOverlay {...defaultProps} formData={{ name: 'Test', description: '' }} />);

      expect(screen.getByRole('button', { name: 'Add' })).toBeInTheDocument();
    });
  });

  describe('Edit mode', () => {
    it('should display "Edit Stage" heading when editing', () => {
      render(<StageFormOverlay {...defaultProps} isEditing={true} />);

      expect(screen.getByText('Edit Stage')).toBeInTheDocument();
    });

    it('should display "Save" on the submit button when editing', () => {
      render(
        <StageFormOverlay
          {...defaultProps}
          isEditing={true}
          formData={{ name: 'Existing', description: '' }}
        />,
      );

      expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument();
    });
  });

  describe('Form behavior', () => {
    it('should render name and description inputs with current formData', () => {
      render(
        <StageFormOverlay
          {...defaultProps}
          formData={{ name: 'Discovery', description: 'Initial stage' }}
        />,
      );

      const nameInput = screen.getByLabelText('Name') as HTMLInputElement;
      const descInput = screen.getByLabelText('Description') as HTMLTextAreaElement;

      expect(nameInput.value).toBe('Discovery');
      expect(descInput.value).toBe('Initial stage');
    });

    it.each([
      { field: 'Name', input: 'New Stage', initial: { name: '', description: '' }, expected: { name: 'New Stage', description: '' } },
      { field: 'Description', input: 'A description', initial: { name: 'Stage', description: '' }, expected: { name: 'Stage', description: 'A description' } },
    ])('should call onFormDataChange when $field input changes', ({ field, input, initial, expected }) => {
      const onFormDataChange = vi.fn();
      render(
        <StageFormOverlay {...defaultProps} onFormDataChange={onFormDataChange} formData={initial} />,
      );

      fireEvent.change(screen.getByLabelText(field), { target: { value: input } });

      expect(onFormDataChange).toHaveBeenCalledWith(expected);
    });

    it('should disable submit button when name is empty', () => {
      render(<StageFormOverlay {...defaultProps} formData={{ name: '', description: '' }} />);

      const addButton = screen.getByRole('button', { name: 'Add' });
      expect(addButton).toBeDisabled();
    });

    it('should disable submit button when name is only whitespace', () => {
      render(<StageFormOverlay {...defaultProps} formData={{ name: '   ', description: '' }} />);

      const addButton = screen.getByRole('button', { name: 'Add' });
      expect(addButton).toBeDisabled();
    });

    it('should enable submit button when name has content', () => {
      render(<StageFormOverlay {...defaultProps} formData={{ name: 'Valid Name', description: '' }} />);

      const addButton = screen.getByRole('button', { name: 'Add' });
      expect(addButton).not.toBeDisabled();
    });

    it('should call onSubmit when submit button is clicked', () => {
      const onSubmit = vi.fn();
      render(
        <StageFormOverlay
          {...defaultProps}
          onSubmit={onSubmit}
          formData={{ name: 'Test Stage', description: '' }}
        />,
      );

      fireEvent.click(screen.getByRole('button', { name: 'Add' }));

      expect(onSubmit).toHaveBeenCalledTimes(1);
    });

    it('should call onCancel when cancel button is clicked', () => {
      const onCancel = vi.fn();
      render(<StageFormOverlay {...defaultProps} onCancel={onCancel} />);

      fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));

      expect(onCancel).toHaveBeenCalledTimes(1);
    });
  });
});
