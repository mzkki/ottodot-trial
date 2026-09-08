package admin

import (
	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(r *gin.RouterGroup, handler *AdminHandler) {
	r.GET("/roster", handler.GetRoster)
	r.GET("/roster/:class_id", handler.GetRoster)
	r.GET("/classes/all", handler.GetAllClasses)
}
