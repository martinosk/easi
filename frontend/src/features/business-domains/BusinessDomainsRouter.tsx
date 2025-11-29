import { useState, useEffect } from 'react';
import { BusinessDomainsPage } from './pages/BusinessDomainsPage';
import { DomainDetailPage } from './pages/DomainDetailPage';
import { DomainVisualizationPage } from './pages/DomainVisualizationPage';
import type { BusinessDomainId } from '../../api/types';

export function BusinessDomainsRouter() {
  const [currentPath, setCurrentPath] = useState(window.location.hash);

  useEffect(() => {
    const handleHashChange = () => {
      setCurrentPath(window.location.hash);
    };

    window.addEventListener('hashchange', handleHashChange);
    return () => window.removeEventListener('hashchange', handleHashChange);
  }, []);

  if (currentPath === '#/business-domains/visualization') {
    return <DomainVisualizationPage />;
  }

  if (currentPath.startsWith('#/business-domains/')) {
    const domainId = currentPath.replace('#/business-domains/', '') as BusinessDomainId;
    return <DomainDetailPage domainId={domainId} />;
  }

  if (currentPath === '#/business-domains' || currentPath === '') {
    return <BusinessDomainsPage />;
  }

  return <BusinessDomainsPage />;
}
