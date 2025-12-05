package cmd

import (
	"context"
	"errors"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"log"
	"minodl/config"
	"minodl/mdb"
	"minodl/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var dl = &cobra.Command{
	Use:   "dl",
	Short: "dl server",
	Long:  "this is a dl server",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()

		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("config error: %v", err)
		}

		_, err = mdb.InitGorm(cfg)
		if err != nil {
			log.Fatalf("db init: %v", err)
		}

		_ = mdb.InitRedis(cfg)
		// 初始化API服务
		r := router.MinoAPI()
		srv := &http.Server{
			Addr:    cfg.ListenAddr,
			Handler: r,
		}

		go func() {
			log.Printf("listening %s", cfg.ListenAddr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("listen: %v", err)
			}
		}()

		// Graceful shutdown
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(dl)
}
