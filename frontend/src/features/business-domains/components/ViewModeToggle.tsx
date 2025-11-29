export type ViewMode = 'treemap' | 'tree' | 'grid';

export interface ViewModeToggleProps {
  mode: ViewMode;
  onModeChange: (mode: ViewMode) => void;
}

export function ViewModeToggle({ mode, onModeChange }: ViewModeToggleProps) {
  const modes: { value: ViewMode; label: string }[] = [
    { value: 'treemap', label: 'Treemap' },
    { value: 'tree', label: 'Tree' },
    { value: 'grid', label: 'Grid' },
  ];

  return (
    <div className="inline-flex rounded-md shadow-sm" role="group">
      {modes.map((modeOption, index) => {
        const isSelected = mode === modeOption.value;
        const isFirst = index === 0;
        const isLast = index === modes.length - 1;

        return (
          <button
            key={modeOption.value}
            type="button"
            onClick={() => onModeChange(modeOption.value)}
            className={`
              px-4 py-2 text-sm font-medium
              ${isFirst ? 'rounded-l-md' : ''}
              ${isLast ? 'rounded-r-md' : ''}
              ${!isFirst && !isLast ? '' : ''}
              ${
                isSelected
                  ? 'bg-blue-600 text-white z-10'
                  : 'bg-white text-gray-700 hover:bg-gray-50'
              }
              border border-gray-300
              ${!isFirst ? '-ml-px' : ''}
              focus:z-10 focus:outline-none focus:ring-2 focus:ring-blue-500
            `}
          >
            {modeOption.label}
          </button>
        );
      })}
    </div>
  );
}
