import { fireEvent, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { type DetailGroup, DetailGroups } from './DetailGroups';
import { useDetailGroupLayout } from './useDetailGroupLayout';

const KEY = 'test-detail-groups';
const DEFAULT_ORDER = ['one', 'two', 'three'] as const;

const ALL_GROUPS: DetailGroup[] = [
  { id: 'one', title: 'One', content: <div data-testid="body-one">one body</div> },
  { id: 'two', title: 'Two', content: <div data-testid="body-two">two body</div> },
  { id: 'three', title: 'Three', content: <div data-testid="body-three">three body</div> },
];

function Harness({ groups }: { groups: DetailGroup[] }) {
  const layout = useDetailGroupLayout(KEY, DEFAULT_ORDER);
  return <DetailGroups groups={groups} layout={layout} />;
}

function headerTitles(): (string | null)[] {
  return screen.getAllByTestId(/^detail-group-/).map((item) => item.getAttribute('data-testid'));
}

describe('DetailGroups', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('renders the supplied groups in layout order with their content expanded', () => {
    renderWithProviders(<Harness groups={ALL_GROUPS} />, { withRouter: false });

    expect(headerTitles()).toEqual(['detail-group-one', 'detail-group-two', 'detail-group-three']);
    expect(screen.getByRole('button', { name: 'One' })).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByTestId('body-two')).toBeVisible();
  });

  it('skips groups the host did not supply without disturbing the order of the rest', () => {
    renderWithProviders(<Harness groups={[ALL_GROUPS[2], ALL_GROUPS[0]]} />, { withRouter: false });

    expect(headerTitles()).toEqual(['detail-group-one', 'detail-group-three']);
    expect(screen.queryByRole('button', { name: 'Two' })).not.toBeInTheDocument();
  });

  it('collapses and expands a group from its header', () => {
    renderWithProviders(<Harness groups={ALL_GROUPS} />, { withRouter: false });

    fireEvent.click(screen.getByRole('button', { name: 'Two' }));
    expect(screen.getByRole('button', { name: 'Two' })).toHaveAttribute('aria-expanded', 'false');
    expect(screen.getByTestId('body-two')).not.toBeVisible();

    fireEvent.click(screen.getByRole('button', { name: 'Two' }));
    expect(screen.getByRole('button', { name: 'Two' })).toHaveAttribute('aria-expanded', 'true');
  });

  it('moves a group with its header controls and disables the moves past either end', () => {
    renderWithProviders(<Harness groups={ALL_GROUPS} />, { withRouter: false });

    expect(screen.getByRole('button', { name: 'Move One up' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Move Three down' })).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: 'Move Three up' }));
    expect(headerTitles()).toEqual(['detail-group-one', 'detail-group-three', 'detail-group-two']);

    fireEvent.click(screen.getByRole('button', { name: 'Move One down' }));
    expect(headerTitles()).toEqual(['detail-group-three', 'detail-group-one', 'detail-group-two']);
    expect(screen.getByRole('button', { name: 'Move Three up' })).toBeDisabled();
  });

  it('shows a count badge in the header when the group carries one', () => {
    renderWithProviders(<Harness groups={[{ ...ALL_GROUPS[0], count: 3 }, ALL_GROUPS[1]]} />, { withRouter: false });

    expect(screen.getByTestId('group-count-one')).toHaveTextContent('3');
    expect(screen.queryByTestId('group-count-two')).not.toBeInTheDocument();
  });

  it('disables the edge moves of the rendered groups, not of the stored order', () => {
    renderWithProviders(<Harness groups={[ALL_GROUPS[1], ALL_GROUPS[2]]} />, { withRouter: false });

    expect(screen.getByRole('button', { name: 'Move Two up' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Move Three down' })).toBeDisabled();
  });
});
