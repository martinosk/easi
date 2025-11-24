import React, { useRef, useEffect, useState } from 'react';
import { apiClient } from '../api/client';
import type { Release } from '../api/types';
import { ReleaseNotesSidebar } from './ReleaseNotesSidebar';
import { ReleaseNotesContent } from './ReleaseNotesContent';

interface ReleaseNotesBrowserProps {
  isOpen: boolean;
  onClose: () => void;
}

export const ReleaseNotesBrowser: React.FC<ReleaseNotesBrowserProps> = ({
  isOpen,
  onClose,
}) => {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [releases, setReleases] = useState<Release[]>([]);
  const [currentVersion, setCurrentVersion] = useState<string | null>(null);
  const [selectedRelease, setSelectedRelease] = useState<Release | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchData() {
      try {
        setIsLoading(true);
        setError(null);
        const [releasesData, version] = await Promise.all([
          apiClient.getReleases(),
          apiClient.getVersion(),
        ]);
        const sortedReleases = [...releasesData].sort((a, b) =>
          new Date(b.releaseDate).getTime() - new Date(a.releaseDate).getTime()
        );
        setReleases(sortedReleases);
        setCurrentVersion(version);
        if (sortedReleases.length > 0) {
          setSelectedRelease(sortedReleases[0]);
        }
      } catch {
        setError('Failed to load release notes');
      } finally {
        setIsLoading(false);
      }
    }

    if (isOpen) {
      fetchData();
    }
  }, [isOpen]);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  const handleDialogClose = () => {
    onClose();
  };

  return (
    <dialog
      ref={dialogRef}
      className="dialog release-browser-dialog"
      onClose={handleDialogClose}
      data-testid="release-notes-browser"
    >
      <div className="release-browser-content">
        <div className="release-browser-header">
          <h2 className="dialog-title">Release Notes</h2>
          {currentVersion && (
            <span className="release-browser-current-version">
              Current: v{currentVersion}
            </span>
          )}
          <button
            type="button"
            className="release-browser-close"
            onClick={onClose}
            aria-label="Close"
          >
            ×
          </button>
        </div>

        {isLoading && (
          <div className="release-browser-loading">
            <div className="loading-spinner" />
            <p>Loading release notes...</p>
          </div>
        )}

        {error && (
          <div className="release-browser-error">
            <p>{error}</p>
          </div>
        )}

        {!isLoading && !error && (
          <div className="release-browser-body">
            <ReleaseNotesSidebar
              releases={releases}
              selectedRelease={selectedRelease}
              currentVersion={currentVersion}
              onSelectRelease={setSelectedRelease}
            />

            <main className="release-browser-main">
              <ReleaseNotesContent selectedRelease={selectedRelease} />
            </main>
          </div>
        )}

        <div className="dialog-actions">
          <button
            type="button"
            className="btn btn-primary"
            onClick={onClose}
          >
            Close
          </button>
        </div>
      </div>
    </dialog>
  );
};