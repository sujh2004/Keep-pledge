package handler

import (
	"errors"
	"net/http"
	"strconv"

	"keep-pledge/backend/internal/config"
	"keep-pledge/backend/internal/dto"
	"keep-pledge/backend/internal/middleware"
	"keep-pledge/backend/internal/service"
	authpkg "keep-pledge/backend/pkg/auth"
	"keep-pledge/backend/pkg/response"
	"keep-pledge/backend/pkg/upload"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.Services
	cfg      *config.Config
	jwt      *authpkg.Manager
}

func New(services *service.Services, cfg *config.Config, jwt *authpkg.Manager) *Handler {
	return &Handler{services: services, cfg: cfg, jwt: jwt}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	api.POST("/auth/register", h.register)
	api.POST("/auth/login", h.login)

	protected := api.Group("")
	protected.Use(middleware.Auth(h.jwt))
	{
		protected.POST("/auth/refresh", h.refresh)
		protected.GET("/auth/me", h.me)

		protected.GET("/users/:id", h.publicUser)
		protected.PUT("/users/profile", h.updateProfile)
		protected.PUT("/users/avatar", h.uploadAvatar)
		protected.GET("/users/:id/stats", h.userStats)
		protected.GET("/users/:id/challenges", h.userChallenges)
		protected.GET("/users/:id/achievements", h.userAchievements)
		protected.GET("/users/:id/heatmap", h.userHeatmap)

		protected.POST("/challenges", h.createChallenge)
		protected.GET("/challenges", h.listChallenges)
		protected.GET("/challenges/:id", h.getChallenge)
		protected.DELETE("/challenges/:id", h.cancelChallenge)
		protected.POST("/challenges/:id/join", h.joinChallenge)
		protected.POST("/challenges/:id/invite", h.inviteChallenge)
		protected.GET("/challenges/:id/participants", h.challengeParticipants)
		protected.GET("/challenges/:id/progress", h.challengeProgress)
		protected.PUT("/challenges/:id/status", h.updateChallengeStatus)
		protected.POST("/challenges/:id/appeal", h.appealChallenge)

		protected.POST("/challenges/:id/checkin", h.submitCheckIn)
		protected.GET("/challenges/:id/checkins", h.listCheckIns)
		protected.GET("/challenges/:id/checkins/today", h.todayCheckIn)
		protected.GET("/challenges/:id/streak", h.streak)
		protected.POST("/challenges/:id/streak-freeze", h.useFreeze)

		protected.POST("/checkins/:id/like", h.likeCheckIn)
		protected.DELETE("/checkins/:id/like", h.unlikeCheckIn)
		protected.POST("/checkins/:id/comment", h.commentCheckIn)
		protected.GET("/checkins/:id/interactions", h.checkInInteractions)
		protected.POST("/checkins/:id/report", h.reportCheckIn)

		protected.GET("/friends", h.friends)
		protected.POST("/friends/request", h.friendRequest)
		protected.PUT("/friends/:id/accept", h.acceptFriend)
		protected.DELETE("/friends/:id", h.deleteFriend)

		protected.GET("/leaderboard/weekly", h.weeklyLeaderboard)
		protected.GET("/leaderboard/friends", h.friendsLeaderboard)

		protected.GET("/achievements", h.achievements)
		protected.GET("/achievements/my", h.myAchievements)

		protected.GET("/certificates", h.certificates)
		protected.GET("/certificates/:id", h.certificate)

		protected.GET("/notifications", h.notifications)
		protected.GET("/notifications/unread-count", h.unreadCount)
		protected.PUT("/notifications/:id", h.notificationAction)
		protected.PUT("/notifications/:id/read", h.markNotificationRead)
	}
}

func (h *Handler) register(c *gin.Context) {
	var req dto.RegisterRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.services.Auth.Register(req)
	writeResult(c, result, err)
}

func (h *Handler) login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.services.Auth.Login(req)
	writeResult(c, result, err)
}

func (h *Handler) refresh(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Auth.Refresh(userID)
	writeResult(c, result, err)
}

func (h *Handler) me(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Auth.Me(userID)
	writeResult(c, result, err)
}

func (h *Handler) publicUser(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.Auth.PublicUser(id)
	writeResult(c, result, err)
}

func (h *Handler) updateProfile(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.UpdateProfileRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.services.Auth.UpdateProfile(userID, req)
	writeResult(c, result, err)
}

func (h *Handler) uploadAvatar(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	url, err := upload.SaveOptional(c, "avatar", h.cfg.Upload.Dir, h.cfg.Upload.PublicPath, h.cfg.Upload.MaxSizeMB)
	if err != nil {
		writeError(c, err)
		return
	}
	result, err := h.services.Auth.UpdateProfile(userID, dto.UpdateProfileRequest{Avatar: url})
	writeResult(c, result, err)
}

func (h *Handler) userStats(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.Auth.Stats(id)
	writeResult(c, result, err)
}

func (h *Handler) userChallenges(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.Challenge.List(id, c.Query("category"))
	writeResult(c, result, err)
}

func (h *Handler) userAchievements(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.Achievement.List(id)
	writeResult(c, result, err)
}

func (h *Handler) userHeatmap(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.Auth.Heatmap(id)
	writeResult(c, result, err)
}

func (h *Handler) createChallenge(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.CreateChallengeRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.services.Challenge.Create(userID, req)
	writeResult(c, result, err)
}

func (h *Handler) listChallenges(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Challenge.List(userID, c.Query("category"))
	writeResult(c, result, err)
}

func (h *Handler) exploreChallenges(c *gin.Context) {
	result, err := h.services.Challenge.Explore(c.Query("category"))
	writeResult(c, result, err)
}

func (h *Handler) getChallenge(c *gin.Context) {
	if c.Param("id") == "explore" {
		h.exploreChallenges(c)
		return
	}
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.Challenge.Get(userID, id)
	writeResult(c, result, err)
}

func (h *Handler) cancelChallenge(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.services.Challenge.Cancel(userID, id); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "挑战已撤回")
}

func (h *Handler) joinChallenge(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.services.Challenge.Join(userID, id); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "加入成功")
}

func (h *Handler) inviteChallenge(c *gin.Context) {
	response.Message(c, "邀请入口已预留，请在好友模块选择见证人")
}

func (h *Handler) challengeParticipants(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.Challenge.Participants(userID, id)
	writeResult(c, result, err)
}

func (h *Handler) challengeProgress(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.Challenge.Progress(userID, id)
	writeResult(c, result, err)
}

func (h *Handler) updateChallengeStatus(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if !bindJSON(c, &req) {
		return
	}
	if err := h.services.Challenge.UpdateStatus(userID, id, req.Status); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "状态已更新")
}

func (h *Handler) appealChallenge(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req dto.AppealRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.services.Challenge.Appeal(userID, id, req.Reason); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "申诉已提交，等待见证人确认")
}

func (h *Handler) submitCheckIn(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	imageURL, err := upload.SaveOptional(c, "image", h.cfg.Upload.Dir, h.cfg.Upload.PublicPath, h.cfg.Upload.MaxSizeMB)
	if err != nil {
		writeError(c, err)
		return
	}
	result, err := h.services.CheckIn.Submit(userID, id, c.PostForm("content"), imageURL)
	writeResult(c, result, err)
}

func (h *Handler) listCheckIns(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.CheckIn.List(id)
	writeResult(c, result, err)
}

func (h *Handler) todayCheckIn(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.CheckIn.Today(userID, id)
	writeResult(c, result, err)
}

func (h *Handler) streak(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.CheckIn.Streak(userID, id)
	writeResult(c, result, err)
}

func (h *Handler) useFreeze(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.CheckIn.UseFreeze(userID, id)
	writeResult(c, result, err)
}

func (h *Handler) likeCheckIn(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.services.CheckIn.Like(userID, id); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "点赞成功")
}

func (h *Handler) unlikeCheckIn(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.services.CheckIn.Unlike(userID, id); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "已取消点赞")
}

func (h *Handler) commentCheckIn(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req dto.CommentRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.services.CheckIn.Comment(userID, id, req.Content); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "评论成功")
}

func (h *Handler) checkInInteractions(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.CheckIn.Interactions(id)
	writeResult(c, result, err)
}

func (h *Handler) reportCheckIn(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req dto.ReportRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.services.CheckIn.Report(userID, id, req.Reason); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "举报已提交")
}

func (h *Handler) friends(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Social.Friends(userID)
	writeResult(c, result, err)
}

func (h *Handler) friendRequest(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.FriendRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.services.Social.Request(userID, req.FriendID); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "好友请求已发送")
}

func (h *Handler) acceptFriend(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.services.Social.Accept(userID, id); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "已接受好友请求")
}

func (h *Handler) deleteFriend(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.services.Social.Delete(userID, id); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "好友已删除")
}

func (h *Handler) weeklyLeaderboard(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Leaderboard.Weekly(userID)
	writeResult(c, result, err)
}

func (h *Handler) friendsLeaderboard(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Leaderboard.Friends(userID)
	writeResult(c, result, err)
}

func (h *Handler) achievements(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Achievement.List(userID)
	writeResult(c, result, err)
}

func (h *Handler) myAchievements(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Achievement.Mine(userID)
	writeResult(c, result, err)
}

func (h *Handler) certificates(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Certificate.List(userID)
	writeResult(c, result, err)
}

func (h *Handler) certificate(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	result, err := h.services.Certificate.Get(userID, id)
	writeResult(c, result, err)
}

func (h *Handler) notifications(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Notification.List(userID, c.Query("unread") == "true")
	writeResult(c, result, err)
}

func (h *Handler) unreadCount(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	result, err := h.services.Notification.UnreadCount(userID)
	writeResult(c, result, err)
}

func (h *Handler) markNotificationRead(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.services.Notification.MarkRead(userID, id); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "已标记为已读")
}

func (h *Handler) notificationAction(c *gin.Context) {
	if c.Param("id") != "read-all" {
		response.Fail(c, http.StatusNotFound, response.CodeValidationFailed, "通知操作不存在")
		return
	}
	h.markAllNotificationsRead(c)
}

func (h *Handler) markAllNotificationsRead(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	if err := h.services.Notification.MarkAllRead(userID); err != nil {
		writeError(c, err)
		return
	}
	response.Message(c, "已全部标记为已读")
}

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		response.Validation(c, err)
		return false
	}
	return true
}

func currentUser(c *gin.Context) (uint64, bool) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未授权")
		return 0, false
	}
	return userID, true
}

func paramID(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeValidationFailed, "路径参数不合法")
		return 0, false
	}
	return id, true
}

func writeResult(c *gin.Context, data any, err error) {
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, data)
}

func writeError(c *gin.Context, err error) {
	var appErr *service.AppError
	if errors.As(err, &appErr) {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, response.CodeInternal, err.Error())
}
