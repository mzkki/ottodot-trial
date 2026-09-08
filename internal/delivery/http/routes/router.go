package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mzkki/ottodot-trial/internal/bootstrap"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/admin"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/booking"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/page"
)

func RegisterRoutes(r *gin.Engine, h *bootstrap.Handlers) {
	// Full HTML pages (top-level routes)
	page.RegisterPageRoutes(r, h.Page)

	// API v1 routes (HTMX partials + JSON)
	v1 := r.Group("/api/v1")
	booking.RegisterBookingRoutes(v1, h.Booking)
	admin.RegisterAdminRoutes(v1, h.Admin)
}
