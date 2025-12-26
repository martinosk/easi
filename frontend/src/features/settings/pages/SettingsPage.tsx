import { Routes, Route, Navigate } from 'react-router-dom';
import { MaturityScaleSettings } from '../components/MaturityScaleSettings';
import { useUserStore } from '../../../store/userStore';
import './SettingsPage.css';

export function SettingsPage() {
  const hasPermission = useUserStore((state) => state.hasPermission);
  const canManageMetaModel = hasPermission('metamodel:write');

  if (!canManageMetaModel) {
    return (
      <div className="settings-page">
        <div className="settings-container">
          <div className="error-message">
            You do not have permission to manage settings.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="settings-page">
      <div className="settings-container">
        <div className="settings-header">
          <h1 className="settings-title">Settings</h1>
          <p className="settings-subtitle">
            Configure system-wide settings for your organization.
          </p>
        </div>

        <Routes>
          <Route path="/" element={<Navigate to="/settings/maturity-scale" replace />} />
          <Route path="/maturity-scale" element={<MaturityScaleSettings />} />
        </Routes>
      </div>
    </div>
  );
}
