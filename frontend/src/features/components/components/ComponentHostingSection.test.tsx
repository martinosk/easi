import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ComponentId, HATEOASLinks } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { ComponentHostingSection } from './ComponentHostingSection';

vi.mock('../api', () => ({
  componentsApi: {
    classifyHosting: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import { componentsApi } from '../api';

const readOnlyLinks: HATEOASLinks = { self: { href: '/api/v1/components/comp-1', method: 'GET' } };

const classifiableLinks: HATEOASLinks = {
  ...readOnlyLinks,
  'x-classify-hosting': { href: '/api/v1/components/comp-1/hosting', method: 'PUT' },
};

describe('ComponentHostingSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows the hosting classification without a control when classification is not permitted', () => {
    const component = buildComponent({ id: 'comp-1' as ComponentId, hosting: 'saas', _links: readOnlyLinks });

    renderWithProviders(<ComponentHostingSection component={component} />, { withRouter: false });

    expect(screen.getByTestId('hosting-badge')).toHaveTextContent('SaaS');
    expect(screen.queryByTestId('hosting-select')).not.toBeInTheDocument();
  });

  it('classifies hosting through the select control', async () => {
    const component = buildComponent({ id: 'comp-1' as ComponentId, hosting: 'unknown', _links: classifiableLinks });
    vi.mocked(componentsApi.classifyHosting).mockResolvedValue({ ...component, hosting: 'cloud' });

    renderWithProviders(<ComponentHostingSection component={component} />, { withRouter: false });

    await userEvent.click(screen.getByTestId('hosting-select'));
    await userEvent.click(await screen.findByRole('option', { name: 'Cloud', hidden: true }));

    await waitFor(() => {
      expect(componentsApi.classifyHosting).toHaveBeenCalledWith(component, 'cloud');
    });
  });

  it('does not reclassify when the current value is picked again', async () => {
    const component = buildComponent({ id: 'comp-1' as ComponentId, hosting: 'saas', _links: classifiableLinks });

    renderWithProviders(<ComponentHostingSection component={component} />, { withRouter: false });

    await userEvent.click(screen.getByTestId('hosting-select'));
    await userEvent.click(await screen.findByRole('option', { name: 'SaaS', hidden: true }));

    expect(componentsApi.classifyHosting).not.toHaveBeenCalled();
  });
});
