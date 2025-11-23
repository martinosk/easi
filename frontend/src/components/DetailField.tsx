import React from 'react';

interface DetailFieldProps {
  label: string;
  children: React.ReactNode;
}

export const DetailField: React.FC<DetailFieldProps> = ({ label, children }) => (
  <div className="detail-field">
    <label className="detail-label">{label}</label>
    <div className="detail-value">{children}</div>
  </div>
);
