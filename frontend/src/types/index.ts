export interface UserSummary {
  id: number
  username: string
  email?: string
  avatar: string
  xp: number
  level: number
  credit_score: number
  streak_freezes: number
}

export interface AuthResponse {
  token: string
  refresh_token: string
  user: UserSummary
}

export interface ChallengeResponse {
  id: number
  creator_id: number
  title: string
  description: string
  category: string
  pledge: string
  penalty_type: string
  penalty_detail: string
  challenge_type: string
  target_days: number
  start_date: string
  end_date: string
  status: string
  is_public: boolean
  xp_reward: number
  can_cancel_before: string
  progress: number
  participant_count: number
  checkin_count: number
}

export interface CheckInResponse {
  id: number
  challenge_id: number
  user_id: number
  content: string
  image_url: string
  checkin_date: string
  streak_count: number
  xp_earned: number
  like_count: number
  created_at: string
  user: UserSummary
}

export interface SubmitCheckInResponse {
  checkin_id: number
  streak_count: number
  xp_earned: number
  xp_multiplier: number
  total_xp: number
  achievements_unlocked: AchievementResponse[]
}

export interface AchievementResponse {
  id: number
  name: string
  description: string
  icon: string
  category: string
  is_hidden: boolean
  unlocked: boolean
  unlocked_at?: string
}

export interface Notification {
  id: number
  user_id: number
  type: string
  title: string
  content: string
  related_id: number
  related_type: string
  is_read: boolean
  created_at: string
}

export interface Certificate {
  id: number
  user_id: number
  challenge_id: number
  certificate_no: string
  image_url: string
  issued_at: string
  challenge: ChallengeResponse
}

export interface Friendship {
  id: number
  status: string
  direction: 'incoming' | 'outgoing'
  created_at: string
  friend: UserSummary
}

export interface Participant {
  id: number
  role: string
  status: string
  current_streak: number
  max_streak: number
  total_checkins: number
  joined_at: string
  user: UserSummary
}

export interface ChallengeDetail {
  challenge: ChallengeResponse
  creator: UserSummary
  participants: Participant[]
  checkins: CheckInResponse[]
}

export interface HeatmapEntry {
  date: string
  count: number
}

export interface UserStats {
  user: UserSummary
  active_challenges: number
  total_checkins: number
  completed: number
}

export interface Interaction {
  id: number
  type: 'like' | 'comment'
  content: string
  created_at: string
  user: UserSummary
}

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export type ChallengeCategory = 'fitness' | 'study' | 'habit' | 'health' | 'social' | 'other'

export const CATEGORY_LABELS: Record<ChallengeCategory, string> = {
  fitness: '健身',
  study: '学习',
  habit: '习惯',
  health: '健康',
  social: '社交',
  other: '其他',
}
