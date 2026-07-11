import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { AppearanceSettings } from './AppearanceSettings';

const { setSkinMock } = vi.hoisted(() => ({ setSkinMock: vi.fn() }));

vi.mock('../../../theme/skin', async () => {
  const actual = await vi.importActual<typeof import('../../../theme/skin')>('../../../theme/skin');
  return {
    ...actual,
    setSkin: setSkinMock,
    getSkin: () => 'easi',
  };
});

const render = () => renderWithProviders(<AppearanceSettings />, { withRouter: false });

describe('AppearanceSettings', () => {
  beforeEach(() => {
    setSkinMock.mockClear();
  });

  it('renders an option for every shipped skin', () => {
    render();

    expect(screen.getByRole('radio', { name: 'EASI graphite' })).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: 'Harbor' })).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: 'Evergreen' })).toBeInTheDocument();
  });

  it('selects the current skin', () => {
    render();

    expect(screen.getByRole('radio', { name: 'EASI graphite' })).toBeChecked();
  });

  it('shows the tenant-scoping description', () => {
    render();

    expect(
      screen.getByText('Chrome colours only — status colours stay the same for every tenant.'),
    ).toBeInTheDocument();
  });

  it('calls setSkin when a different skin is selected', async () => {
    render();

    await userEvent.click(screen.getByRole('radio', { name: 'Harbor' }));

    expect(setSkinMock).toHaveBeenCalledWith('harbor');
  });
});
