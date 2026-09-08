package bootstrap

import (
	"time"

	"github.com/mzkki/ottodot-trial/internal/config"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/admin"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/booking"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/page"
	"github.com/mzkki/ottodot-trial/internal/infrastructure/payment"
	tx "github.com/mzkki/ottodot-trial/internal/pkg/tx"
	"github.com/mzkki/ottodot-trial/internal/repository"
	"github.com/mzkki/ottodot-trial/internal/usecase"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ─── Infrastructure Providers ─────────────────────────────────────────────

func ProvideLogger(cfg *viper.Viper) (*zap.Logger, error) {
	return config.NewLogger(cfg)
}

func ProvideTxManager(db *gorm.DB) *tx.GormTransactionManager {
	return tx.NewGormTransactionManager(db)
}

func ProvidePaymentGateway() usecase.PaymentGateway {
	return payment.NewMockPaymentGateway(2 * time.Second)
}

// ─── Repository Providers ─────────────────────────────────────────────────

func ProvideParentRepo(db *gorm.DB) *repository.ParentRepository {
	return repository.NewParentRepository(db)
}

func ProvideStudentRepo(db *gorm.DB) *repository.StudentRepository {
	return repository.NewStudentRepository(db)
}

func ProvideTrialClassRepo(db *gorm.DB) *repository.TrialClassRepository {
	return repository.NewTrialClassRepository(db)
}

func ProvideBookingRepo(db *gorm.DB) *repository.BookingRepository {
	return repository.NewBookingRepository(db)
}

func ProvidePaymentAttemptRepo(db *gorm.DB) *repository.PaymentAttemptRepository {
	return repository.NewPaymentAttemptRepository(db)
}

// ─── Usecase Providers ────────────────────────────────────────────────────

func ProvideBookingUseCase(
	parentRepo *repository.ParentRepository,
	studentRepo *repository.StudentRepository,
	classRepo *repository.TrialClassRepository,
	bookingRepo *repository.BookingRepository,
	paymentRepo *repository.PaymentAttemptRepository,
	paymentGateway usecase.PaymentGateway,
	txManager *tx.GormTransactionManager,
) *usecase.BookingUseCase {
	return usecase.NewBookingUseCase(parentRepo, studentRepo, classRepo, bookingRepo, paymentRepo, paymentGateway, txManager)
}

// ─── Handler Providers ────────────────────────────────────────────────────

func ProvidePageHandler(bookingUseCase *usecase.BookingUseCase, logger *zap.Logger) *page.PageHandler {
	return page.NewPageHandler(bookingUseCase, logger)
}

func ProvideBookingHandler(bookingUseCase *usecase.BookingUseCase, logger *zap.Logger) *booking.BookingHandler {
	return booking.NewBookingHandler(bookingUseCase, logger)
}

func ProvideAdminHandler(bookingUseCase *usecase.BookingUseCase, logger *zap.Logger) *admin.AdminHandler {
	return admin.NewAdminHandler(bookingUseCase, logger)
}
