package page

import (
	"github.com/gin-gonic/gin"
)

func RegisterPageRoutes(r *gin.Engine, handler *PageHandler) {
	r.GET("/", handler.Home)
	r.GET("/admin/roster", handler.AdminRoster)
}
