package main

import (
	"context"
	"database/sql"
	"github.com/Guram-Gurych/gophermart.git/internal/auth"
	"github.com/Guram-Gurych/gophermart.git/internal/balance"
	"github.com/Guram-Gurych/gophermart.git/internal/orders"
	authMiddleware "github.com/Guram-Gurych/gophermart.git/internal/platform/auth"
	"github.com/Guram-Gurych/gophermart.git/internal/platform/config"
	platformLogger "github.com/Guram-Gurych/gophermart.git/internal/platform/logger"
	"github.com/Guram-Gurych/gophermart.git/internal/platform/storage"
	"github.com/Guram-Gurych/gophermart.git/internal/worker"
	"github.com/Guram-Gurych/gophermart.git/migrations"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cnf, err := config.InitConfig()
	if err != nil {
		log.Fatalf("error when initializing the config: %v", err)
	}

	opts := &slog.HandlerOptions{
		Level: config.ParseLogLevel(cnf.LogLevel),
	}

	baseHandler := slog.NewJSONHandler(os.Stdout, opts)
	finalHandler := &platformLogger.ReqIDHandler{Handler: baseHandler}

	logger := slog.New(finalHandler)
	slog.SetDefault(logger)

	logger.Info("server is starting",
		slog.String("address", cnf.ServerAddress),
		slog.String("db", cnf.DBAddress))

	db, err := sql.Open("pgx", cnf.DBAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		logger.Error("starting the database returned an error", slog.Any("error", err))
	}

	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = migrations.RunMigrations(initCtx, db, logger); err != nil {
		logger.Error("failed to run migrations", slog.Any("error", err))
	}

	baseRepo := storage.NewRepository(db)

	authRepo := auth.NewRepository(baseRepo)
	authService := auth.NewService(authRepo, cnf.SecretKey)
	authHandler := auth.NewAuthHandler(authService, logger)

	orderRepo := orders.NewRepository(baseRepo)
	orderService := orders.NewService(orderRepo)
	orderHandler := orders.NewOrderHandler(orderService, logger)

	balanceRepo := balance.NewRepository(baseRepo)
	balanceService := balance.NewService(balanceRepo)
	balanceHandler := balance.NewBalanceHandler(balanceService, logger)

	r := chi.NewRouter()

	r.Use(platformLogger.RequestMiddleware())
	r.Use(platformLogger.LoggerMiddleware(logger))

	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.AuthMiddleware(cnf.SecretKey))
		r.Post("/api/user/orders", orderHandler.SetOrder)
		r.Get("/api/user/orders", orderHandler.GetOrders)
		r.Get("/api/user/balance", balanceHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", balanceHandler.SetWithdraw)
		r.Get("/api/user/withdrawals", balanceHandler.GetWithdraws)
	})

	client := http.Client{Timeout: 10 * time.Second}
	worker := worker.NewWorker(orderRepo, &client, logger, cnf.AccrualAddress)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go worker.StartScheduler(ctx)

	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		go worker.StartWorker(ctx)
	}

	srv := &http.Server{
		Addr:    cnf.ServerAddress,
		Handler: r,
	}

	logger.Info("Server started")

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error when starting the server", slog.Any("error", err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
	}

	worker.Wait()
	logger.Info("All workers stopped")
}
