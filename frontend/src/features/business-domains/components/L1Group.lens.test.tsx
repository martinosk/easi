import { screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { toCapabilityId } from '../../../api/types';
import { buildCapabilityJourney, renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import type { CapabilityJourney } from '../../journeys/types';
import type { BoardLens } from '../lens/boardLens';
import { buildJourneyIndex } from '../lens/journeyIndex';
import { BoardLensProvider } from './BoardLensContext';
import { L1Group } from './L1Group';

const groupProps = {
  distinctAppCount: 2,
  searchQuery: '',
  selectedCapabilities: new Set<ReturnType<typeof toCapabilityId>>(),
  getColorForValue: () => '#000',
  getRealizationsForCapability: () => [],
  onCapabilityClick: vi.fn(),
  onCapabilityContextMenu: vi.fn(),
  onChipClick: vi.fn(),
};

function renderGroup(
  node: ReturnType<typeof buildCapabilityTree>[0],
  lens: BoardLens,
  journeys: CapabilityJourney[],
  changesOnly = false,
) {
  const index = buildJourneyIndex({ journeys, capabilityDomainNames: new Map() });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <BoardLensProvider lens={lens} changesOnly={changesOnly} index={index} openCapabilityById={vi.fn()}>
      {children}
    </BoardLensProvider>
  );
  return renderWithProviders(wrapper({ children: <L1Group node={node} {...groupProps} distinctAppCount={0} /> }), {
    withRouter: false,
  });
}

const moveJourney = (capabilityId: string) =>
  buildCapabilityJourney({
    capabilityId,
    kind: 'move',
    status: 'planned',
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

describe('L1Group — lens behaviour', () => {
  it('hides a childless move L1 entirely in the Target lens', () => {
    const [node] = buildCapabilityTree([cap('inv', 'Invoicing', 'L1')], { orphanRoots: 'any-level' });
    const { container } = renderGroup(node, 'target', [moveJourney('inv')]);
    expect(container.querySelector('[data-testid="l1-group-inv"]')).toBeNull();
  });

  it('renders the childless move L1 as a ghost in the Journey lens', () => {
    const [node] = buildCapabilityTree([cap('inv', 'Invoicing', 'L1')], { orphanRoots: 'any-level' });
    renderGroup(node, 'journey', [moveJourney('inv')]);
    expect(screen.getByTestId('move-ghost-inv')).toBeInTheDocument();
  });

  it('dims a group with no change and force-opens a group with a change under changes-only', () => {
    const [changed] = buildCapabilityTree(
      [cap('fb', 'Ferry booking', 'L1'), cap('bm', 'Booking management', 'L2', 'fb')],
      { orphanRoots: 'any-level' },
    );
    const [steady] = buildCapabilityTree(
      [cap('cc', 'Customs compliance', 'L1'), cap('cl', 'Customs clearance', 'L2', 'cc')],
      { orphanRoots: 'any-level' },
    );
    const journeys = [buildCapabilityJourney({ capabilityId: 'bm', kind: 'migration', status: 'in-flight' })];

    const { rerender } = renderGroup(changed, 'journey', journeys, true);
    // changed group is force-expanded and not dimmed
    expect(screen.getByTestId('capability-card-bm')).toBeInTheDocument();
    expect(screen.getByTestId('l1-group-fb')).not.toHaveAttribute('data-dimmed');

    const index = buildJourneyIndex({ journeys, capabilityDomainNames: new Map() });
    rerender(
      <BoardLensProvider lens="journey" changesOnly index={index} openCapabilityById={vi.fn()}>
        <L1Group node={steady} {...groupProps} distinctAppCount={0} />
      </BoardLensProvider>,
    );
    expect(screen.getByTestId('l1-group-cc')).toHaveAttribute('data-dimmed', 'true');
  });
});
