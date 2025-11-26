import React from 'react';

export const LoadingScreen: React.FC = () => {
  return (
    <div className="loading-container">
      <div className="loading-spinner"></div>
      <p>Loading component modeler...</p>
    </div>
  );
};
