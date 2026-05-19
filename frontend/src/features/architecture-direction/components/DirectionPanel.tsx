import { useMemo, useState } from 'react';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { useEnterpriseCapabilityLinks } from '../../enterprise-architecture/hooks/useEnterpriseCapabilities';
import { useAdvanceDirection, useDirectionForEnterpriseCapability, useRejectDirection } from '../hooks/useDirection';
import type { Direction } from '../types';
import { CaptureDirectionForm } from './CaptureDirectionForm';
import './DirectionPanel.css';
import { DirectionStatusBadge } from './DirectionStatusBadge';

interface DirectionPanelProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
}

const TYPE_LABELS: Record<Direction['type'], string> = {
  consolidate: 'Consolidate',
  decompose: 'Decompose',
  stay: 'Stay',
};

const HORIZON_LABELS: Record<Direction['horizon'], string> = {
  now: 'Now',
  next: 'Next',
  later: 'Later',
};

export function DirectionPanel({ enterpriseCapabilityId }: DirectionPanelProps) {
  const { data, isLoading, error } = useDirectionForEnterpriseCapability(enterpriseCapabilityId);
  const [isCapturing, setIsCapturing] = useState(false);

  if (isLoading) {
    return (
      <section className="direction-panel" aria-busy="true">
        <h3>Direction</h3>
        <p>Loading direction…</p>
      </section>
    );
  }

  if (error) {
    return (
      <section className="direction-panel">
        <h3>Direction</h3>
        <p className="direction-panel__error">Failed to load direction.</p>
      </section>
    );
  }

  if (isCapturing) {
    return (
      <section className="direction-panel" data-testid="direction-panel">
        <header className="direction-panel__header">
          <h3 className="direction-panel__title">Capture a direction</h3>
        </header>
        <CaptureDirectionForm
          enterpriseCapabilityId={enterpriseCapabilityId}
          onCaptured={() => setIsCapturing(false)}
          onCancel={() => setIsCapturing(false)}
        />
      </section>
    );
  }

  const direction = data?.direction ?? null;
  if (!direction) {
    const canCapture = !!data?._links?.['x-capture-direction'];
    return <NoDirectionView canCapture={canCapture} onCapture={() => setIsCapturing(true)} />;
  }

  return <DirectionDetail direction={direction} enterpriseCapabilityId={enterpriseCapabilityId} />;
}

function NoDirectionView({ canCapture, onCapture }: { canCapture: boolean; onCapture: () => void }) {
  return (
    <section className="direction-panel" data-testid="direction-panel">
      <header className="direction-panel__header">
        <h3 className="direction-panel__title">Direction</h3>
        <span data-testid="direction-empty-state" className="direction-panel__empty">
          No direction set
        </span>
      </header>
      <p className="direction-panel__empty-body">
        The architecture group has not captured a direction on this enterprise capability.
      </p>
      {canCapture && (
        <button type="button" className="btn btn-primary" onClick={onCapture}>
          Capture direction
        </button>
      )}
    </section>
  );
}

interface DirectionDetailProps {
  direction: Direction;
  enterpriseCapabilityId: EnterpriseCapabilityId;
}

function indexBy<T>(items: T[] | undefined, key: (item: T) => string, value: (item: T) => string | undefined): (id: string) => string | undefined {
  const lookup = new Map<string, string>();
  for (const item of items ?? []) {
    const v = value(item);
    if (v) lookup.set(key(item), v);
  }
  return (id: string) => lookup.get(id);
}

function useNameResolvers(enterpriseCapabilityId: EnterpriseCapabilityId) {
  const { data: links } = useEnterpriseCapabilityLinks(enterpriseCapabilityId);
  const { data: domainsResponse } = useBusinessDomainsQuery();

  const capabilityName = useMemo(
    () => indexBy(links, (l) => l.domainCapabilityId, (l) => l.domainCapabilityName),
    [links],
  );
  const sourceDomainName = useMemo(
    () => indexBy(links, (l) => l.domainCapabilityId, (l) => l.businessDomainName),
    [links],
  );
  const domainName = useMemo(
    () => indexBy(domainsResponse?.data, (d) => d.id, (d) => d.name),
    [domainsResponse],
  );

  return { capabilityName, sourceDomainName, domainName };
}

function DirectionDetail({ direction, enterpriseCapabilityId }: DirectionDetailProps) {
  const { capabilityName, sourceDomainName, domainName } = useNameResolvers(enterpriseCapabilityId);
  return (
    <section className="direction-panel" data-testid="direction-panel">
      <DirectionHeader direction={direction} />
      <DirectionNarrative narrative={direction.narrative} />
      <DirectionFacts
        direction={direction}
        resolveCapabilityName={capabilityName}
        resolveSourceDomainName={sourceDomainName}
        resolveDomainName={domainName}
      />
      <DirectionActions direction={direction} enterpriseCapabilityId={enterpriseCapabilityId} />
    </section>
  );
}

function DirectionHeader({ direction }: { direction: Direction }) {
  return (
    <header className="direction-panel__header">
      <div className="direction-panel__title">
        <h3 style={{ margin: 0 }}>Direction</h3>
        <span data-testid="direction-type">{TYPE_LABELS[direction.type]}</span>
      </div>
      <DirectionStatusBadge status={direction.status} />
    </header>
  );
}

function DirectionNarrative({ narrative }: { narrative: Direction['narrative'] }) {
  if (narrative) {
    return (
      <p data-testid="direction-narrative" className="direction-panel__narrative">
        {narrative}
      </p>
    );
  }
  return (
    <p className="direction-panel__narrative direction-panel__narrative--missing">
      No narrative yet. Add one before advancing this direction to proposed.
    </p>
  );
}

interface DirectionFactsProps {
  direction: Direction;
  resolveCapabilityName: (id: string) => string | undefined;
  resolveSourceDomainName: (capabilityId: string) => string | undefined;
  resolveDomainName: (id: string) => string | undefined;
}

function DirectionFacts({
  direction,
  resolveCapabilityName,
  resolveSourceDomainName,
  resolveDomainName,
}: DirectionFactsProps) {
  return (
    <dl className="direction-facts">
      <dt>Horizon</dt>
      <dd>{HORIZON_LABELS[direction.horizon]}</dd>
      <dt>Sources</dt>
      <dd>
        <SourceList
          sources={direction.sourceCapabilities}
          resolveName={resolveCapabilityName}
          resolveDomain={resolveSourceDomainName}
        />
      </dd>
      {direction.placements.length > 0 && (
        <>
          <dt>Placements</dt>
          <dd>
            <PlacementList placements={direction.placements} resolveDomainName={resolveDomainName} />
          </dd>
        </>
      )}
    </dl>
  );
}

interface SourceListProps {
  sources: Direction['sourceCapabilities'];
  resolveName: (id: string) => string | undefined;
  resolveDomain: (capabilityId: string) => string | undefined;
}

function SourceList({ sources, resolveName, resolveDomain }: SourceListProps) {
  return (
    <ul data-testid="direction-sources" className="direction-list">
      {sources.map((source) => {
        const name = resolveName(source.id);
        const domain = resolveDomain(source.id);
        return (
          <li key={source.id}>
            {name ? (
              <span className="direction-list__name">{name}</span>
            ) : (
              <code className="direction-list__name">{source.id}</code>
            )}
            {domain && <span className="direction-list__domain">· {domain}</span>}
            {source.stale && (
              <span data-testid="stale-reference" className="direction-list__stale">
                (stale — source capability deleted)
              </span>
            )}
          </li>
        );
      })}
    </ul>
  );
}

interface PlacementListProps {
  placements: Direction['placements'];
  resolveDomainName: (id: string) => string | undefined;
}

function PlacementList({ placements, resolveDomainName }: PlacementListProps) {
  return (
    <ul className="direction-list">
      {placements.map((placement, i) => {
        const domain = resolveDomainName(placement.targetBusinessDomainId);
        const name = placement.resultingName;
        return (
          <li key={`${placement.targetBusinessDomainId}-${i}`}>
            {name ? (
              <span className="direction-list__name">{name}</span>
            ) : (
              <code className="direction-list__name">{placement.targetBusinessDomainId}</code>
            )}
            {domain && <span className="direction-list__domain">· {domain}</span>}
          </li>
        );
      })}
    </ul>
  );
}

function DirectionActions({ direction, enterpriseCapabilityId }: DirectionDetailProps) {
  const advanceMutation = useAdvanceDirection();
  const rejectMutation = useRejectDirection();
  const links = direction._links ?? {};
  const dispatch = (target: 'proposed' | 'agreed') =>
    advanceMutation.mutate({ directionId: direction.id, enterpriseCapabilityId, target });

  return (
    <div className="direction-actions">
      {links['x-advance-proposed'] && (
        <button
          type="button"
          className="btn btn-primary"
          data-testid="advance-to-proposed"
          disabled={advanceMutation.isPending}
          onClick={() => dispatch('proposed')}
        >
          Advance to proposed
        </button>
      )}
      {links['x-advance-agreed'] && (
        <button
          type="button"
          className="btn btn-primary"
          data-testid="advance-to-agreed"
          disabled={advanceMutation.isPending}
          onClick={() => dispatch('agreed')}
        >
          Advance to agreed
        </button>
      )}
      {links['x-reject'] && (
        <button
          type="button"
          className="btn btn-secondary"
          data-testid="reject-direction"
          disabled={rejectMutation.isPending}
          onClick={() => rejectMutation.mutate({ directionId: direction.id, enterpriseCapabilityId })}
        >
          Reject
        </button>
      )}
    </div>
  );
}
