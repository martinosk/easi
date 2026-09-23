import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import { beforeEach, describe, expect, it } from 'vitest';
import type { AcquiredEntity, HATEOASLinks, InternalTeam, Vendor } from '../../../api/types';
import { toAcquiredEntityId, toInternalTeamId, toVendorId } from '../../../api/types';
import {
  buildAcquiredEntity,
  buildInternalTeam,
  buildOriginRelationship,
  buildVendor,
  detailGroupIds,
  fieldLabelsIn,
  renderWithProviders,
  seedDb,
  server,
} from '../../../test/helpers';
import { OriginEntityDetailsPanel } from './OriginEntityDetailsPanel';

const API_BASE = 'http://localhost:8080';

function readOnlyLinks(path: string): HATEOASLinks {
  return { self: { href: `/api/v1/${path}`, method: 'GET' } };
}

function editableLinks(path: string): HATEOASLinks {
  return {
    self: { href: `/api/v1/${path}`, method: 'GET' },
    edit: { href: `/api/v1/${path}`, method: 'PUT' },
    'x-one-pager': { href: `/api/v1/one-pagers/${path}`, method: 'GET' },
  };
}

function acquired(links: HATEOASLinks): AcquiredEntity {
  return buildAcquiredEntity({
    id: toAcquiredEntityId('ae-1'),
    name: 'Nordic Cargo',
    acquisitionDate: '2023-06-01',
    integrationStatus: 'IN_PROGRESS',
    notes: 'Integration underway',
    _links: links,
  });
}

function vendor(links: HATEOASLinks): Vendor {
  return buildVendor({
    id: toVendorId('vendor-1'),
    name: 'SAP',
    implementationPartner: 'Accenture',
    notes: 'Enterprise agreement',
    _links: links,
  });
}

function team(links: HATEOASLinks): InternalTeam {
  return buildInternalTeam({
    id: toInternalTeamId('team-1'),
    name: 'Platform Team',
    department: 'Engineering',
    contactPerson: 'Kim',
    notes: 'Owns the platform',
    _links: links,
  });
}

const FIELD_LABEL =
  /^(Acquisition Date|Integration Status|Implementation Partner|Department|Contact Person|Notes|Created|Type)$/;

function labelsInGroup(groupId: string): (string | null)[] {
  return fieldLabelsIn(within(screen.getByTestId(`detail-group-${groupId}`)), FIELD_LABEL);
}

describe('OriginEntityDetailsPanel', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  describe('acquired entity', () => {
    it('renders the name as the heading and every section in order', async () => {
      seedDb({
        acquiredEntities: [acquired(editableLinks('acquired-entities/ae-1'))],
        originRelationships: [
          buildOriginRelationship({
            relationshipType: 'AcquiredVia',
            originEntityId: 'ae-1',
            componentName: 'Phoenix',
          }),
        ],
      });
      renderWithProviders(<OriginEntityDetailsPanel entityType="acquired" entityId="ae-1" />);

      const heading = await screen.findByRole('heading', { name: 'Nordic Cargo' });
      expect(screen.queryByText('Acquired Entity Details')).not.toBeInTheDocument();
      expect(detailGroupIds()).toEqual([
        'detail-group-description',
        'detail-group-metadata',
        'detail-group-applications',
      ]);
      const description = screen.getByTestId('detail-group-description');
      expect(heading.compareDocumentPosition(description) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
      expect(within(description).queryByRole('heading')).not.toBeInTheDocument();
      expect(labelsInGroup('description')).toEqual(['Acquisition Date', 'Integration Status', 'Notes']);
      expect(labelsInGroup('metadata')).toEqual(['Created', 'Type']);
      expect(screen.getByTestId('group-count-applications')).toHaveTextContent('1');
      expect(within(screen.getByTestId('detail-group-applications')).getByText('Phoenix')).toBeInTheDocument();
      expect(screen.getByText('In Progress')).toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'One-Pager' })).toBeInTheDocument();
    });

    it('renames in place through the edit link, sending the other fields unchanged', async () => {
      let received: unknown;
      server.use(
        http.put(`${API_BASE}/api/v1/acquired-entities/ae-1`, async ({ request }) => {
          received = await request.json();
          return HttpResponse.json({ ...acquired(editableLinks('acquired-entities/ae-1')), name: 'Nordic Cargo AS' });
        }),
      );
      seedDb({ acquiredEntities: [acquired(editableLinks('acquired-entities/ae-1'))] });
      renderWithProviders(<OriginEntityDetailsPanel entityType="acquired" entityId="ae-1" />);

      fireEvent.click(await screen.findByRole('button', { name: 'Edit name' }));
      fireEvent.change(screen.getByTestId('origin-entity-name-input'), { target: { value: 'Nordic Cargo AS' } });
      fireEvent.keyDown(screen.getByTestId('origin-entity-name-input'), { key: 'Enter' });

      await waitFor(() =>
        expect(received).toEqual({
          name: 'Nordic Cargo AS',
          acquisitionDate: '2023-06-01',
          integrationStatus: 'IN_PROGRESS',
          notes: 'Integration underway',
        }),
      );
    });

    it('changes the integration status and the acquisition date in place', async () => {
      const bodies: unknown[] = [];
      server.use(
        http.put(`${API_BASE}/api/v1/acquired-entities/ae-1`, async ({ request }) => {
          const body = (await request.json()) as Record<string, unknown>;
          bodies.push(body);
          return HttpResponse.json({ ...acquired(editableLinks('acquired-entities/ae-1')), ...body });
        }),
      );
      seedDb({ acquiredEntities: [acquired(editableLinks('acquired-entities/ae-1'))] });
      renderWithProviders(<OriginEntityDetailsPanel entityType="acquired" entityId="ae-1" />);

      fireEvent.click(await screen.findByRole('button', { name: 'Edit integration status' }));
      fireEvent.click(screen.getByTestId('origin-entity-integration-status-input'));
      const option = await waitFor(() => {
        const found = Array.from(document.querySelectorAll('[data-combobox-option]')).find(
          (element) => element.textContent === 'Completed',
        );
        if (!found) throw new Error('option not rendered');
        return found;
      });
      fireEvent.click(option);
      fireEvent.click(screen.getByTestId('origin-entity-integration-status-save'));
      await waitFor(() => expect(bodies[0]).toMatchObject({ integrationStatus: 'COMPLETED', name: 'Nordic Cargo' }));

      fireEvent.click(await screen.findByRole('button', { name: 'Edit acquisition date' }));
      fireEvent.change(screen.getByTestId('origin-entity-acquisition-date-input'), { target: { value: '2024-01-15' } });
      fireEvent.keyDown(screen.getByTestId('origin-entity-acquisition-date-input'), { key: 'Enter' });
      await waitFor(() => expect(bodies[1]).toMatchObject({ acquisitionDate: '2024-01-15', name: 'Nordic Cargo' }));
    });

    it('renders every field read-only with no controls when the edit link is absent', async () => {
      seedDb({ acquiredEntities: [acquired(readOnlyLinks('acquired-entities/ae-1'))] });
      renderWithProviders(<OriginEntityDetailsPanel entityType="acquired" entityId="ae-1" />);

      await screen.findByRole('heading', { name: 'Nordic Cargo' });
      expect(screen.getByText('Integration underway')).toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /^(Edit|Set|Add) / })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'One-Pager' })).not.toBeInTheDocument();
    });
  });

  describe('vendor', () => {
    it('renders the vendor sections in order and edits the implementation partner in place', async () => {
      let received: unknown;
      server.use(
        http.put(`${API_BASE}/api/v1/vendors/vendor-1`, async ({ request }) => {
          received = await request.json();
          return HttpResponse.json({
            ...vendor(editableLinks('vendors/vendor-1')),
            implementationPartner: 'Capgemini',
          });
        }),
      );
      seedDb({ vendors: [vendor(editableLinks('vendors/vendor-1'))] });
      renderWithProviders(<OriginEntityDetailsPanel entityType="vendor" entityId="vendor-1" />);

      expect(await screen.findByRole('heading', { name: 'SAP' })).toBeInTheDocument();
      expect(screen.queryByText('Vendor Details')).not.toBeInTheDocument();
      expect(labelsInGroup('description')).toEqual(['Implementation Partner', 'Notes']);
      expect(labelsInGroup('metadata')).toEqual(['Created', 'Type']);

      fireEvent.click(screen.getByRole('button', { name: 'Edit implementation partner' }));
      fireEvent.change(screen.getByTestId('origin-entity-implementation-partner-input'), {
        target: { value: 'Capgemini' },
      });
      fireEvent.keyDown(screen.getByTestId('origin-entity-implementation-partner-input'), { key: 'Enter' });

      await waitFor(() =>
        expect(received).toEqual({ name: 'SAP', implementationPartner: 'Capgemini', notes: 'Enterprise agreement' }),
      );
    });
  });

  describe('internal team', () => {
    it('renders the team sections in order and prompts for an empty department only when editable', async () => {
      seedDb({ internalTeams: [{ ...team(editableLinks('internal-teams/team-1')), department: undefined }] });
      const { unmount } = renderWithProviders(<OriginEntityDetailsPanel entityType="team" entityId="team-1" />);

      expect(await screen.findByRole('heading', { name: 'Platform Team' })).toBeInTheDocument();
      expect(screen.queryByText('Internal Team Details')).not.toBeInTheDocument();
      expect(labelsInGroup('description')).toEqual(['Department', 'Contact Person', 'Notes']);
      expect(screen.getByRole('button', { name: 'Add a department' })).toBeInTheDocument();
      unmount();

      seedDb({ internalTeams: [{ ...team(readOnlyLinks('internal-teams/team-1')), department: undefined }] });
      renderWithProviders(<OriginEntityDetailsPanel entityType="team" entityId="team-1" />);
      await screen.findByRole('heading', { name: 'Platform Team' });
      expect(screen.queryByText('Department', { selector: 'label' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'Add a department' })).not.toBeInTheDocument();
    });
  });

  it('renders the view-membership slot as an "In this view" group after Applications and nothing view-scoped without it', async () => {
    seedDb({
      vendors: [vendor(editableLinks('vendors/vendor-1'))],
      originRelationships: [
        buildOriginRelationship({
          relationshipType: 'PurchasedFrom',
          originEntityId: 'vendor-1',
          componentName: 'Phoenix',
        }),
      ],
    });
    const { unmount } = renderWithProviders(
      <OriginEntityDetailsPanel
        entityType="vendor"
        entityId="vendor-1"
        viewMembership={<div data-testid="view-slot">remove</div>}
      />,
    );

    await screen.findByRole('heading', { name: 'SAP' });
    expect(detailGroupIds().slice(-2)).toEqual(['detail-group-applications', 'detail-group-view']);
    const view = within(screen.getByTestId('detail-group-view'));
    expect(view.getByRole('button', { name: 'In this view' })).toBeInTheDocument();
    expect(view.getByTestId('view-slot')).toBeInTheDocument();
    unmount();

    renderWithProviders(<OriginEntityDetailsPanel entityType="vendor" entityId="vendor-1" />);
    await screen.findByRole('heading', { name: 'SAP' });
    expect(screen.queryByTestId('detail-group-view')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Remove from View' })).not.toBeInTheDocument();
  });

  it('keeps the Applications group with an empty state and a zero count when nothing is related', async () => {
    seedDb({ vendors: [vendor(readOnlyLinks('vendors/vendor-1'))] });
    renderWithProviders(<OriginEntityDetailsPanel entityType="vendor" entityId="vendor-1" />);

    await screen.findByRole('heading', { name: 'SAP' });
    expect(screen.getByTestId('group-count-applications')).toHaveTextContent('0');
    expect(
      within(screen.getByTestId('detail-group-applications')).getByText('no related application'),
    ).toBeInTheDocument();
  });

  it('renders the one-pager action and history below the groups', async () => {
    seedDb({ vendors: [vendor(editableLinks('vendors/vendor-1'))] });
    renderWithProviders(<OriginEntityDetailsPanel entityType="vendor" entityId="vendor-1" />);

    const lastGroup = (await screen.findAllByTestId(/^detail-group-/)).at(-1) as HTMLElement;
    const onePager = screen.getByRole('button', { name: 'One-Pager' });
    const history = screen.getByRole('button', { name: /^History/ });
    expect(lastGroup.compareDocumentPosition(onePager) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(onePager.compareDocumentPosition(history) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });
});
