import { useState, useEffect, useCallback } from 'react';
import { getPersistedBoolean, getPersistedSet, persistBoolean, persistSet } from '../utils/treeUtils';

function usePersistedBoolean(key: string, defaultValue: boolean): [boolean, React.Dispatch<React.SetStateAction<boolean>>] {
  const [value, setValue] = useState(() => getPersistedBoolean(key, defaultValue));

  useEffect(() => {
    persistBoolean(key, value);
  }, [key, value]);

  return [value, setValue];
}

function usePersistedSet(key: string): [Set<string>, React.Dispatch<React.SetStateAction<Set<string>>>] {
  const [value, setValue] = useState(() => getPersistedSet(key));

  useEffect(() => {
    persistSet(key, value);
  }, [key, value]);

  return [value, setValue];
}

export function useNavigationTreeState() {
  const [isOpen, setIsOpen] = usePersistedBoolean('navigationTreeOpen', true);
  const [isModelsExpanded, setIsModelsExpanded] = usePersistedBoolean('navigationTreeModelsExpanded', true);
  const [isViewsExpanded, setIsViewsExpanded] = usePersistedBoolean('navigationTreeViewsExpanded', true);
  const [isCapabilitiesExpanded, setIsCapabilitiesExpanded] = usePersistedBoolean('navigationTreeCapabilitiesExpanded', true);
  const [expandedCapabilities, setExpandedCapabilities] = usePersistedSet('navigationTreeExpandedCapabilities');
  const [isAcquiredEntitiesExpanded, setIsAcquiredEntitiesExpanded] = usePersistedBoolean('navigationTreeAcquiredEntitiesExpanded', false);
  const [isVendorsExpanded, setIsVendorsExpanded] = usePersistedBoolean('navigationTreeVendorsExpanded', false);
  const [isInternalTeamsExpanded, setIsInternalTeamsExpanded] = usePersistedBoolean('navigationTreeInternalTeamsExpanded', false);

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
    isAcquiredEntitiesExpanded,
    setIsAcquiredEntitiesExpanded,
    isVendorsExpanded,
    setIsVendorsExpanded,
    isInternalTeamsExpanded,
    setIsInternalTeamsExpanded,
  };
}
