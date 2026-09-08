package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func JSON(c *gin.Context, status int, message string, data ...interface{}) {
	body := gin.H{
		"error":   false,
		"message": message,
	}

	if len(data) > 0 {
		body["data"] = data[0]
	}

	c.JSON(status, body)
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{
		Error:   true,
		Message: message,
	})
}

func HTMLError(c *gin.Context, status int, message string) {
	c.HTML(status, "error", gin.H{
		"Title":   "Error",
		"Message": message,
		"Status":  status,
	})
}

func Redirect(c *gin.Context, url string) {
	c.Redirect(http.StatusFound, url)
}
