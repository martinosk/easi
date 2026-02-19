import { useState } from 'react';

const toolDisplayNames: Record<string, string> = {
  list_applications: 'Searching applications',
  get_application_details: 'Looking up application',
  list_application_relations: 'Finding relations',
  list_capabilities: 'Searching capabilities',
  get_capability_details: 'Looking up capability',
  list_business_domains: 'Searching business domains',
  get_business_domain_details: 'Looking up business domain',
  list_value_streams: 'Searching value streams',
  get_value_stream_details: 'Looking up value stream',
  search_architecture: 'Searching architecture',
  get_portfolio_summary: 'Getting portfolio summary',
  create_application: 'Creating application',
  update_application: 'Updating application',
  delete_application: 'Deleting application',
  create_capability: 'Creating capability',
  update_capability: 'Updating capability',
  delete_capability: 'Deleting capability',
  create_business_domain: 'Creating business domain',
  update_business_domain: 'Updating business domain',
  create_application_relation: 'Creating relation',
  delete_application_relation: 'Deleting relation',
  realize_capability: 'Linking to capability',
  unrealize_capability: 'Unlinking from capability',
};

type ToolCategory = 'query' | 'mutate' | 'delete';

const deletePrefixes = ['delete_', 'unrealize_'];
const mutatePrefixes = ['create_', 'update_', 'realize_'];

function matchesAnyPrefix(name: string, prefixes: string[]): boolean {
  return prefixes.some(p => name.startsWith(p));
}

function categorize(name: string): ToolCategory {
  if (matchesAnyPrefix(name, deletePrefixes)) return 'delete';
  if (matchesAnyPrefix(name, mutatePrefixes)) return 'mutate';
  return 'query';
}

const categoryIcons: Record<ToolCategory, string> = {
  query: '\uD83D\uDD0D',
  mutate: '\u270F',
  delete: '\uD83D\uDDD1',
};

const statusIcons: Record<string, string> = {
  completed: '\u2713',
  error: '\u26A0',
};

interface ToolCallIndicatorProps {
  status: 'running' | 'completed' | 'error';
  name: string;
  resultPreview?: string;
  errorMessage?: string;
}

function StatusIndicator({ status }: { status: string }) {
  if (status === 'running') return <span className="tool-call-pulse" />;
  const icon = statusIcons[status];
  if (!icon) return null;
  return <span className="tool-call-status-icon">{icon}</span>;
}

function hasVisiblePreview(expanded: boolean, resultPreview?: string): boolean {
  return expanded && Boolean(resultPreview);
}

function StatusDetail({ status, errorMessage, expanded, resultPreview }: {
  status: string;
  errorMessage?: string;
  expanded: boolean;
  resultPreview?: string;
}) {
  if (status === 'running') return <span className="tool-call-activity">Looking up data...</span>;
  if (status === 'error' && errorMessage) return <span className="tool-call-error-message">{errorMessage}</span>;
  if (status === 'completed' && hasVisiblePreview(expanded, resultPreview)) return <div className="tool-call-preview">{resultPreview}</div>;
  return null;
}

export function ToolCallIndicator({ status, name, resultPreview, errorMessage }: ToolCallIndicatorProps) {
  const [expanded, setExpanded] = useState(false);
  const label = toolDisplayNames[name] ?? name;
  const icon = categoryIcons[categorize(name)];

  return (
    <div
      className={`tool-call-indicator tool-call-${status}`}
      onClick={() => setExpanded(prev => !prev)}
    >
      <span className="tool-call-category-icon">{icon}</span>
      <StatusIndicator status={status} />
      <span className="tool-call-label">{label}</span>
      <StatusDetail status={status} errorMessage={errorMessage} expanded={expanded} resultPreview={resultPreview} />
    </div>
  );
}
