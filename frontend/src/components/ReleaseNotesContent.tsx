import React, { useMemo } from 'react';
import type { Release } from '../api/types';
import {
  parseMarkdownSections,
  formatInlineMarkdown,
  formatReleaseDate,
  getSectionStyle,
} from './releaseNotesUtils';

interface ReleaseNotesContentProps {
  selectedRelease: Release | null;
}

export const ReleaseNotesContent: React.FC<ReleaseNotesContentProps> = ({
  selectedRelease,
}) => {
  const sections = useMemo(() => {
    if (!selectedRelease) return [];
    return parseMarkdownSections(selectedRelease.notes);
  }, [selectedRelease]);

  if (!selectedRelease) {
    return (
      <div className="release-browser-empty">
        <p>No release selected</p>
      </div>
    );
  }

  return (
    <>
      <div className="release-browser-main-header">
        <h3 className="release-browser-main-title">
          Version {selectedRelease.version}
        </h3>
        <span className="release-browser-main-date">
          {formatReleaseDate(selectedRelease.releaseDate)}
        </span>
      </div>

      <div className="release-browser-notes">
        {sections.length > 0 ? (
          sections.map((section, index) => {
            const sectionStyle = getSectionStyle(section.title);
            return (
              <div
                key={index}
                className={`release-notes-section ${sectionStyle.className}`}
              >
                <h4 className="release-notes-section-title">
                  <span className="release-notes-section-icon">
                    {sectionStyle.icon}
                  </span>
                  {section.title}
                </h4>
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
            {selectedRelease.notes.split('\n').map((line, index) => (
              <p key={index}>{formatInlineMarkdown(line) || '\u00A0'}</p>
            ))}
          </div>
        )}
      </div>
    </>
  );
};