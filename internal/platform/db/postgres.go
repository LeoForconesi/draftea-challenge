package db

import (
	"fmt"

	"draftea-challenge/internal/platform/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgres opens a Postgres connection using GORM.
func NewPostgres(cfg config.DBConfig, log *zap.Logger) (*gorm.DB, func() error, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
		cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		if err := sqlDB.Close(); err != nil {
			log.Warn("failed to close db", zap.Error(err))
			return err
		}
		return nil
	}

	return db, cleanup, nil
}
