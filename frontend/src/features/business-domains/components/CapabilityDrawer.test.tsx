import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { toComponentId } from '../../../api/types';
import { buildBusinessDomain, buildCapabilityRealization, renderWithProviders } from '../../../test/helpers';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import type {
  RealizationRoleAssignment,
  TimeAssessment,
  TimeAssessmentGradeCounts,
  TimeGrade,
} from '../../architecture-direction/types';
import { CapabilityDrawer } from './CapabilityDrawer';

vi.mock('../../../hooks/useStrategyPillarsSettings', () => ({
  useStrategyPillarsConfig: () => ({ data: { data: [] } }),
}));

vi.mock('../hooks/useStrategyImportance', () => ({
  useStrategyImportanceByDomainAndCapability: () => ({ data: { data: [] }, isLoading: false }),
  useSetStrategyImportance: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateStrategyImportance: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRemoveStrategyImportance: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

const mockGetAssessment = vi.fn<(componentId: string) => TimeAssessment | undefined>();
const mockGetRollup = vi.fn<(componentId: string) => TimeAssessmentGradeCounts | undefined>();
let mockCanAssess = false;

vi.mock('../hooks/useCapabilityAssessments', () => ({
  useCapabilityAssessments: () => ({
    getAssessment: mockGetAssessment,
    getRollup: mockGetRollup,
    canAssess: mockCanAssess,
  }),
}));

let mockSuggestions: { capabilityId: string; componentId: string; suggestedTime: string | null }[] = [];

vi.mock('../../enterprise-architecture/hooks/useTimeSuggestions', () => ({
  useTimeSuggestions: () => ({
    suggestions: mockSuggestions.map((s) => ({
      capabilityId: s.capabilityId,
      capabilityName: '',
      componentId: s.componentId,
      componentName: '',
      suggestedTime: s.suggestedTime,
      technicalGap: null,
      functionalGap: null,
    })),
    isLoading: false,
    error: null,
    refetch: vi.fn(),
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
    _links: { self: { href: '', method: 'GET' } },
    ...overrides,
  };
}

const domain = buildBusinessDomain({ name: 'Ferry Freight' });

describe('CapabilityDrawer', () => {
  beforeEach(() => {
    mockGetAssessment.mockReset().mockReturnValue(undefined);
    mockGetRollup.mockReset().mockReturnValue(undefined);
    mockCanAssess = false;
    mockSuggestions = [];
    mockGetRole.mockReset().mockReturnValue(undefined);
    mockCanAssign = false;
  });

  it('renders no capability content when no capability is selected', () => {
    renderWithProviders(
      <CapabilityDrawer
        capability={null}
        domain={null}
        l1Name={null}
        getRealizationsForCapability={() => []}
        onClose={vi.fn()}
        onChipClick={vi.fn()}
      />,
    );

    expect(screen.queryByRole('heading')).not.toBeInTheDocument();
    expect(screen.queryByText('no realising application mapped')).not.toBeInTheDocument();
  });

  it('renders the breadcrumb, name, level, and empty realisations state', () => {
    const capability = cap('l2-a', 'Booking Management', 'L2');
    renderWithProviders(
      <CapabilityDrawer
        capability={capability}
        domain={domain}
        l1Name="Ferry Booking"
        getRealizationsForCapability={() => []}
        onClose={vi.fn()}
        onChipClick={vi.fn()}
      />,
    );

    expect(screen.getByTestId('capability-drawer')).toHaveTextContent('Ferry Freight');
    expect(screen.getByTestId('capability-drawer')).toHaveTextContent('Ferry Booking');
    expect(screen.getByRole('heading', { name: 'Booking Management' })).toBeInTheDocument();
    expect(screen.getByText('L2')).toBeInTheDocument();
    expect(screen.getByText('no realising application mapped')).toBeInTheDocument();
  });

  it('renders a row per realising application with level, origin note, and notes', () => {
    const capability = cap('l2-a', 'Booking Management', 'L2');
    const realizations = [
      buildCapabilityRealization({
        componentId: toComponentId('comp-1'),
        componentName: 'Phoenix',
        realizationLevel: 'Full',
        origin: 'Direct',
        notes: 'Primary booking engine',
      }),
      buildCapabilityRealization({
        componentId: toComponentId('comp-2'),
        componentName: 'Seabook',
        realizationLevel: 'Partial',
        origin: 'Inherited',
        sourceCapabilityName: 'Ferry Booking',
      }),
    ];

    renderWithProviders(
      <CapabilityDrawer
        capability={capability}
        domain={domain}
        l1Name="Ferry Booking"
        getRealizationsForCapability={() => realizations}
        onClose={vi.fn()}
        onChipClick={vi.fn()}
      />,
    );

    expect(screen.getByTestId('drawer-realization-real-1')).toHaveTextContent('Primary booking engine');
    expect(screen.getByTestId('drawer-realization-real-2')).toHaveTextContent('Inherited from Ferry Booking');
  });

  it('calls onChipClick when a realising application chip is clicked', async () => {
    const capability = cap('l2-a', 'Booking Management', 'L2');
    const onChipClick = vi.fn();
    const realizations = [
      buildCapabilityRealization({ componentId: toComponentId('comp-1'), componentName: 'Phoenix' }),
    ];

    renderWithProviders(
      <CapabilityDrawer
        capability={capability}
        domain={domain}
        l1Name="Ferry Booking"
        getRealizationsForCapability={() => realizations}
        onClose={vi.fn()}
        onChipClick={onChipClick}
      />,
    );

    await userEvent.click(screen.getByTestId('app-chip-comp-1'));

    expect(onChipClick).toHaveBeenCalledWith('comp-1');
  });

  it('renders the journey section after the realising applications', () => {
    const capability = cap('l2-a', 'Booking Management', 'L2');
    renderWithProviders(
      <CapabilityDrawer
        capability={capability}
        domain={domain}
        l1Name="Ferry Booking"
        getRealizationsForCapability={() => []}
        onClose={vi.fn()}
        onChipClick={vi.fn()}
      />,
    );

    expect(screen.getByTestId('journey-section')).toBeInTheDocument();
    expect(screen.getByText('Transition')).toBeInTheDocument();
    expect(screen.getByText('No change planned.')).toBeInTheDocument();
  });

  it('renders description, tags and owners under Details', () => {
    const capability = {
      ...cap('l2-a', 'Booking Management', 'L2'),
      description: 'Handles route bookings',
      primaryOwner: 'Jane Doe',
      tags: ['core', 'freight'],
    };

    renderWithProviders(
      <CapabilityDrawer
        capability={capability}
        domain={domain}
        l1Name="Ferry Booking"
        getRealizationsForCapability={() => []}
        onClose={vi.fn()}
        onChipClick={vi.fn()}
      />,
    );

    expect(screen.getByText('Handles route bookings')).toBeInTheDocument();
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    expect(screen.getByText('core')).toBeInTheDocument();
    expect(screen.getByText('freight')).toBeInTheDocument();
  });

  describe('TIME assessment on a Direct realising application', () => {
    function renderAssessmentDrawer(realizationOverrides: Partial<Parameters<typeof buildCapabilityRealization>[0]> = {}) {
      const capability = cap('l2-a', 'Booking Management', 'L2');
      const realizations = [
        buildCapabilityRealization({
          componentId: toComponentId('comp-1'),
          componentName: 'Phoenix',
          origin: 'Direct',
          ...realizationOverrides,
        }),
      ];

      renderWithProviders(
        <CapabilityDrawer
          capability={capability}
          domain={domain}
          l1Name="Ferry Booking"
          getRealizationsForCapability={() => realizations}
          onClose={vi.fn()}
          onChipClick={vi.fn()}
        />,
      );
    }

    it('shows the current grade, assessor, and date for an assessed realization', () => {
      mockGetAssessment.mockReturnValue(buildAssessment());

      renderAssessmentDrawer();

      expect(screen.getByTestId('assessment-comp-1')).toHaveTextContent('Migrate — for this capability');
      expect(screen.getByTestId('assessment-comp-1')).toHaveTextContent('Domain Architect');
    });

    it('shows an explicit unassessed state for a Direct realization with no assessment', () => {
      renderAssessmentDrawer();

      expect(screen.getByTestId('assessment-comp-1')).toHaveTextContent('unassessed');
    });

    it('shows the stale marker and landscape rollup line for an assessed realization', () => {
      mockGetAssessment.mockReturnValue(buildAssessment({ stale: true }));
      mockGetRollup.mockReturnValue({ Invest: 1, Tolerate: 1, Migrate: 1, Eliminate: 1 });

      renderAssessmentDrawer();

      expect(screen.getByTestId('assessment-stale-comp-1')).toHaveTextContent('stale');
      expect(screen.getByTestId('assessment-rollup-comp-1')).toHaveTextContent(
        'Across landscape: I×1 · T×1 · M×1 · E×1',
      );
    });

    it('shows no assess/remove control for a read-only caller with no write links', () => {
      mockGetAssessment.mockReturnValue(buildAssessment({ _links: { self: { href: '', method: 'GET' } } }));
      mockCanAssess = false;

      renderAssessmentDrawer();

      expect(screen.getByTestId('assessment-comp-1')).toBeInTheDocument();
      expect(screen.queryByTestId('reassess-btn-comp-1')).not.toBeInTheDocument();
      expect(screen.queryByTestId('remove-assessment-btn-comp-1')).not.toBeInTheDocument();
      expect(screen.queryByTestId('assess-btn-comp-1')).not.toBeInTheDocument();
    });

    it('shows no assessment UI at all for an Inherited realization', () => {
      mockGetAssessment.mockReturnValue(buildAssessment());

      renderAssessmentDrawer({ origin: 'Inherited', sourceCapabilityName: 'Ferry Booking' });

      expect(screen.queryByTestId('assessment-comp-1')).not.toBeInTheDocument();
    });

    it('normalises the computed suggestion for the realized component and pre-fills it as reference when opening the assess control', async () => {
      mockCanAssess = true;
      mockSuggestions = [{ capabilityId: 'l2-a', componentId: 'comp-1', suggestedTime: 'ELIMINATE' }];

      renderAssessmentDrawer();

      await userEvent.click(screen.getByTestId('assess-btn-comp-1'));

      expect(screen.getByTestId('assessment-suggestion-comp-1')).toHaveTextContent('Eliminate');
      expect(screen.getByRole('radio', { name: 'Eliminate' })).toBeChecked();
    });
  });

  describe('Realization role on a Direct realising application', () => {
    function renderRoleDrawer(realizationOverrides: Partial<Parameters<typeof buildCapabilityRealization>[0]> = {}) {
      const capability = cap('l2-a', 'Booking Management', 'L2');
      const realizations = [
        buildCapabilityRealization({
          componentId: toComponentId('comp-1'),
          componentName: 'Phoenix',
          origin: 'Direct',
          ...realizationOverrides,
        }),
      ];

      renderWithProviders(
        <CapabilityDrawer
          capability={capability}
          domain={domain}
          l1Name="Ferry Booking"
          getRealizationsForCapability={() => realizations}
          onClose={vi.fn()}
          onChipClick={vi.fn()}
        />,
      );
    }

    it('shows the current role badge for a classified Direct realization', () => {
      mockGetRole.mockReturnValue(buildRole({ role: 'standard' }));

      renderRoleDrawer();

      expect(screen.getByTestId('role-badge-comp-1')).toHaveTextContent('standard');
    });

    it('shows no role UI for an unclassified realization when the caller cannot assign', () => {
      mockCanAssign = false;

      renderRoleDrawer();

      expect(screen.queryByTestId('role-comp-1')).not.toBeInTheDocument();
    });

    it('shows assign controls for an unclassified realization when the caller can assign', () => {
      mockCanAssign = true;

      renderRoleDrawer();

      expect(screen.getByTestId('assign-standard-btn-comp-1')).toBeInTheDocument();
      expect(screen.getByTestId('assign-legacy-btn-comp-1')).toBeInTheDocument();
    });

    it('shows no role UI at all for an Inherited realization', () => {
      mockGetRole.mockReturnValue(buildRole({ role: 'standard' }));

      renderRoleDrawer({ origin: 'Inherited', sourceCapabilityName: 'Ferry Booking' });

      expect(screen.queryByTestId('role-comp-1')).not.toBeInTheDocument();
      expect(screen.queryByTestId('role-badge-comp-1')).not.toBeInTheDocument();
    });
  });
});
