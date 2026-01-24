import React, { useState, useMemo } from 'react';
import type { Vendor, View } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { TreeItemList } from '../shared/TreeItemList';
import { useCanvasLayoutContext } from '../../../canvas/context/CanvasLayoutContext';
import { ORIGIN_ENTITY_PREFIXES } from '../../../canvas/utils/nodeFactory';

interface VendorsSectionProps {
  vendors: Vendor[];
  currentView: View | null;
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

export const VendorsSection: React.FC<VendorsSectionProps> = ({
  vendors,
  currentView,
  selectedVendorId,
  isExpanded,
  onToggle,
  onAddVendor,
  onVendorSelect,
  onVendorContextMenu,
}) => {
  const [search, setSearch] = useState('');
  const { positions: layoutPositions } = useCanvasLayoutContext();

  const vendorIdsOnCanvas = useMemo(() => {
    const onCanvas = new Set<string>();
    for (const vendor of vendors) {
      const nodeId = `${ORIGIN_ENTITY_PREFIXES.vendor}${vendor.id}`;
      if (layoutPositions[nodeId] !== undefined) {
        onCanvas.add(vendor.id);
      }
    }
    return onCanvas;
  }, [vendors, layoutPositions]);

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
          isInView={(vendor) => !currentView || vendorIdsOnCanvas.has(vendor.id)}
          getTitle={(vendor, isInView) =>
            isInView ? vendor.name : `${vendor.name} (not on canvas)`
          }
          renderLabel={(vendor) => vendor.name}
          onSelect={(vendor) => onVendorSelect?.(vendor.id)}
          onContextMenu={onVendorContextMenu}
        />
      </div>
    </TreeSection>
  );
};
