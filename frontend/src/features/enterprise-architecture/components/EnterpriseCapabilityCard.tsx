import { useState } from 'react';
import toast from 'react-hot-toast';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import type { EnterpriseCapability } from '../types';
import type { Capability } from '../../../api/types';

export interface EnterpriseCapabilityCardProps {
  capability: EnterpriseCapability;
  onDrop: (capability: Capability) => void;
}

export function EnterpriseCapabilityCard({ capability, onDrop }: EnterpriseCapabilityCardProps) {
  const [isDragOver, setIsDragOver] = useState(false);

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setIsDragOver(true);
  };

  const handleDragLeave = () => {
    setIsDragOver(false);
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragOver(false);

    try {
      const data = e.dataTransfer.getData('application/json');
      const domainCapability = JSON.parse(data) as Capability;
      onDrop(domainCapability);
    } catch (error) {
      toast.error('Failed to link capability. Invalid data format.');
    }
  };

  return (
    <div
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      style={{
        padding: '1rem',
        marginBottom: '1rem',
        backgroundColor: isDragOver ? '#dbeafe' : '#ffffff',
        border: isDragOver ? '2px dashed #3b82f6' : '1px solid #e5e7eb',
        borderRadius: '0.5rem',
        transition: 'all 0.2s ease',
      }}
    >
      <div style={{ marginBottom: '0.5rem' }}>
        <h3 style={{ fontSize: '1.125rem', fontWeight: 600, margin: 0 }}>
          {capability.name}
        </h3>
        {capability.category && (
          <span
            style={{
              display: 'inline-block',
              marginTop: '0.25rem',
              fontSize: '0.75rem',
              color: '#6b7280',
              backgroundColor: '#f3f4f6',
              padding: '0.125rem 0.5rem',
              borderRadius: '0.25rem',
            }}
          >
            {capability.category}
          </span>
        )}
      </div>

      {capability.description && (
        <p style={{ fontSize: '0.875rem', color: '#6b7280', margin: '0.5rem 0' }}>
          {capability.description}
        </p>
      )}

      <div style={{ display: 'flex', gap: '1rem', marginTop: '0.75rem' }}>
        <div style={{ fontSize: '0.875rem', color: '#374151', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
          <span style={{ fontWeight: 500 }}>Links:</span>{' '}
          <span style={{ color: '#3b82f6' }}>{capability.linkCount}</span>
          <HelpTooltip content="Number of domain capabilities linked to this enterprise capability" iconOnly />
        </div>
        <div style={{ fontSize: '0.875rem', color: '#374151', display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
          <span style={{ fontWeight: 500 }}>Domains:</span>{' '}
          <span style={{ color: '#3b82f6' }}>{capability.domainCount}</span>
          <HelpTooltip content="Number of business domains containing linked capabilities" iconOnly />
        </div>
      </div>
      {capability.linkCount === 0 && (
        <div style={{
          fontSize: '0.75rem',
          color: '#6b7280',
          marginTop: '0.5rem',
          fontStyle: 'italic',
          display: 'flex',
          alignItems: 'center',
          gap: '0.25rem'
        }}>
          Drag domain capabilities here to link them
        </div>
      )}
    </div>
  );
}
