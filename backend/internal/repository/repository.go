package repository

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func New(db *gorm.DB, redisClient *redis.Client) *Repository {
	return &Repository{DB: db, Redis: redisClient}
}
