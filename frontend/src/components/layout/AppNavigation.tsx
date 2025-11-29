import { useState, useEffect } from 'react';

type AppView = 'canvas' | 'business-domains';

interface AppNavigationProps {
  onViewChange: (view: AppView) => void;
}

export function AppNavigation({ onViewChange }: AppNavigationProps) {
  const [currentView, setCurrentView] = useState<AppView>('canvas');

  useEffect(() => {
    const hash = window.location.hash;
    if (hash.startsWith('#/business-domains')) {
      setCurrentView('business-domains');
      onViewChange('business-domains');
    } else {
      setCurrentView('canvas');
      onViewChange('canvas');
    }

    const handleHashChange = () => {
      const newHash = window.location.hash;
      if (newHash.startsWith('#/business-domains')) {
        setCurrentView('business-domains');
        onViewChange('business-domains');
      } else {
        setCurrentView('canvas');
        onViewChange('canvas');
      }
    };

    window.addEventListener('hashchange', handleHashChange);
    return () => window.removeEventListener('hashchange', handleHashChange);
  }, [onViewChange]);

  const handleNavigate = (view: AppView) => {
    if (view === 'business-domains') {
      window.location.hash = '#/business-domains';
    } else {
      window.location.hash = '#/';
    }
  };

  return (
    <nav className="app-navigation" data-testid="app-navigation">
      <button
        type="button"
        className={`nav-tab ${currentView === 'canvas' ? 'nav-tab-active' : ''}`}
        onClick={() => handleNavigate('canvas')}
        data-testid="nav-canvas"
      >
        Architecture Canvas
      </button>
      <button
        type="button"
        className={`nav-tab ${currentView === 'business-domains' ? 'nav-tab-active' : ''}`}
        onClick={() => handleNavigate('business-domains')}
        data-testid="nav-business-domains"
      >
        Business Domains
      </button>
    </nav>
  );
}
