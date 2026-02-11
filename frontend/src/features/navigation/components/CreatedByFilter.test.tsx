import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { createMantineTestWrapper } from '../../../test/helpers';
import { CreatedByFilter } from './CreatedByFilter';
import type { ArtifactCreator } from '../utils/filterByCreator';

describe('CreatedByFilter', () => {
  const onSelectionChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const defaultUsers = [
    { id: 'user-alice', name: 'Alice Johnson', email: 'alice@example.com' },
    { id: 'user-bob', name: 'Bob Smith', email: 'bob@example.com' },
    { id: 'user-carol', email: 'carol@example.com' },
  ];

  function renderFilter(props: {
    artifactCreators?: ArtifactCreator[];
    users?: Array<{ id: string; name?: string; email: string }>;
    selectedCreatorIds?: string[];
  } = {}) {
    const { Wrapper } = createMantineTestWrapper();
    return render(
      <CreatedByFilter
        artifactCreators={props.artifactCreators ?? []}
        users={props.users ?? defaultUsers}
        selectedCreatorIds={props.selectedCreatorIds ?? []}
        onSelectionChange={onSelectionChange}
      />,
      { wrapper: Wrapper }
    );
  }

  it('should render the filter with a label', () => {
    renderFilter({
      artifactCreators: [
        { aggregateId: 'comp-1', creatorId: 'user-alice' },
      ],
    });

    expect(screen.getByText(/created by/i)).toBeInTheDocument();
  });

  it('should show unique creators as options deduplicated by creatorId', () => {
    const artifactCreators: ArtifactCreator[] = [
      { aggregateId: 'comp-1', creatorId: 'user-alice' },
      { aggregateId: 'comp-2', creatorId: 'user-alice' },
      { aggregateId: 'cap-1', creatorId: 'user-bob' },
      { aggregateId: 'ae-1', creatorId: 'user-alice' },
    ];

    renderFilter({ artifactCreators });

    const aliceOption = screen.getByText('Alice Johnson');
    const bobOption = screen.getByText('Bob Smith');
    expect(aliceOption).toBeInTheDocument();
    expect(bobOption).toBeInTheDocument();

    const allAlice = screen.getAllByText('Alice Johnson');
    expect(allAlice).toHaveLength(1);
  });

  it('should call onSelectionChange when a creator is selected', () => {
    const artifactCreators: ArtifactCreator[] = [
      { aggregateId: 'comp-1', creatorId: 'user-alice' },
      { aggregateId: 'comp-2', creatorId: 'user-bob' },
    ];

    renderFilter({ artifactCreators });

    fireEvent.click(screen.getByText('Alice Johnson'));

    expect(onSelectionChange).toHaveBeenCalledWith(['user-alice']);
  });

  it('should call onSelectionChange with empty array when selection is cleared', () => {
    const artifactCreators: ArtifactCreator[] = [
      { aggregateId: 'comp-1', creatorId: 'user-alice' },
    ];

    renderFilter({
      artifactCreators,
      selectedCreatorIds: ['user-alice'],
    });

    const clearButton = screen.getByRole('button', { name: /clear/i });
    fireEvent.click(clearButton);

    expect(onSelectionChange).toHaveBeenCalledWith([]);
  });

  it('should display email as fallback when user has no name', () => {
    const artifactCreators: ArtifactCreator[] = [
      { aggregateId: 'comp-1', creatorId: 'user-carol' },
    ];

    renderFilter({ artifactCreators });

    expect(screen.getByText('carol@example.com')).toBeInTheDocument();
  });
});
