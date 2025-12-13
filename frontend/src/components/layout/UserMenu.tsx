import { useState, useRef, useEffect } from 'react';
import { useUserStore } from '../../store/userStore';

export function UserMenu() {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  const user = useUserStore((state) => state.user);
  const tenant = useUserStore((state) => state.tenant);
  const logout = useUserStore((state) => state.logout);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  if (!user || !tenant) {
    return null;
  }

  const initials = user.name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const handleLogout = async () => {
    setIsOpen(false);
    await logout();
    const basePath = import.meta.env.BASE_URL || '/';
    window.location.href = `${basePath}login`;
  };

  return (
    <div className="user-menu" ref={menuRef}>
      <button
        type="button"
        className="user-menu-trigger"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
        data-testid="user-menu-trigger"
      >
        <span className="user-menu-avatar">{initials}</span>
        <svg
          className={`user-menu-chevron ${isOpen ? 'user-menu-chevron-open' : ''}`}
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path d="M6 9L12 15L18 9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
      </button>

      {isOpen && (
        <div className="user-menu-dropdown" data-testid="user-menu-dropdown">
          <div className="user-menu-header">
            <div className="user-menu-name">{user.name}</div>
            <div className="user-menu-email">{user.email}</div>
          </div>

          <div className="user-menu-divider" />

          <div className="user-menu-info">
            <div className="user-menu-info-row">
              <span className="user-menu-info-label">Organization</span>
              <span className="user-menu-info-value">{tenant.name}</span>
            </div>
            <div className="user-menu-info-row">
              <span className="user-menu-info-label">Role</span>
              <span className="user-menu-role-badge">{user.role}</span>
            </div>
          </div>

          <div className="user-menu-divider" />

          <button
            type="button"
            className="user-menu-item user-menu-logout"
            onClick={handleLogout}
            data-testid="user-menu-logout"
          >
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M9 21H5C4.46957 21 3.96086 20.7893 3.58579 20.4142C3.21071 20.0391 3 19.5304 3 19V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M16 17L21 12L16 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M21 12H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Sign out
          </button>
        </div>
      )}
    </div>
  );
}
