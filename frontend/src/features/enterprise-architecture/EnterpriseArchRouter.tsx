import { useState } from 'react';
import { EnterpriseArchPage } from './pages/EnterpriseArchPage';
import { EnterpriseArchLinkingPage } from './pages/EnterpriseArchLinkingPage';

type View = 'list' | 'linking';

export function EnterpriseArchRouter() {
  const [view, setView] = useState<View>('list');

  const handleNavigateToLinking = () => setView('linking');
  const handleNavigateToList = () => setView('list');

  if (view === 'linking') {
    return <EnterpriseArchLinkingPage onNavigateBack={handleNavigateToList} />;
  }

  return <EnterpriseArchPage onNavigateToLinking={handleNavigateToLinking} />;
}
