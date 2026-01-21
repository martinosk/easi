import React from 'react';
import { DetailField } from '../../../components/shared/DetailField';
import { AuditHistorySection } from '../../audit';
import { hasLink } from '../../../utils/hateoas';
import type { Vendor, OriginRelationship } from '../../../api/types';

interface VendorDetailsProps {
  vendor: Vendor;
  relationships: OriginRelationship[];
  onEdit: () => void;
  onDelete: () => void;
}

export const VendorDetails: React.FC<VendorDetailsProps> = ({
  vendor,
  relationships,
  onEdit,
  onDelete,
}) => {
  const canEdit = hasLink(vendor, 'edit');
  const canDelete = hasLink(vendor, 'delete');
  const formattedCreatedAt = new Date(vendor.createdAt).toLocaleString();

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Vendor Details</h3>
      </div>

      <div className="detail-content">
        <div className="detail-actions">
          {canEdit && (
            <button className="btn btn-secondary btn-small" onClick={onEdit}>
              Edit
            </button>
          )}
          {canDelete && (
            <button className="btn btn-danger btn-small" onClick={onDelete}>
              Delete
            </button>
          )}
        </div>

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
