import { Navigate, Route, Routes } from 'react-router-dom'
import { I18nProvider } from './i18n'
import { HomePage } from './pages/HomePage'

export default function App() {
  return (
    <I18nProvider>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </I18nProvider>
  )
}
