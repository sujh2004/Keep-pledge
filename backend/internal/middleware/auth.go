package middleware

import (
	"net/http"
	"strings"

	authpkg "keep-pledge/backend/pkg/auth"
	"keep-pledge/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

func Auth(jwt *authpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未授权")
			c.Abort()
			return
		}
		claims, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "Token 无效或已过期")
			c.Abort()
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (uint64, bool) {
	value, ok := c.Get(UserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint64)
	return userID, ok
}
