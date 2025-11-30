import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { DepthSelector } from './DepthSelector';

describe('DepthSelector', () => {
  it('should render all depth options', () => {
    const onChange = vi.fn();
    render(<DepthSelector value={1} onChange={onChange} />);

    expect(screen.getByRole('button', { name: 'L1' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'L1-L2' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'L1-L3' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'L1-L4' })).toBeInTheDocument();
  });

  it('should highlight selected depth', () => {
    const onChange = vi.fn();
    render(<DepthSelector value={2} onChange={onChange} />);

    const selectedButton = screen.getByRole('button', { name: 'L1-L2' });
    expect(selectedButton).toHaveAttribute('data-selected', 'true');
  });

  it('should call onChange when depth is selected', () => {
    const onChange = vi.fn();
    render(<DepthSelector value={1} onChange={onChange} />);

    fireEvent.click(screen.getByRole('button', { name: 'L1-L3' }));

    expect(onChange).toHaveBeenCalledWith(3);
  });

  it('should not call onChange when current depth is clicked', () => {
    const onChange = vi.fn();
    render(<DepthSelector value={2} onChange={onChange} />);

    fireEvent.click(screen.getByRole('button', { name: 'L1-L2' }));

    expect(onChange).not.toHaveBeenCalled();
  });

  it('should display depth 1 as L1 only', () => {
    const onChange = vi.fn();
    render(<DepthSelector value={1} onChange={onChange} />);

    const l1Button = screen.getByRole('button', { name: 'L1' });
    expect(l1Button).toHaveAttribute('data-selected', 'true');
  });

  it('should display depth 4 as L1-L4', () => {
    const onChange = vi.fn();
    render(<DepthSelector value={4} onChange={onChange} />);

    const l4Button = screen.getByRole('button', { name: 'L1-L4' });
    expect(l4Button).toHaveAttribute('data-selected', 'true');
  });
});
