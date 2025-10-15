package cmd

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/database"
	"github.com/frahmantamala/jadiles/pkg/logger"
	"github.com/gomodule/redigo/redis"
	slogmulti "github.com/samber/slog-multi"
	"gorm.io/gorm"

	goRedis "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	goredistrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/redis/go-redis.v9"
)

var (
	rootCmd = &cobra.Command{
		Use:   "server",
		Short: "Jadi Les CMD",
		Long:  "A command-line tool for managing the From De Hands application",
	}

	configPath string = "."
	config     internal.Config
	appLogger  *logger.Logger
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(httpServerCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(versionCmd)
}

// initConfig loads the configuration
func initConfig() {
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	config = cfg
}

func initLogger(serviceName string, cfg internal.Config) {
	formatterHandler := logger.SlogOption{
		Resource: map[string]string{
			"service.name":        serviceName,
			"service.ns":          cfg.Namespace,
			"service.instance_id": cfg.InstanceID,
			"service.version":     version,
			"service.env":         cfg.Env,
		},
		ContextExtractor:   internal.SlogContextExtractor,
		AttributeFormatter: internal.LogAttributeFmter,
		Writer:             os.Stdout,
		Leveler:            cfg.Logger.ParseSlogLevel(),
	}.NewHandler()

	slogger := slog.New(
		slogmulti.Fanout(formatterHandler),
	)
	slog.SetDefault(slogger)
}

func loadConfig(path string) (internal.Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("env")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return internal.Config{}, err
	}

	var cfg internal.Config
	if err = viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func initDatabase(cfg internal.Config) (*gorm.DB, error) {
	db, err := database.NewGormDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}

func initGoRedis(cfg internal.RedisConfig) (goRedis.UniversalClient, error) {
	u, err := goRedis.ParseURL(cfg.Source)
	if err != nil {
		return nil, err
	}

	return goredistrace.NewClient(&goRedis.Options{
		Addr:            u.Addr,
		ClientName:      u.ClientName,
		Username:        u.Username,
		Password:        u.Password,
		DB:              u.DB,
		MaxIdleConns:    cfg.PoolMaxIdle,
		PoolSize:        cfg.PoolMaxActive,
		ConnMaxIdleTime: cfg.PoolIdleTimeout,
		ConnMaxLifetime: cfg.PoolMaxConnLifetime,
	},
		goredistrace.WithServiceName("fromdehands-redis"),
		goredistrace.WithErrorCheck(func(err error) bool {
			return !errors.Is(err, goRedis.Nil)
		}),
	), nil
}

func initRedisPool(cfg internal.RedisConfig, isEnableTrace bool) (*redis.Pool, error) {
	dialFn := func() (redis.Conn, error) {
		return redis.DialURL(cfg.Source)
	}

	r := &redis.Pool{
		MaxActive:       cfg.PoolMaxActive,
		MaxIdle:         cfg.PoolMaxIdle,
		IdleTimeout:     cfg.PoolIdleTimeout,
		MaxConnLifetime: cfg.PoolMaxConnLifetime,
		Wait:            true,
		Dial:            dialFn,
	}

	_, err := r.Get().Do("PING")
	return r, err
}
