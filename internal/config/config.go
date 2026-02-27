package config

import (
	"authentication_service/internal/logger"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofor-little/env"
)

type Config struct {
	GRPCServer  *GRPCServer  `validate:"required"`
	HTTPServer  *HTTPServer  `validate:"required"`
	Database    *Database    `validate:"required"`
	Redis       *Redis       `validate:"required"`
	AuthService *AuthService `validate:"required"`
	Metrics     *Metrics     `validate:"required"`
	Tracing     *Tracing     `validate:"required"`
}

type GRPCServer struct {
	URL string `validate:"required,hostname_port"`
}

type HTTPServer struct {
	URL string `validate:"required,hostname_port"`
}

type Database struct {
	DSN                 string        `validate:"required"`
	Driver              string        `validate:"required,oneof=postgres sqlite"`
	PoolMaxIdleConns    int           `validate:"gte=0"`
	PoolMaxOpenConns    int           `validate:"gte=0"`
	PoolConnMaxLifetime time.Duration `validate:"gte=0"`
	Logger              logger.Logger
}

type Redis struct {
	Addr         string        `validate:"required"`
	Password     string        `validate:"required"`
	DB           int           `validate:"gte=0"`
	DialTimeout  time.Duration `validate:"gte=0"`
	ReadTimeout  time.Duration `validate:"gte=0"`
	WriteTimeout time.Duration `validate:"gte=0"`
	PoolSize     int           `validate:"gte=0"`
	MinIdleConns int           `validate:"gte=0"`
}

type AuthService struct {
	SecretKey []byte        `validate:"required"`
	TokenTTL  time.Duration `validate:"required"`
}

type Metrics struct {
	EnableDefaultMetrics bool
}

type Tracing struct {
	ServiceName  string `validate:"required"`
	CollectorURL string `validate:"required,hostname_port"`
}

type LoaderOpts struct {
	EnvPath   string
	EnvLoader func(string) error
	Logger    logger.Logger
}

func NewConfig() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return NewConfigWithOpts(LoaderOpts{
		EnvPath: getEnv("ENV_CONFIG", path.Join(cwd, ".env")),
	})
}

func NewConfigWithOpts(opts LoaderOpts) (*Config, error) {
	log := opts.Logger
	if log == nil {
		log = logger.NewLogger("info", os.Stderr)
	}

	envLoader := opts.EnvLoader
	if envLoader == nil {
		envLoader = func(path string) error {
			_, err := os.Stat(path)
			if err != nil {
				return err
			}
			return env.Load(path)
		}
	}

	if err := envLoader(opts.EnvPath); err == nil {
		log.Info("Loaded environment variables from " + opts.EnvPath)
	} else {
		log.Info("failed to load .env file, using system environment variables")
	}

	cfg := &Config{
		GRPCServer: &GRPCServer{
			URL: getEnv("GPRC_SERVER_URL", ":8080"),
		},
		HTTPServer: &HTTPServer{
			URL: getEnv("HTTP_SERVER_URL", ":8081"),
		},
		Database: &Database{
			DSN:                 getEnv("DATABASE_URL", ""),
			Driver:              getEnv("DATABASE_DRIVER", "postgres"),
			PoolMaxIdleConns:    getEnvInt("DATABASE_POOL_MAX_IDLE", 10),
			PoolMaxOpenConns:    getEnvInt("DATABASE_POOL_MAX_OPEN", 100),
			PoolConnMaxLifetime: getEnvDuration("DATABASE_POOL_MAX_LIFETIME", time.Hour),
		},
		Redis: &Redis{
			Addr:         getEnv("REDIS_ADDR", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", "default"),
			DB:           getEnvInt("REDIS_DB", 0),
			DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 20),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNECTIONS", 5),
		},
		AuthService: &AuthService{
			SecretKey: []byte(getEnv("JWT_SECRET", "super_secret_key")),
			TokenTTL:  getEnvDuration("TOKEN_TTL", 15*time.Minute),
		},
		Metrics: &Metrics{
			EnableDefaultMetrics: getEnvBool("METRICS_ENABLE_DEFAULT_METRICS", false),
		},
		Tracing: &Tracing{
			ServiceName:  getEnv("TRACING_SERVICE_NAME", "auth-service"),
			CollectorURL: getEnv("TRACING_COLLECTOR_URL", "localhost:4318"),
		},
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, err := strconv.ParseInt(os.Getenv(key), 10, 32); err == nil {
		return int(value)
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return value
	}
	return defaultValue
}
