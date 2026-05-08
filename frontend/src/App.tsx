import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './store/auth'
import AppLayout from './components/Layout'
import LoginPage from './pages/Login'
import RegisterPage from './pages/Login/Register'
import HomePage from './pages/Home'
import ChallengePage from './pages/Challenge'
import CreateChallengePage from './pages/CreateChallenge'
import CheckInPage from './pages/CheckIn'
import ProfilePage from './pages/Profile'
import LeaderboardPage from './pages/Leaderboard'
import AchievementsPage from './pages/Achievements'
import CertificatePage from './pages/Certificate'
import FriendsPage from './pages/Friends'
import ExplorePage from './pages/Explore'
import NotificationsPage from './pages/Notifications'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token)
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route
        path="/*"
        element={
          <PrivateRoute>
            <AppLayout>
              <Routes>
                <Route path="/" element={<HomePage />} />
                <Route path="/challenge/create" element={<CreateChallengePage />} />
                <Route path="/challenge/:id" element={<ChallengePage />} />
                <Route path="/challenge/:id/checkin" element={<CheckInPage />} />
                <Route path="/profile" element={<ProfilePage />} />
                <Route path="/profile/:id" element={<ProfilePage />} />
                <Route path="/leaderboard" element={<LeaderboardPage />} />
                <Route path="/achievements" element={<AchievementsPage />} />
                <Route path="/certificate/:id" element={<CertificatePage />} />
                <Route path="/friends" element={<FriendsPage />} />
                <Route path="/explore" element={<ExplorePage />} />
                <Route path="/notifications" element={<NotificationsPage />} />
              </Routes>
            </AppLayout>
          </PrivateRoute>
        }
      />
    </Routes>
  )
}
