import { TextInput, UnstyledButton } from '@mantine/core';
import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { HexColorPicker } from 'react-colorful';
import classes from './ColorPicker.module.css';

interface ColorPickerProps {
  color: string | null;
  onChange: (color: string) => void;
  disabled: boolean;
  disabledTooltip?: string;
}

export function ColorPicker({ color, onChange, disabled, disabledTooltip }: ColorPickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [tempColor, setTempColor] = useState(color || '#E0E0E0');
  const displayColor = tempColor;
  const commitTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pickerRef = useRef<HTMLDivElement>(null);

  useLayoutEffect(() => {
    const normalized = color || '#E0E0E0';
    if (tempColor !== normalized) queueMicrotask(() => setTempColor(normalized));
  }, [color, tempColor]);

  useEffect(() => {
    return () => {
      if (commitTimeoutRef.current) {
        clearTimeout(commitTimeoutRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (event: MouseEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(event.target as Node)) {
        if (commitTimeoutRef.current) {
          clearTimeout(commitTimeoutRef.current);
        }
        const upperColor = tempColor.toUpperCase();
        if (upperColor !== color?.toUpperCase()) {
          onChange(upperColor);
        }
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen, tempColor, color, onChange]);

  const handleButtonClick = () => {
    if (!disabled) {
      setIsOpen(!isOpen);
    }
  };

  const commitColor = (newColor: string) => {
    const upperColor = newColor.toUpperCase();
    if (upperColor !== color?.toUpperCase()) {
      onChange(upperColor);
    }
  };

  const handleColorChange = (newColor: string) => {
    setTempColor(newColor.toUpperCase());

    if (commitTimeoutRef.current) {
      clearTimeout(commitTimeoutRef.current);
    }

    commitTimeoutRef.current = setTimeout(() => {
      commitColor(newColor);
    }, 300);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setTempColor(value);
    if (/^#[0-9A-Fa-f]{6}$/.test(value)) {
      commitColor(value);
      setIsOpen(false);
    }
  };

  return (
    <div ref={pickerRef} className={classes.wrapper}>
      <UnstyledButton
        component="button"
        type="button"
        data-testid="color-picker-button"
        onClick={handleButtonClick}
        disabled={disabled}
        title={disabled && disabledTooltip ? disabledTooltip : undefined}
        className={classes.button}
      >
        <div data-testid="color-picker-display" className={classes.swatch} style={{ backgroundColor: displayColor }} />
        <span>{displayColor}</span>
      </UnstyledButton>

      {!disabled && (
        <div
          data-testid="color-picker-popover"
          className={classes.popover}
          style={{ display: isOpen ? 'block' : 'none' }}
        >
          <HexColorPicker color={displayColor} onChange={handleColorChange} />
          <TextInput
            data-testid="color-picker-input"
            value={displayColor}
            onChange={handleInputChange}
            mt="xs"
            styles={{ input: { fontFamily: 'monospace' } }}
          />
        </div>
      )}

      {disabled && disabledTooltip && (
        <div data-testid="color-picker-tooltip" style={{ display: 'none' }}>
          {disabledTooltip}
        </div>
      )}
    </div>
  );
}
