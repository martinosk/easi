import { useMemo } from 'react';
import type { EnterpriseCapability, EnterpriseCapabilityLink } from '../types';
import { useEnterpriseCapabilityLinks, useUnlinkDomainCapability } from '../hooks/useEnterpriseCapabilities';

interface EnterpriseCapabilityDetailPanelProps {
  capability: EnterpriseCapability;
  canWrite: boolean;
  onClose: () => void;
}


export function EnterpriseCapabilityDetailPanel({
  capability,
  canWrite,
  onClose,
}: EnterpriseCapabilityDetailPanelProps) {
  const { data: links, isLoading } = useEnterpriseCapabilityLinks(capability.id);
  const unlinkMutation = useUnlinkDomainCapability();

  const groupedLinks = useMemo(() => {
    if (!links) return [];

    const grouped = new Map<string, EnterpriseCapabilityLink[]>();

    for (const link of links) {
      const domainName = link.businessDomainName || 'Unassigned';
      if (!grouped.has(domainName)) {
        grouped.set(domainName, []);
      }
      grouped.get(domainName)!.push(link);
    }

    return Array.from(grouped.entries())
      .map(([domainName, domainLinks]) => ({ domainName, links: domainLinks }))
      .sort((a, b) => {
        if (a.domainName === 'Unassigned') return 1;
        if (b.domainName === 'Unassigned') return -1;
        return a.domainName.localeCompare(b.domainName);
      });
  }, [links]);

  const handleUnlink = async (link: EnterpriseCapabilityLink) => {
    await unlinkMutation.mutateAsync({
      enterpriseCapabilityId: capability.id,
      linkId: link.id,
    });
  };

  return (
    <div className="detail-panel">
      <div className="detail-panel-header">
        <div className="detail-panel-title">
          <h2>{capability.name}</h2>
          {capability.category && (
            <span className="category-badge">{capability.category}</span>
          )}
        </div>
        <button
          type="button"
          className="btn btn-icon"
          onClick={onClose}
          aria-label="Close detail panel"
        >
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" width="20" height="20">
            <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
      </div>

      {capability.description && (
        <p className="detail-panel-description">{capability.description}</p>
      )}

      <div className="detail-panel-stats">
        <div className="stat">
          <span className="stat-label">Links:</span>
          <span className="stat-value">{capability.linkCount}</span>
        </div>
        <div className="stat">
          <span className="stat-label">Domains:</span>
          <span className="stat-value">{capability.domainCount}</span>
        </div>
      </div>

      <div className="detail-panel-section">
        <h3>Linked Capabilities</h3>

        {isLoading ? (
          <div className="loading-state">Loading linked capabilities...</div>
        ) : groupedLinks.length === 0 ? (
          <p className="empty-state">No capabilities linked yet. Use the Manage Links page to link domain capabilities.</p>
        ) : (
          <div className="linked-capabilities-list">
            {groupedLinks.map((group) => (
              <div key={group.domainName} className="domain-group">
                <h4 className="domain-group-header">{group.domainName}</h4>
                <ul className="link-list">
                  {group.links.map((link) => (
                    <li key={link.id} className="link-item">
                      <span className="link-name">
                        {link.domainCapabilityName || link.domainCapabilityId}
                      </span>
                      {canWrite && (
                        <button
                          type="button"
                          className="btn btn-sm btn-ghost btn-danger"
                          onClick={() => handleUnlink(link)}
                          disabled={unlinkMutation.isPending}
                          title="Unlink capability"
                        >
                          Unlink
                        </button>
                      )}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
