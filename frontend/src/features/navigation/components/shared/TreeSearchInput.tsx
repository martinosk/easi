import { ActionIcon, TextInput } from '@mantine/core';
import React from 'react';

interface TreeSearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
}

export const TreeSearchInput: React.FC<TreeSearchInputProps> = ({ value, onChange, placeholder }) => (
  <div className="tree-search">
    <TextInput
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.currentTarget.value)}
      size="xs"
      rightSection={
        value ? (
          <ActionIcon variant="subtle" color="gray" size="xs" onClick={() => onChange('')} aria-label="Clear search">
            ×
          </ActionIcon>
        ) : null
      }
    />
  </div>
);
