import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ReactFlowProvider } from '@xyflow/react';
import { MantineProvider } from '@mantine/core';
import type { ReactNode } from 'react';
import { useAppStore } from '../../../store/appStore';
import { withDynamicExpansion } from './withDynamicExpansion';

function InnerStub({ data }: { data: { label: string }; id: string; selected?: boolean }) {
  return <div data-testid="inner-node">{data.label}</div>;
}

function Wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return (
    <QueryClientProvider client={qc}>
      <MantineProvider>
        <ReactFlowProvider>{children}</ReactFlowProvider>
      </MantineProvider>
    </QueryClientProvider>
  );
}

describe('withDynamicExpansion', () => {
  it('renders the inner node directly when dynamic mode is disabled', () => {
    useAppStore.setState({ dynamicEnabled: false });
    const Wrapped = withDynamicExpansion(InnerStub);

    render(
      <Wrapper>
        <Wrapped data={{ label: 'Order Service' }} id="comp-1" />
      </Wrapper>,
    );

    expect(screen.getByTestId('inner-node')).toHaveTextContent('Order Service');
    expect(screen.queryByLabelText(/Expand Order Service/i)).not.toBeInTheDocument();
  });

  it('does not enter an infinite render loop when disabled', () => {
    useAppStore.setState({ dynamicEnabled: false });
    let renderCount = 0;
    function CountingInner(props: { data: { label: string }; id: string; selected?: boolean }) {
      renderCount++;
      return <div>{props.data.label}</div>;
    }
    const Wrapped = withDynamicExpansion(CountingInner);

    render(
      <Wrapper>
        <Wrapped data={{ label: 'X' }} id="comp-1" />
      </Wrapper>,
    );

    expect(renderCount).toBeLessThan(5);
  });

  it('does not enter an infinite render loop when enabled with no neighbors', async () => {
    act(() => {
      useAppStore.setState({
        dynamicEnabled: true,
        dynamicEntities: [{ id: 'comp-1', type: 'component' }],
        dynamicOriginal: { entities: [{ id: 'comp-1', type: 'component' }], positions: { 'comp-1': { x: 0, y: 0 } } },
      });
    });
    let renderCount = 0;
    function CountingInner(props: { data: { label: string }; id: string; selected?: boolean }) {
      renderCount++;
      return <div>{props.data.label}</div>;
    }
    const Wrapped = withDynamicExpansion(CountingInner);

    render(
      <Wrapper>
        <Wrapped data={{ label: 'X' }} id="comp-1" />
      </Wrapper>,
    );

    await waitFor(() => expect(renderCount).toBeGreaterThan(0));
    expect(renderCount).toBeLessThan(10);

    act(() => {
      useAppStore.setState({ dynamicEnabled: false, dynamicOriginal: null, dynamicEntities: [] });
    });
  });
});
