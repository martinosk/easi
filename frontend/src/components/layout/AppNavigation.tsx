import { useNavigate } from 'react-router-dom';
import logo from '../../assets/logo.svg';
import { UserMenu } from './UserMenu';
import { ROUTES } from '../../routes/routes';
import { useUserStore } from '../../store/userStore';

type AppView = 'canvas' | 'business-domains' | 'invitations' | 'users';

interface AppNavigationProps {
  currentView: AppView;
  onOpenReleaseNotes?: () => void;
}

export function AppNavigation({ currentView, onOpenReleaseNotes }: AppNavigationProps) {
  const navigate = useNavigate();
  const hasPermission = useUserStore((state) => state.hasPermission);

  const canViewUsers = hasPermission('users:read');
  const canManageInvitations = hasPermission('invitations:manage');

  const handleNavigate = (view: AppView) => {
    if (view === 'business-domains') {
      navigate(ROUTES.BUSINESS_DOMAINS);
    } else if (view === 'invitations') {
      navigate(ROUTES.INVITATIONS);
    } else if (view === 'users') {
      navigate(ROUTES.USERS);
    } else {
      navigate(ROUTES.HOME);
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
        {canViewUsers && (
          <button
            type="button"
            className={`app-header-nav-item ${currentView === 'users' ? 'app-header-nav-item-active' : ''}`}
            onClick={() => handleNavigate('users')}
            data-testid="nav-users"
          >
            <svg className="app-header-nav-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M17 21V19C17 17.9391 16.5786 16.9217 15.8284 16.1716C15.0783 15.4214 14.0609 15 13 15H5C3.93913 15 2.92172 15.4214 2.17157 16.1716C1.42143 16.9217 1 17.9391 1 19V21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M9 11C11.2091 11 13 9.20914 13 7C13 4.79086 11.2091 3 9 3C6.79086 3 5 4.79086 5 7C5 9.20914 6.79086 11 9 11Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Users
          </button>
        )}
        {canManageInvitations && (
          <button
            type="button"
            className={`app-header-nav-item ${currentView === 'invitations' ? 'app-header-nav-item-active' : ''}`}
            onClick={() => handleNavigate('invitations')}
            data-testid="nav-invitations"
          >
            <svg className="app-header-nav-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M22 2L11 13" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M22 2L15 22L11 13L2 9L22 2Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Invitations
          </button>
        )}
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
        <UserMenu />
      </div>
    </header>
  );
}
