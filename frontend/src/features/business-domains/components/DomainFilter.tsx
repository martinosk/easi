import { useState } from 'react';
import type { BusinessDomain, BusinessDomainId } from '../../../api/types';

interface DomainFilterProps {
  domains: BusinessDomain[];
  selected: BusinessDomainId | null;
  onSelect: (domainId: BusinessDomainId | null) => void;
}

export function DomainFilter({ domains, selected, onSelect }: DomainFilterProps) {
  const [searchTerm, setSearchTerm] = useState('');

  const filteredDomains = domains.filter(domain =>
    domain.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    domain.description.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="w-64 bg-white border-r border-gray-200 h-full overflow-hidden flex flex-col">
      <div className="p-4 border-b border-gray-200">
        <h2 className="text-lg font-semibold text-gray-900 mb-3">Business Domains</h2>
        <input
          type="text"
          placeholder="Search domains..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <div className="flex-1 overflow-y-auto">
        <button
          onClick={() => onSelect(null)}
          className={`w-full px-4 py-3 text-left border-b border-gray-100 hover:bg-gray-50 transition-colors ${
            selected === null ? 'bg-blue-50 border-l-4 border-l-blue-600' : ''
          }`}
        >
          <div className="flex items-center">
            <svg
              className="w-5 h-5 mr-2 text-orange-500"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
            <span className="text-sm font-medium text-gray-900">Orphaned Capabilities</span>
          </div>
        </button>

        {filteredDomains.map((domain) => (
          <button
            key={domain.id}
            onClick={() => onSelect(domain.id)}
            title={domain.description}
            className={`w-full px-4 py-3 text-left border-b border-gray-100 hover:bg-gray-50 transition-colors ${
              selected === domain.id ? 'bg-blue-50 border-l-4 border-l-blue-600' : ''
            }`}
          >
            <div className="text-sm font-medium text-gray-900 truncate">{domain.name}</div>
            <div className="text-xs text-gray-500 mt-1">
              {domain.capabilityCount} {domain.capabilityCount === 1 ? 'capability' : 'capabilities'}
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
