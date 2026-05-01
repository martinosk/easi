import React, { useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import type { RelatedLink } from '../../../utils/xRelated';

interface HandleCreatePickerProps {
  x: number;
  y: number;
  entries: RelatedLink[];
  onSelect: (entry: RelatedLink) => void;
  onClose: () => void;
}

export const HandleCreatePicker: React.FC<HandleCreatePickerProps> = ({ x, y, entries, onSelect, onClose }) => {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) onClose();
    };
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleEscape);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [onClose]);

  if (entries.length === 0) return null;

  const handleSelect = (entry: RelatedLink) => {
    onSelect(entry);
    onClose();
  };

  return createPortal(
    <div
      ref={ref}
      className="context-menu handle-create-picker"
      role="menu"
      aria-label="Create related entity"
      data-testid="handle-create-picker"
      style={{ left: x, top: y }}
    >
      {entries.map((entry) => (
        <button
          key={entry.relationType}
          type="button"
          className="context-menu-item"
          role="menuitem"
          onClick={() => handleSelect(entry)}
        >
          <span>{entry.title}</span>
        </button>
      ))}
      <button
        type="button"
        className="context-menu-item handle-create-picker-cancel"
        onClick={onClose}
        aria-label="Cancel"
      >
        Cancel
      </button>
    </div>,
    document.body,
  );
};
