package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess              = 0
	CodeValidationFailed     = 1001
	CodeUnauthorized         = 1002
	CodeForbidden            = 1003
	CodeUserNotFound         = 2001
	CodeUserExists           = 2002
	CodePasswordWrong        = 2003
	CodeChallengeNotFound    = 3001
	CodeChallengeEnded       = 3002
	CodeAlreadyCheckedIn     = 3003
	CodeCoolingPeriodExpired = 3004
	CodeCreditNotEnough      = 3005
	CodeNoFreezeAvailable    = 3006
	CodeAchievementNotFound  = 4001
	CodeInternal             = 5001
)

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func Message(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Body{
		Code:    CodeSuccess,
		Message: message,
	})
}

func Fail(c *gin.Context, status int, code int, message string) {
	c.JSON(status, Body{
		Code:    code,
		Message: message,
	})
}

func Validation(c *gin.Context, err error) {
	Fail(c, http.StatusBadRequest, CodeValidationFailed, err.Error())
}
