import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateRelationDialog } from './CreateRelationDialog';
import { useAppStore } from '../store/appStore';

// Mock the store
vi.mock('../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

describe('CreateRelationDialog', () => {
  const mockOnClose = vi.fn();
  const mockCreateRelation = vi.fn();

  const mockComponents = [
    { id: '1', name: 'Component A' },
    { id: '2', name: 'Component B' },
    { id: '3', name: 'Component C' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useAppStore).mockImplementation((selector: any) =>
      selector({
        components: mockComponents,
        createRelation: mockCreateRelation,
      })
    );

    // Mock HTMLDialogElement methods
    HTMLDialogElement.prototype.showModal = vi.fn();
    HTMLDialogElement.prototype.close = vi.fn();
  });

  it('should render dialog when open', () => {
    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />);

    expect(screen.getByRole('heading', { name: 'Create Relation', hidden: true })).toBeInTheDocument();
    expect(screen.getByLabelText(/Source Component/, { hidden: true })).toBeInTheDocument();
    expect(screen.getByLabelText(/Target Component/, { hidden: true })).toBeInTheDocument();
    expect(screen.getByLabelText(/Relation Type/, { hidden: true })).toBeInTheDocument();
  });

  it('should display all components in dropdowns', () => {
    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />);

    const sourceSelect = screen.getByLabelText(/Source Component/, { hidden: true }) as HTMLSelectElement;
    const targetSelect = screen.getByLabelText(/Target Component/, { hidden: true }) as HTMLSelectElement;

    expect(sourceSelect.options).toHaveLength(4); // 3 components + 1 placeholder
    expect(targetSelect.options).toHaveLength(4);
  });

  it('should display error when source and target are not selected', async () => {
    const { container } = render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />);

    const form = container.querySelector('form')!;
    fireEvent.submit(form);

    await waitFor(() => {
      expect(screen.getByText('Both source and target components are required', { hidden: true })).toBeInTheDocument();
    });

    expect(mockCreateRelation).not.toHaveBeenCalled();
  });

  it('should display error when source and target are the same', async () => {
    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />);

    const sourceSelect = screen.getByLabelText(/Source Component/, { hidden: true });
    const targetSelect = screen.getByLabelText(/Target Component/, { hidden: true });
    const buttons = screen.getAllByRole('button', { hidden: true });
    const submitButton = buttons.find(btn => btn.textContent === 'Create Relation');

    fireEvent.change(sourceSelect, { target: { value: '1' } });
    fireEvent.change(targetSelect, { target: { value: '1' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(screen.getByText('Source and target components must be different', { hidden: true })).toBeInTheDocument();
    });

    expect(mockCreateRelation).not.toHaveBeenCalled();
  });

  it('should call createRelation with valid data', async () => {
    mockCreateRelation.mockResolvedValueOnce({
      id: 'rel-1',
      sourceComponentId: '1',
      targetComponentId: '2',
      relationType: 'Triggers',
    });

    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />);

    const sourceSelect = screen.getByLabelText(/Source Component/, { hidden: true });
    const targetSelect = screen.getByLabelText(/Target Component/, { hidden: true });
    const typeSelect = screen.getByLabelText(/Relation Type/, { hidden: true });
    const nameInput = screen.getByLabelText(/Name/, { hidden: true });
    const buttons = screen.getAllByRole('button', { hidden: true });
    const submitButton = buttons.find(btn => btn.textContent === 'Create Relation');

    fireEvent.change(sourceSelect, { target: { value: '1' } });
    fireEvent.change(targetSelect, { target: { value: '2' } });
    fireEvent.change(typeSelect, { target: { value: 'Triggers' } });
    fireEvent.change(nameInput, { target: { value: 'Test Relation' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(mockCreateRelation).toHaveBeenCalledWith('1', '2', 'Triggers', 'Test Relation', undefined);
    });

    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should handle Serves relation type', async () => {
    mockCreateRelation.mockResolvedValueOnce({
      id: 'rel-1',
      sourceComponentId: '1',
      targetComponentId: '2',
      relationType: 'Serves',
    });

    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />);

    const sourceSelect = screen.getByLabelText(/Source Component/, { hidden: true });
    const targetSelect = screen.getByLabelText(/Target Component/, { hidden: true });
    const typeSelect = screen.getByLabelText(/Relation Type/, { hidden: true });
    const buttons = screen.getAllByRole('button', { hidden: true });
    const submitButton = buttons.find(btn => btn.textContent === 'Create Relation');

    fireEvent.change(sourceSelect, { target: { value: '1' } });
    fireEvent.change(targetSelect, { target: { value: '2' } });
    fireEvent.change(typeSelect, { target: { value: 'Serves' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(mockCreateRelation).toHaveBeenCalledWith('1', '2', 'Serves', undefined, undefined);
    });
  });

  it('should pre-fill source and target when provided', () => {
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="2"
      />
    );

    const sourceSelect = screen.getByLabelText(/Source Component/, { hidden: true }) as HTMLSelectElement;
    const targetSelect = screen.getByLabelText(/Target Component/, { hidden: true }) as HTMLSelectElement;

    expect(sourceSelect.value).toBe('1');
    expect(targetSelect.value).toBe('2');
  });

  it('should disable source and target when pre-filled', () => {
    render(
      <CreateRelationDialog
        isOpen={true}
        onClose={mockOnClose}
        sourceComponentId="1"
        targetComponentId="2"
      />
    );

    const sourceSelect = screen.getByLabelText(/Source Component/, { hidden: true }) as HTMLSelectElement;
    const targetSelect = screen.getByLabelText(/Target Component/, { hidden: true }) as HTMLSelectElement;

    expect(sourceSelect.disabled).toBe(true);
    expect(targetSelect.disabled).toBe(true);
  });

  it('should handle create relation error', async () => {
    mockCreateRelation.mockRejectedValueOnce(new Error('Creation failed'));

    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />);

    const sourceSelect = screen.getByLabelText(/Source Component/, { hidden: true });
    const targetSelect = screen.getByLabelText(/Target Component/, { hidden: true });
    const buttons = screen.getAllByRole('button', { hidden: true });
    const submitButton = buttons.find(btn => btn.textContent === 'Create Relation');

    fireEvent.change(sourceSelect, { target: { value: '1' } });
    fireEvent.change(targetSelect, { target: { value: '2' } });
    fireEvent.click(submitButton!);

    await waitFor(() => {
      expect(screen.getByText('Creation failed', { hidden: true })).toBeInTheDocument();
    });

    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it('should disable submit button when required fields are empty', () => {
    render(<CreateRelationDialog isOpen={true} onClose={mockOnClose} />);

    const buttons = screen.getAllByRole('button', { hidden: true });
    const submitButton = buttons.find(btn => btn.textContent === 'Create Relation') as HTMLButtonElement;

    expect(submitButton!.disabled).toBe(true);
  });
});
