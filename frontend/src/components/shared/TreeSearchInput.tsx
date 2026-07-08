import { ActionIcon, Box, TextInput } from '@mantine/core';
import React from 'react';
import classes from './TreeSearchInput.module.css';

interface TreeSearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  testId?: string;
}

export const TreeSearchInput: React.FC<TreeSearchInputProps> = ({ value, onChange, placeholder, testId }) => (
  <Box className={classes.root}>
    <TextInput
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.currentTarget.value)}
      size="xs"
      data-testid={testId}
      rightSection={
        value ? (
          <ActionIcon variant="subtle" color="gray" size="xs" onClick={() => onChange('')} aria-label="Clear search">
            ×
          </ActionIcon>
        ) : null
      }
    />
  </Box>
);
