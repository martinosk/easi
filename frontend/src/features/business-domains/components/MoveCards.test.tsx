import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { buildCapabilityJourney, renderWithProviders } from '../../../test/helpers';
import type { CapabilityJourney } from '../../journeys/types';
import type { BoardLens } from '../lens/boardLens';
import { buildJourneyIndex } from '../lens/journeyIndex';
import { BoardLensProvider } from './BoardLensContext';
import { ArrivingMoves, GhostCard } from './MoveCards';

function moveJourney(): CapabilityJourney {
  return buildCapabilityJourney({
    id: 'mv-inv',
    capabilityId: 'inv',
    capabilityName: 'Invoicing',
    kind: 'move',
    status: 'planned',
    progress: null,
    targetPeriod: { year: 2027, quarter: 1 },
    toApplication: { componentId: 'sap', componentName: 'SAP S/4', stale: false },
    move: {
      targetDomainId: 'gf',
      targetDomainName: 'Group functions',
      targetDomainStale: false,
      targetParentId: 'ap',
      targetParentName: 'Accounts payable',
      targetParentStale: false,
      resultingName: 'Freight invoicing',
    },
  });
}

function renderMove(lens: BoardLens, openCapabilityById = vi.fn()) {
  const journey = moveJourney();
  const index = buildJourneyIndex({
    journeys: [journey],
    capabilityDomainNames: new Map([['inv', 'Ferry freight']]),
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <BoardLensProvider lens={lens} changesOnly={false} index={index} openCapabilityById={openCapabilityById}>
      {children}
    </BoardLensProvider>
  );
  return renderWithProviders(
    wrapper({
      children: (
        <>
          <GhostCard journey={journey} realizations={[]} onChipClick={vi.fn()} />
          <ArrivingMoves journeys={[journey]} />
        </>
      ),
    }),
    { withRouter: false },
  );
}

describe('MoveCards — Journey lens', () => {
  it('renders a ghost "moving out" card with a trace link to the destination domain', () => {
    renderMove('journey');
    const ghost = screen.getByTestId('move-ghost-inv');
    expect(ghost).toHaveTextContent('moving out');
    expect(screen.getByTestId('move-trace-source-mv-inv')).toHaveTextContent('to Group functions');
  });

  it('renders an arriving card with the target app and a trace link from the source domain', () => {
    renderMove('journey');
    const arriving = screen.getByTestId('move-arriving-inv');
    expect(arriving).toHaveTextContent('Freight invoicing');
    expect(arriving).toHaveTextContent('arriving Q1 2027');
    expect(screen.getByTestId('move-trace-dest-mv-inv')).toHaveTextContent('from Ferry freight');
  });

  it('highlights both ends when a trace link is activated', async () => {
    renderMove('journey');
    expect(screen.getByTestId('move-ghost-inv')).not.toHaveAttribute('data-traced');

    await userEvent.click(screen.getByTestId('move-trace-source-mv-inv'));

    expect(screen.getByTestId('move-ghost-inv')).toHaveAttribute('data-traced', 'true');
    expect(screen.getByTestId('move-arriving-inv')).toHaveAttribute('data-traced', 'true');
  });

  it('opens the capability drawer for the moving capability when the ghost card is clicked', async () => {
    const openCapabilityById = vi.fn();
    renderMove('journey', openCapabilityById);

    await userEvent.click(screen.getByTestId('move-ghost-inv'));

    expect(openCapabilityById).toHaveBeenCalledWith('inv');
  });
});

describe('MoveCards — Target lens', () => {
  it('renders the arriving card as a moved-from card with the target app, no ghost', () => {
    renderMove('target');
    const arriving = screen.getByTestId('move-arriving-inv');
    expect(arriving).toHaveTextContent('Freight invoicing');
    expect(arriving).toHaveTextContent('SAP S/4');
    expect(arriving).toHaveTextContent('moved from Ferry freight');
  });
});
