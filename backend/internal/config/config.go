package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Upload   UploadConfig   `yaml:"upload"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Charset  string `yaml:"charset"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret          string `yaml:"secret"`
	AccessTokenTTL  string `yaml:"access_token_ttl"`
	RefreshTokenTTL string `yaml:"refresh_token_ttl"`
}

type UploadConfig struct {
	Dir        string `yaml:"dir"`
	PublicPath string `yaml:"public_path"`
	MaxSizeMB  int64  `yaml:"max_size_mb"`
}

func Load(path string) (*Config, error) {
	cfg := defaultConfig()
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			applyEnv(cfg)
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, err
	}
	applyEnv(cfg)
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{Address: ":8080"},
		Database: DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     3306,
			Username: "keep_pledge",
			Password: "keep_pledge",
			Name:     "keep_pledge",
			Charset:  "utf8mb4",
		},
		Redis: RedisConfig{
			Address: "127.0.0.1:6379",
			DB:      0,
		},
		JWT: JWTConfig{
			Secret:          "change-me-in-production",
			AccessTokenTTL:  "2h",
			RefreshTokenTTL: "168h",
		},
		Upload: UploadConfig{
			Dir:        "uploads",
			PublicPath: "/uploads",
			MaxSizeMB:  5,
		},
	}
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("SERVER_ADDRESS"); v != "" {
		cfg.Server.Address = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.Redis.Address = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
}

func (d DatabaseConfig) DSN() string {
	charset := d.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.Name, charset)
}

func (j JWTConfig) AccessTTL() time.Duration {
	duration, err := time.ParseDuration(j.AccessTokenTTL)
	if err != nil {
		return 2 * time.Hour
	}
	return duration
}

func (j JWTConfig) RefreshTTL() time.Duration {
	duration, err := time.ParseDuration(j.RefreshTokenTTL)
	if err != nil {
		return 7 * 24 * time.Hour
	}
	return duration
}
