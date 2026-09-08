package booking

type CreateBookingRequest struct {
	StudentID    string `form:"student_id" json:"student_id" binding:"required,uuid" example:"b0000000-0000-0000-0000-000000000001"`
	TrialClassID string `form:"trial_class_id" json:"trial_class_id" binding:"required,uuid" example:"c0000000-0000-0000-0000-000000000001"`
}

type ProcessPaymentRequest struct {
	SimulateFailure string `form:"simulate_failure" json:"simulate_failure" example:"false"`
}
