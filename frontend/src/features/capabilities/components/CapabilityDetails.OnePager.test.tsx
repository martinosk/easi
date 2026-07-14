import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import type { Capability, CapabilityId } from '../../../api/types';
import type { AppStore } from '../../../store/appStore';
import { useAppStore } from '../../../store/appStore';
import { createMantineTestWrapper, seedDb } from '../../../test/helpers';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { CapabilityDetails } from './CapabilityDetails';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: vi.fn(),
}));

function capability(hasOnePagerLink: boolean): Capability {
  return {
    id: 'cap-1' as CapabilityId,
    name: 'Test Capability',
    level: 'L2',
    createdAt: '2024-01-01T00:00:00Z',
    _links: {
      self: { href: '/api/v1/capabilities/cap-1', method: 'GET' },
      ...(hasOnePagerLink ? { 'x-one-pager': { href: '/api/v1/one-pagers/capability/cap-1', method: 'GET' } } : {}),
    },
  };
}

function renderPanel(hasOnePagerLink: boolean) {
  seedDb({ capabilities: [capability(hasOnePagerLink)], components: [] });
  vi.mocked(useAppStore).mockImplementation((selector: (state: AppStore) => unknown) =>
    selector({ selectedCapabilityId: 'cap-1', selectCapability: vi.fn() } as unknown as AppStore),
  );
  vi.mocked(useCurrentView).mockReturnValue({ currentView: null, currentViewId: null, isLoading: false, error: null });

  const { Wrapper } = createMantineTestWrapper();
  return render(<CapabilityDetails onRemoveFromView={vi.fn()} />, {
    wrapper: ({ children }) => (
      <MemoryRouter>
        <Wrapper>{children}</Wrapper>
      </MemoryRouter>
    ),
  });
}

describe('CapabilityDetails one-pager surface', () => {
  it('shows the One-Pager action and no inline edit form', async () => {
    renderPanel(true);

    expect(await screen.findByRole('button', { name: 'One-Pager' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Save one-pager' })).not.toBeInTheDocument();
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
  });

  it('shows no One-Pager action when the subject lacks the link', async () => {
    renderPanel(false);

    await screen.findByText('Test Capability');
    expect(screen.queryByRole('button', { name: 'One-Pager' })).not.toBeInTheDocument();
  });
});
