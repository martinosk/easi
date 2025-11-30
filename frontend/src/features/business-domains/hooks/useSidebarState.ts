import { useState } from 'react';

export function useSidebarState() {
  const [isDomainsSidebarCollapsed, setIsDomainsSidebarCollapsed] = useState(false);
  const [isExplorerSidebarCollapsed, setIsExplorerSidebarCollapsed] = useState(false);

  return {
    isDomainsSidebarCollapsed,
    isExplorerSidebarCollapsed,
    toggleDomainsSidebar: () => setIsDomainsSidebarCollapsed((prev) => !prev),
    toggleExplorerSidebar: () => setIsExplorerSidebarCollapsed((prev) => !prev),
    setIsDomainsSidebarCollapsed,
    setIsExplorerSidebarCollapsed,
  };
}
