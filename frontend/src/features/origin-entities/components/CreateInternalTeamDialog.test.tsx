import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { CreateInternalTeamDialog } from './CreateInternalTeamDialog';

const mockMutateAsync = vi.fn();

vi.mock('../hooks/useInternalTeams', () => ({
  useCreateInternalTeam: () => ({ mutateAsync: mockMutateAsync, isPending: false }),
}));

describe('CreateInternalTeamDialog onCreated handoff', () => {
  let qc: QueryClient;
  beforeEach(() => {
    vi.clearAllMocks();
    qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  });

  const renderDialog = (props: { onCreated?: (t: { id: string; name: string }) => void | Promise<void> } = {}) =>
    render(
      <QueryClientProvider client={qc}>
        <CreateInternalTeamDialog isOpen onClose={vi.fn()} onCreated={props.onCreated} />
      </QueryClientProvider>,
      { wrapper: MantineTestWrapper },
    );

  it('passes the new internal team to onCreated when supplied', async () => {
    mockMutateAsync.mockResolvedValueOnce({ id: 'team-1', name: 'Platform' });
    const onCreated = vi.fn();

    renderDialog({ onCreated });

    fireEvent.change(screen.getByTestId('internal-team-name-input'), { target: { value: 'Platform' } });
    const submit = screen.getByTestId('create-internal-team-submit') as HTMLButtonElement;
    await waitFor(() => expect(submit.disabled).toBe(false));
    fireEvent.click(submit);

    await waitFor(() => {
      expect(onCreated).toHaveBeenCalledWith(expect.objectContaining({ id: 'team-1' }));
    });
  });

  it('keeps the submit button disabled while onCreated is still running', async () => {
    mockMutateAsync.mockResolvedValueOnce({ id: 'team-1', name: 'Platform' });
    let resolveHandoff!: () => void;
    const onCreated = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          resolveHandoff = resolve;
        }),
    );

    renderDialog({ onCreated });

    fireEvent.change(screen.getByTestId('internal-team-name-input'), { target: { value: 'Platform' } });
    const submit = screen.getByTestId('create-internal-team-submit') as HTMLButtonElement;
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

    const nameInput = screen.getByTestId('internal-team-name-input') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'Platform' } });
    const submit = screen.getByTestId('create-internal-team-submit') as HTMLButtonElement;
    await waitFor(() => expect(submit.disabled).toBe(false));
    fireEvent.click(submit);

    await waitFor(() => expect(screen.getByTestId('create-internal-team-error')).toBeInTheDocument());
    expect(nameInput.value).toBe('Platform');
  });
});
