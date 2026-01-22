import React, { useState, useMemo } from 'react';
import type { Vendor, View, OriginRelationship } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { TreeItemList } from '../shared/TreeItemList';

interface VendorsSectionProps {
  vendors: Vendor[];
  currentView: View | null;
  originRelationships: OriginRelationship[];
  selectedVendorId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddVendor?: () => void;
  onVendorSelect?: (vendorId: string) => void;
  onVendorContextMenu: (e: React.MouseEvent, vendor: Vendor) => void;
}

function filterVendors(vendors: Vendor[], search: string): Vendor[] {
  if (!search.trim()) return vendors;
  const searchLower = search.toLowerCase();
  return vendors.filter(
    (v) =>
      v.name.toLowerCase().includes(searchLower) ||
      (v.implementationPartner && v.implementationPartner.toLowerCase().includes(searchLower)) ||
      (v.notes && v.notes.toLowerCase().includes(searchLower))
  );
}

function buildVendorIdsInView(
  relationships: OriginRelationship[],
  componentIdsInView: Set<string>
): Set<string> {
  const inView = new Set<string>();
  for (const rel of relationships) {
    if (rel.relationshipType === 'PurchasedFrom' && componentIdsInView.has(rel.componentId)) {
      inView.add(rel.originEntityId);
    }
  }
  return inView;
}

export const VendorsSection: React.FC<VendorsSectionProps> = ({
  vendors,
  currentView,
  originRelationships,
  selectedVendorId,
  isExpanded,
  onToggle,
  onAddVendor,
  onVendorSelect,
  onVendorContextMenu,
}) => {
  const [search, setSearch] = useState('');

  const componentIdsInView = useMemo(() => {
    if (!currentView) return new Set<string>();
    return new Set(currentView.components.map(vc => vc.componentId));
  }, [currentView]);

  const vendorIdsInView = useMemo(
    () => buildVendorIdsInView(originRelationships, componentIdsInView),
    [originRelationships, componentIdsInView]
  );

  const filteredVendors = useMemo(
    () => filterVendors(vendors, search),
    [vendors, search]
  );

  const hasNoVendors = vendors.length === 0;
  const emptyMessage = hasNoVendors ? 'No vendors' : 'No matches';

  return (
    <TreeSection
      label="Vendors"
      count={vendors.length}
      isExpanded={isExpanded}
      onToggle={onToggle}
      onAdd={onAddVendor}
      addTitle="Create new vendor"
      addTestId="create-vendor-button"
    >
      <TreeSearchInput
        value={search}
        onChange={setSearch}
        placeholder="Search vendors..."
      />
      <div className="tree-items">
        <TreeItemList
          items={filteredVendors}
          emptyMessage={emptyMessage}
          icon="🏪"
          dragDataKey="vendorId"
          isSelected={(vendor) => selectedVendorId === vendor.id}
          isInView={(vendor) => !currentView || vendorIdsInView.has(vendor.id)}
          getTitle={(vendor, isInView) =>
            isInView ? vendor.name : `${vendor.name} (not linked to components in current view)`
          }
          renderLabel={(vendor) => vendor.name}
          onSelect={(vendor) => onVendorSelect?.(vendor.id)}
          onContextMenu={onVendorContextMenu}
        />
      </div>
    </TreeSection>
  );
};
