import type { CapabilityTreeNode } from '../types/visualization';

interface CapabilityDetailPanelProps {
  capability: CapabilityTreeNode;
  onClose: () => void;
}

const LEVEL_COLORS = {
  L1: 'bg-blue-100 text-blue-800 border-blue-300',
  L2: 'bg-purple-100 text-purple-800 border-purple-300',
  L3: 'bg-pink-100 text-pink-800 border-pink-300',
  L4: 'bg-orange-100 text-orange-800 border-orange-300',
};

export function CapabilityDetailPanel({ capability, onClose }: CapabilityDetailPanelProps) {
  return (
    <div className="w-96 bg-white border-l border-gray-200 h-full overflow-hidden flex flex-col">
      <div className="p-4 border-b border-gray-200 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900">Capability Details</h2>
        <button
          onClick={onClose}
          className="p-1 text-gray-400 hover:text-gray-600 rounded-full hover:bg-gray-100"
          aria-label="Close panel"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-6">
        <div>
          <div className="flex items-center gap-2 mb-3">
            <span className={`px-2 py-1 text-xs font-semibold rounded border ${LEVEL_COLORS[capability.level]}`}>
              {capability.level}
            </span>
            <span className="text-sm text-gray-500">{capability.code}</span>
          </div>
          <h3 className="text-xl font-semibold text-gray-900 mb-2">{capability.name}</h3>
          {capability.description && (
            <p className="text-sm text-gray-600">{capability.description}</p>
          )}
        </div>

        {capability.parentId && (
          <div>
            <h4 className="text-sm font-semibold text-gray-700 mb-2">Parent Capability</h4>
            <div className="p-3 bg-gray-50 rounded-lg">
              <div className="text-sm text-gray-900">View parent in hierarchy</div>
            </div>
          </div>
        )}

        {capability.children.length > 0 && (
          <div>
            <h4 className="text-sm font-semibold text-gray-700 mb-2">
              Child Capabilities ({capability.children.length})
            </h4>
            <div className="space-y-2">
              {capability.children.map((child) => (
                <div key={child.id} className="p-3 bg-gray-50 rounded-lg">
                  <div className="flex items-center gap-2 mb-1">
                    <span className={`px-2 py-1 text-xs font-semibold rounded border ${LEVEL_COLORS[child.level]}`}>
                      {child.level}
                    </span>
                    <span className="text-sm font-medium text-gray-900">{child.name}</span>
                  </div>
                  {child.description && (
                    <p className="text-xs text-gray-600 mt-1">{child.description}</p>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {capability.realizingSystems.length > 0 && (
          <div>
            <h4 className="text-sm font-semibold text-gray-700 mb-2">
              Realizing Systems ({capability.realizingSystems.length})
            </h4>
            <div className="space-y-2">
              {capability.realizingSystems.map((system) => (
                <div key={system.componentId} className="p-3 bg-blue-50 rounded-lg">
                  <div className="flex items-center gap-2">
                    <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"
                      />
                    </svg>
                    <span className="text-sm font-medium text-gray-900">{system.componentName}</span>
                  </div>
                  <div className="text-xs text-gray-600 mt-1">{system.componentType}</div>
                </div>
              ))}
            </div>
          </div>
        )}

        {capability.associatedDomains.length > 0 && (
          <div>
            <h4 className="text-sm font-semibold text-gray-700 mb-2">
              Business Domains ({capability.associatedDomains.length})
            </h4>
            <div className="space-y-2">
              {capability.associatedDomains.map((domain, idx) => (
                <div key={idx} className="p-3 bg-purple-50 rounded-lg">
                  <span className="text-sm text-gray-900">{domain}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {capability.isOrphaned && (
          <div className="p-4 bg-orange-50 border border-orange-200 rounded-lg">
            <div className="flex items-center gap-2 text-orange-800">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
              <div className="text-sm font-semibold">Unassigned Capability</div>
            </div>
            <p className="text-xs text-orange-700 mt-2">
              This L1 capability is not assigned to any business domain.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
