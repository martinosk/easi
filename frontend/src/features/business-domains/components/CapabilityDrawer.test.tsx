import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Capability, CapabilityRealization } from '../../../api/types';
import { toCapabilityId, toComponentId } from '../../../api/types';
import { buildBusinessDomain, buildCapabilityRealization, renderWithProviders, seedDb } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import type {
  RealizationRoleAssignment,
  TimeAssessment,
  TimeAssessmentGradeCounts,
  TimeGrade,
  TimeSuggestion,
} from '../../architecture-direction/types';
import { NO_HIERARCHY_JOURNEYS } from '../lens/hierarchyJourneys';
import { CapabilityDrawer } from './CapabilityDrawer';

vi.mock('../../../hooks/useStrategyPillarsSettings', () => ({
  useStrategyPillarsConfig: () => ({ data: { data: [] } }),
}));

vi.mock('../hooks/useStrategyImportance', () => ({
  useStrategyImportanceByDomainAndCapability: () => ({ data: { data: [] }, isLoading: false }),
  useStrategyImportanceByCapability: () => ({ data: [], isLoading: false }),
  useSetStrategyImportance: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateStrategyImportance: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRemoveStrategyImportance: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

const mockGetAssessment = vi.fn<(componentId: string) => TimeAssessment | undefined>();
const mockGetRollup = vi.fn<(componentId: string) => TimeAssessmentGradeCounts | undefined>();
const mockGetSuggestion = vi.fn<(componentId: string) => TimeSuggestion | null>();
let mockCanAssess = false;

vi.mock('../hooks/useCapabilityAssessments', () => ({
  useCapabilityAssessments: () => ({
    getAssessment: mockGetAssessment,
    getRollup: mockGetRollup,
    getSuggestion: mockGetSuggestion,
    canAssess: mockCanAssess,
  }),
}));

vi.mock('../../architecture-direction/hooks/useTimeAssessments', () => ({
  useAssessRealization: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRemoveTimeAssessment: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

const mockGetRole = vi.fn<(componentId: string) => RealizationRoleAssignment | undefined>();
let mockCanAssign = false;

vi.mock('../hooks/useCapabilityRoles', () => ({
  useCapabilityRoles: () => ({
    getRole: mockGetRole,
    canAssign: mockCanAssign,
  }),
}));

vi.mock('../../architecture-direction/hooks/useRealizationRoles', () => ({
  useAssignRealizationRole: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useClearRealizationRole: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../journeys/hooks/useJourneys', () => ({
  useJourneyForCapability: () => ({ data: { journey: null, _links: {} } }),
  useJourneyHistory: () => ({ data: undefined }),
  useCaptureJourney: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useStartJourney: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useCompleteJourney: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useAbandonJourney: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateJourneyDetails: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateJourneyProgress: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useChangeJourneySourceApplications: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useAddJourneyMilestone: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateJourneyMilestone: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRemoveJourneyMilestone: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

function buildRole(overrides: Partial<RealizationRoleAssignment> = {}): RealizationRoleAssignment {
  return {
    capabilityId: 'l2-a',
    capabilityName: 'Booking Management',
    componentId: 'comp-1',
    componentName: 'Phoenix',
    role: 'standard',
    assignedBy: 'user-1',
    assignedAt: '2026-02-01T00:00:00Z',
    _links: { self: { href: '', method: 'GET' } },
    ...overrides,
  };
}

function buildAssessment(overrides: Partial<TimeAssessment> = {}): TimeAssessment {
  return {
    id: 'ta-1',
    capabilityId: 'l2-a',
    capabilityName: 'Booking Management',
    componentId: 'comp-1',
    componentName: 'Phoenix',
    grade: 'Migrate' as TimeGrade,
    rationale: '',
    assessedBy: 'user-1',
    assessedByName: 'Domain Architect',
    assessedAt: '2026-02-01T00:00:00Z',
    stale: false,
    suggestion: null,
    _links: { self: { href: '', method: 'GET' } },
    ...overrides,
  };
}

const domain = buildBusinessDomain({ name: 'Ferry Freight' });

function realizationFor(overrides: Partial<CapabilityRealization> = {}): CapabilityRealization {
  return buildCapabilityRealization({
    capabilityId: toCapabilityId('l2-a'),
    componentId: toComponentId('comp-1'),
    componentName: 'Phoenix',
    origin: 'Direct',
    ...overrides,
  });
}

interface DrawerOptions {
  realizations?: CapabilityRealization[];
  onChipClick?: (componentId: string) => void;
}

function renderDrawer(capability: Capability | null, { realizations = [], onChipClick = vi.fn() }: DrawerOptions = {}) {
  seedDb({ capabilities: capability ? [capability] : [], components: [], capabilityRealizations: realizations });
  return renderWithProviders(
    <CapabilityDrawer
      capability={capability}
      domain={domain}
      l1Name="Ferry Booking"
      getRealizationsForCapability={() => realizations}
      hierarchyJourneys={NO_HIERARCHY_JOURNEYS}
      onClose={vi.fn()}
      onChipClick={onChipClick}
      onNavigateToCapability={vi.fn()}
    />,
  );
}

describe('CapabilityDrawer', () => {
  beforeEach(() => {
    mockGetAssessment.mockReset().mockReturnValue(undefined);
    mockGetRollup.mockReset().mockReturnValue(undefined);
    mockCanAssess = false;
    mockGetSuggestion.mockReset().mockReturnValue(null);
    mockGetRole.mockReset().mockReturnValue(undefined);
    mockCanAssign = false;
  });

  it('renders no capability content when no capability is selected', () => {
    renderDrawer(null);

    expect(screen.queryByRole('heading')).not.toBeInTheDocument();
    expect(screen.queryByText('no realising application mapped')).not.toBeInTheDocument();
  });

  it('renders the breadcrumb, the shared panel with one heading, and the empty realisations state', async () => {
    renderDrawer(cap('l2-a', 'Booking Management', 'L2'));

    expect(await screen.findByRole('heading', { name: 'Booking Management' })).toBeInTheDocument();
    expect(screen.getByTestId('capability-drawer')).toHaveTextContent('Ferry Freight');
    expect(screen.getByTestId('capability-drawer')).toHaveTextContent('Ferry Booking');
    expect(screen.getAllByRole('heading', { name: 'Booking Management' })).toHaveLength(1);
    expect(screen.getByText('L2')).toBeInTheDocument();
    expect(screen.getByText('no realising application mapped')).toBeInTheDocument();
    expect(screen.queryByText('Capability Details')).not.toBeInTheDocument();
  });

  it('never offers a whole-record Edit action', async () => {
    renderDrawer(cap('l2-a', 'Booking Management', 'L2'));

    await screen.findByRole('heading', { name: 'Booking Management' });
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Edit name' })).toBeInTheDocument();
  });

  it('renders the journey and strategic importance sections above the capability fields', async () => {
    renderDrawer(cap('l2-a', 'Booking Management', 'L2'));

    const heading = await screen.findByRole('heading', { name: 'Booking Management' });
    const journey = screen.getByTestId('journey-section');
    expect(journey.compareDocumentPosition(heading) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(screen.getByText('Transition')).toBeInTheDocument();
    expect(screen.getByText('No change planned.')).toBeInTheDocument();
  });

  it('renders a row per realising application with level, origin note, and notes', async () => {
    renderDrawer(cap('l2-a', 'Booking Management', 'L2'), {
      realizations: [
        realizationFor({ realizationLevel: 'Full', notes: 'Primary booking engine' }),
        realizationFor({
          componentId: toComponentId('comp-2'),
          componentName: 'Seabook',
          realizationLevel: 'Partial',
          origin: 'Inherited',
          sourceCapabilityName: 'Ferry Booking',
        }),
      ],
    });

    const rows = await screen.findAllByTestId(/^drawer-realization-/);
    expect(rows[0]).toHaveTextContent('Primary booking engine');
    expect(rows[1]).toHaveTextContent('Inherited from Ferry Booking');
  });

  it('calls onChipClick when a realising application chip is clicked', async () => {
    const onChipClick = vi.fn();
    renderDrawer(cap('l2-a', 'Booking Management', 'L2'), { realizations: [realizationFor()], onChipClick });

    await userEvent.click(await screen.findByTestId('app-chip-comp-1'));

    expect(onChipClick).toHaveBeenCalledWith('comp-1');
  });

  it('renders description, tags and owners', async () => {
    renderDrawer({
      ...cap('l2-a', 'Booking Management', 'L2'),
      description: 'Handles route bookings',
      primaryOwner: 'Jane Doe',
      tags: ['core', 'freight'],
    });

    expect(await screen.findByText('Handles route bookings')).toBeInTheDocument();
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    expect(screen.getByText('core')).toBeInTheDocument();
    expect(screen.getByText('freight')).toBeInTheDocument();
  });

  it('renders the EA owner display name instead of the stored user id', async () => {
    renderDrawer({
      ...cap('l2-a', 'Booking Management', 'L2'),
      eaOwner: '2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b',
      eaOwnerName: 'Alice Smith',
    });

    expect(await screen.findByText('Alice Smith')).toBeInTheDocument();
    expect(screen.queryByText('2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b')).not.toBeInTheDocument();
  });

  it('falls back to the stored EA owner value when no display name is provided', async () => {
    renderDrawer({ ...cap('l2-a', 'Booking Management', 'L2'), eaOwner: 'Bob Jones' });

    expect(await screen.findByText('Bob Jones')).toBeInTheDocument();
  });

  describe('TIME assessment on a Direct realising application', () => {
    function renderAssessmentDrawer(realizationOverrides: Partial<CapabilityRealization> = {}) {
      renderDrawer(cap('l2-a', 'Booking Management', 'L2'), { realizations: [realizationFor(realizationOverrides)] });
    }

    it('shows the current grade, assessor, and date for an assessed realization', async () => {
      mockGetAssessment.mockReturnValue(buildAssessment());

      renderAssessmentDrawer();

      expect(await screen.findByTestId('assessment-comp-1')).toHaveTextContent('Migrate — for this capability');
      expect(screen.getByTestId('assessment-comp-1')).toHaveTextContent('Domain Architect');
    });

    it('shows an explicit unassessed state for a Direct realization with no assessment', async () => {
      renderAssessmentDrawer();

      expect(await screen.findByTestId('assessment-comp-1')).toHaveTextContent('unassessed');
    });

    it('shows the stale marker and landscape rollup line for an assessed realization', async () => {
      mockGetAssessment.mockReturnValue(buildAssessment({ stale: true }));
      mockGetRollup.mockReturnValue({ Invest: 1, Tolerate: 1, Migrate: 1, Eliminate: 1 });

      renderAssessmentDrawer();

      expect(await screen.findByTestId('assessment-stale-comp-1')).toHaveTextContent('stale');
      expect(screen.getByTestId('assessment-rollup-comp-1')).toHaveTextContent(
        'Across landscape: I×1 · T×1 · M×1 · E×1',
      );
    });

    it('shows no assess/remove control for a read-only caller with no write links', async () => {
      mockGetAssessment.mockReturnValue(buildAssessment({ _links: { self: { href: '', method: 'GET' } } }));
      mockCanAssess = false;

      renderAssessmentDrawer();

      expect(await screen.findByTestId('assessment-comp-1')).toBeInTheDocument();
      expect(screen.queryByTestId('reassess-btn-comp-1')).not.toBeInTheDocument();
      expect(screen.queryByTestId('remove-assessment-btn-comp-1')).not.toBeInTheDocument();
      expect(screen.queryByTestId('assess-btn-comp-1')).not.toBeInTheDocument();
    });

    it('shows no assessment UI at all for an Inherited realization', async () => {
      mockGetAssessment.mockReturnValue(buildAssessment());

      renderAssessmentDrawer({ origin: 'Inherited', sourceCapabilityName: 'Ferry Booking' });

      await screen.findByTestId('app-chip-comp-1');
      expect(screen.queryByTestId('assessment-comp-1')).not.toBeInTheDocument();
    });

    it('shows the computed suggestion for the realized component beside the grade choices without pre-selecting it', async () => {
      mockCanAssess = true;
      mockGetSuggestion.mockReturnValue({
        grade: 'Eliminate',
        confidence: 'HIGH',
        technicalGap: 2.5,
        functionalGap: 2,
      });

      renderAssessmentDrawer();

      expect(await screen.findByTestId('assessment-suggestion-comp-1')).toHaveTextContent('Suggested: Eliminate');

      await userEvent.click(screen.getByTestId('assess-btn-comp-1'));

      expect(screen.getByTestId('assessment-suggestion-comp-1')).toHaveTextContent('high confidence');
      expect(screen.getByRole('radio', { name: 'Eliminate' })).not.toBeChecked();
    });
  });

  describe('Realization role on a Direct realising application', () => {
    function renderRoleDrawer(realizationOverrides: Partial<CapabilityRealization> = {}) {
      renderDrawer(cap('l2-a', 'Booking Management', 'L2'), { realizations: [realizationFor(realizationOverrides)] });
    }

    it('shows the current role badge for a classified Direct realization', async () => {
      mockGetRole.mockReturnValue(buildRole({ role: 'standard' }));

      renderRoleDrawer();

      expect(await screen.findByTestId('role-badge-comp-1')).toHaveTextContent('standard');
    });

    it('shows no role UI for an unclassified realization when the caller cannot assign', async () => {
      mockCanAssign = false;

      renderRoleDrawer();

      await screen.findByTestId('app-chip-comp-1');
      expect(screen.queryByTestId('role-comp-1')).not.toBeInTheDocument();
    });

    it('shows assign controls for an unclassified realization when the caller can assign', async () => {
      mockCanAssign = true;

      renderRoleDrawer();

      expect(await screen.findByTestId('assign-standard-btn-comp-1')).toBeInTheDocument();
      expect(screen.getByTestId('assign-legacy-btn-comp-1')).toBeInTheDocument();
    });

    it('shows no role UI at all for an Inherited realization', async () => {
      mockGetRole.mockReturnValue(buildRole({ role: 'standard' }));

      renderRoleDrawer({ origin: 'Inherited', sourceCapabilityName: 'Ferry Booking' });

      await screen.findByTestId('app-chip-comp-1');
      expect(screen.queryByTestId('role-comp-1')).not.toBeInTheDocument();
      expect(screen.queryByTestId('role-badge-comp-1')).not.toBeInTheDocument();
    });
  });
});
