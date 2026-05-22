import { Switch } from '@mantine/core';

export interface ShowApplicationsToggleProps {
  showApplications: boolean;
  onShowApplicationsChange: (value: boolean) => void;
}

export function ShowApplicationsToggle({ showApplications, onShowApplicationsChange }: ShowApplicationsToggleProps) {
  return (
    <Switch
      label="Apps"
      checked={showApplications}
      onChange={(e) => onShowApplicationsChange(e.currentTarget.checked)}
    />
  );
}
