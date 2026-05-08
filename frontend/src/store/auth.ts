import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { UserSummary } from '../types'

interface AuthState {
  token: string | null
  refreshToken: string | null
  user: UserSummary | null
  setAuth: (token: string, refreshToken: string, user: UserSummary) => void
  setUser: (user: UserSummary) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      refreshToken: null,
      user: null,
      setAuth: (token, refreshToken, user) => set({ token, refreshToken, user }),
      setUser: (user) => set({ user }),
      logout: () => set({ token: null, refreshToken: null, user: null }),
    }),
    { name: 'keep-pledge-auth' },
  ),
)
