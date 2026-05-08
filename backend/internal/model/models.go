package model

import "time"

type User struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	Email         string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"email"`
	PasswordHash  string    `gorm:"type:varchar(255);not null" json:"-"`
	Avatar        string    `gorm:"type:varchar(512);default:''" json:"avatar"`
	XP            uint      `gorm:"default:0" json:"xp"`
	Level         uint      `gorm:"default:1" json:"level"`
	CreditScore   int       `gorm:"default:100" json:"credit_score"`
	StreakFreezes uint      `gorm:"default:1" json:"streak_freezes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Challenge struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatorID       uint64    `gorm:"index;not null" json:"creator_id"`
	Creator         User      `gorm:"foreignKey:CreatorID" json:"creator"`
	Title           string    `gorm:"type:varchar(128);not null" json:"title"`
	Description     string    `gorm:"type:text" json:"description"`
	Category        string    `gorm:"type:enum('fitness','study','habit','health','social','other');default:'other'" json:"category"`
	Pledge          string    `gorm:"type:text;not null" json:"pledge"`
	PenaltyType     string    `gorm:"type:enum('donate','social','custom');not null" json:"penalty_type"`
	PenaltyDetail   string    `gorm:"type:varchar(512)" json:"penalty_detail"`
	ChallengeType   string    `gorm:"type:enum('daily','total');not null" json:"challenge_type"`
	TargetDays      uint      `gorm:"not null" json:"target_days"`
	StartDate       time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate         time.Time `gorm:"type:date;not null" json:"end_date"`
	Status          string    `gorm:"type:enum('pending','active','completed','failed','cancelled');default:'pending';index" json:"status"`
	IsPublic        bool      `gorm:"default:false;index" json:"is_public"`
	XPReward        uint      `gorm:"default:200" json:"xp_reward"`
	CanCancelBefore time.Time `json:"can_cancel_before"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ChallengeParticipant struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ChallengeID   uint64    `gorm:"index;uniqueIndex:uk_challenge_user_role;not null" json:"challenge_id"`
	Challenge     Challenge `gorm:"foreignKey:ChallengeID" json:"challenge"`
	UserID        uint64    `gorm:"index;uniqueIndex:uk_challenge_user_role;not null" json:"user_id"`
	User          User      `gorm:"foreignKey:UserID" json:"user"`
	Role          string    `gorm:"type:enum('creator','participant','witness');uniqueIndex:uk_challenge_user_role;not null" json:"role"`
	Status        string    `gorm:"type:enum('pending','accepted','rejected');default:'pending'" json:"status"`
	CurrentStreak uint      `gorm:"default:0" json:"current_streak"`
	MaxStreak     uint      `gorm:"default:0" json:"max_streak"`
	TotalCheckins uint      `gorm:"default:0" json:"total_checkins"`
	JoinedAt      time.Time `gorm:"not null" json:"joined_at"`
}

type CheckIn struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ChallengeID uint64    `gorm:"index;uniqueIndex:uk_checkin_daily;not null" json:"challenge_id"`
	UserID      uint64    `gorm:"index;uniqueIndex:uk_checkin_daily;not null" json:"user_id"`
	Content     string    `gorm:"type:text" json:"content"`
	ImageURL    string    `gorm:"type:varchar(512)" json:"image_url"`
	CheckinDate time.Time `gorm:"type:date;index;uniqueIndex:uk_checkin_daily;not null" json:"checkin_date"`
	StreakCount uint      `gorm:"default:1" json:"streak_count"`
	XPEarned    uint      `gorm:"default:10" json:"xp_earned"`
	LikeCount   uint      `gorm:"default:0" json:"like_count"`
	IsReported  bool      `gorm:"default:false" json:"is_reported"`
	CreatedAt   time.Time `json:"created_at"`
	User        User      `gorm:"foreignKey:UserID" json:"user"`
}

type CheckInInteraction struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CheckinID uint64    `gorm:"index;not null" json:"checkin_id"`
	UserID    uint64    `gorm:"index;not null" json:"user_id"`
	Type      string    `gorm:"type:enum('like','comment');not null" json:"type"`
	Content   string    `gorm:"type:varchar(512)" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
}

type StreakFreezeLog struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"index;uniqueIndex:uk_freeze_daily;not null" json:"user_id"`
	ChallengeID uint64    `gorm:"index;uniqueIndex:uk_freeze_daily;not null" json:"challenge_id"`
	FreezeDate  time.Time `gorm:"type:date;index;uniqueIndex:uk_freeze_daily;not null" json:"freeze_date"`
	CreatedAt   time.Time `json:"created_at"`
}

type Achievement struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Description    string `gorm:"type:varchar(256);not null" json:"description"`
	Icon           string `gorm:"type:varchar(256)" json:"icon"`
	Category       string `gorm:"type:enum('checkin','challenge','social','hidden');not null" json:"category"`
	ConditionType  string `gorm:"type:varchar(64);not null" json:"condition_type"`
	ConditionValue int    `gorm:"not null" json:"condition_value"`
	IsHidden       bool   `gorm:"default:false" json:"is_hidden"`
	XPReward       uint   `gorm:"default:50" json:"xp_reward"`
}

type UserAchievement struct {
	ID            uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64      `gorm:"index;uniqueIndex:uk_user_achievement;not null" json:"user_id"`
	AchievementID uint64      `gorm:"index;uniqueIndex:uk_user_achievement;not null" json:"achievement_id"`
	Achievement   Achievement `gorm:"foreignKey:AchievementID" json:"achievement"`
	UnlockedAt    time.Time   `gorm:"not null" json:"unlocked_at"`
}

type Friendship struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"index;uniqueIndex:uk_friendship;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	FriendID  uint64    `gorm:"index;uniqueIndex:uk_friendship;not null" json:"friend_id"`
	Status    string    `gorm:"type:enum('pending','accepted','blocked');default:'pending'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Friend    User      `gorm:"foreignKey:FriendID" json:"friend"`
}

type Certificate struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"index;not null" json:"user_id"`
	ChallengeID   uint64    `gorm:"index;not null" json:"challenge_id"`
	CertificateNo string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"certificate_no"`
	ImageURL      string    `gorm:"type:varchar(512)" json:"image_url"`
	IssuedAt      time.Time `gorm:"not null" json:"issued_at"`
	Challenge     Challenge `gorm:"foreignKey:ChallengeID" json:"challenge"`
}

type Notification struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"index:idx_notification_user_read;not null" json:"user_id"`
	Type        string    `gorm:"type:enum('invite','checkin_remind','streak_warn','achievement','witness','friend_request','like','comment','system');not null" json:"type"`
	Title       string    `gorm:"type:varchar(128);not null" json:"title"`
	Content     string    `gorm:"type:varchar(512);not null" json:"content"`
	RelatedID   uint64    `json:"related_id"`
	RelatedType string    `gorm:"type:varchar(32)" json:"related_type"`
	IsRead      bool      `gorm:"index:idx_notification_user_read;default:false" json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

type Report struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ReporterID uint64    `gorm:"index;not null" json:"reporter_id"`
	TargetType string    `gorm:"type:enum('checkin','challenge','user');not null" json:"target_type"`
	TargetID   uint64    `gorm:"index;not null" json:"target_id"`
	Reason     string    `gorm:"type:varchar(512);not null" json:"reason"`
	Status     string    `gorm:"type:enum('pending','resolved','dismissed');default:'pending'" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

func AllModels() []any {
	return []any{
		&User{},
		&Challenge{},
		&ChallengeParticipant{},
		&CheckIn{},
		&CheckInInteraction{},
		&StreakFreezeLog{},
		&Achievement{},
		&UserAchievement{},
		&Friendship{},
		&Certificate{},
		&Notification{},
		&Report{},
	}
}
