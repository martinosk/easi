import { Routes, Route } from 'react-router-dom';
import { ValueStreamsPage } from './pages/ValueStreamsPage';
import { ValueStreamDetailPage } from './pages/ValueStreamDetailPage';

export function ValueStreamsRouter() {
  return (
    <Routes>
      <Route index element={<ValueStreamsPage />} />
      <Route path=":valueStreamId" element={<ValueStreamDetailPage />} />
    </Routes>
  );
}
