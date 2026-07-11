import { MantineProvider } from '@mantine/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { theme } from '../../../theme/mantine';
import { OnePagerActionButton } from './OnePagerActionButton';

function renderButton(hasLink: boolean) {
  const subject = {
    _links: hasLink ? { 'x-one-pager': { href: '/api/v1/one-pagers/vendor/vendor-1', method: 'GET' as const } } : {},
  };

  return render(
    <MemoryRouter>
      <MantineProvider theme={theme}>
        <OnePagerActionButton subject={subject} subjectType="vendor" subjectId="vendor-1" />
      </MantineProvider>
    </MemoryRouter>,
  );
}

describe('OnePagerActionButton', () => {
  it('renders the One-Pager action when the x-one-pager link is present', () => {
    renderButton(true);

    expect(screen.getByRole('button', { name: 'One-Pager' })).toBeInTheDocument();
  });

  it('renders nothing when the x-one-pager link is absent', () => {
    renderButton(false);

    expect(screen.queryByRole('button', { name: 'One-Pager' })).not.toBeInTheDocument();
  });

  it('navigates to the one-pager route when clicked', async () => {
    const user = userEvent.setup();
    const subject = {
      _links: { 'x-one-pager': { href: '/api/v1/one-pagers/vendor/vendor-1', method: 'GET' as const } },
    };

    render(
      <MemoryRouter initialEntries={['/']}>
        <MantineProvider theme={theme}>
          <Routes>
            <Route
              path="/"
              element={<OnePagerActionButton subject={subject} subjectType="vendor" subjectId="vendor-1" />}
            />
            <Route path="/one-pagers/:subjectType/:subjectId" element={<div>One-Pager Page</div>} />
          </Routes>
        </MantineProvider>
      </MemoryRouter>,
    );

    await user.click(screen.getByRole('button', { name: 'One-Pager' }));

    expect(screen.getByText('One-Pager Page')).toBeInTheDocument();
  });
});
