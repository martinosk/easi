import { screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { detailGroupIds, renderWithProviders } from '../../test/helpers';
import { DetailsShell } from './DetailsShell';

const KEY = 'test-details-shell';

function renderShell(transition: React.ReactNode, viewMembership?: React.ReactNode) {
  return renderWithProviders(
    <DetailsShell
      layoutKey={KEY}
      heading={<h2>Order Management</h2>}
      groups={[
        { id: 'description', title: 'Description', content: <div>description body</div> },
        { id: 'transition', title: 'Transition', content: transition },
        { id: 'metadata', title: 'Metadata', content: <div>metadata body</div> },
      ]}
      viewMembership={viewMembership}
      footer={<div data-testid="footer">history</div>}
    />,
    { withRouter: false },
  );
}

describe('DetailsShell', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('renders the heading before the groups and the footer after them', () => {
    renderShell(<div>plan</div>);

    const heading = screen.getByRole('heading', { name: 'Order Management' });
    const groups = screen.getAllByTestId(/^detail-group-/);
    const footer = screen.getByTestId('footer');
    expect(heading.compareDocumentPosition(groups[0]) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(groups.at(-1)?.compareDocumentPosition(footer) ?? 0 & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it('omits a group without content but keeps its default position for a surface that supplies it', () => {
    const first = renderShell(undefined);
    expect(detailGroupIds()).toEqual(['detail-group-description', 'detail-group-metadata']);
    first.unmount();

    renderShell(<div>plan</div>);
    expect(detailGroupIds()).toEqual(['detail-group-description', 'detail-group-transition', 'detail-group-metadata']);
  });

  it('appends an "In this view" group only when the host supplies view membership', () => {
    const without = renderShell(<div>plan</div>);
    expect(screen.queryByTestId('detail-group-view')).not.toBeInTheDocument();
    without.unmount();

    renderShell(<div>plan</div>, <div>colour</div>);
    expect(detailGroupIds().at(-1)).toBe('detail-group-view');
    expect(screen.getByRole('button', { name: 'In this view' })).toBeInTheDocument();
  });
});
