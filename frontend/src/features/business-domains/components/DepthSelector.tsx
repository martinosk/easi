export type DepthLevel = 1 | 2 | 3 | 4;

export interface DepthSelectorProps {
  value: DepthLevel;
  onChange: (depth: DepthLevel) => void;
}

const DEPTH_OPTIONS: { value: DepthLevel; label: string }[] = [
  { value: 1, label: 'L1' },
  { value: 2, label: 'L1-L2' },
  { value: 3, label: 'L1-L3' },
  { value: 4, label: 'L1-L4' },
];

export function DepthSelector({ value, onChange }: DepthSelectorProps) {
  return (
    <div
      style={{
        display: 'flex',
        gap: '0.25rem',
        padding: '0.25rem',
        backgroundColor: '#f3f4f6',
        borderRadius: '0.5rem',
      }}
    >
      {DEPTH_OPTIONS.map((option) => (
        <button
          key={option.value}
          type="button"
          data-selected={value === option.value}
          onClick={() => {
            if (value !== option.value) {
              onChange(option.value);
            }
          }}
          style={{
            padding: '0.5rem 0.75rem',
            borderRadius: '0.375rem',
            border: 'none',
            cursor: value === option.value ? 'default' : 'pointer',
            backgroundColor: value === option.value ? '#3b82f6' : 'transparent',
            color: value === option.value ? 'white' : '#374151',
            fontWeight: 500,
            fontSize: '0.875rem',
            transition: 'all 0.15s ease',
          }}
        >
          {option.label}
        </button>
      ))}
    </div>
  );
}
