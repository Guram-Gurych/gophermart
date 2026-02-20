package main

import (
	"context"
	"database/sql"
	"github.com/Guram-Gurych/gophermart.git/internal/config"
	"github.com/Guram-Gurych/gophermart.git/internal/handlers"
	"github.com/Guram-Gurych/gophermart.git/internal/middleware"
	"github.com/Guram-Gurych/gophermart.git/internal/repository"
	"github.com/Guram-Gurych/gophermart.git/internal/services"
	"github.com/Guram-Gurych/gophermart.git/migrations"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
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

	db, err := sql.Open("pgx", cnf.DBAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := migrations.RunMigrations(initCtx, db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	rep := repository.NewRepository(db)

	authHandler := handlers.NewAuthHandler(services.NewAuthService(rep, cnf.SecretKey))
	orderHandler := handlers.NewOrderHandler(services.NewOrderService(rep))
	balanceHandler := handlers.NewBalanceHandler(services.NewBalanceService(rep))

	r := chi.NewRouter()
	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cnf.SecretKey))
		r.Post("/api/user/orders", orderHandler.SetOrder)
		r.Get("/api/user/orders", orderHandler.GetOrders)
		r.Get("/api/user/balance", balanceHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", balanceHandler.SetWithdraw)
		r.Get("/api/user/withdrawals", balanceHandler.GetWithdraws)
	})

	client := http.Client{Timeout: 10 * time.Second}
	worker := services.NewWorker(rep, &client, cnf.AccrualAddress)

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

	log.Println("Server started")

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	worker.Wait()
	log.Println("All workers stopped")
}
