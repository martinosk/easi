import React, { useMemo, useState } from 'react';
import type { Vendor, View } from '../../../../api/types';
import type { TreeMultiSelectProps } from '../../types';
import { TreeItemList } from '../shared/TreeItemList';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { TreeSection } from '../TreeSection';

interface VendorsSectionProps {
  vendors: Vendor[];
  currentView: View | null;
  originEntitiesInView?: Set<string>;
  selectedVendorId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddVendor?: () => void;
  onVendorSelect?: (vendorId: string) => void;
  onVendorContextMenu: (e: React.MouseEvent, vendor: Vendor) => void;
  multiSelect: TreeMultiSelectProps;
}

function defaultOriginEntitiesInView(currentView: View | null): Set<string> {
  return new Set((currentView?.originEntities ?? []).map((oe) => oe.originEntityId));
}

function filterVendors(vendors: Vendor[], search: string): Vendor[] {
  if (!search.trim()) return vendors;
  const searchLower = search.toLowerCase();
  return vendors.filter(
    (v) =>
      v.name.toLowerCase().includes(searchLower) ||
      (v.implementationPartner && v.implementationPartner.toLowerCase().includes(searchLower)) ||
      (v.notes && v.notes.toLowerCase().includes(searchLower)),
  );
}

export const VendorsSection: React.FC<VendorsSectionProps> = ({
  vendors,
  currentView,
  originEntitiesInView,
  selectedVendorId,
  isExpanded,
  onToggle,
  onAddVendor,
  onVendorSelect,
  onVendorContextMenu,
  multiSelect,
}) => {
  const [search, setSearch] = useState('');
  const vendorIdsOnCanvas = useMemo(
    () => originEntitiesInView ?? defaultOriginEntitiesInView(currentView),
    [originEntitiesInView, currentView],
  );

  const filteredVendors = useMemo(() => filterVendors(vendors, search), [vendors, search]);

  const visibleItems = useMemo(
    () =>
      filteredVendors.map((v) => ({
        id: v.id,
        name: v.name,
        type: 'vendor' as const,
        links: v._links,
      })),
    [filteredVendors],
  );

  const hasNoVendors = vendors.length === 0;
  const emptyMessage = hasNoVendors ? 'No vendors' : 'No matches';

  const handleSelect = (vendor: Vendor, event: React.MouseEvent) => {
    const result = multiSelect.handleItemClick(
      { id: vendor.id, name: vendor.name, type: 'vendor', links: vendor._links },
      'vendors',
      visibleItems,
      event,
    );
    if (result === 'single') {
      onVendorSelect?.(vendor.id);
    }
  };

  const handleContextMenu = (e: React.MouseEvent, vendor: Vendor) => {
    const handled = multiSelect.handleContextMenu(e, vendor.id, multiSelect.selectedItems);
    if (!handled) {
      onVendorContextMenu(e, vendor);
    }
  };

  const handleDragStart = (e: React.DragEvent, vendor: Vendor) => {
    const handled = multiSelect.handleDragStart(e, vendor.id);
    if (!handled) {
      e.dataTransfer.setData('vendorId', vendor.id);
      e.dataTransfer.effectAllowed = 'copy';
    }
  };

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
      <TreeSearchInput value={search} onChange={setSearch} placeholder="Search vendors..." />
      <div className="tree-items">
        <TreeItemList
          items={filteredVendors}
          emptyMessage={emptyMessage}
          icon="🏪"
          dragDataKey="vendorId"
          isSelected={(vendor) => selectedVendorId === vendor.id || multiSelect.isMultiSelected(vendor.id)}
          isInView={(vendor) => !currentView || vendorIdsOnCanvas.has(vendor.id)}
          getTitle={(vendor, isInView) => (isInView ? vendor.name : `${vendor.name} (not on canvas)`)}
          renderLabel={(vendor) => vendor.name}
          onSelect={handleSelect}
          onContextMenu={handleContextMenu}
          onDragStart={handleDragStart}
        />
      </div>
    </TreeSection>
  );
};
