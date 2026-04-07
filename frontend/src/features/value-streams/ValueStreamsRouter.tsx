import { Route, Routes } from 'react-router-dom';
import { ValueStreamDetailPage } from './pages/ValueStreamDetailPage';
import { ValueStreamsPage } from './pages/ValueStreamsPage';

export function ValueStreamsRouter() {
  return (
    <Routes>
      <Route index element={<ValueStreamsPage />} />
      <Route path=":valueStreamId" element={<ValueStreamDetailPage />} />
    </Routes>
  );
}
