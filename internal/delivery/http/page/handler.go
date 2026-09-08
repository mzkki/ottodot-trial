package page

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mzkki/ottodot-trial/internal/usecase"
	"go.uber.org/zap"
)

type PageHandler struct {
	bookingUseCase usecase.BookingUseCaseInterface
	logger         *zap.Logger
}

func NewPageHandler(bookingUseCase usecase.BookingUseCaseInterface, logger *zap.Logger) *PageHandler {
	return &PageHandler{
		bookingUseCase: bookingUseCase,
		logger:         logger.Named("PageHandler"),
	}
}

// Home renders the main booking page with parent selector.
func (h *PageHandler) Home(c *gin.Context) {
	parents, err := h.bookingUseCase.GetAllParents(c.Request.Context())
	if err != nil {
		h.logger.Error("failed_to_get_parents", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title":   "Error",
			"Message": "Failed to load parents",
		})
		return
	}

	c.HTML(http.StatusOK, "home", gin.H{
		"Title":   "Book a Trial Class",
		"Parents": parents,
	})
}

// AdminRoster renders the admin class roster page.
func (h *PageHandler) AdminRoster(c *gin.Context) {
	classes, err := h.bookingUseCase.GetAllClasses(c.Request.Context())
	if err != nil {
		h.logger.Error("failed_to_get_classes", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title":   "Error",
			"Message": "Failed to load classes",
		})
		return
	}

	c.HTML(http.StatusOK, "admin_roster", gin.H{
		"Title":   "Class Roster",
		"Classes": classes,
	})
}
