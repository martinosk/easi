import { useState, useEffect } from 'react';
import logo from '../../assets/logo.svg';

type AppView = 'canvas' | 'business-domains';

interface AppNavigationProps {
  onViewChange: (view: AppView) => void;
  onOpenReleaseNotes?: () => void;
}

export function AppNavigation({ onViewChange, onOpenReleaseNotes }: AppNavigationProps) {
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
    <header className="app-header" data-testid="app-navigation">
      <div className="app-header-brand">
        <img src={logo} alt="easi logo" className="app-header-logo" />
        <span className="app-header-title">easi</span>
      </div>

      <nav className="app-header-nav">
        <button
          type="button"
          className={`app-header-nav-item ${currentView === 'canvas' ? 'app-header-nav-item-active' : ''}`}
          onClick={() => handleNavigate('canvas')}
          data-testid="nav-canvas"
        >
          <svg className="app-header-nav-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="3" y="3" width="7" height="7" rx="1" stroke="currentColor" strokeWidth="2"/>
            <rect x="14" y="3" width="7" height="7" rx="1" stroke="currentColor" strokeWidth="2"/>
            <rect x="3" y="14" width="7" height="7" rx="1" stroke="currentColor" strokeWidth="2"/>
            <rect x="14" y="14" width="7" height="7" rx="1" stroke="currentColor" strokeWidth="2"/>
          </svg>
          Architecture Canvas
        </button>
        <button
          type="button"
          className={`app-header-nav-item ${currentView === 'business-domains' ? 'app-header-nav-item-active' : ''}`}
          onClick={() => handleNavigate('business-domains')}
          data-testid="nav-business-domains"
        >
          <svg className="app-header-nav-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2"/>
            <path d="M12 3C12 3 8 8 8 12C8 16 12 21 12 21" stroke="currentColor" strokeWidth="2"/>
            <path d="M12 3C12 3 16 8 16 12C16 16 12 21 12 21" stroke="currentColor" strokeWidth="2"/>
            <path d="M3 12H21" stroke="currentColor" strokeWidth="2"/>
          </svg>
          Business Domains
        </button>
      </nav>

      <div className="app-header-actions">
        {onOpenReleaseNotes && (
          <button
            type="button"
            className="app-header-action-btn"
            onClick={onOpenReleaseNotes}
            title="View release notes"
          >
            <svg className="app-header-action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 8V12M12 16H12.01M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            What's New
          </button>
        )}
      </div>
    </header>
  );
}
