import React, { useRef, useEffect, useMemo } from 'react';
import type { Release } from '../../../api/types';
import {
  parseMarkdownSections,
  formatInlineMarkdown,
  formatReleaseDate,
  getSectionStyle,
} from './releaseNotesUtils';

interface ReleaseNotesOverlayProps {
  isOpen: boolean;
  release: Release;
  onDismiss: (mode: 'forever' | 'untilNext') => void;
}

export const ReleaseNotesOverlay: React.FC<ReleaseNotesOverlayProps> = ({
  isOpen,
  release,
  onDismiss,
}) => {
  const dialogRef = useRef<HTMLDialogElement>(null);

  const sections = useMemo(() => parseMarkdownSections(release.notes), [release.notes]);

  const formattedDate = useMemo(
    () => formatReleaseDate(release.releaseDate),
    [release.releaseDate]
  );

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
    onDismiss('untilNext');
  };

  return (
    <dialog
      ref={dialogRef}
      className="dialog release-notes-dialog"
      onClose={handleDialogClose}
      data-testid="release-notes-overlay"
    >
      <div className="dialog-content">
        <div className="release-notes-header">
          <h2 className="dialog-title">What's New in {release.version}</h2>
          <span className="release-notes-date">{formattedDate}</span>
        </div>

        <div className="release-notes-body">
          {sections.length > 0 ? (
            sections.map((section, index) => {
              const sectionStyle = getSectionStyle(section.title);
              return (
                <div
                  key={index}
                  className={`release-notes-section ${sectionStyle.className}`}
                >
                  <h3 className="release-notes-section-title">
                    <span className="release-notes-section-icon">
                      {sectionStyle.icon}
                    </span>
                    {section.title}
                  </h3>
                  <ul className="release-notes-list">
                    {section.items.map((item, itemIndex) => (
                      <li key={itemIndex} className="release-notes-item">
                        {formatInlineMarkdown(item)}
                      </li>
                    ))}
                  </ul>
                </div>
              );
            })
          ) : (
            <div className="release-notes-raw">
              {release.notes.split('\n').map((line, index) => (
                <p key={index}>{formatInlineMarkdown(line) || '\u00A0'}</p>
              ))}
            </div>
          )}
        </div>

        <div className="dialog-actions release-notes-actions">
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => onDismiss('forever')}
            data-testid="release-notes-hide-forever"
          >
            Don't show again
          </button>
          <button
            type="button"
            className="btn btn-primary"
            onClick={() => onDismiss('untilNext')}
            data-testid="release-notes-dismiss"
          >
            Got it
          </button>
        </div>
      </div>
    </dialog>
  );
};