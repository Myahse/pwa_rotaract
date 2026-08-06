import { Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './auth/AuthContext'
import { SocialLayout } from './components/SocialLayout'
import { FeedPage } from './pages/FeedPage'
import { FollowingPage } from './pages/FollowingPage'
import { FriendsPage } from './pages/FriendsPage'
import { GroupDetailPage } from './pages/GroupDetailPage'
import { GroupsPage } from './pages/GroupsPage'
import { MessagesPage } from './pages/MessagesPage'
import { PostDetailPage } from './pages/PostDetailPage'
import { UserProfilePage } from './pages/UserProfilePage'

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route element={<SocialLayout />}>
          <Route index element={<FeedPage />} />
          <Route path="posts/:postID" element={<PostDetailPage />} />
          <Route path="following" element={<FollowingPage />} />
          <Route path="friends" element={<FriendsPage />} />
          <Route path="groups" element={<GroupsPage />} />
          <Route path="groups/:groupID" element={<GroupDetailPage />} />
          <Route path="messages" element={<MessagesPage />} />
          <Route path="messages/:conversationID" element={<MessagesPage />} />
          <Route path="me" element={<UserProfilePage />} />
          <Route path="users/:userID" element={<UserProfilePage />} />
          <Route path="compose" element={<Navigate to="/" replace />} />
          <Route path="login" element={<Navigate to="/" replace />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  )
}
