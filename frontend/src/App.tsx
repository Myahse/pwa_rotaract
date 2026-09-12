import { Navigate, Route, Routes } from 'react-router-dom'
import type { ReactNode } from 'react'
import { AuthProvider, useAuth } from './auth/AuthContext'
import { AppLayout } from './components/AppLayout'
import { ClubAccessRoute } from './components/ClubAccessRoute'
import { ProtectedRoute } from './components/ProtectedRoute'
import { AccessRequestPage } from './pages/AccessRequestPage'
import { ChatPage } from './pages/ChatPage'
import { MessagesInboxPage } from './pages/MessagesInboxPage'
import { ClubPage } from './pages/ClubPage'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { ProfilePage } from './pages/ProfilePage'
import { HeadManagePage } from './pages/HeadManagePage'
import { CotisationsPage } from './pages/CotisationsPage'
import { ForgotPasswordPage } from './pages/ForgotPasswordPage'
import { ResetPasswordPage } from './pages/ResetPasswordPage'
import { ClubRegisterCompletePage } from './pages/ClubRegisterCompletePage'

function PublicOnly({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <div className="page-center">Chargement…</div>
  if (user) return <Navigate to="/home" replace />
  return children
}

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/" element={<PublicOnly><LoginPage /></PublicOnly>} />
        <Route path="/login" element={<Navigate to="/" replace />} />
        <Route path="/register" element={<PublicOnly><RegisterPage /></PublicOnly>} />
        <Route path="/register-club" element={<PublicOnly><ClubRegisterCompletePage /></PublicOnly>} />
        <Route path="/access-request" element={<PublicOnly><AccessRequestPage /></PublicOnly>} />
        <Route path="/forgot-password" element={<PublicOnly><ForgotPasswordPage /></PublicOnly>} />
        <Route path="/reset-password" element={<PublicOnly><ResetPasswordPage /></PublicOnly>} />
        <Route element={<ProtectedRoute />}>
          <Route element={<AppLayout />}>
            <Route path="home" element={<HomePage />} />
            <Route path="profile" element={<ProfilePage />} />
            <Route path="clubs/:clubId" element={<ClubAccessRoute section="club"><ClubPage /></ClubAccessRoute>} />
            <Route path="clubs/:clubId/manage" element={<ClubAccessRoute section="manage"><HeadManagePage /></ClubAccessRoute>} />
            <Route path="clubs/:clubId/cotisations" element={<ClubAccessRoute section="cotisations"><CotisationsPage /></ClubAccessRoute>} />
            <Route path="clubs/:clubId/messages" element={<ClubAccessRoute section="messages"><MessagesInboxPage /></ClubAccessRoute>} />
            <Route path="chat/:groupId" element={<ChatPage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  )
}
