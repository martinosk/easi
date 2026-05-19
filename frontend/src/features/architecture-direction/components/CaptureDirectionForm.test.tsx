import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { toEnterpriseCapabilityId } from '../../../api/types';

vi.mock('../../enterprise-architecture/hooks/useEnterpriseCapabilities', () => ({
  useEnterpriseCapability: vi.fn(),
  useEnterpriseCapabilityLinks: vi.fn(),
}));

vi.mock('../../business-domains/hooks/useBusinessDomains', () => ({
  useBusinessDomainsQuery: vi.fn(),
}));

vi.mock('../hooks/useDirection', () => ({
  useCaptureDirection: vi.fn(),
}));

import {
  useEnterpriseCapability,
  useEnterpriseCapabilityLinks,
} from '../../enterprise-architecture/hooks/useEnterpriseCapabilities';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { useCaptureDirection } from '../hooks/useDirection';
import { CaptureDirectionForm } from './CaptureDirectionForm';

const mockedEC = vi.mocked(useEnterpriseCapability);
const mockedLinks = vi.mocked(useEnterpriseCapabilityLinks);
const mockedDomains = vi.mocked(useBusinessDomainsQuery);
const mockedCapture = vi.mocked(useCaptureDirection);

function setupHooks(ecName: string) {
  mockedEC.mockReturnValue({ data: { id: 'ec-1', name: ecName } } as never);
  mockedLinks.mockReturnValue({ data: [], isLoading: false } as never);
  mockedDomains.mockReturnValue({
    data: { data: [{ id: 'dom-1', name: 'Passenger' }] },
    isLoading: false,
  } as never);
  mockedCapture.mockReturnValue({ mutateAsync: vi.fn(), isPending: false } as never);
}

function renderForm() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <CaptureDirectionForm
        enterpriseCapabilityId={toEnterpriseCapabilityId('ec-1')}
        onCaptured={() => undefined}
        onCancel={() => undefined}
      />
    </QueryClientProvider>,
  );
}

describe('CaptureDirectionForm', () => {
  it('pre-fills resulting name from the Enterprise Capability when a placement is added', async () => {
    setupHooks('Customer Service');
    renderForm();

    const user = userEvent.setup();
    await user.click(screen.getByTestId('add-placement'));

    const nameInput = screen.getByLabelText(/Resulting name for placement 1/i) as HTMLInputElement;
    expect(nameInput.value).toBe('Customer Service');
  });

  it('hides the add-placement button once a consolidate direction has one placement', async () => {
    setupHooks('Customer Service');
    renderForm();

    const user = userEvent.setup();
    await user.click(screen.getByTestId('add-placement'));

    expect(screen.queryByTestId('add-placement')).not.toBeInTheDocument();
  });
});
