import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { CreateAcquiredEntityDialog } from './CreateAcquiredEntityDialog';

const mockMutateAsync = vi.fn();

vi.mock('../hooks/useAcquiredEntities', () => ({
  useCreateAcquiredEntity: () => ({ mutateAsync: mockMutateAsync, isPending: false }),
}));

describe('CreateAcquiredEntityDialog onCreated handoff', () => {
  let qc: QueryClient;
  beforeEach(() => {
    vi.clearAllMocks();
    qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  });

  it('passes the new entity to onCreated when supplied', async () => {
    mockMutateAsync.mockResolvedValueOnce({ id: 'acq-1', name: 'Acme' });
    const onCreated = vi.fn();
    const onClose = vi.fn();

    render(
      <QueryClientProvider client={qc}>
        <CreateAcquiredEntityDialog isOpen onClose={onClose} onCreated={onCreated} />
      </QueryClientProvider>,
      { wrapper: MantineTestWrapper },
    );

    fireEvent.change(screen.getByTestId('acquired-entity-name-input'), { target: { value: 'Acme' } });
    const submit = screen.getByTestId('create-acquired-entity-submit') as HTMLButtonElement;
    await waitFor(() => expect(submit.disabled).toBe(false));
    fireEvent.click(submit);

    await waitFor(() => {
      expect(onCreated).toHaveBeenCalledWith(expect.objectContaining({ id: 'acq-1' }));
    });
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });
});
