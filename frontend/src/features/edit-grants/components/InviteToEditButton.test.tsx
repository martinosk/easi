import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { InviteToEditButton } from './InviteToEditButton';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import type { ResourceWithLinks } from '../../../utils/hateoas';

vi.mock('../hooks/useEditGrants', () => ({
  useCreateEditGrant: vi.fn(),
}));

import { useCreateEditGrant } from '../hooks/useEditGrants';

describe('InviteToEditButton', () => {
  const mockMutateAsync = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(useCreateEditGrant).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
      isError: false,
      isSuccess: false,
      error: null,
      data: undefined,
      mutate: vi.fn(),
      reset: vi.fn(),
    } as unknown as ReturnType<typeof useCreateEditGrant>);
  });

  function renderButton(resource: ResourceWithLinks) {
    return render(
      <MantineTestWrapper>
        <InviteToEditButton
          resource={resource}
          artifactType="capability"
          artifactId="cap-123"
        />
      </MantineTestWrapper>
    );
  }

  it('should render when x-edit-grants link is present', () => {
    const resource: ResourceWithLinks = {
      _links: {
        'x-edit-grants': { href: '/api/v1/edit-grants', method: 'POST' },
      },
    };

    renderButton(resource);

    const button = screen.getByTestId('invite-to-edit-btn');
    expect(button).toBeInTheDocument();
    expect(button).toHaveTextContent('Invite to Edit...');
  });

  it('should not render when x-edit-grants link is absent', () => {
    const resource: ResourceWithLinks = {
      _links: {
        self: { href: '/api/v1/capabilities/cap-123', method: 'GET' },
      },
    };

    renderButton(resource);

    expect(screen.queryByTestId('invite-to-edit-btn')).not.toBeInTheDocument();
  });

  it('should not render when _links is undefined', () => {
    const resource: ResourceWithLinks = {};

    renderButton(resource);

    expect(screen.queryByTestId('invite-to-edit-btn')).not.toBeInTheDocument();
  });

  it('should open dialog when clicked', () => {
    const resource: ResourceWithLinks = {
      _links: {
        'x-edit-grants': { href: '/api/v1/edit-grants', method: 'POST' },
      },
    };

    renderButton(resource);

    fireEvent.click(screen.getByTestId('invite-to-edit-btn'));

    expect(screen.getByTestId('invite-to-edit-dialog')).toBeInTheDocument();
  });
});
