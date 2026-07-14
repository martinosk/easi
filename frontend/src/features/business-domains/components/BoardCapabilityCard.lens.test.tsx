import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import type { CapabilityRealization, RealizationLevel, TimeGrade } from '../../../api/types';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { buildCapabilityJourney, buildCapabilityRealization, renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import type { RealizationRole } from '../../architecture-direction/types';
import { buildCapabilityTree } from '../../capabilities/hooks/useCapabilityTree';
import type { CapabilityJourney } from '../../journeys/types';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import type { BoardLens } from '../lens/boardLens';
import { buildJourneyIndex } from '../lens/journeyIndex';
import { BoardCapabilityCard } from './BoardCapabilityCard';
import { BoardLensProvider } from './BoardLensContext';

interface AssessedInput {
  componentId: string;
  componentName?: string;
  realizationLevel?: RealizationLevel;
  origin?: 'Direct' | 'Inherited';
  role?: RealizationRole;
  timeGrade?: TimeGrade;
}

function assessed(input: AssessedInput): AssessedRealization {
  return {
    ...buildCapabilityRealization({
      capabilityId: toCapabilityId('bm'),
      componentId: toComponentId(input.componentId),
      componentName: input.componentName ?? input.componentId,
      realizationLevel: input.realizationLevel ?? 'Full',
      origin: input.origin ?? 'Direct',
    }),
    role: input.role,
    timeGrade: input.timeGrade,
  };
}

interface RenderOptions {
  lens: BoardLens;
  journeys?: CapabilityJourney[];
  realizationsByCapability?: Record<string, AssessedRealization[]>;
  changesOnly?: boolean;
  node?: ReturnType<typeof buildCapabilityTree>[0];
  openCapabilityById?: (capabilityId: string) => void;
}

function bookingNode() {
  const [node] = buildCapabilityTree([cap('bm', 'Booking management', 'L2')], { orphanRoots: 'any-level' });
  return node;
}

function renderCard({
  lens,
  journeys = [],
  realizationsByCapability = {},
  changesOnly = false,
  node = bookingNode(),
  openCapabilityById = vi.fn(),
}: RenderOptions) {
  const index = buildJourneyIndex({
    journeys,
    capabilityDomainNames: new Map([['bm', 'Ferry freight']]),
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <BoardLensProvider lens={lens} changesOnly={changesOnly} index={index} openCapabilityById={openCapabilityById}>
      {children}
    </BoardLensProvider>
  );
  return renderWithProviders(
    wrapper({
      children: (
        <BoardCapabilityCard
          node={node}
          isSelected={false}
          getColorForValue={() => '#000'}
          getRealizationsForCapability={(id) => (realizationsByCapability[id] ?? []) as CapabilityRealization[]}
          onClick={vi.fn()}
          onContextMenu={vi.fn()}
          onChipClick={vi.fn()}
        />
      ),
    }),
    { withRouter: false },
  );
}

const migration = () =>
  buildCapabilityJourney({
    id: 'j-bm',
    capabilityId: 'bm',
    kind: 'migration',
    status: 'in-flight',
    progress: 60,
    fromApplications: [{ componentId: 'seabook', componentName: 'Seabook', stale: false }],
    toApplication: { componentId: 'phoenix', componentName: 'Phoenix', stale: false },
  });

describe('BoardCapabilityCard — Journey lens', () => {
  it('renders an active migration with from/to chips, kind, in-flight pill and a 60% progress bar', () => {
    renderCard({ lens: 'journey', journeys: [migration()] });

    const chips = screen.getAllByTestId('plan-chip');
    expect(chips.find((c) => c.textContent === 'Seabook')).toHaveAttribute('data-variant', 'legacy');
    expect(chips.find((c) => c.textContent === 'Phoenix')).toHaveAttribute('data-variant', 'future');
    expect(screen.getByTestId('status-pill-in-flight')).toHaveTextContent('in flight');
    expect(screen.getByTestId('journey-progress-fill')).toHaveStyle('--journey-progress: 60%');
  });

  it('renders a completed journey with the standard chip, a full bar and a done pill', () => {
    const done = buildCapabilityJourney({
      id: 'j-done',
      capabilityId: 'bm',
      kind: 'consolidation',
      status: 'done',
      progress: 100,
      toApplication: { componentId: 'cargoflow', componentName: 'CargoFlow', stale: false },
    });
    renderCard({
      lens: 'journey',
      journeys: [done],
      realizationsByCapability: {
        bm: [assessed({ componentId: 'cargoflow', componentName: 'CargoFlow', role: 'standard' })],
      },
    });

    expect(screen.getByTestId('status-pill-done')).toHaveTextContent('done');
    expect(screen.getByTestId('app-chip-cargoflow')).toBeInTheDocument();
    expect(screen.getByTestId('journey-progress-fill')).toHaveAttribute('data-fill', 'done');
  });

  it('renders a capability with no journey as "no change planned"', () => {
    renderCard({
      lens: 'journey',
      realizationsByCapability: { bm: [assessed({ componentId: 'phoenix', componentName: 'Phoenix' })] },
    });

    expect(screen.getByText('no change planned')).toBeInTheDocument();
    expect(screen.queryByTestId('status-pill-in-flight')).not.toBeInTheDocument();
  });

  it('derives a sub-capability breakdown naming the apps behind each status', async () => {
    const [node] = buildCapabilityTree(
      [
        cap('bm', 'Booking management', 'L2'),
        cap('q', 'Quotation', 'L3', 'bm'),
        cap('bc', 'Booking capture', 'L3', 'bm'),
        cap('ac', 'Amendments', 'L3', 'bm'),
      ],
      { orphanRoots: 'any-level' },
    );
    renderCard({
      lens: 'journey',
      node,
      journeys: [migration()],
      realizationsByCapability: {
        q: [assessed({ componentId: 'phoenix', componentName: 'Phoenix' })],
        bc: [
          assessed({ componentId: 'seabook', componentName: 'Seabook' }),
          assessed({ componentId: 'phoenix', componentName: 'Phoenix', realizationLevel: 'Planned' }),
        ],
        ac: [assessed({ componentId: 'seabook', componentName: 'Seabook' })],
      },
    });

    await userEvent.click(screen.getByTestId('subcap-expander-bm'));

    expect(screen.getByTestId('subcap-q')).toHaveTextContent('Phoenix');
    expect(screen.getByTestId('subcap-bc')).toHaveTextContent('Seabook → Phoenix');
    expect(screen.getByTestId('subcap-ac')).toHaveTextContent('Seabook');
  });
});

describe('BoardCapabilityCard — Target lens', () => {
  it('shows the target app as the standard chip, tagged consolidated for consolidations', () => {
    const consolidation = buildCapabilityJourney({
      capabilityId: 'bm',
      kind: 'consolidation',
      status: 'in-flight',
      toApplication: { componentId: 'notifyhub', componentName: 'Notify Hub', stale: false },
    });
    renderCard({ lens: 'target', journeys: [consolidation] });

    expect(screen.getByTestId('plan-chip')).toHaveTextContent('Notify Hub');
    expect(screen.getByText('consolidated')).toBeInTheDocument();
  });

  it('shows the standard-role chips for a capability without a journey', () => {
    renderCard({
      lens: 'target',
      realizationsByCapability: {
        bm: [
          assessed({ componentId: 'phoenix', componentName: 'Phoenix', role: 'standard' }),
          assessed({ componentId: 'seabook', componentName: 'Seabook', role: 'legacy' }),
        ],
      },
    });

    expect(screen.getByTestId('app-chip-phoenix')).toBeInTheDocument();
    expect(screen.queryByTestId('app-chip-seabook')).not.toBeInTheDocument();
  });

  it('shows "no standard defined" for a capability with no apps and no journey', () => {
    renderCard({ lens: 'target' });
    expect(screen.getByTestId('target-no-standard')).toBeInTheDocument();
  });
});

describe('BoardCapabilityCard — TIME badges are Now-only', () => {
  const graded = () => ({
    bm: [assessed({ componentId: 'phoenix', componentName: 'Phoenix', timeGrade: 'Invest' })],
  });

  it('shows the grade badge in the Now lens', () => {
    renderCard({ lens: 'now', realizationsByCapability: graded() });
    expect(screen.getByTestId('app-chip-grade-phoenix')).toBeInTheDocument();
  });

  it('hides the grade badge in the Journey lens', () => {
    renderCard({ lens: 'journey', realizationsByCapability: graded() });
    expect(screen.queryByTestId('app-chip-grade-phoenix')).not.toBeInTheDocument();
  });
});

describe('BoardCapabilityCard — changes-only dimming', () => {
  it('dims a capability without a change when the toggle is on', () => {
    renderCard({ lens: 'journey', changesOnly: true });
    expect(screen.getByTestId('capability-card-bm')).toHaveAttribute('data-dimmed', 'true');
  });

  it('does not dim a capability with an active journey', () => {
    renderCard({ lens: 'journey', changesOnly: true, journeys: [migration()] });
    expect(screen.getByTestId('capability-card-bm')).not.toHaveAttribute('data-dimmed');
  });
});
