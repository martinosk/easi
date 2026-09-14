import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { ComponentId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { ComponentContainmentSection } from './ComponentContainmentSection';

const crm = { id: 'crm' as ComponentId, name: 'CRM Suite', kind: 'composition' as const };

describe('ComponentContainmentSection', () => {
  it('shows a part as part of its parent with the containment kind', () => {
    const component = buildComponent({ id: 'comp-1' as ComponentId, partOf: crm });

    renderWithProviders(<ComponentContainmentSection component={component} />, { withRouter: false });

    expect(screen.getByTestId('part-of-line')).toHaveTextContent('Part of CRM Suite');
    expect(screen.getByTestId('containment-kind-badge')).toHaveTextContent('Composition');
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('lists the parts of a parent with each kind', () => {
    const component = buildComponent({
      id: 'comp-1' as ComponentId,
      parts: [
        { id: 'quoting' as ComponentId, name: 'Quoting', kind: 'composition' },
        { id: 'billing' as ComponentId, name: 'Billing', kind: 'aggregation' },
      ],
    });

    renderWithProviders(<ComponentContainmentSection component={component} />, { withRouter: false });

    const list = screen.getByTestId('parts-list');
    expect(list).toHaveTextContent('Quoting');
    expect(list).toHaveTextContent('Billing');
    const badges = screen.getAllByTestId('containment-kind-badge').map((badge) => badge.textContent);
    expect(badges).toEqual(['Composition', 'Aggregation']);
  });

  it('shows a standalone component as standalone without any action', () => {
    const component = buildComponent({ id: 'comp-1' as ComponentId });

    renderWithProviders(<ComponentContainmentSection component={component} />, { withRouter: false });

    expect(screen.getByTestId('containment-section')).toHaveTextContent('Standalone');
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });
});
