package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzkki/ottodot-trial/internal/usecase"
	"go.uber.org/zap"
)

type AdminHandler struct {
	bookingUseCase usecase.BookingUseCaseInterface
	logger         *zap.Logger
}

func NewAdminHandler(bookingUseCase usecase.BookingUseCaseInterface, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		bookingUseCase: bookingUseCase,
		logger:         logger.Named("AdminHandler"),
	}
}

// GetRoster returns an HTMX partial with the confirmed roster for a class.
// @Summary      Get class roster
// @Description  Returns confirmed bookings roster for a trial class (HTMX partial)
// @Tags         Admin
// @Produce      html
// @Param        class_id query string false "Trial Class UUID" Format(uuid) example("c0000000-0000-0000-0000-000000000001")
// @Success      200 {string} string "HTML roster_table partial"
// @Failure      400 {string} string "HTML error partial"
// @Router       /roster [get]
func (h *AdminHandler) GetRoster(c *gin.Context) {
	classID := c.Param("class_id")
	if classID == "" {
		classID = c.Query("class_id")
	}
	if classID == "" {
		c.HTML(http.StatusOK, "roster_table", gin.H{
			"Class":    nil,
			"Bookings": nil,
			"Count":    0,
		})
		return
	}

	if _, err := uuid.Parse(classID); err != nil {
		c.HTML(http.StatusBadRequest, "error_partial", gin.H{"Message": "Invalid class ID"})
		return
	}

	class, err := h.bookingUseCase.GetClassByID(c.Request.Context(), classID)
	if err != nil {
		h.logger.Error("failed_to_get_class", zap.Error(err))
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Class not found"})
		return
	}

	bookings, err := h.bookingUseCase.GetClassRoster(c.Request.Context(), classID)
	if err != nil {
		h.logger.Error("failed_to_get_roster", zap.Error(err))
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Failed to load roster"})
		return
	}

	c.HTML(http.StatusOK, "roster_table", gin.H{
		"Class":    class,
		"Bookings": bookings,
		"Count":    len(bookings),
	})
}

// GetAllClasses returns all classes as JSON (for API consumers).
// @Summary      Get all trial classes
// @Description  Returns all trial classes in JSON format
// @Tags         Admin
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      500 {object} response.ErrorResponse
// @Router       /classes/all [get]
func (h *AdminHandler) GetAllClasses(c *gin.Context) {
	classes, err := h.bookingUseCase.GetAllClasses(c.Request.Context())
	if err != nil {
		h.logger.Error("failed_to_get_classes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to load classes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": false, "data": classes})
}
