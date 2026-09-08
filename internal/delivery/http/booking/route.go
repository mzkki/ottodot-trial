package booking

import (
	"github.com/gin-gonic/gin"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/middleware"
	"golang.org/x/time/rate"
)

func RegisterBookingRoutes(r *gin.RouterGroup, handler *BookingHandler) {
	// Mutating actions (booking creation & payment) have strict rate limiting
	mutatingLimiter := middleware.RateLimit(middleware.NewIPRateLimiter(rate.Limit(5), 10))

	r.GET("/students", handler.GetStudents)
	r.GET("/classes", handler.GetClasses)
	r.POST("/bookings", mutatingLimiter, handler.CreateBooking)
	r.POST("/bookings/:id/pay", mutatingLimiter, handler.ProcessPayment)
	r.GET("/bookings/:id/status", handler.GetBookingStatus)
}
