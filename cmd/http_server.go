package cmd

import (
	"fmt"
	"log"
	"time"

	// "github.com/frahmantamala/jadiles/internal/transport"
	"github.com/spf13/cobra"
)

const (
	defautlWaitShutdownDuration = 10 * time.Second
)

var httpServerCmd = &cobra.Command{
	RunE:  runHTTPServer,
	Use:   "http_server",
	Short: "to run http server",
}

func runHTTPServer(_ *cobra.Command, _ []string) error {
	cfg, err := loadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	initLogger(cfg.Name, cfg)

	dbConn, err := initDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		sqlDB, err := dbConn.DB()
		if err != nil {
			fmt.Errorf("failed to get sql.DB for cleanup: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			fmt.Errorf("failed to close database connection: %v", err)
		}
	}()

	redisConn, err := initRedisPool(cfg.Redis, true)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = redisConn.Close() }()

	goRedisClient, err := initGoRedis(cfg.Redis)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = goRedisClient.Close() }()

	// server, err := transport.NewRESTServer(
	// 	dbConn,
	// 	goRedisClient,
	// 	redisConn,
	// 	cfg,
	// )
	// if err != nil {
	// 	log.Fatalf("failed to initiate http server: %s", err)
	// }

	// errCh := make(chan error, 1)
	// signalCh := make(chan os.Signal, 1)
	// signal.Notify(signalCh, os.Interrupt)

	// go func() {
	// 	log.Println("http server is running")
	// 	if err := server.Start(); err != nil {
	// 		errCh <- fmt.Errorf("failed to run http server: %w", err)
	// 	}
	// }()

	// go func() {
	// 	<-signalCh
	// 	signal.Reset(os.Interrupt)
	// 	errCh <- fmt.Errorf("interrupted") //nolint:goerr113
	// }()

	// <-errCh

	// shutdownCtx, cancel := context.WithTimeout(context.Background(), defautlWaitShutdownDuration)
	// defer cancel()

	// wg := new(sync.WaitGroup)

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	if err := server.Stop(shutdownCtx); err != nil {
	// 		log.Println(err)
	// 	}
	// }()

	// wg.Wait()
	return nil
}
