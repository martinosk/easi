import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { createMantineTestWrapper } from '../../../test/helpers';
import { DomainFilter } from './DomainFilter';

describe('DomainFilter', () => {
  const onSelectionChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const defaultDomains = [
    { id: 'domain-sales', name: 'Sales' },
    { id: 'domain-engineering', name: 'Engineering' },
    { id: 'domain-hr', name: 'Human Resources' },
  ];

  function renderFilter(props: {
    domains?: Array<{ id: string; name: string }>;
    selectedDomainIds?: string[];
  } = {}) {
    const { Wrapper } = createMantineTestWrapper();
    return render(
      <DomainFilter
        domains={props.domains ?? defaultDomains}
        selectedDomainIds={props.selectedDomainIds ?? []}
        onSelectionChange={onSelectionChange}
      />,
      { wrapper: Wrapper }
    );
  }

  it('should render the filter with "Assigned to domain" label', () => {
    renderFilter();

    expect(screen.getByText(/assigned to domain/i)).toBeInTheDocument();
  });

  it('should show all domains as options', () => {
    renderFilter();

    expect(screen.getByText('Sales')).toBeInTheDocument();
    expect(screen.getByText('Engineering')).toBeInTheDocument();
    expect(screen.getByText('Human Resources')).toBeInTheDocument();
  });

  it('should show "Unassigned" as an option', () => {
    renderFilter();

    expect(screen.getByText('Unassigned')).toBeInTheDocument();
  });

  it('should call onSelectionChange when a domain is selected', () => {
    renderFilter();

    fireEvent.click(screen.getByText('Sales'));

    expect(onSelectionChange).toHaveBeenCalledWith(['domain-sales']);
  });

  it('should call onSelectionChange to deselect when an already-selected domain is clicked', () => {
    renderFilter({
      selectedDomainIds: ['domain-sales', 'domain-engineering'],
    });

    fireEvent.click(screen.getByText('Sales'));

    expect(onSelectionChange).toHaveBeenCalledWith(['domain-engineering']);
  });

  it('should call onSelectionChange with empty array when Clear is clicked', () => {
    renderFilter({
      selectedDomainIds: ['domain-sales'],
    });

    const clearButton = screen.getByRole('button', { name: /clear/i });
    fireEvent.click(clearButton);

    expect(onSelectionChange).toHaveBeenCalledWith([]);
  });

  it('should not show Clear button when no selections are active', () => {
    renderFilter({
      selectedDomainIds: [],
    });

    expect(screen.queryByRole('button', { name: /clear/i })).not.toBeInTheDocument();
  });

  it('should show Clear button when selections are active', () => {
    renderFilter({
      selectedDomainIds: ['domain-sales'],
    });

    expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument();
  });
});
