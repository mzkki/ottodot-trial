//go:build wireinject
// +build wireinject

package bootstrap

import (
	"github.com/mzkki/ottodot-trial/internal/delivery/http/admin"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/booking"
	"github.com/mzkki/ottodot-trial/internal/delivery/http/page"
	"github.com/google/wire"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handlers struct {
	Page    *page.PageHandler
	Booking *booking.BookingHandler
	Admin   *admin.AdminHandler
	Logger  *zap.Logger
	Config  *viper.Viper
}

func InitializeHandlers(db *gorm.DB, config *viper.Viper) (*Handlers, error) {
	wire.Build(
		ProvideLogger,
		ProvideTxManager,

		// Repositories
		ProvideParentRepo,
		ProvideStudentRepo,
		ProvideTrialClassRepo,
		ProvideBookingRepo,
		ProvidePaymentAttemptRepo,

		// Usecases
		ProvidePaymentGateway,
		ProvideBookingUseCase,

		// Handlers
		ProvidePageHandler,
		ProvideBookingHandler,
		ProvideAdminHandler,

		wire.Struct(new(Handlers), "*"),
	)
	return nil, nil
}
