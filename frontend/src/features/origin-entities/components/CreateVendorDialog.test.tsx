import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { CreateVendorDialog } from './CreateVendorDialog';

const mockMutateAsync = vi.fn();

vi.mock('../hooks/useVendors', () => ({
  useCreateVendor: () => ({ mutateAsync: mockMutateAsync, isPending: false }),
}));

describe('CreateVendorDialog onCreated handoff', () => {
  let qc: QueryClient;
  beforeEach(() => {
    vi.clearAllMocks();
    qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  });

  const renderDialog = (props: { onCreated?: (v: { id: string; name: string }) => void | Promise<void> } = {}) =>
    render(
      <QueryClientProvider client={qc}>
        <CreateVendorDialog isOpen onClose={vi.fn()} onCreated={props.onCreated} />
      </QueryClientProvider>,
      { wrapper: MantineTestWrapper },
    );

  it('passes the new vendor to onCreated when supplied', async () => {
    mockMutateAsync.mockResolvedValueOnce({ id: 'vendor-1', name: 'SAP' });
    const onCreated = vi.fn();

    renderDialog({ onCreated });

    fireEvent.change(screen.getByTestId('vendor-name-input'), { target: { value: 'SAP' } });
    const submit = screen.getByTestId('create-vendor-submit') as HTMLButtonElement;
    await waitFor(() => expect(submit.disabled).toBe(false));
    fireEvent.click(submit);

    await waitFor(() => {
      expect(onCreated).toHaveBeenCalledWith(expect.objectContaining({ id: 'vendor-1' }));
    });
  });

  it('keeps the submit button disabled while onCreated is still running', async () => {
    mockMutateAsync.mockResolvedValueOnce({ id: 'vendor-1', name: 'SAP' });
    let resolveHandoff!: () => void;
    const onCreated = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveHandoff = resolve;
        }),
    );

    renderDialog({ onCreated });

    fireEvent.change(screen.getByTestId('vendor-name-input'), { target: { value: 'SAP' } });
    const submit = screen.getByTestId('create-vendor-submit') as HTMLButtonElement;
    await waitFor(() => expect(submit.disabled).toBe(false));
    fireEvent.click(submit);

    await waitFor(() => expect(onCreated).toHaveBeenCalled());
    expect(submit.disabled).toBe(true);

    resolveHandoff();
    await waitFor(() => expect(onCreated).toHaveBeenCalledTimes(1));
  });

  it('preserves form input when the create call fails', async () => {
    mockMutateAsync.mockRejectedValueOnce(new Error('boom'));

    renderDialog();

    const nameInput = screen.getByTestId('vendor-name-input') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'SAP' } });
    const submit = screen.getByTestId('create-vendor-submit') as HTMLButtonElement;
    await waitFor(() => expect(submit.disabled).toBe(false));
    fireEvent.click(submit);

    await waitFor(() => expect(screen.getByTestId('create-vendor-error')).toBeInTheDocument());
    expect(nameInput.value).toBe('SAP');
  });
});
