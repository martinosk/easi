import React from 'react';
import type { Release } from '../../../api/types';
import { formatShortDate } from './releaseNotesUtils';

interface ReleaseNotesSidebarProps {
  releases: Release[];
  selectedRelease: Release | null;
  currentVersion: string | null;
  onSelectRelease: (release: Release) => void;
}

export const ReleaseNotesSidebar: React.FC<ReleaseNotesSidebarProps> = ({
  releases,
  selectedRelease,
  currentVersion,
  onSelectRelease,
}) => {
  return (
    <aside className="release-browser-sidebar">
      <h3 className="release-browser-sidebar-title">Releases</h3>
      <ul className="release-browser-list">
        {releases.map((release) => (
          <li key={release.version}>
            <button
              type="button"
              className={`release-browser-item ${
                selectedRelease?.version === release.version ? 'selected' : ''
              } ${release.version === currentVersion ? 'current' : ''}`}
              onClick={() => onSelectRelease(release)}
            >
              <span className="release-browser-item-version">
                v{release.version}
                {release.version === currentVersion && (
                  <span className="release-browser-item-badge">current</span>
                )}
              </span>
              <span className="release-browser-item-date">
                {formatShortDate(release.releaseDate)}
              </span>
            </button>
          </li>
        ))}
      </ul>
    </aside>
  );
};