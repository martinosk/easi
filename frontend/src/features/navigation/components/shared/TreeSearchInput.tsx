import React from 'react';

interface TreeSearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
}

export const TreeSearchInput: React.FC<TreeSearchInputProps> = ({
  value,
  onChange,
  placeholder,
}) => (
  <div className="tree-search">
    <input
      type="text"
      className="tree-search-input"
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
    {value && (
      <button
        className="tree-search-clear"
        onClick={() => onChange('')}
        aria-label="Clear search"
      >
        x
      </button>
    )}
  </div>
);
