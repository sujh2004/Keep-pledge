import client from './client'
import type {
  AuthResponse,
  UserSummary,
  UserStats,
  HeatmapEntry,
  ChallengeResponse,
  ChallengeDetail,
  CheckInResponse,
  SubmitCheckInResponse,
  AchievementResponse,
  Certificate,
  Notification,
  Friendship,
  Interaction,
} from '../types'

function d<T>(res: { data: { data: T } }): T {
  return res.data.data
}

export const authApi = {
  register: (data: { username: string; email: string; password: string }) =>
    client.post('/auth/register', data).then(d<AuthResponse>),
  login: (data: { email: string; password: string }) =>
    client.post('/auth/login', data).then(d<AuthResponse>),
  refresh: () => client.post('/auth/refresh').then(d<AuthResponse>),
  me: () => client.get('/auth/me').then(d<UserSummary>),
}

export const userApi = {
  get: (id: number) => client.get(`/users/${id}`).then(d<UserSummary>),
  updateProfile: (data: { username?: string; avatar?: string }) =>
    client.put('/users/profile', data).then(d<UserSummary>),
  uploadAvatar: (file: File) => {
    const fd = new FormData()
    fd.append('avatar', file)
    return client.put('/users/avatar', fd).then(d<UserSummary>)
  },
  stats: (id: number) => client.get(`/users/${id}/stats`).then(d<UserStats>),
  challenges: (id: number, category?: string) =>
    client.get(`/users/${id}/challenges`, { params: { category } }).then(d<ChallengeResponse[]>),
  achievements: (id: number) => client.get(`/users/${id}/achievements`).then(d<AchievementResponse[]>),
  heatmap: (id: number) => client.get(`/users/${id}/heatmap`).then(d<HeatmapEntry[]>),
}

export const challengeApi = {
  create: (data: {
    title: string
    description: string
    category: string
    pledge: string
    penalty_type: string
    penalty_detail: string
    challenge_type: string
    target_days: number
    start_date: string
    is_public: boolean
    witness_ids?: number[]
  }) => client.post('/challenges', data).then(d<ChallengeResponse>),
  list: (category?: string) =>
    client.get('/challenges', { params: { category } }).then(d<ChallengeResponse[]>),
  get: (id: number) => client.get(`/challenges/${id}`).then(d<ChallengeDetail>),
  cancel: (id: number) => client.delete(`/challenges/${id}`),
  join: (id: number) => client.post(`/challenges/${id}/join`),
  participants: (id: number) => client.get(`/challenges/${id}/participants`).then(d<unknown[]>),
  progress: (id: number) => client.get(`/challenges/${id}/progress`).then(d<{ target_days: number; checkins: number; progress: number; status: string }>),
  explore: (category?: string) =>
    client.get('/challenges/explore', { params: { category } }).then(d<ChallengeResponse[]>),
}

export const checkinApi = {
  submit: (challengeId: number, content: string, image?: File) => {
    const fd = new FormData()
    fd.append('content', content)
    if (image) fd.append('image', image)
    return client.post(`/challenges/${challengeId}/checkin`, fd).then(d<SubmitCheckInResponse>)
  },
  list: (challengeId: number) =>
    client.get(`/challenges/${challengeId}/checkins`).then(d<CheckInResponse[]>),
  today: (challengeId: number) =>
    client.get(`/challenges/${challengeId}/checkins/today`).then(d<{ checked_in: boolean; used_freeze: boolean; checkin?: CheckInResponse }>),
  streak: (challengeId: number) =>
    client.get(`/challenges/${challengeId}/streak`).then(d<{ current_streak: number; max_streak: number; total_checkins: number }>),
  useFreeze: (challengeId: number) =>
    client.post(`/challenges/${challengeId}/streak-freeze`).then(d<{ remaining_freezes: number; freeze_date: string; streak_preserved: number }>),
  like: (checkinId: number) => client.post(`/checkins/${checkinId}/like`),
  unlike: (checkinId: number) => client.delete(`/checkins/${checkinId}/like`),
  comment: (checkinId: number, content: string) =>
    client.post(`/checkins/${checkinId}/comment`, { content }),
  interactions: (checkinId: number) =>
    client.get(`/checkins/${checkinId}/interactions`).then(d<Interaction[]>),
  report: (checkinId: number, reason: string) =>
    client.post(`/checkins/${checkinId}/report`, { reason }),
}

export const socialApi = {
  friends: () => client.get('/friends').then(d<Friendship[]>),
  request: (friendId: number) => client.post('/friends/request', { friend_id: friendId }),
  accept: (id: number) => client.put(`/friends/${id}/accept`),
  delete: (id: number) => client.delete(`/friends/${id}`),
}

export const leaderboardApi = {
  weekly: () => client.get('/leaderboard/weekly').then(d<UserSummary[]>),
  friends: () => client.get('/leaderboard/friends').then(d<UserSummary[]>),
}

export const achievementApi = {
  all: () => client.get('/achievements').then(d<AchievementResponse[]>),
  mine: () => client.get('/achievements/my').then(d<AchievementResponse[]>),
}

export const certificateApi = {
  list: () => client.get('/certificates').then(d<Certificate[]>),
  get: (id: number) => client.get(`/certificates/${id}`).then(d<Certificate>),
}

export const notificationApi = {
  list: (unread?: boolean) =>
    client.get('/notifications', { params: { unread: unread || undefined } }).then(d<Notification[]>),
  unreadCount: () => client.get('/notifications/unread-count').then(d<{ count: number }>),
  markRead: (id: number) => client.put(`/notifications/${id}/read`),
  markAllRead: () => client.put('/notifications/read-all'),
}
