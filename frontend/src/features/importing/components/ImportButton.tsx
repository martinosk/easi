import React from 'react';

interface ImportButtonProps {
  onClick: () => void;
}

export const ImportButton: React.FC<ImportButtonProps> = ({ onClick }) => {
  return (
    <button
      className="btn btn-secondary"
      onClick={onClick}
      data-testid="import-button"
      title="Import from ArchiMate Open Exchange"
    >
      Import
    </button>
  );
};
