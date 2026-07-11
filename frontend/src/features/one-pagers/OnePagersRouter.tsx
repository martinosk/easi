import { Route, Routes } from 'react-router-dom';
import { OnePagerPage } from './pages/OnePagerPage';

export function OnePagersRouter() {
  return (
    <Routes>
      <Route path=":subjectType/:subjectId" element={<OnePagerPage />} />
    </Routes>
  );
}
