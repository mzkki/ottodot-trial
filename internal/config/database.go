package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(v *viper.Viper) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		v.GetString("DB_HOST"),
		v.GetString("DB_USER"),
		v.GetString("DB_PASS"),
		v.GetString("DB_NAME"),
		v.GetString("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	maxOpen := v.GetInt("DB_MAX_OPEN_CONNS")
	if maxOpen == 0 {
		maxOpen = 25
	}
	sqlDB.SetMaxOpenConns(maxOpen)

	maxIdle := v.GetInt("DB_MAX_IDLE_CONNS")
	if maxIdle == 0 {
		maxIdle = 10
	}
	sqlDB.SetMaxIdleConns(maxIdle)

	lifetime := v.GetDuration("DB_CONN_MAX_LIFETIME")
	if lifetime == 0 {
		lifetime = 5 * time.Minute
	}
	sqlDB.SetConnMaxLifetime(lifetime)

	return db, nil
}
