package dto

import "time"

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=32"`
	Avatar   string `json:"avatar" binding:"omitempty,max=512"`
}

type UserSummary struct {
	ID            uint64 `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email,omitempty"`
	Avatar        string `json:"avatar"`
	XP            uint   `json:"xp"`
	Level         uint   `json:"level"`
	CreditScore   int    `json:"credit_score"`
	StreakFreezes uint   `json:"streak_freezes"`
}

type AuthResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         UserSummary `json:"user"`
}

type CreateChallengeRequest struct {
	Title         string   `json:"title" binding:"required,min=2,max=128"`
	Description   string   `json:"description" binding:"max=2000"`
	Category      string   `json:"category" binding:"required,oneof=fitness study habit health social other"`
	Pledge        string   `json:"pledge" binding:"required,min=5"`
	PenaltyType   string   `json:"penalty_type" binding:"required,oneof=donate social custom"`
	PenaltyDetail string   `json:"penalty_detail" binding:"max=512"`
	ChallengeType string   `json:"challenge_type" binding:"required,oneof=daily total"`
	TargetDays    uint     `json:"target_days" binding:"required,min=1,max=365"`
	StartDate     string   `json:"start_date" binding:"required"`
	IsPublic      bool     `json:"is_public"`
	WitnessIDs    []uint64 `json:"witness_ids"`
}

type ChallengeResponse struct {
	ID               uint64  `json:"id"`
	CreatorID        uint64  `json:"creator_id"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	Category         string  `json:"category"`
	Pledge           string  `json:"pledge"`
	PenaltyType      string  `json:"penalty_type"`
	PenaltyDetail    string  `json:"penalty_detail"`
	ChallengeType    string  `json:"challenge_type"`
	TargetDays       uint    `json:"target_days"`
	StartDate        string  `json:"start_date"`
	EndDate          string  `json:"end_date"`
	Status           string  `json:"status"`
	IsPublic         bool    `json:"is_public"`
	XPReward         uint    `json:"xp_reward"`
	CanCancelBefore  string  `json:"can_cancel_before"`
	Progress         float64 `json:"progress"`
	ParticipantCount int64   `json:"participant_count"`
	CheckinCount     int64   `json:"checkin_count"`
}

type CheckInResponse struct {
	ID          uint64      `json:"id"`
	ChallengeID uint64      `json:"challenge_id"`
	UserID      uint64      `json:"user_id"`
	Content     string      `json:"content"`
	ImageURL    string      `json:"image_url"`
	CheckinDate string      `json:"checkin_date"`
	StreakCount uint        `json:"streak_count"`
	XPEarned    uint        `json:"xp_earned"`
	LikeCount   uint        `json:"like_count"`
	CreatedAt   time.Time   `json:"created_at"`
	User        UserSummary `json:"user"`
}

type SubmitCheckInResponse struct {
	CheckinID            uint64                `json:"checkin_id"`
	StreakCount          uint                  `json:"streak_count"`
	XPEarned             uint                  `json:"xp_earned"`
	XPMultiplier         float64               `json:"xp_multiplier"`
	TotalXP              uint                  `json:"total_xp"`
	AchievementsUnlocked []AchievementResponse `json:"achievements_unlocked"`
}

type AchievementResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Category    string `json:"category"`
	IsHidden    bool   `json:"is_hidden"`
	Unlocked    bool   `json:"unlocked"`
	UnlockedAt  string `json:"unlocked_at,omitempty"`
}

type CertificateResponse struct {
	ID            uint64            `json:"id"`
	UserID        uint64            `json:"user_id"`
	ChallengeID   uint64            `json:"challenge_id"`
	CertificateNo string            `json:"certificate_no"`
	ImageURL      string            `json:"image_url"`
	IssuedAt      string            `json:"issued_at"`
	Challenge     ChallengeResponse `json:"challenge"`
}

type CommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=512"`
}

type FriendRequest struct {
	FriendID uint64 `json:"friend_id" binding:"required"`
}

type AppealRequest struct {
	Reason string `json:"reason" binding:"required,min=5,max=512"`
}

type ReportRequest struct {
	Reason string `json:"reason" binding:"required,min=5,max=512"`
}
