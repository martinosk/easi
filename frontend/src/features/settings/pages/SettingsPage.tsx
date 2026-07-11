import { Navigate, NavLink, Route, Routes } from 'react-router-dom';
import { useUserStore } from '../../../store/userStore';
import { OnePagersSettings } from '../../one-pagers';
import { AIConfigurationSettings } from '../components/AIConfigurationSettings';
import { AppearanceSettings } from '../components/AppearanceSettings';
import { MaturityScaleSettings } from '../components/MaturityScaleSettings';
import { StrategyPillarsSettings } from '../components/StrategyPillarsSettings';
import './SettingsPage.css';

export function SettingsPage() {
  const hasPermission = useUserStore((state) => state.hasPermission);
  const canManageMetaModel = hasPermission('metamodel:write');

  if (!canManageMetaModel) {
    return (
      <div className="settings-page">
        <div className="settings-container">
          <div className="error-message">You do not have permission to manage settings.</div>
        </div>
      </div>
    );
  }

  return (
    <div className="settings-page">
      <div className="settings-container">
        <div className="settings-header">
          <h1 className="settings-title">Settings</h1>
          <p className="settings-subtitle">Configure system-wide settings for your organization.</p>
        </div>

        <nav className="settings-tabs">
          <NavLink
            to="/settings/maturity-scale"
            className={({ isActive }) => `settings-tab ${isActive ? 'settings-tab-active' : ''}`}
          >
            Maturity Scale
          </NavLink>
          <NavLink
            to="/settings/strategy-pillars"
            className={({ isActive }) => `settings-tab ${isActive ? 'settings-tab-active' : ''}`}
          >
            Strategy Pillars
          </NavLink>
          <NavLink
            to="/settings/ai-configuration"
            className={({ isActive }) => `settings-tab ${isActive ? 'settings-tab-active' : ''}`}
          >
            AI Configuration
          </NavLink>
          <NavLink
            to="/settings/one-pagers"
            className={({ isActive }) => `settings-tab ${isActive ? 'settings-tab-active' : ''}`}
          >
            One-Pagers
          </NavLink>
          <NavLink
            to="/settings/appearance"
            className={({ isActive }) => `settings-tab ${isActive ? 'settings-tab-active' : ''}`}
          >
            Appearance
          </NavLink>
        </nav>

        <Routes>
          <Route path="/" element={<Navigate to="/settings/maturity-scale" replace />} />
          <Route path="/maturity-scale" element={<MaturityScaleSettings />} />
          <Route path="/strategy-pillars" element={<StrategyPillarsSettings />} />
          <Route path="/ai-configuration" element={<AIConfigurationSettings />} />
          <Route path="/one-pagers" element={<OnePagersSettings />} />
          <Route path="/appearance" element={<AppearanceSettings />} />
        </Routes>
      </div>
    </div>
  );
}
