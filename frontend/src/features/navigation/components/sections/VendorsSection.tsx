import React, { useState, useMemo } from 'react';
import type { Vendor } from '../../../../api/types';
import { TreeSection } from '../TreeSection';

interface VendorsSectionProps {
  vendors: Vendor[];
  selectedVendorId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddVendor?: () => void;
  onVendorSelect?: (vendorId: string) => void;
  onVendorContextMenu: (e: React.MouseEvent, vendor: Vendor) => void;
}

export const VendorsSection: React.FC<VendorsSectionProps> = ({
  vendors,
  selectedVendorId,
  isExpanded,
  onToggle,
  onAddVendor,
  onVendorSelect,
  onVendorContextMenu,
}) => {
  const [search, setSearch] = useState('');

  const filteredVendors = useMemo(() => {
    if (!search.trim()) {
      return vendors;
    }
    const searchLower = search.toLowerCase();
    return vendors.filter(
      (v) =>
        v.name.toLowerCase().includes(searchLower) ||
        (v.implementationPartner && v.implementationPartner.toLowerCase().includes(searchLower)) ||
        (v.notes && v.notes.toLowerCase().includes(searchLower))
    );
  }, [vendors, search]);

  const handleVendorClick = (vendorId: string) => {
    if (onVendorSelect) {
      onVendorSelect(vendorId);
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
      <div className="tree-search">
        <input
          type="text"
          className="tree-search-input"
          placeholder="Search vendors..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        {search && (
          <button
            className="tree-search-clear"
            onClick={() => setSearch('')}
            aria-label="Clear search"
          >
            x
          </button>
        )}
      </div>
      <div className="tree-items">
        {filteredVendors.length === 0 ? (
          <div className="tree-item-empty">
            {vendors.length === 0 ? 'No vendors' : 'No matches'}
          </div>
        ) : (
          filteredVendors.map((vendor) => {
            const isSelected = selectedVendorId === vendor.id;

            return (
              <button
                key={vendor.id}
                className={`tree-item ${isSelected ? 'selected' : ''}`}
                onClick={() => handleVendorClick(vendor.id)}
                onContextMenu={(e) => onVendorContextMenu(e, vendor)}
                title={vendor.name}
                draggable
                onDragStart={(e) => {
                  e.dataTransfer.setData('vendorId', vendor.id);
                  e.dataTransfer.effectAllowed = 'copy';
                }}
              >
                <span className="tree-item-icon">🏪</span>
                <span className="tree-item-label">{vendor.name}</span>
              </button>
            );
          })
        )}
      </div>
    </TreeSection>
  );
};
