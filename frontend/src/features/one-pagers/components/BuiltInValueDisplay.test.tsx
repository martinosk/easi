import { MantineProvider } from '@mantine/core';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { theme } from '../../../theme/mantine';
import type { BuiltInValue } from '../types';
import { BuiltInValueDisplay } from './BuiltInValueDisplay';

function renderValue(value: BuiltInValue | null) {
  return render(
    <MemoryRouter>
      <MantineProvider theme={theme}>
        <BuiltInValueDisplay value={value} />
      </MantineProvider>
    </MemoryRouter>,
  );
}

describe('BuiltInValueDisplay', () => {
  describe('references', () => {
    it('renders a reference with a subjectType as a link to its one-pager', () => {
      renderValue({
        type: 'references',
        references: [{ id: 'app-1', label: 'Billing Service', subjectType: 'application' }],
      });

      const link = screen.getByRole('link', { name: 'Billing Service' });
      expect(link).toHaveAttribute('href', '/one-pagers/application/app-1');
    });

    it('renders a reference without a subjectType as plain text with no link', () => {
      renderValue({
        type: 'references',
        references: [{ id: 'domain-1', label: 'Payments' }],
      });

      expect(screen.getByText('Payments')).toBeInTheDocument();
      expect(screen.queryByRole('link')).not.toBeInTheDocument();
    });

    it('renders every reference in a multi-reference list', () => {
      renderValue({
        type: 'references',
        references: [
          { id: 'app-1', label: 'Billing Service', subjectType: 'application' },
          { id: 'app-2', label: 'Order Service', subjectType: 'application' },
        ],
      });

      expect(screen.getByRole('link', { name: 'Billing Service' })).toHaveAttribute(
        'href',
        '/one-pagers/application/app-1',
      );
      expect(screen.getByRole('link', { name: 'Order Service' })).toHaveAttribute(
        'href',
        '/one-pagers/application/app-2',
      );
    });
  });
});
