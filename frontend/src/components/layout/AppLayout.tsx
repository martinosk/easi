import React from 'react';
import { Toaster } from 'react-hot-toast';

interface AppLayoutProps {
  children: React.ReactNode;
}

export const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  return (
    <div className="app-container">
      {children}
      <Toaster
        position="bottom-right"
        toastOptions={{
          duration: 3000,
          style: {
            background: 'var(--ink)',
            color: 'var(--surface)',
          },
          success: {
            iconTheme: {
              primary: 'var(--status-positive)',
              secondary: 'var(--surface)',
            },
          },
          error: {
            iconTheme: {
              primary: 'var(--status-danger)',
              secondary: 'var(--surface)',
            },
          },
        }}
      />
    </div>
  );
};
