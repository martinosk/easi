import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AdvancedSettings } from './AdvancedSettings';

const defaultProps = {
  maxTokens: 4096,
  onMaxTokensChange: vi.fn(),
  temperature: 0.3,
  onTemperatureChange: vi.fn(),
  systemPromptOverride: '',
  onSystemPromptOverrideChange: vi.fn(),
};

describe('AdvancedSettings', () => {
  it('renders all advanced fields', () => {
    render(<AdvancedSettings {...defaultProps} />);

    expect(screen.getByLabelText('Max Tokens')).toBeInTheDocument();
    expect(screen.getByLabelText('Temperature')).toBeInTheDocument();
    expect(screen.getByLabelText('System Prompt Override')).toBeInTheDocument();
  });

  it('displays current max tokens value', () => {
    render(<AdvancedSettings {...defaultProps} maxTokens={8192} />);

    const input = screen.getByLabelText('Max Tokens') as HTMLInputElement;
    expect(input.value).toBe('8192');
  });

  it('calls onMaxTokensChange when value changes', () => {
    const onMaxTokensChange = vi.fn();
    render(<AdvancedSettings {...defaultProps} onMaxTokensChange={onMaxTokensChange} />);

    fireEvent.change(screen.getByLabelText('Max Tokens'), { target: { value: '8192' } });
    expect(onMaxTokensChange).toHaveBeenCalledWith(8192);
  });

  it('displays current temperature value', () => {
    render(<AdvancedSettings {...defaultProps} temperature={1.5} />);

    const input = screen.getByLabelText('Temperature') as HTMLInputElement;
    expect(input.value).toBe('1.5');
  });

  it('calls onTemperatureChange when value changes', () => {
    const onTemperatureChange = vi.fn();
    render(<AdvancedSettings {...defaultProps} onTemperatureChange={onTemperatureChange} />);

    fireEvent.change(screen.getByLabelText('Temperature'), { target: { value: '0.7' } });
    expect(onTemperatureChange).toHaveBeenCalledWith(0.7);
  });

  it('displays system prompt override text', () => {
    render(<AdvancedSettings {...defaultProps} systemPromptOverride="Custom context" />);

    const textarea = screen.getByLabelText('System Prompt Override') as HTMLTextAreaElement;
    expect(textarea.value).toBe('Custom context');
  });

  it('calls onSystemPromptOverrideChange when value changes', () => {
    const onSystemPromptOverrideChange = vi.fn();
    render(<AdvancedSettings {...defaultProps} onSystemPromptOverrideChange={onSystemPromptOverrideChange} />);

    fireEvent.change(screen.getByLabelText('System Prompt Override'), { target: { value: 'New prompt' } });
    expect(onSystemPromptOverrideChange).toHaveBeenCalledWith('New prompt');
  });

  it('shows system prompt help text', () => {
    render(<AdvancedSettings {...defaultProps} />);

    expect(screen.getByText(/appended to the built-in system prompt/i)).toBeInTheDocument();
  });

  it('shows max tokens range hint', () => {
    render(<AdvancedSettings {...defaultProps} />);

    expect(screen.getByText(/256 - 32,768/)).toBeInTheDocument();
  });

  it('shows temperature range hint', () => {
    render(<AdvancedSettings {...defaultProps} />);

    expect(screen.getByText(/0\.0 - 2\.0/)).toBeInTheDocument();
  });

  it('falls back to default max tokens when input is cleared', () => {
    const onMaxTokensChange = vi.fn();
    render(<AdvancedSettings {...defaultProps} onMaxTokensChange={onMaxTokensChange} />);

    fireEvent.change(screen.getByLabelText('Max Tokens'), { target: { value: '' } });
    expect(onMaxTokensChange).toHaveBeenCalledWith(4096);
  });

  it('falls back to default temperature when input is cleared', () => {
    const onTemperatureChange = vi.fn();
    render(<AdvancedSettings {...defaultProps} onTemperatureChange={onTemperatureChange} />);

    fireEvent.change(screen.getByLabelText('Temperature'), { target: { value: '' } });
    expect(onTemperatureChange).toHaveBeenCalledWith(0.3);
  });
});
