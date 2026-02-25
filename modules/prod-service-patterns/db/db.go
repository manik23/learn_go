package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	DB    *gorm.DB
	Token chan struct{}
}

type User struct {
	ID        uint64 `gorm:"primaryKey"`
	Name      string
	Email     string
	CreatedAt time.Time
}

func NewDatabase(ctx context.Context, capacity int) (*Database, error) {
	db, err := gorm.Open(sqlite.Open("prod-service-patterns.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database %w", err)
	}

	// Automigrate User
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate user table: %w", err)
	}

	tokens := make(chan struct{}, capacity)
	for range capacity {
		tokens <- struct{}{}
	}
	return &Database{DB: db, Token: tokens}, nil
}
