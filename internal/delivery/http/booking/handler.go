package booking

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mzkki/ottodot-trial/internal/usecase"
	"go.uber.org/zap"
)

type BookingHandler struct {
	bookingUseCase usecase.BookingUseCaseInterface
	logger         *zap.Logger
}

func NewBookingHandler(bookingUseCase usecase.BookingUseCaseInterface, logger *zap.Logger) *BookingHandler {
	return &BookingHandler{
		bookingUseCase: bookingUseCase,
		logger:         logger.Named("BookingHandler"),
	}
}

// GetStudents returns an HTMX partial with students for a given parent.
// @Summary      Get students by parent
// @Description  Returns student list options for a selected parent (HTMX partial)
// @Tags         Booking
// @Produce      html
// @Param        parent_id query string true "Parent UUID" Format(uuid) example("a0000000-0000-0000-0000-000000000001")
// @Success      200 {string} string "HTML student_select partial"
// @Failure      400 {string} string "HTML error partial"
// @Router       /students [get]
func (h *BookingHandler) GetStudents(c *gin.Context) {
	parentID := c.Query("parent_id")
	if parentID == "" {
		c.HTML(http.StatusOK, "student_select", gin.H{"Students": nil})
		return
	}

	if _, err := uuid.Parse(parentID); err != nil {
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Invalid parent ID"})
		return
	}

	students, err := h.bookingUseCase.GetStudentsByParent(c.Request.Context(), parentID)
	if err != nil {
		h.logger.Error("failed_to_get_students", zap.Error(err))
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Failed to load students"})
		return
	}

	c.HTML(http.StatusOK, "student_select", gin.H{"Students": students})
}

// GetClasses returns an HTMX partial with available classes and booking status for a student.
// It leverages GetAvailableClassesWithStatus to execute in a single batch query (no N+1).
// @Summary      Get available trial classes
// @Description  Returns class list enriched with real-time spot availability and booking status for a student (HTMX partial)
// @Tags         Booking
// @Produce      html
// @Param        student_id query string true "Student UUID" Format(uuid) example("b0000000-0000-0000-0000-000000000001")
// @Success      200 {string} string "HTML class_list partial"
// @Failure      400 {string} string "HTML error partial"
// @Router       /classes [get]
func (h *BookingHandler) GetClasses(c *gin.Context) {
	studentID := c.Query("student_id")
	if studentID == "" {
		c.HTML(http.StatusOK, "class_list", gin.H{"Classes": nil})
		return
	}

	if _, err := uuid.Parse(studentID); err != nil {
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Invalid student ID"})
		return
	}

	classesWithCount, err := h.bookingUseCase.GetAvailableClassesWithStatus(c.Request.Context(), studentID)
	if err != nil {
		h.logger.Error("failed_to_get_classes", zap.Error(err))
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Failed to load classes"})
		return
	}

	c.HTML(http.StatusOK, "class_list", gin.H{
		"Classes":   classesWithCount,
		"StudentID": studentID,
	})
}

// CreateBooking creates a pending booking and returns the payment form.
// @Summary      Create a trial class booking
// @Description  Creates a pending booking for a student and trial class, returning the payment form
// @Tags         Booking
// @Accept       x-www-form-urlencoded,json
// @Produce      html
// @Param        request body booking.CreateBookingRequest true "Booking payload"
// @Success      200 {string} string "HTML booking_form partial"
// @Failure      400 {string} string "HTML error partial"
// @Router       /bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Please select a valid student and class"})
		return
	}

	booking, err := h.bookingUseCase.CreateBooking(c.Request.Context(), req.StudentID, req.TrialClassID)
	if err != nil {
		h.logger.Error("failed_to_create_booking", zap.Error(err))

		msg := "Failed to create booking"
		if errors.Is(err, usecase.ErrDuplicateBooking) {
			msg = "This student already has an active booking for this class"
		} else if errors.Is(err, usecase.ErrClassFull) {
			msg = "This class is full — no spots available"
		}

		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": msg})
		return
	}

	c.HTML(http.StatusOK, "booking_form", gin.H{
		"Booking": booking,
	})
}

// ProcessPayment handles the payment flow with race condition protection.
// @Summary      Process booking payment
// @Description  Processes payment with pessimistic row-locking concurrency guard and deterministic simulation toggle
// @Tags         Booking
// @Accept       x-www-form-urlencoded,json
// @Produce      html
// @Param        id path string true "Booking UUID" Format(uuid) example("d0000000-0000-0000-0000-000000000001")
// @Param        request body booking.ProcessPaymentRequest false "Payment simulation options"
// @Success      200 {string} string "HTML booking_status partial"
// @Failure      400 {string} string "HTML error partial"
// @Router       /bookings/{id}/pay [post]
func (h *BookingHandler) ProcessPayment(c *gin.Context) {
	bookingID := c.Param("id")
	if _, err := uuid.Parse(bookingID); err != nil {
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Invalid booking ID"})
		return
	}

	var req ProcessPaymentRequest
	_ = c.ShouldBind(&req)
	simulateFailure := req.SimulateFailure == "true"

	booking, err := h.bookingUseCase.ProcessPayment(c.Request.Context(), bookingID, simulateFailure)
	if err != nil {
		h.logger.Error("payment_processing_failed",
			zap.String("booking_id", bookingID),
			zap.Error(err),
		)

		// Even on error, we may have a booking to display
		if booking != nil {
			msg := "Payment failed"
			if errors.Is(err, usecase.ErrClassFull) {
				msg = "Sorry, this class just filled up! Another student took the last spot."
			} else if errors.Is(err, usecase.ErrBookingNotPending) {
				msg = "This booking has already been processed"
			} else if errors.Is(err, usecase.ErrPaymentFailed) {
				msg = "Payment was declined. You can try again."
			}

			c.HTML(http.StatusOK, "booking_status", gin.H{
				"Booking":      booking,
				"ErrorMessage": msg,
			})
			return
		}

		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Failed to process payment"})
		return
	}

	c.HTML(http.StatusOK, "booking_status", gin.H{
		"Booking":      booking,
		"ErrorMessage": "",
	})
}

// GetBookingStatus returns the current status of a booking.
// @Summary      Get booking status
// @Description  Returns the current booking status and payment attempt history (HTMX partial)
// @Tags         Booking
// @Produce      html
// @Param        id path string true "Booking UUID" Format(uuid) example("d0000000-0000-0000-0000-000000000001")
// @Success      200 {string} string "HTML booking_status partial"
// @Failure      400 {string} string "HTML error partial"
// @Router       /bookings/{id}/status [get]
func (h *BookingHandler) GetBookingStatus(c *gin.Context) {
	bookingID := c.Param("id")
	if _, err := uuid.Parse(bookingID); err != nil {
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Invalid booking ID"})
		return
	}

	booking, err := h.bookingUseCase.GetBookingByID(c.Request.Context(), bookingID)
	if err != nil {
		h.logger.Error("failed_to_get_booking", zap.Error(err))
		c.HTML(http.StatusOK, "error_partial", gin.H{"Message": "Booking not found"})
		return
	}

	c.HTML(http.StatusOK, "booking_status", gin.H{
		"Booking":      booking,
		"ErrorMessage": "",
	})
}
