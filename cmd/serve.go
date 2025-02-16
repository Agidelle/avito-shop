package cmd

import (
	"log/slog"
	"net/http"
	"os"

	"avito-shop/internal/config"
	urls "avito-shop/internal/http-server/handlers/url"
	mwJWT "avito-shop/internal/http-server/middleware"
	mwLogger "avito-shop/internal/http-server/middleware/logger"
	"avito-shop/internal/service/shop"
	"avito-shop/internal/service/shop/storage/mysql"

	"github.com/spf13/cobra"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case envLocal:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return logger
}

func protectedRoutes(r chi.Router, jwtSecret string, handlers *urls.Handlers) {
	r.Use(mwJWT.JWTMiddleware(jwtSecret))
	r.Get("/api/info", handlers.Info())
	r.Post("/api/sendCoin", handlers.SendCoin())
	r.Get("/api/buy/{item}", handlers.BuyItem())
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  `Start the HTTP server with the configured settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.MustLoad(cfgFile)
		if err != nil {
			return err
		}

		log := setupLogger(cfg.Env)
		log.Info("Start service", slog.String("env", cfg.Env))
		log.Debug("Debug messages are enabled")
		db, err := mysql.New(cfg.DB)
		if err != nil {
			log.Error("Failed to connect to database", "error", err)
			return err
		}

		r := chi.NewRouter()
		r.Use(middleware.Recoverer)
		r.Use(mwLogger.New(log))
		r.Use(middleware.Logger)
		r.Use(middleware.URLFormat)

		jwtSecret := cfg.AuthKey
		service := shop.NewService(db)
		handlers := urls.NewHandlers(db, service, log)
		r.Post("/api/auth", handlers.Auth(jwtSecret))

		r.Group(func(r chi.Router) {
			protectedRoutes(r, jwtSecret, handlers)
		})

		log.Info("Starting server", "address", cfg.Address)
		if err = http.ListenAndServe("localhost:8080", r); err != nil {
			log.Error("Failed to start server", "error", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
