import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId, toEnterpriseCapabilityId } from '../../../api/types';

vi.mock('../api/standardApplicationApi', () => ({
  standardApplicationApi: {
    getForEnterpriseCapability: vi.fn(),
    getHistory: vi.fn(),
    set: vi.fn(),
  },
}));

vi.mock('../../components/hooks/useComponents', () => ({
  useComponents: vi.fn(),
}));

import { standardApplicationApi } from '../api/standardApplicationApi';
import { useComponents } from '../../components/hooks/useComponents';
import { SetStandardApplicationForm } from './SetStandardApplicationForm';

const mockedSet = vi.mocked(standardApplicationApi.set);
const mockedComponents = vi.mocked(useComponents);

function renderForm(props: Partial<React.ComponentProps<typeof SetStandardApplicationForm>> = {}) {
  mockedComponents.mockReturnValue({
    data: [
      { id: toComponentId('app-a'), name: 'Acme ERP', createdAt: '2025-01-01', _links: {} },
      { id: toComponentId('app-b'), name: 'Beta Suite', createdAt: '2025-01-01', _links: {} },
    ],
    isLoading: false,
    error: null,
  } as never);
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const onSubmitted = vi.fn();
  const onCancel = vi.fn();
  const utils = render(
    <MantineProvider>
      <QueryClientProvider client={queryClient}>
        <SetStandardApplicationForm
          enterpriseCapabilityId={toEnterpriseCapabilityId('ec-1')}
          onSubmitted={onSubmitted}
          onCancel={onCancel}
          {...props}
        />
      </QueryClientProvider>
    </MantineProvider>,
  );
  return { ...utils, onSubmitted, onCancel };
}

describe('SetStandardApplicationForm', () => {
  it.each([
    {
      name: 'narrative is empty even when an application is picked',
      fillFields: async (user: ReturnType<typeof userEvent.setup>) => {
        await user.click(screen.getByTestId('standard-application-picker'));
        await user.click(await screen.findByText('Acme ERP'));
      },
      errorMessage: /narrative is required/i,
    },
    {
      name: 'application is unpicked even when a narrative is typed',
      fillFields: async (user: ReturnType<typeof userEvent.setup>) => {
        await user.type(screen.getByTestId('standard-application-narrative'), 'covers the operational layer');
      },
      errorMessage: /pick an application/i,
    },
  ])('rejects the submit when $name', async ({ fillFields, errorMessage }) => {
    const user = userEvent.setup();
    renderForm();

    await fillFields(user);
    await user.click(screen.getByTestId('standard-application-submit'));

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
    expect(mockedSet).not.toHaveBeenCalled();
  });

  it('submits the picked application and trimmed narrative on save', async () => {
    const user = userEvent.setup();
    mockedSet.mockResolvedValueOnce({} as never);
    const { onSubmitted } = renderForm();

    await user.click(screen.getByTestId('standard-application-picker'));
    await user.click(await screen.findByText('Beta Suite'));
    await user.type(screen.getByTestId('standard-application-narrative'), '  covers reporting  ');
    await user.click(screen.getByTestId('standard-application-submit'));

    await waitFor(() => {
      expect(mockedSet).toHaveBeenCalledWith('ec-1', {
        applicationId: 'app-b',
        narrative: 'covers reporting',
      });
    });
    await waitFor(() => expect(onSubmitted).toHaveBeenCalled());
  });

  it('invokes onCancel when the user clicks Cancel', async () => {
    const user = userEvent.setup();
    const { onCancel } = renderForm();

    await user.click(screen.getByRole('button', { name: /cancel/i }));

    expect(onCancel).toHaveBeenCalled();
  });

  it('keeps Save enabled on open when change-flow defaults are valid', async () => {
    renderForm({ initialApplicationId: 'app-a', initialNarrative: 'previously set narrative' });

    const submit = await screen.findByTestId('standard-application-submit');
    expect(submit).not.toBeDisabled();
  });
});
