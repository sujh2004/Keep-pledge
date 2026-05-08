package service

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"keep-pledge/backend/internal/config"
	"keep-pledge/backend/internal/dto"
	"keep-pledge/backend/internal/model"
	"keep-pledge/backend/internal/repository"
	authpkg "keep-pledge/backend/pkg/auth"
	"keep-pledge/backend/pkg/response"
	"keep-pledge/backend/pkg/validator"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	fontdraw "golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"gorm.io/gorm"
)

type AppError struct {
	HTTPStatus int
	Code       int
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

func newAppError(status int, code int, message string) *AppError {
	return &AppError{HTTPStatus: status, Code: code, Message: message}
}

type Services struct {
	Auth         *AuthService
	Challenge    *ChallengeService
	CheckIn      *CheckInService
	Notification *NotificationService
	Social       *SocialService
	Achievement  *AchievementService
	Leaderboard  *LeaderboardService
	Certificate  *CertificateService
	Task         *TaskService
}

func NewServices(repo *repository.Repository, cfg *config.Config, jwt *authpkg.Manager) *Services {
	return &Services{
		Auth:         &AuthService{repo: repo, cfg: cfg, jwt: jwt},
		Challenge:    &ChallengeService{repo: repo, cfg: cfg},
		CheckIn:      &CheckInService{repo: repo},
		Notification: &NotificationService{repo: repo},
		Social:       &SocialService{repo: repo},
		Achievement:  &AchievementService{repo: repo},
		Leaderboard:  &LeaderboardService{repo: repo},
		Certificate:  &CertificateService{repo: repo},
		Task:         &TaskService{repo: repo},
	}
}

type AuthService struct {
	repo *repository.Repository
	cfg  *config.Config
	jwt  *authpkg.Manager
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var count int64
	if err := s.repo.DB.Model(&model.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, newAppError(http.StatusConflict, response.CodeUserExists, "用户名或邮箱已存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username:      username,
		Email:         email,
		PasswordHash:  string(hash),
		CreditScore:   100,
		Level:         1,
		StreakFreezes: 1,
	}
	if err := s.repo.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return s.issueTokens(user)
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	var user model.User
	if err := s.repo.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, newAppError(http.StatusNotFound, response.CodeUserNotFound, "用户不存在")
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, newAppError(http.StatusUnauthorized, response.CodePasswordWrong, "密码错误")
	}
	return s.issueTokens(user)
}

func (s *AuthService) Refresh(userID uint64) (*dto.AuthResponse, error) {
	var user model.User
	if err := s.repo.DB.First(&user, userID).Error; err != nil {
		return nil, notFound(response.CodeUserNotFound, "用户不存在")
	}
	return s.issueTokens(user)
}

func (s *AuthService) Me(userID uint64) (*dto.UserSummary, error) {
	var user model.User
	if err := s.repo.DB.First(&user, userID).Error; err != nil {
		return nil, notFound(response.CodeUserNotFound, "用户不存在")
	}
	summary := userSummary(user, true)
	return &summary, nil
}

func (s *AuthService) UpdateProfile(userID uint64, req dto.UpdateProfileRequest) (*dto.UserSummary, error) {
	var user model.User
	if err := s.repo.DB.First(&user, userID).Error; err != nil {
		return nil, notFound(response.CodeUserNotFound, "用户不存在")
	}
	if strings.TrimSpace(req.Username) != "" {
		user.Username = strings.TrimSpace(req.Username)
	}
	if strings.TrimSpace(req.Avatar) != "" {
		user.Avatar = strings.TrimSpace(req.Avatar)
	}
	if err := s.repo.DB.Save(&user).Error; err != nil {
		return nil, err
	}
	summary := userSummary(user, true)
	return &summary, nil
}

func (s *AuthService) PublicUser(userID uint64) (*dto.UserSummary, error) {
	var user model.User
	if err := s.repo.DB.First(&user, userID).Error; err != nil {
		return nil, notFound(response.CodeUserNotFound, "用户不存在")
	}
	summary := userSummary(user, false)
	return &summary, nil
}

func (s *AuthService) Stats(userID uint64) (map[string]any, error) {
	var user model.User
	if err := s.repo.DB.First(&user, userID).Error; err != nil {
		return nil, notFound(response.CodeUserNotFound, "用户不存在")
	}
	var activeChallenges int64
	var checkins int64
	var completed int64
	s.repo.DB.Model(&model.ChallengeParticipant{}).Where("user_id = ? AND role <> ? AND status = ?", userID, "witness", "accepted").Count(&activeChallenges)
	s.repo.DB.Model(&model.CheckIn{}).Where("user_id = ?", userID).Count(&checkins)
	s.repo.DB.Model(&model.Challenge{}).Where("creator_id = ? AND status = ?", userID, "completed").Count(&completed)
	return map[string]any{
		"user":              userSummary(user, true),
		"active_challenges": activeChallenges,
		"total_checkins":    checkins,
		"completed":         completed,
	}, nil
}

func (s *AuthService) Heatmap(userID uint64) ([]map[string]any, error) {
	var rows []struct {
		Date  time.Time
		Count int
	}
	start := beginningOfDay(time.Now()).AddDate(0, -3, 0)
	if err := s.repo.DB.Model(&model.CheckIn{}).
		Select("checkin_date as date, count(*) as count").
		Where("user_id = ? AND checkin_date >= ?", userID, start).
		Group("checkin_date").
		Order("checkin_date asc").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		result = append(result, map[string]any{
			"date":  row.Date.Format("2006-01-02"),
			"count": row.Count,
		})
	}
	return result, nil
}

func (s *AuthService) issueTokens(user model.User) (*dto.AuthResponse, error) {
	token, err := s.jwt.Generate(user.ID, s.cfg.JWT.AccessTTL())
	if err != nil {
		return nil, err
	}
	refresh, err := s.jwt.Generate(user.ID, s.cfg.JWT.RefreshTTL())
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{
		Token:        token,
		RefreshToken: refresh,
		User:         userSummary(user, true),
	}, nil
}

type ChallengeService struct {
	repo *repository.Repository
	cfg  *config.Config
}

func (s *ChallengeService) Create(userID uint64, req dto.CreateChallengeRequest) (*dto.ChallengeResponse, error) {
	var user model.User
	if err := s.repo.DB.First(&user, userID).Error; err != nil {
		return nil, notFound(response.CodeUserNotFound, "用户不存在")
	}
	if user.CreditScore < 60 {
		var active int64
		s.repo.DB.Model(&model.Challenge{}).
			Where("creator_id = ? AND status IN ?", userID, []string{"pending", "active"}).
			Count(&active)
		if active >= 1 {
			return nil, newAppError(http.StatusForbidden, response.CodeCreditNotEnough, "信用分不足，最多只能同时创建 1 个挑战")
		}
	}
	if len(req.WitnessIDs) > 3 {
		return nil, newAppError(http.StatusBadRequest, response.CodeValidationFailed, "见证人最多 3 名")
	}
	start, err := time.ParseInLocation("2006-01-02", req.StartDate, time.Local)
	if err != nil {
		return nil, newAppError(http.StatusBadRequest, response.CodeValidationFailed, "start_date 格式必须为 YYYY-MM-DD")
	}
	end := start.AddDate(0, 0, int(req.TargetDays)-1)
	now := time.Now()

	challenge := model.Challenge{
		CreatorID:       userID,
		Title:           strings.TrimSpace(req.Title),
		Description:     strings.TrimSpace(req.Description),
		Category:        req.Category,
		Pledge:          strings.TrimSpace(req.Pledge),
		PenaltyType:     req.PenaltyType,
		PenaltyDetail:   strings.TrimSpace(req.PenaltyDetail),
		ChallengeType:   req.ChallengeType,
		TargetDays:      req.TargetDays,
		StartDate:       start,
		EndDate:         end,
		Status:          "pending",
		IsPublic:        req.IsPublic,
		XPReward:        200,
		CanCancelBefore: now.Add(24 * time.Hour),
	}

	err = s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&challenge).Error; err != nil {
			return err
		}
		creator := model.ChallengeParticipant{
			ChallengeID: challenge.ID,
			UserID:      userID,
			Role:        "creator",
			Status:      "accepted",
			JoinedAt:    now,
		}
		if err := tx.Create(&creator).Error; err != nil {
			return err
		}
		for _, witnessID := range req.WitnessIDs {
			if witnessID == userID {
				continue
			}
			participant := model.ChallengeParticipant{
				ChallengeID: challenge.ID,
				UserID:      witnessID,
				Role:        "witness",
				Status:      "pending",
				JoinedAt:    now,
			}
			if err := tx.Create(&participant).Error; err != nil {
				return err
			}
			notice := model.Notification{
				UserID:      witnessID,
				Type:        "invite",
				Title:       "新的见证邀请",
				Content:     fmt.Sprintf("%s 邀请你见证挑战「%s」", user.Username, challenge.Title),
				RelatedID:   challenge.ID,
				RelatedType: "challenge",
			}
			if err := tx.Create(&notice).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := mapChallenge(challenge, 0, 1)
	return &result, nil
}

func (s *ChallengeService) List(userID uint64, category string) ([]dto.ChallengeResponse, error) {
	var challenges []model.Challenge
	query := s.repo.DB.Model(&model.Challenge{}).
		Joins("JOIN challenge_participants ON challenge_participants.challenge_id = challenges.id").
		Where("challenge_participants.user_id = ?", userID).
		Order("challenges.created_at desc")
	if category != "" {
		query = query.Where("challenges.category = ?", category)
	}
	if err := query.Find(&challenges).Error; err != nil {
		return nil, err
	}
	return s.mapChallengeList(challenges), nil
}

func (s *ChallengeService) Explore(category string) ([]dto.ChallengeResponse, error) {
	var challenges []model.Challenge
	query := s.repo.DB.Where("is_public = ? AND status <> ?", true, "cancelled").Order("created_at desc")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if err := query.Find(&challenges).Error; err != nil {
		return nil, err
	}
	return s.mapChallengeList(challenges), nil
}

func (s *ChallengeService) Get(userID uint64, challengeID uint64) (map[string]any, error) {
	var challenge model.Challenge
	if err := s.repo.DB.Preload("Creator").First(&challenge, challengeID).Error; err != nil {
		return nil, notFound(response.CodeChallengeNotFound, "挑战不存在")
	}
	if !challenge.IsPublic && !s.isParticipant(userID, challengeID) {
		return nil, newAppError(http.StatusForbidden, response.CodeForbidden, "无权查看该挑战")
	}
	participants, err := s.Participants(userID, challengeID)
	if err != nil {
		return nil, err
	}
	checkins, err := (&CheckInService{repo: s.repo}).List(challengeID)
	if err != nil {
		return nil, err
	}
	resp := mapChallenge(challenge, int64(len(checkins)), int64(len(participants)))
	return map[string]any{
		"challenge":    resp,
		"creator":      userSummary(challenge.Creator, false),
		"participants": participants,
		"checkins":     checkins,
	}, nil
}

func (s *ChallengeService) Cancel(userID uint64, challengeID uint64) error {
	var challenge model.Challenge
	if err := s.repo.DB.First(&challenge, challengeID).Error; err != nil {
		return notFound(response.CodeChallengeNotFound, "挑战不存在")
	}
	if challenge.CreatorID != userID {
		return newAppError(http.StatusForbidden, response.CodeForbidden, "只有创建者可以撤回挑战")
	}
	if time.Now().After(challenge.CanCancelBefore) {
		return newAppError(http.StatusBadRequest, response.CodeCoolingPeriodExpired, "冷静期已过，无法撤回誓约")
	}
	if challenge.Status == "cancelled" {
		return nil
	}
	return s.repo.DB.Model(&challenge).Update("status", "cancelled").Error
}

func (s *ChallengeService) Join(userID uint64, challengeID uint64) error {
	var challenge model.Challenge
	if err := s.repo.DB.First(&challenge, challengeID).Error; err != nil {
		return notFound(response.CodeChallengeNotFound, "挑战不存在")
	}
	if !challenge.IsPublic {
		return newAppError(http.StatusForbidden, response.CodeForbidden, "该挑战不是公开挑战")
	}
	var count int64
	s.repo.DB.Model(&model.ChallengeParticipant{}).
		Where("challenge_id = ? AND user_id = ?", challengeID, userID).
		Count(&count)
	if count > 0 {
		return nil
	}
	participant := model.ChallengeParticipant{
		ChallengeID: challengeID,
		UserID:      userID,
		Role:        "participant",
		Status:      "accepted",
		JoinedAt:    time.Now(),
	}
	return s.repo.DB.Create(&participant).Error
}

func (s *ChallengeService) Participants(userID uint64, challengeID uint64) ([]map[string]any, error) {
	var challenge model.Challenge
	if err := s.repo.DB.First(&challenge, challengeID).Error; err != nil {
		return nil, notFound(response.CodeChallengeNotFound, "挑战不存在")
	}
	if !challenge.IsPublic && !s.isParticipant(userID, challengeID) {
		return nil, newAppError(http.StatusForbidden, response.CodeForbidden, "无权查看参与者")
	}
	var participants []model.ChallengeParticipant
	if err := s.repo.DB.Preload("User").Where("challenge_id = ?", challengeID).Order("role asc, joined_at asc").Find(&participants).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(participants))
	for _, p := range participants {
		result = append(result, map[string]any{
			"id":             p.ID,
			"role":           p.Role,
			"status":         p.Status,
			"current_streak": p.CurrentStreak,
			"max_streak":     p.MaxStreak,
			"total_checkins": p.TotalCheckins,
			"joined_at":      p.JoinedAt,
			"user":           userSummary(p.User, false),
		})
	}
	return result, nil
}

func (s *ChallengeService) Progress(userID uint64, challengeID uint64) (map[string]any, error) {
	var challenge model.Challenge
	if err := s.repo.DB.First(&challenge, challengeID).Error; err != nil {
		return nil, notFound(response.CodeChallengeNotFound, "挑战不存在")
	}
	if !challenge.IsPublic && !s.isParticipant(userID, challengeID) {
		return nil, newAppError(http.StatusForbidden, response.CodeForbidden, "无权查看进度")
	}
	var checkins int64
	s.repo.DB.Model(&model.CheckIn{}).Where("challenge_id = ?", challengeID).Count(&checkins)
	return map[string]any{
		"target_days": challenge.TargetDays,
		"checkins":    checkins,
		"progress":    progressPercent(checkins, challenge.TargetDays),
		"status":      challenge.Status,
	}, nil
}

func (s *ChallengeService) UpdateStatus(userID uint64, challengeID uint64, status string) error {
	if status != "pending" && status != "active" && status != "completed" && status != "failed" && status != "cancelled" {
		return newAppError(http.StatusBadRequest, response.CodeValidationFailed, "状态值不合法")
	}
	var challenge model.Challenge
	if err := s.repo.DB.First(&challenge, challengeID).Error; err != nil {
		return notFound(response.CodeChallengeNotFound, "挑战不存在")
	}
	if challenge.CreatorID != userID {
		return newAppError(http.StatusForbidden, response.CodeForbidden, "只有创建者可以更新状态")
	}
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&challenge).Update("status", status).Error; err != nil {
			return err
		}
		if status == "completed" {
			return applyChallengeCompletion(tx, s.cfg, challenge)
		}
		if status == "failed" {
			return tx.Model(&model.User{}).Where("id = ?", challenge.CreatorID).
				Update("credit_score", gorm.Expr("GREATEST(credit_score - ?, 0)", 15)).Error
		}
		return nil
	})
}

func (s *ChallengeService) Appeal(userID uint64, challengeID uint64, reason string) error {
	var challenge model.Challenge
	if err := s.repo.DB.First(&challenge, challengeID).Error; err != nil {
		return notFound(response.CodeChallengeNotFound, "挑战不存在")
	}
	if challenge.CreatorID != userID {
		return newAppError(http.StatusForbidden, response.CodeForbidden, "只有创建者可以提交申诉")
	}
	var witnesses []model.ChallengeParticipant
	if err := s.repo.DB.Where("challenge_id = ? AND role = ?", challengeID, "witness").Find(&witnesses).Error; err != nil {
		return err
	}
	if len(witnesses) == 0 {
		return newAppError(http.StatusBadRequest, response.CodeValidationFailed, "无见证人的挑战不支持申诉")
	}
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		for _, w := range witnesses {
			notice := model.Notification{
				UserID:      w.UserID,
				Type:        "witness",
				Title:       "不可抗力申诉待确认",
				Content:     fmt.Sprintf("挑战「%s」提交了申诉：%s", challenge.Title, reason),
				RelatedID:   challenge.ID,
				RelatedType: "challenge",
			}
			if err := tx.Create(&notice).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ChallengeService) mapChallengeList(challenges []model.Challenge) []dto.ChallengeResponse {
	result := make([]dto.ChallengeResponse, 0, len(challenges))
	for _, challenge := range challenges {
		var checkins int64
		var participants int64
		s.repo.DB.Model(&model.CheckIn{}).Where("challenge_id = ?", challenge.ID).Count(&checkins)
		s.repo.DB.Model(&model.ChallengeParticipant{}).Where("challenge_id = ?", challenge.ID).Count(&participants)
		result = append(result, mapChallenge(challenge, checkins, participants))
	}
	return result
}

func (s *ChallengeService) isParticipant(userID uint64, challengeID uint64) bool {
	var count int64
	s.repo.DB.Model(&model.ChallengeParticipant{}).
		Where("challenge_id = ? AND user_id = ?", challengeID, userID).
		Count(&count)
	return count > 0
}

type CheckInService struct {
	repo *repository.Repository
}

func (s *CheckInService) Submit(userID uint64, challengeID uint64, content string, imageURL string) (*dto.SubmitCheckInResponse, error) {
	content = strings.TrimSpace(content)
	imageURL = strings.TrimSpace(imageURL)
	if !validator.HasTextOrImage(content, imageURL) {
		return nil, newAppError(http.StatusBadRequest, response.CodeValidationFailed, "打卡需要提供文字或图片证明")
	}
	var challenge model.Challenge
	if err := s.repo.DB.First(&challenge, challengeID).Error; err != nil {
		return nil, notFound(response.CodeChallengeNotFound, "挑战不存在")
	}
	if challenge.Status == "completed" || challenge.Status == "failed" || challenge.Status == "cancelled" {
		return nil, newAppError(http.StatusBadRequest, response.CodeChallengeEnded, "挑战已结束")
	}
	var participant model.ChallengeParticipant
	if err := s.repo.DB.Where("challenge_id = ? AND user_id = ? AND role <> ?", challengeID, userID, "witness").First(&participant).Error; err != nil {
		return nil, newAppError(http.StatusForbidden, response.CodeForbidden, "你尚未参与该挑战")
	}
	today := beginningOfDay(time.Now())
	var existing int64
	s.repo.DB.Model(&model.CheckIn{}).Where("challenge_id = ? AND user_id = ? AND checkin_date = ?", challengeID, userID, today).Count(&existing)
	if existing > 0 {
		return nil, newAppError(http.StatusBadRequest, response.CodeAlreadyCheckedIn, "今日已打卡")
	}

	streak := uint(1)
	var latest model.CheckIn
	if err := s.repo.DB.Where("challenge_id = ? AND user_id = ? AND checkin_date < ?", challengeID, userID, today).
		Order("checkin_date desc").First(&latest).Error; err == nil {
		if sameDate(latest.CheckinDate, today.AddDate(0, 0, -1)) {
			streak = latest.StreakCount + 1
		}
	}
	multiplier := xpMultiplier(streak)
	xpEarned := uint(math.Round(10 * multiplier))
	checkin := model.CheckIn{
		ChallengeID: challengeID,
		UserID:      userID,
		Content:     content,
		ImageURL:    imageURL,
		CheckinDate: today,
		StreakCount: streak,
		XPEarned:    xpEarned,
	}

	var unlocked []dto.AchievementResponse
	var totalXP uint
	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&checkin).Error; err != nil {
			return err
		}
		participant.CurrentStreak = streak
		if streak > participant.MaxStreak {
			participant.MaxStreak = streak
		}
		participant.TotalCheckins++
		if err := tx.Save(&participant).Error; err != nil {
			return err
		}
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		user.XP += xpEarned
		if streak == 7 {
			user.CreditScore = clampCredit(user.CreditScore + 5)
		}
		user.Level = calculateLevel(user.XP)
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		var err error
		unlocked, err = unlockCheckInAchievements(tx, userID, streak, time.Now(), user.CreditScore)
		if err != nil {
			return err
		}
		if len(unlocked) > 0 {
			var after model.User
			if err := tx.First(&after, userID).Error; err != nil {
				return err
			}
			totalXP = after.XP
		} else {
			totalXP = user.XP
		}
		if challenge.Status == "pending" && !time.Now().Before(challenge.StartDate) {
			if err := tx.Model(&challenge).Update("status", "active").Error; err != nil {
				return err
			}
		}
		return notifyWitnesses(tx, challenge, "witness", "监督对象完成打卡", fmt.Sprintf("挑战「%s」有新的打卡记录", challenge.Title))
	})
	if err != nil {
		return nil, err
	}
	return &dto.SubmitCheckInResponse{
		CheckinID:            checkin.ID,
		StreakCount:          streak,
		XPEarned:             xpEarned,
		XPMultiplier:         multiplier,
		TotalXP:              totalXP,
		AchievementsUnlocked: unlocked,
	}, nil
}

func (s *CheckInService) List(challengeID uint64) ([]dto.CheckInResponse, error) {
	var checkins []model.CheckIn
	if err := s.repo.DB.Preload("User").Where("challenge_id = ?", challengeID).Order("checkin_date desc, created_at desc").Find(&checkins).Error; err != nil {
		return nil, err
	}
	result := make([]dto.CheckInResponse, 0, len(checkins))
	for _, checkin := range checkins {
		result = append(result, mapCheckIn(checkin))
	}
	return result, nil
}

func (s *CheckInService) Today(userID uint64, challengeID uint64) (map[string]any, error) {
	today := beginningOfDay(time.Now())
	var checkin model.CheckIn
	if err := s.repo.DB.Where("challenge_id = ? AND user_id = ? AND checkin_date = ?", challengeID, userID, today).First(&checkin).Error; err == nil {
		return map[string]any{"checked_in": true, "used_freeze": false, "checkin": mapCheckIn(checkin)}, nil
	}
	var freeze model.StreakFreezeLog
	if err := s.repo.DB.Where("challenge_id = ? AND user_id = ? AND freeze_date = ?", challengeID, userID, today).First(&freeze).Error; err == nil {
		return map[string]any{"checked_in": false, "used_freeze": true}, nil
	}
	return map[string]any{"checked_in": false, "used_freeze": false}, nil
}

func (s *CheckInService) Streak(userID uint64, challengeID uint64) (map[string]any, error) {
	var participant model.ChallengeParticipant
	if err := s.repo.DB.Where("challenge_id = ? AND user_id = ?", challengeID, userID).First(&participant).Error; err != nil {
		return nil, newAppError(http.StatusForbidden, response.CodeForbidden, "你尚未参与该挑战")
	}
	return map[string]any{
		"current_streak": participant.CurrentStreak,
		"max_streak":     participant.MaxStreak,
		"total_checkins": participant.TotalCheckins,
	}, nil
}

func (s *CheckInService) UseFreeze(userID uint64, challengeID uint64) (map[string]any, error) {
	today := beginningOfDay(time.Now())
	var user model.User
	if err := s.repo.DB.First(&user, userID).Error; err != nil {
		return nil, notFound(response.CodeUserNotFound, "用户不存在")
	}
	if user.StreakFreezes == 0 {
		return nil, newAppError(http.StatusBadRequest, response.CodeNoFreezeAvailable, "没有可用的连击保护次数")
	}
	var participant model.ChallengeParticipant
	if err := s.repo.DB.Where("challenge_id = ? AND user_id = ?", challengeID, userID).First(&participant).Error; err != nil {
		return nil, newAppError(http.StatusForbidden, response.CodeForbidden, "你尚未参与该挑战")
	}
	var exists int64
	s.repo.DB.Model(&model.StreakFreezeLog{}).Where("challenge_id = ? AND user_id = ? AND freeze_date = ?", challengeID, userID, today).Count(&exists)
	if exists > 0 {
		return nil, newAppError(http.StatusBadRequest, response.CodeValidationFailed, "今日已使用连击保护")
	}
	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		log := model.StreakFreezeLog{UserID: userID, ChallengeID: challengeID, FreezeDate: today}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		return tx.Model(&user).Update("streak_freezes", gorm.Expr("streak_freezes - 1")).Error
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"remaining_freezes": user.StreakFreezes - 1,
		"freeze_date":       today.Format("2006-01-02"),
		"streak_preserved":  participant.CurrentStreak,
	}, nil
}

func (s *CheckInService) Like(userID uint64, checkinID uint64) error {
	var checkin model.CheckIn
	if err := s.repo.DB.First(&checkin, checkinID).Error; err != nil {
		return notFound(response.CodeChallengeNotFound, "打卡记录不存在")
	}
	var count int64
	s.repo.DB.Model(&model.CheckInInteraction{}).Where("checkin_id = ? AND user_id = ? AND type = ?", checkinID, userID, "like").Count(&count)
	if count > 0 {
		return nil
	}
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.CheckInInteraction{CheckinID: checkinID, UserID: userID, Type: "like"}).Error; err != nil {
			return err
		}
		if err := tx.Model(&checkin).Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			return err
		}
		if checkin.UserID != userID {
			return tx.Create(&model.Notification{
				UserID:      checkin.UserID,
				Type:        "like",
				Title:       "打卡获得点赞",
				Content:     "你的打卡记录获得了新的点赞",
				RelatedID:   checkin.ID,
				RelatedType: "checkin",
			}).Error
		}
		return nil
	})
}

func (s *CheckInService) Unlike(userID uint64, checkinID uint64) error {
	var deleted int64
	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("checkin_id = ? AND user_id = ? AND type = ?", checkinID, userID, "like").Delete(&model.CheckInInteraction{})
		if result.Error != nil {
			return result.Error
		}
		deleted = result.RowsAffected
		if deleted > 0 {
			return tx.Model(&model.CheckIn{}).Where("id = ? AND like_count > 0", checkinID).
				Update("like_count", gorm.Expr("like_count - 1")).Error
		}
		return nil
	})
	return err
}

func (s *CheckInService) Comment(userID uint64, checkinID uint64, content string) error {
	var checkin model.CheckIn
	if err := s.repo.DB.First(&checkin, checkinID).Error; err != nil {
		return notFound(response.CodeChallengeNotFound, "打卡记录不存在")
	}
	content = strings.TrimSpace(content)
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.CheckInInteraction{CheckinID: checkinID, UserID: userID, Type: "comment", Content: content}).Error; err != nil {
			return err
		}
		if checkin.UserID != userID {
			return tx.Create(&model.Notification{
				UserID:      checkin.UserID,
				Type:        "comment",
				Title:       "打卡收到评论",
				Content:     content,
				RelatedID:   checkin.ID,
				RelatedType: "checkin",
			}).Error
		}
		return nil
	})
}

func (s *CheckInService) Interactions(checkinID uint64) ([]map[string]any, error) {
	var interactions []model.CheckInInteraction
	if err := s.repo.DB.Preload("User").Where("checkin_id = ?", checkinID).Order("created_at asc").Find(&interactions).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(interactions))
	for _, item := range interactions {
		result = append(result, map[string]any{
			"id":         item.ID,
			"type":       item.Type,
			"content":    item.Content,
			"created_at": item.CreatedAt,
			"user":       userSummary(item.User, false),
		})
	}
	return result, nil
}

func (s *CheckInService) Report(userID uint64, checkinID uint64, reason string) error {
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Report{ReporterID: userID, TargetType: "checkin", TargetID: checkinID, Reason: strings.TrimSpace(reason), Status: "pending"}).Error; err != nil {
			return err
		}
		return tx.Model(&model.CheckIn{}).Where("id = ?", checkinID).Update("is_reported", true).Error
	})
}

type NotificationService struct {
	repo *repository.Repository
}

func (s *NotificationService) List(userID uint64, unreadOnly bool) ([]model.Notification, error) {
	var notifications []model.Notification
	query := s.repo.DB.Where("user_id = ?", userID).Order("created_at desc")
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}
	if err := query.Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (s *NotificationService) UnreadCount(userID uint64) (map[string]int64, error) {
	var count int64
	if err := s.repo.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error; err != nil {
		return nil, err
	}
	return map[string]int64{"count": count}, nil
}

func (s *NotificationService) MarkRead(userID uint64, notificationID uint64) error {
	return s.repo.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

func (s *NotificationService) MarkAllRead(userID uint64) error {
	return s.repo.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

type SocialService struct {
	repo *repository.Repository
}

func (s *SocialService) Friends(userID uint64) ([]map[string]any, error) {
	var outgoing []model.Friendship
	if err := s.repo.DB.Preload("Friend").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&outgoing).Error; err != nil {
		return nil, err
	}
	var incomingPending []model.Friendship
	if err := s.repo.DB.Preload("User").
		Where("friend_id = ? AND status = ?", userID, "pending").
		Order("created_at desc").
		Find(&incomingPending).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]any, 0, len(outgoing)+len(incomingPending))
	for _, friend := range outgoing {
		result = append(result, map[string]any{
			"id":         friend.ID,
			"status":     friend.Status,
			"direction":  "outgoing",
			"created_at": friend.CreatedAt,
			"friend":     userSummary(friend.Friend, false),
		})
	}
	for _, friend := range incomingPending {
		result = append(result, map[string]any{
			"id":         friend.ID,
			"status":     friend.Status,
			"direction":  "incoming",
			"created_at": friend.CreatedAt,
			"friend":     userSummary(friend.User, false),
		})
	}
	return result, nil
}

func (s *SocialService) Request(userID uint64, friendID uint64) error {
	if userID == friendID {
		return newAppError(http.StatusBadRequest, response.CodeValidationFailed, "不能添加自己为好友")
	}
	var friend model.User
	if err := s.repo.DB.First(&friend, friendID).Error; err != nil {
		return notFound(response.CodeUserNotFound, "用户不存在")
	}
	var existing model.Friendship
	err := s.repo.DB.
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userID, friendID, friendID, userID).
		First(&existing).Error
	if err == nil {
		if existing.Status == "accepted" {
			return nil
		}
		if existing.UserID == userID {
			return nil
		}
		return newAppError(http.StatusBadRequest, response.CodeValidationFailed, "对方已向你发送好友请求，请在待处理请求中接受")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Friendship{UserID: userID, FriendID: friendID, Status: "pending"}).Error; err != nil {
			return err
		}
		return tx.Create(&model.Notification{
			UserID:      friendID,
			Type:        "friend_request",
			Title:       "新的好友请求",
			Content:     "你收到一条新的好友请求",
			RelatedID:   userID,
			RelatedType: "user",
		}).Error
	})
}

func (s *SocialService) Accept(userID uint64, friendshipID uint64) error {
	var f model.Friendship
	if err := s.repo.DB.First(&f, friendshipID).Error; err != nil {
		return notFound(response.CodeUserNotFound, "好友请求不存在")
	}
	if f.FriendID != userID {
		return newAppError(http.StatusForbidden, response.CodeForbidden, "无权接受该请求")
	}
	if f.Status == "accepted" {
		return nil
	}
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&f).Update("status", "accepted").Error; err != nil {
			return err
		}
		reverse := model.Friendship{UserID: f.FriendID, FriendID: f.UserID}
		return tx.Where("user_id = ? AND friend_id = ?", f.FriendID, f.UserID).
			Assign(model.Friendship{Status: "accepted"}).
			FirstOrCreate(&reverse).Error
	})
}

func (s *SocialService) Delete(userID uint64, friendshipID uint64) error {
	var f model.Friendship
	if err := s.repo.DB.First(&f, friendshipID).Error; err != nil {
		return notFound(response.CodeUserNotFound, "好友关系不存在")
	}
	if f.UserID != userID && f.FriendID != userID {
		return newAppError(http.StatusForbidden, response.CodeForbidden, "无权删除该好友关系")
	}
	return s.repo.DB.
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", f.UserID, f.FriendID, f.FriendID, f.UserID).
		Delete(&model.Friendship{}).Error
}

type AchievementService struct {
	repo *repository.Repository
}

func (s *AchievementService) Seed() error {
	for _, achievement := range defaultAchievements() {
		var existing model.Achievement
		err := s.repo.DB.Where("name = ?", achievement.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.repo.DB.Create(&achievement).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *AchievementService) List(userID uint64) ([]dto.AchievementResponse, error) {
	var achievements []model.Achievement
	if err := s.repo.DB.Order("id asc").Find(&achievements).Error; err != nil {
		return nil, err
	}
	var unlocked []model.UserAchievement
	if err := s.repo.DB.Where("user_id = ?", userID).Find(&unlocked).Error; err != nil {
		return nil, err
	}
	unlockedMap := map[uint64]model.UserAchievement{}
	for _, item := range unlocked {
		unlockedMap[item.AchievementID] = item
	}
	result := make([]dto.AchievementResponse, 0, len(achievements))
	for _, achievement := range achievements {
		item, ok := unlockedMap[achievement.ID]
		resp := achievementResponse(achievement, ok, "")
		if ok {
			resp.UnlockedAt = item.UnlockedAt.Format(time.RFC3339)
		}
		result = append(result, resp)
	}
	return result, nil
}

func (s *AchievementService) Mine(userID uint64) ([]dto.AchievementResponse, error) {
	var items []model.UserAchievement
	if err := s.repo.DB.Preload("Achievement").Where("user_id = ?", userID).Order("unlocked_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	result := make([]dto.AchievementResponse, 0, len(items))
	for _, item := range items {
		result = append(result, achievementResponse(item.Achievement, true, item.UnlockedAt.Format(time.RFC3339)))
	}
	return result, nil
}

type LeaderboardService struct {
	repo *repository.Repository
}

func (s *LeaderboardService) Weekly(userID uint64) ([]dto.UserSummary, error) {
	if s.repo.Redis != nil {
		users, err := s.weeklyFromRedis()
		if err == nil {
			return users, nil
		}
	}
	var users []model.User
	if err := s.repo.DB.Where("credit_score >= ?", 60).Order("xp desc, credit_score desc").Limit(20).Find(&users).Error; err != nil {
		return nil, err
	}
	return mapUsers(users, false), nil
}

func (s *LeaderboardService) weeklyFromRedis() ([]dto.UserSummary, error) {
	const key = "leaderboard:weekly"
	ctx := context.Background()

	var users []model.User
	if err := s.repo.DB.Where("credit_score >= ?", 60).Order("xp desc, credit_score desc").Limit(100).Find(&users).Error; err != nil {
		return nil, err
	}
	pipe := s.repo.Redis.TxPipeline()
	pipe.Del(ctx, key)
	for _, user := range users {
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(user.XP),
			Member: fmt.Sprintf("%d", user.ID),
		})
	}
	pipe.Expire(ctx, key, 10*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	ranked, err := s.repo.Redis.ZRevRange(ctx, key, 0, 19).Result()
	if err != nil {
		return nil, err
	}
	if len(ranked) == 0 {
		return []dto.UserSummary{}, nil
	}
	ids := make([]uint64, 0, len(ranked))
	for _, raw := range ranked {
		var id uint64
		if _, err := fmt.Sscanf(raw, "%d", &id); err == nil {
			ids = append(ids, id)
		}
	}
	var dbUsers []model.User
	if err := s.repo.DB.Where("id IN ?", ids).Find(&dbUsers).Error; err != nil {
		return nil, err
	}
	userMap := make(map[uint64]model.User, len(dbUsers))
	for _, user := range dbUsers {
		userMap[user.ID] = user
	}
	result := make([]dto.UserSummary, 0, len(ids))
	for _, id := range ids {
		if user, ok := userMap[id]; ok {
			result = append(result, userSummary(user, false))
		}
	}
	return result, nil
}

func (s *LeaderboardService) Friends(userID uint64) ([]dto.UserSummary, error) {
	var friendIDs []uint64
	if err := s.repo.DB.Model(&model.Friendship{}).Where("user_id = ? AND status = ?", userID, "accepted").Pluck("friend_id", &friendIDs).Error; err != nil {
		return nil, err
	}
	friendIDs = append(friendIDs, userID)
	var users []model.User
	if err := s.repo.DB.Where("id IN ? AND credit_score >= ?", friendIDs, 60).Order("xp desc, credit_score desc").Find(&users).Error; err != nil {
		return nil, err
	}
	return mapUsers(users, false), nil
}

type CertificateService struct {
	repo *repository.Repository
}

func (s *CertificateService) List(userID uint64) ([]dto.CertificateResponse, error) {
	var certificates []model.Certificate
	if err := s.repo.DB.Preload("Challenge").Where("user_id = ?", userID).Order("issued_at desc").Find(&certificates).Error; err != nil {
		return nil, err
	}
	result := make([]dto.CertificateResponse, 0, len(certificates))
	for _, certificate := range certificates {
		result = append(result, mapCertificate(certificate))
	}
	return result, nil
}

func (s *CertificateService) Get(userID uint64, id uint64) (*dto.CertificateResponse, error) {
	var certificate model.Certificate
	if err := s.repo.DB.Preload("Challenge").Where("id = ? AND user_id = ?", id, userID).First(&certificate).Error; err != nil {
		return nil, notFound(response.CodeChallengeNotFound, "证书不存在")
	}
	result := mapCertificate(certificate)
	return &result, nil
}

type TaskService struct {
	repo *repository.Repository
}

func (s *TaskService) RunMissedCheckinSweep(now time.Time) (int, error) {
	checkDate := beginningOfDay(now).AddDate(0, 0, -1)
	var participants []model.ChallengeParticipant
	if err := s.repo.DB.
		Joins("JOIN challenges ON challenges.id = challenge_participants.challenge_id").
		Where("challenge_participants.role <> ?", "witness").
		Where("challenge_participants.status = ?", "accepted").
		Where("challenge_participants.current_streak > 0").
		Where("challenges.status IN ?", []string{"pending", "active"}).
		Where("challenges.start_date <= ? AND challenges.end_date >= ?", checkDate, checkDate).
		Find(&participants).Error; err != nil {
		return 0, err
	}

	affected := 0
	for _, participant := range participants {
		var checkins int64
		if err := s.repo.DB.Model(&model.CheckIn{}).
			Where("challenge_id = ? AND user_id = ? AND checkin_date = ?", participant.ChallengeID, participant.UserID, checkDate).
			Count(&checkins).Error; err != nil {
			return affected, err
		}
		if checkins > 0 {
			continue
		}
		var freezes int64
		if err := s.repo.DB.Model(&model.StreakFreezeLog{}).
			Where("challenge_id = ? AND user_id = ? AND freeze_date = ?", participant.ChallengeID, participant.UserID, checkDate).
			Count(&freezes).Error; err != nil {
			return affected, err
		}
		if freezes > 0 {
			continue
		}

		if err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&model.ChallengeParticipant{}).
				Where("id = ?", participant.ID).
				Update("current_streak", 0).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.User{}).
				Where("id = ?", participant.UserID).
				Update("credit_score", gorm.Expr("GREATEST(credit_score - ?, 0)", 5)).Error; err != nil {
				return err
			}
			return tx.Create(&model.Notification{
				UserID:      participant.UserID,
				Type:        "streak_warn",
				Title:       "断签提醒",
				Content:     fmt.Sprintf("%s 未完成打卡，连击已归零并扣减 5 点信用分", checkDate.Format("2006-01-02")),
				RelatedID:   participant.ChallengeID,
				RelatedType: "challenge",
			}).Error
		}); err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func userSummary(user model.User, includeEmail bool) dto.UserSummary {
	summary := dto.UserSummary{
		ID:            user.ID,
		Username:      user.Username,
		Avatar:        user.Avatar,
		XP:            user.XP,
		Level:         user.Level,
		CreditScore:   user.CreditScore,
		StreakFreezes: user.StreakFreezes,
	}
	if includeEmail {
		summary.Email = user.Email
	}
	return summary
}

func mapUsers(users []model.User, includeEmail bool) []dto.UserSummary {
	result := make([]dto.UserSummary, 0, len(users))
	for _, user := range users {
		result = append(result, userSummary(user, includeEmail))
	}
	return result
}

func mapChallenge(ch model.Challenge, checkins int64, participants int64) dto.ChallengeResponse {
	return dto.ChallengeResponse{
		ID:               ch.ID,
		CreatorID:        ch.CreatorID,
		Title:            ch.Title,
		Description:      ch.Description,
		Category:         ch.Category,
		Pledge:           ch.Pledge,
		PenaltyType:      ch.PenaltyType,
		PenaltyDetail:    ch.PenaltyDetail,
		ChallengeType:    ch.ChallengeType,
		TargetDays:       ch.TargetDays,
		StartDate:        ch.StartDate.Format("2006-01-02"),
		EndDate:          ch.EndDate.Format("2006-01-02"),
		Status:           ch.Status,
		IsPublic:         ch.IsPublic,
		XPReward:         ch.XPReward,
		CanCancelBefore:  ch.CanCancelBefore.Format(time.RFC3339),
		Progress:         progressPercent(checkins, ch.TargetDays),
		ParticipantCount: participants,
		CheckinCount:     checkins,
	}
}

func mapCheckIn(checkin model.CheckIn) dto.CheckInResponse {
	return dto.CheckInResponse{
		ID:          checkin.ID,
		ChallengeID: checkin.ChallengeID,
		UserID:      checkin.UserID,
		Content:     checkin.Content,
		ImageURL:    checkin.ImageURL,
		CheckinDate: checkin.CheckinDate.Format("2006-01-02"),
		StreakCount: checkin.StreakCount,
		XPEarned:    checkin.XPEarned,
		LikeCount:   checkin.LikeCount,
		CreatedAt:   checkin.CreatedAt,
		User:        userSummary(checkin.User, false),
	}
}

func mapCertificate(certificate model.Certificate) dto.CertificateResponse {
	return dto.CertificateResponse{
		ID:            certificate.ID,
		UserID:        certificate.UserID,
		ChallengeID:   certificate.ChallengeID,
		CertificateNo: certificate.CertificateNo,
		ImageURL:      certificate.ImageURL,
		IssuedAt:      certificate.IssuedAt.Format(time.RFC3339),
		Challenge:     mapChallenge(certificate.Challenge, 0, 0),
	}
}

func progressPercent(checkins int64, targetDays uint) float64 {
	if targetDays == 0 {
		return 0
	}
	progress := float64(checkins) / float64(targetDays) * 100
	if progress > 100 {
		return 100
	}
	return math.Round(progress*10) / 10
}

func beginningOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func sameDate(a time.Time, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func xpMultiplier(streak uint) float64 {
	switch {
	case streak >= 30:
		return 3
	case streak >= 14:
		return 2.5
	case streak >= 7:
		return 2
	case streak >= 3:
		return 1.5
	default:
		return 1
	}
}

func calculateLevel(xp uint) uint {
	return uint(math.Floor(math.Sqrt(float64(xp)/100))) + 1
}

func clampCredit(score int) int {
	if score > 200 {
		return 200
	}
	if score < 0 {
		return 0
	}
	return score
}

func notFound(code int, message string) *AppError {
	return newAppError(http.StatusNotFound, code, message)
}

func defaultAchievements() []model.Achievement {
	return []model.Achievement{
		{Name: "初次打卡", Description: "完成第一次打卡", Icon: "check-circle", Category: "checkin", ConditionType: "checkin_count", ConditionValue: 1, XPReward: 50},
		{Name: "三天连击", Description: "连续打卡 3 天", Icon: "fire", Category: "checkin", ConditionType: "streak", ConditionValue: 3, XPReward: 50},
		{Name: "七天连击", Description: "连续打卡 7 天", Icon: "thunderbolt", Category: "checkin", ConditionType: "streak", ConditionValue: 7, XPReward: 50},
		{Name: "两周坚持", Description: "连续打卡 14 天", Icon: "calendar", Category: "checkin", ConditionType: "streak", ConditionValue: 14, XPReward: 50},
		{Name: "月度之星", Description: "连续打卡 30 天", Icon: "star", Category: "checkin", ConditionType: "streak", ConditionValue: 30, XPReward: 50},
		{Name: "初次守约", Description: "完成第一个挑战", Icon: "trophy", Category: "challenge", ConditionType: "completed_challenges", ConditionValue: 1, XPReward: 50},
		{Name: "三战三捷", Description: "完成 3 个挑战", Icon: "crown", Category: "challenge", ConditionType: "completed_challenges", ConditionValue: 3, XPReward: 50},
		{Name: "挑战达人", Description: "完成 5 个挑战", Icon: "medal", Category: "challenge", ConditionType: "completed_challenges", ConditionValue: 5, XPReward: 50},
		{Name: "交个朋友", Description: "添加第一个好友", Icon: "team", Category: "social", ConditionType: "friends", ConditionValue: 1, XPReward: 50},
		{Name: "社交蝴蝶", Description: "拥有 5 个好友", Icon: "share-alt", Category: "social", ConditionType: "friends", ConditionValue: 5, XPReward: 50},
		{Name: "最佳见证", Description: "作为见证人确认 10 次打卡", Icon: "safety", Category: "social", ConditionType: "witness_confirm", ConditionValue: 10, XPReward: 50},
		{Name: "夜猫子", Description: "在 23:00-23:59 打卡", Icon: "moon", Category: "hidden", ConditionType: "checkin_hour_late", ConditionValue: 23, IsHidden: true, XPReward: 50},
		{Name: "早起鸟", Description: "在 05:00-07:00 打卡", Icon: "sun", Category: "hidden", ConditionType: "checkin_hour_early", ConditionValue: 5, IsHidden: true, XPReward: 50},
		{Name: "绝地反击", Description: "使用连击保护后继续保持 7 天连击", Icon: "reload", Category: "hidden", ConditionType: "freeze_recover", ConditionValue: 7, IsHidden: true, XPReward: 50},
		{Name: "金牌信用", Description: "信用分达到 180", Icon: "gold", Category: "hidden", ConditionType: "credit_score", ConditionValue: 180, IsHidden: true, XPReward: 50},
	}
}

func unlockCheckInAchievements(tx *gorm.DB, userID uint64, streak uint, now time.Time, creditScore int) ([]dto.AchievementResponse, error) {
	var totalCheckins int64
	if err := tx.Model(&model.CheckIn{}).Where("user_id = ?", userID).Count(&totalCheckins).Error; err != nil {
		return nil, err
	}
	targets := []string{}
	if totalCheckins >= 1 {
		targets = append(targets, "初次打卡")
	}
	if streak >= 3 {
		targets = append(targets, "三天连击")
	}
	if streak >= 7 {
		targets = append(targets, "七天连击")
	}
	if streak >= 14 {
		targets = append(targets, "两周坚持")
	}
	if streak >= 30 {
		targets = append(targets, "月度之星")
	}
	if now.Hour() == 23 {
		targets = append(targets, "夜猫子")
	}
	if now.Hour() >= 5 && now.Hour() <= 7 {
		targets = append(targets, "早起鸟")
	}
	if creditScore >= 180 {
		targets = append(targets, "金牌信用")
	}
	return unlockByNames(tx, userID, targets)
}

func unlockByNames(tx *gorm.DB, userID uint64, names []string) ([]dto.AchievementResponse, error) {
	unlocked := []dto.AchievementResponse{}
	for _, name := range names {
		var achievement model.Achievement
		if err := tx.Where("name = ?", name).First(&achievement).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		var count int64
		tx.Model(&model.UserAchievement{}).Where("user_id = ? AND achievement_id = ?", userID, achievement.ID).Count(&count)
		if count > 0 {
			continue
		}
		if err := tx.Create(&model.UserAchievement{UserID: userID, AchievementID: achievement.ID, UnlockedAt: time.Now()}).Error; err != nil {
			return nil, err
		}
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			Updates(map[string]any{
				"xp":    gorm.Expr("xp + ?", achievement.XPReward),
				"level": gorm.Expr("FLOOR(SQRT((xp + ?) / 100)) + 1", achievement.XPReward),
			}).Error; err != nil {
			return nil, err
		}
		if err := tx.Create(&model.Notification{
			UserID:      userID,
			Type:        "achievement",
			Title:       "解锁新成就",
			Content:     "你解锁了「" + achievement.Name + "」",
			RelatedID:   achievement.ID,
			RelatedType: "achievement",
		}).Error; err != nil {
			return nil, err
		}
		unlocked = append(unlocked, achievementResponse(achievement, true, time.Now().Format(time.RFC3339)))
	}
	return unlocked, nil
}

func achievementResponse(achievement model.Achievement, unlocked bool, unlockedAt string) dto.AchievementResponse {
	return dto.AchievementResponse{
		ID:          achievement.ID,
		Name:        achievement.Name,
		Description: achievement.Description,
		Icon:        achievement.Icon,
		Category:    achievement.Category,
		IsHidden:    achievement.IsHidden,
		Unlocked:    unlocked,
		UnlockedAt:  unlockedAt,
	}
}

func applyChallengeCompletion(tx *gorm.DB, cfg *config.Config, challenge model.Challenge) error {
	var user model.User
	if err := tx.First(&user, challenge.CreatorID).Error; err != nil {
		return err
	}
	user.XP += challenge.XPReward
	user.CreditScore = clampCredit(user.CreditScore + 10)
	user.Level = calculateLevel(user.XP)
	if err := tx.Save(&user).Error; err != nil {
		return err
	}
	var completed int64
	if err := tx.Model(&model.Challenge{}).Where("creator_id = ? AND status = ?", challenge.CreatorID, "completed").Count(&completed).Error; err != nil {
		return err
	}
	targets := []string{}
	if completed >= 1 {
		targets = append(targets, "初次守约")
	}
	if completed >= 3 {
		targets = append(targets, "三战三捷")
	}
	if completed >= 5 {
		targets = append(targets, "挑战达人")
	}
	if _, err := unlockByNames(tx, challenge.CreatorID, targets); err != nil {
		return err
	}
	cert := model.Certificate{
		UserID:        challenge.CreatorID,
		ChallengeID:   challenge.ID,
		CertificateNo: fmt.Sprintf("KP-%s-%05d", time.Now().Format("20060102"), challenge.ID),
		IssuedAt:      time.Now(),
	}
	if err := tx.Where("challenge_id = ? AND user_id = ?", challenge.ID, challenge.CreatorID).FirstOrCreate(&cert).Error; err != nil {
		return err
	}
	imageURL, err := generateCertificateImage(cfg, cert, challenge, user)
	if err != nil {
		return err
	}
	return tx.Model(&cert).Update("image_url", imageURL).Error
}

func notifyWitnesses(tx *gorm.DB, challenge model.Challenge, noticeType string, title string, content string) error {
	var witnesses []model.ChallengeParticipant
	if err := tx.Where("challenge_id = ? AND role = ?", challenge.ID, "witness").Find(&witnesses).Error; err != nil {
		return err
	}
	for _, witness := range witnesses {
		notice := model.Notification{
			UserID:      witness.UserID,
			Type:        noticeType,
			Title:       title,
			Content:     content,
			RelatedID:   challenge.ID,
			RelatedType: "challenge",
		}
		if err := tx.Create(&notice).Error; err != nil {
			return err
		}
	}
	return nil
}

func generateCertificateImage(cfg *config.Config, cert model.Certificate, challenge model.Challenge, user model.User) (string, error) {
	dir := filepath.Join(cfg.Upload.Dir, "certificates")
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", err
	}

	width, height := 1200, 800
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 248, G: 250, B: 252, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 60, width-60, height-60), &image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(80, 80, width-80, 110), &image.Uniform{C: color.RGBA{R: 99, G: 102, B: 241, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(80, height-110, width-80, height-80), &image.Uniform{C: color.RGBA{R: 99, G: 102, B: 241, A: 255}}, image.Point{}, draw.Src)

	titleColor := color.RGBA{R: 30, G: 41, B: 59, A: 255}
	textColor := color.RGBA{R: 71, G: 85, B: 105, A: 255}
	brandColor := color.RGBA{R: 79, G: 70, B: 229, A: 255}

	drawText(img, 420, 180, "KEEP PLEDGE", brandColor)
	drawText(img, 378, 230, "COMPLETION CERTIFICATE", titleColor)
	drawText(img, 140, 330, "Certificate No: "+cert.CertificateNo, textColor)
	drawText(img, 140, 380, "User: "+asciiSafe(user.Username), textColor)
	drawText(img, 140, 430, "Challenge: "+asciiSafe(challenge.Title), textColor)
	drawText(img, 140, 480, "Period: "+challenge.StartDate.Format("2006-01-02")+" to "+challenge.EndDate.Format("2006-01-02"), textColor)
	drawText(img, 140, 530, fmt.Sprintf("Target Days: %d", challenge.TargetDays), textColor)
	drawText(img, 140, 620, "Issued At: "+cert.IssuedAt.Format("2006-01-02"), textColor)
	drawText(img, 760, 620, "Status: COMPLETED", brandColor)

	fileName := cert.CertificateNo + ".png"
	fullPath := filepath.Join(dir, fileName)
	file, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		return "", err
	}

	publicPath := strings.TrimRight(cfg.Upload.PublicPath, "/")
	return publicPath + "/certificates/" + fileName, nil
}

func drawText(img *image.RGBA, x int, y int, text string, c color.Color) {
	d := &fontdraw.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}

func asciiSafe(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 32 && r <= 126 {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('?')
	}
	return b.String()
}
