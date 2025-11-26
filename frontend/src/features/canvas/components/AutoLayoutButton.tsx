import React, { useState } from 'react';
import { useAppStore } from '../../../store/appStore';

export const AutoLayoutButton: React.FC = () => {
  const applyAutoLayout = useAppStore((state) => state.applyAutoLayout);
  const [isApplying, setIsApplying] = useState(false);

  const handleClick = async () => {
    setIsApplying(true);
    try {
      await applyAutoLayout();
    } finally {
      setIsApplying(false);
    }
  };

  return (
    <button
      className="btn btn-secondary btn-small"
      onClick={handleClick}
      disabled={isApplying}
      aria-label="Apply automatic layout to components"
    >
      {isApplying ? (
        <>
          <span className="spinner-small"></span>
          Applying...
        </>
      ) : (
        <>
          <span>⚡</span>
          Auto Layout
        </>
      )}
    </button>
  );
};
