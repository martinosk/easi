import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { ToolCallIndicator } from './ToolCallIndicator';

const render = (ui: React.ReactElement) => renderWithProviders(ui, { withRouter: false });

describe('ToolCallIndicator', () => {
  it('should render running state with pulsing dot and activity text', () => {
    render(<ToolCallIndicator status="running" name="list_applications" />);
    expect(screen.getByRole('button')).toHaveAttribute('data-status', 'running');
    expect(screen.getByTestId('tool-call-pulse')).toBeInTheDocument();
    expect(screen.getByText('Looking up data...')).toBeInTheDocument();
  });

  it('should render completed state with check icon and preview', () => {
    render(<ToolCallIndicator status="completed" name="list_applications" resultPreview="Found 3 applications" />);
    expect(screen.getByRole('button')).toHaveAttribute('data-status', 'completed');
    expect(screen.getByText('\u2713')).toBeInTheDocument();
  });

  it('should render error state with warning icon and error message', () => {
    render(<ToolCallIndicator status="error" name="list_applications" errorMessage="Service unavailable" />);
    expect(screen.getByRole('button')).toHaveAttribute('data-status', 'error');
    expect(screen.getByText('\u26A0')).toBeInTheDocument();
    expect(screen.getByText('Service unavailable')).toBeInTheDocument();
  });

  it('should show search icon for read tools', () => {
    render(<ToolCallIndicator status="running" name="list_applications" />);
    expect(screen.getByText('\uD83D\uDD0D')).toBeInTheDocument();
  });

  it('should show pencil icon for write tools', () => {
    render(<ToolCallIndicator status="running" name="create_application" />);
    expect(screen.getByText('\u270F')).toBeInTheDocument();
  });

  it('should show trash icon for delete tools', () => {
    render(<ToolCallIndicator status="running" name="delete_application" />);
    expect(screen.getByText('\uD83D\uDDD1')).toBeInTheDocument();
  });

  it('should show trash icon for unrealize tools', () => {
    render(<ToolCallIndicator status="running" name="unrealize_capability" />);
    expect(screen.getByText('\uD83D\uDDD1')).toBeInTheDocument();
  });

  it('should map tool name to friendly label', () => {
    render(<ToolCallIndicator status="running" name="list_applications" />);
    expect(screen.getByText('Searching applications')).toBeInTheDocument();
  });

  it('should use tool name as fallback for unknown tools', () => {
    render(<ToolCallIndicator status="running" name="some_unknown_tool" />);
    expect(screen.getByText('some_unknown_tool')).toBeInTheDocument();
  });

  it('should expand preview on click when completed', () => {
    render(<ToolCallIndicator status="completed" name="list_applications" resultPreview="Found 3 applications" />);

    expect(screen.queryByText('Found 3 applications')).not.toBeInTheDocument();

    const indicator = screen.getByRole('button', { name: /Searching applications/ });
    fireEvent.click(indicator);

    expect(screen.getByText('Found 3 applications')).toBeInTheDocument();
  });

  it('should collapse preview on second click', () => {
    render(<ToolCallIndicator status="completed" name="list_applications" resultPreview="Found 3 applications" />);

    const indicator = screen.getByRole('button', { name: /Searching applications/ });
    fireEvent.click(indicator);
    expect(screen.getByText('Found 3 applications')).toBeInTheDocument();

    fireEvent.click(indicator);
    expect(screen.queryByText('Found 3 applications')).not.toBeInTheDocument();
  });
});
