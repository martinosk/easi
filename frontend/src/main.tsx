import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { MantineProvider } from '@mantine/core'
import 'dockview/dist/styles/dockview.css'
import '@mantine/core/styles.css'
import './index.css'
import App from './App.tsx'
import { LoginPage } from './features/auth/pages/LoginPage.tsx'
import { theme } from './theme/mantine'

const basename = import.meta.env.BASE_URL.replace(/\/$/, '') || '';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <MantineProvider theme={theme} defaultColorScheme="light">
      <BrowserRouter basename={basename}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<App />} />
        </Routes>
      </BrowserRouter>
    </MantineProvider>
  </StrictMode>,
)
