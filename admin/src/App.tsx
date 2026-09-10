import { Navigate, Route, Routes } from 'react-router-dom'
import type { ReactNode } from 'react'
import { AuthProvider, useAuth } from './auth/AuthContext'
import { AdminLayout } from './components/AdminLayout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { AccessRequestsPage } from './pages/AccessRequestsPage'
import { ClubDetailPage } from './pages/ClubDetailPage'
import { ClubRegistrationsPage } from './pages/ClubRegistrationsPage'
import { ClubsPage } from './pages/ClubsPage'
import { DashboardPage } from './pages/DashboardPage'
import { DonationsPage } from './pages/DonationsPage'
import { EventEditorPage } from './pages/EventEditorPage'
import { EventsPage } from './pages/EventsPage'
import { WebsiteHomePage } from './pages/WebsiteHomePage'
import { LandingPage } from './pages/LandingPage'
import { LoginPage } from './pages/LoginPage'

function PublicOnly({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <div className="page-center">Chargement…</div>
  if (user) return <Navigate to="/dashboard" replace />
  return children
}

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/" element={<PublicOnly><LandingPage /></PublicOnly>} />
        <Route path="/login" element={<PublicOnly><LoginPage /></PublicOnly>} />
        <Route element={<ProtectedRoute />}>
          <Route element={<AdminLayout />}>
            <Route path="dashboard" element={<DashboardPage />} />
            <Route path="clubs" element={<ClubsPage />} />
            <Route path="clubs/:clubId" element={<ClubDetailPage />} />
            <Route path="access-requests" element={<AccessRequestsPage />} />
            <Route path="events" element={<EventsPage />} />
            <Route path="events/:eventId" element={<EventEditorPage />} />
            <Route path="website-home" element={<WebsiteHomePage />} />
            <Route path="donations" element={<DonationsPage />} />
            <Route path="club-registrations" element={<ClubRegistrationsPage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  )
}
