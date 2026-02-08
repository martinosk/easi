import React from 'react';
import { DetailField } from '../../../components/shared/DetailField';
import { AuditHistorySection } from '../../audit';
import { hasLink } from '../../../utils/hateoas';
import type { Vendor, OriginRelationship } from '../../../api/types';

interface VendorDetailsProps {
  vendor: Vendor;
  relationships: OriginRelationship[];
  canRemoveFromView: boolean;
  onEdit: () => void;
  onRemoveFromView: () => void;
}

export const VendorDetails: React.FC<VendorDetailsProps> = ({
  vendor,
  relationships,
  canRemoveFromView,
  onEdit,
  onRemoveFromView,
}) => {
  const canEdit = hasLink(vendor, 'edit');
  const formattedCreatedAt = new Date(vendor.createdAt).toLocaleString();
  const showActionButtons = canEdit || canRemoveFromView;

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Vendor Details</h3>
      </div>

      <div className="detail-content">
        {showActionButtons && (
          <div className="detail-actions">
            {canEdit && (
              <button className="btn btn-secondary btn-small" onClick={onEdit}>
                Edit
              </button>
            )}
            {canRemoveFromView && (
              <button className="btn btn-secondary btn-small" onClick={onRemoveFromView}>
                Remove from View
              </button>
            )}
          </div>
        )}

        <DetailField label="Name">{vendor.name}</DetailField>

        {vendor.implementationPartner && (
          <DetailField label="Implementation Partner">
            {vendor.implementationPartner}
          </DetailField>
        )}

        {vendor.notes && <DetailField label="Notes">{vendor.notes}</DetailField>}

        <DetailField label="Created">
          <span className="detail-date">{formattedCreatedAt}</span>
        </DetailField>

        <DetailField label="Type">Vendor</DetailField>

        {relationships.length > 0 && (
          <DetailField label={`Applications (${relationships.length})`}>
            <ul className="realization-list">
              {relationships.map((rel) => (
                <li key={rel.id} className="realization-item">
                  <span className="realization-name">{rel.componentName}</span>
                  <span className="realization-level">Purchased from</span>
                </li>
              ))}
            </ul>
          </DetailField>
        )}

        <AuditHistorySection aggregateId={vendor.id} />
      </div>
    </div>
  );
};
