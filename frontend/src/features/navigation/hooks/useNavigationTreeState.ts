import { useState, useEffect, useCallback } from 'react';
import { getPersistedBoolean, getPersistedSet } from '../utils/treeUtils';

export function useNavigationTreeState() {
  const [isOpen, setIsOpen] = useState(() => getPersistedBoolean('navigationTreeOpen', true));
  const [isModelsExpanded, setIsModelsExpanded] = useState(() => getPersistedBoolean('navigationTreeModelsExpanded', true));
  const [isViewsExpanded, setIsViewsExpanded] = useState(() => getPersistedBoolean('navigationTreeViewsExpanded', true));
  const [isCapabilitiesExpanded, setIsCapabilitiesExpanded] = useState(() => getPersistedBoolean('navigationTreeCapabilitiesExpanded', true));
  const [expandedCapabilities, setExpandedCapabilities] = useState<Set<string>>(() => getPersistedSet('navigationTreeExpandedCapabilities'));

  useEffect(() => {
    localStorage.setItem('navigationTreeOpen', JSON.stringify(isOpen));
  }, [isOpen]);

  useEffect(() => {
    localStorage.setItem('navigationTreeModelsExpanded', JSON.stringify(isModelsExpanded));
  }, [isModelsExpanded]);

  useEffect(() => {
    localStorage.setItem('navigationTreeViewsExpanded', JSON.stringify(isViewsExpanded));
  }, [isViewsExpanded]);

  useEffect(() => {
    localStorage.setItem('navigationTreeCapabilitiesExpanded', JSON.stringify(isCapabilitiesExpanded));
  }, [isCapabilitiesExpanded]);

  useEffect(() => {
    localStorage.setItem('navigationTreeExpandedCapabilities', JSON.stringify([...expandedCapabilities]));
  }, [expandedCapabilities]);

  const toggleCapabilityExpanded = useCallback((capabilityId: string) => {
    setExpandedCapabilities((prev) => {
      const next = new Set(prev);
      if (next.has(capabilityId)) {
        next.delete(capabilityId);
      } else {
        next.add(capabilityId);
      }
      return next;
    });
  }, []);

  return {
    isOpen,
    setIsOpen,
    isModelsExpanded,
    setIsModelsExpanded,
    isViewsExpanded,
    setIsViewsExpanded,
    isCapabilitiesExpanded,
    setIsCapabilitiesExpanded,
    expandedCapabilities,
    toggleCapabilityExpanded,
  };
}
