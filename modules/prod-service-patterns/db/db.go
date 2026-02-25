package db

import (
	"context"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	DB    *gorm.DB
	Token chan struct{}
}

func NewDatabase(ctx context.Context, capacity int) (*Database, error) {
	db, err := gorm.Open(sqlite.Open("prod-service-patterns.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database %w", err)
	}

	tokens := make(chan struct{}, capacity)
	for range capacity {
		tokens <- struct{}{}
	}
	return &Database{DB: db, Token: tokens}, nil
}
