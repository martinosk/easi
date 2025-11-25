import { useState, useRef, useEffect } from 'react';
import { HexColorPicker } from 'react-colorful';

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
  const commitTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const pickerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setTempColor(color || '#E0E0E0');
  }, [color]);

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
    <div ref={pickerRef} style={{ position: 'relative', display: 'inline-block' }}>
      <button
        data-testid="color-picker-button"
        onClick={handleButtonClick}
        disabled={disabled}
        title={disabled && disabledTooltip ? disabledTooltip : undefined}
        style={{
          padding: '8px',
          border: '1px solid #ccc',
          borderRadius: '4px',
          background: 'white',
          cursor: disabled ? 'not-allowed' : 'pointer',
          opacity: disabled ? 0.6 : 1,
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
        }}
      >
        <div
          data-testid="color-picker-display"
          style={{
            width: '32px',
            height: '32px',
            backgroundColor: displayColor,
            border: '1px solid #ccc',
            borderRadius: '4px',
          }}
        />
        <span>{displayColor}</span>
      </button>

      {!disabled && (
        <div
          data-testid="color-picker-popover"
          style={{
            display: isOpen ? 'block' : 'none',
            position: 'absolute',
            top: '100%',
            left: 0,
            marginTop: '4px',
            zIndex: 1000,
            padding: '16px',
            background: 'white',
            border: '1px solid #ccc',
            borderRadius: '8px',
            boxShadow: '0 4px 8px rgba(0, 0, 0, 0.2)',
          }}
        >
          <HexColorPicker color={displayColor} onChange={handleColorChange} />
          <input
            data-testid="color-picker-input"
            type="text"
            value={displayColor}
            onChange={handleInputChange}
            style={{
              marginTop: '8px',
              width: '100%',
              padding: '4px 8px',
              border: '1px solid #ccc',
              borderRadius: '4px',
              fontFamily: 'monospace',
            }}
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
