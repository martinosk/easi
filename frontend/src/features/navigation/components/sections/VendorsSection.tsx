import React, { useState, useMemo } from 'react';
import type { Vendor, View } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { TreeItemList } from '../shared/TreeItemList';
import { ORIGIN_ENTITY_PREFIXES } from '../../../canvas/utils/nodeFactory';
import type { TreeMultiSelectProps } from '../../types';

interface VendorsSectionProps {
  vendors: Vendor[];
  currentView: View | null;
  selectedVendorId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddVendor?: () => void;
  onVendorSelect?: (vendorId: string) => void;
  onVendorContextMenu: (e: React.MouseEvent, vendor: Vendor) => void;
  multiSelect: TreeMultiSelectProps;
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

function buildVendorIdsOnCanvas(vendors: Vendor[], currentView: View | null): Set<string> {
  const viewOriginEntityIds = new Set(
    (currentView?.originEntities ?? []).map((oe) => oe.originEntityId)
  );
  const onCanvas = new Set<string>();
  for (const vendor of vendors) {
    if (viewOriginEntityIds.has(`${ORIGIN_ENTITY_PREFIXES.vendor}${vendor.id}`)) {
      onCanvas.add(vendor.id);
    }
  }
  return onCanvas;
}

export const VendorsSection: React.FC<VendorsSectionProps> = ({
  vendors,
  currentView,
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
    () => buildVendorIdsOnCanvas(vendors, currentView),
    [vendors, currentView]
  );

  const filteredVendors = useMemo(
    () => filterVendors(vendors, search),
    [vendors, search]
  );

  const visibleItems = useMemo(
    () => filteredVendors.map((v) => ({
      id: v.id, name: v.name, type: 'vendor' as const, links: v._links,
    })),
    [filteredVendors]
  );

  const hasNoVendors = vendors.length === 0;
  const emptyMessage = hasNoVendors ? 'No vendors' : 'No matches';

  const handleSelect = (vendor: Vendor, event: React.MouseEvent) => {
    const result = multiSelect.handleItemClick(
      { id: vendor.id, name: vendor.name, type: 'vendor', links: vendor._links },
      'vendors',
      visibleItems,
      event
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
          isSelected={(vendor) => selectedVendorId === vendor.id || multiSelect.isMultiSelected(vendor.id)}
          isInView={(vendor) => !currentView || vendorIdsOnCanvas.has(vendor.id)}
          getTitle={(vendor, isInView) =>
            isInView ? vendor.name : `${vendor.name} (not on canvas)`
          }
          renderLabel={(vendor) => vendor.name}
          onSelect={handleSelect}
          onContextMenu={handleContextMenu}
          onDragStart={handleDragStart}
        />
      </div>
    </TreeSection>
  );
};
