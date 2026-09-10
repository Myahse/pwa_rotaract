import { Navigate, Route, Routes } from 'react-router-dom'
import { I18nProvider } from './i18n'
import { Layout } from './components/Layout'
import { HomePage } from './pages/HomePage'
import { DonatePage } from './pages/DonatePage'
import { EventDetailPage } from './pages/EventDetailPage'
import { EventsPage } from './pages/EventsPage'
import { JoinPage } from './pages/JoinPage'

export default function App() {
  return (
    <I18nProvider>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/events" element={<EventsPage />} />
          <Route path="/events/:eventId" element={<EventDetailPage />} />
          <Route path="/donate" element={<DonatePage />} />
          <Route path="/join" element={<JoinPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </I18nProvider>
  )
}
