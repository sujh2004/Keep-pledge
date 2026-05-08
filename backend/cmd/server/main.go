package main

import (
	"context"
	"log"
	"os"
	"time"

	"keep-pledge/backend/internal/config"
	"keep-pledge/backend/internal/handler"
	"keep-pledge/backend/internal/middleware"
	"keep-pledge/backend/internal/model"
	"keep-pledge/backend/internal/repository"
	"keep-pledge/backend/internal/service"
	authpkg "keep-pledge/backend/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	if err := os.MkdirAll(cfg.Upload.Dir, 0750); err != nil {
		log.Fatalf("create upload dir: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("redis unavailable, leaderboard will fallback to database: %v", err)
		redisClient = nil
	}

	repo := repository.New(db, redisClient)
	jwt := authpkg.NewManager(cfg.JWT.Secret)
	services := service.NewServices(repo, cfg, jwt)
	if err := services.Achievement.Seed(); err != nil {
		log.Fatalf("seed achievements: %v", err)
	}
	startDailyStreakSweep(services)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORS(), middleware.RateLimit(60))
	router.Static(cfg.Upload.PublicPath, cfg.Upload.Dir)

	api := router.Group("/api/v1")
	handler.New(services, cfg, jwt).RegisterRoutes(api)

	log.Printf("KeepPledge backend listening on %s", cfg.Server.Address)
	if err := router.Run(cfg.Server.Address); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func startDailyStreakSweep(services *service.Services) {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		lastRun := ""
		for {
			now := time.Now()
			runKey := now.Format("2006-01-02")
			if now.Hour() == 1 && lastRun != runKey {
				affected, err := services.Task.RunMissedCheckinSweep(now)
				if err != nil {
					log.Printf("daily streak sweep failed: %v", err)
				} else {
					log.Printf("daily streak sweep finished, affected=%d", affected)
					lastRun = runKey
				}
			}
			<-ticker.C
		}
	}()
}
