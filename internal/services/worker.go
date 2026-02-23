package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type AccrualResponse struct {
	Order   string              `json:"order"`
	Status  models.OrderStatus  `json:"status"`
	Accrual *models.JSONBalance `json:"accrual,omitempty"`
}

type Worker struct {
	repository  WorkerRepository
	client      *http.Client
	logger      *slog.Logger
	address     string
	taskChan    chan string
	stopTimeout chan time.Duration
	wg          sync.WaitGroup
}

func NewWorker(rep WorkerRepository, client *http.Client, logger *slog.Logger, addr string) *Worker {
	task := make(chan string)
	stop := make(chan time.Duration)

	return &Worker{
		repository:  rep,
		client:      client,
		logger:      logger,
		address:     addr,
		taskChan:    task,
		stopTimeout: stop,
	}
}

func (w *Worker) StartScheduler(ctx context.Context) {
	w.logger.Info("scheduler started")
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("scheduler stopping")
			return
		case timeout := <-w.stopTimeout:
			w.logger.Info("scheduler timeout")
			pauseTimer := time.NewTimer(timeout)
			select {
			case <-pauseTimer.C:
			case <-ctx.Done():
				pauseTimer.Stop()
				return
			}
		default:
			tasks, err := w.repository.GetPendingOrders(ctx, 10)
			if err != nil {
				w.logger.Error("Error when receiving orders from the database", slog.Any("error", err))
				continue
			}

			if len(tasks) == 0 {
				w.logger.Debug("no pending orders found, sleeping")
				pauseTimer := time.NewTimer(time.Second * 5)
				select {
				case <-pauseTimer.C:
				case <-ctx.Done():
					pauseTimer.Stop()
					return
				}
			}

			if err = w.repository.UpdateOrdersStatus(ctx, tasks, models.StatusProcessing); err != nil {
				w.logger.Error("Error when changing orders statuses in the database", slog.Any("error", err))
				continue
			}

			w.logger.Info("found orders for processing", slog.Int("count", len(tasks)))
			for _, task := range tasks {
				select {
				case w.taskChan <- task:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (w *Worker) StartWorker(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case orderNumber, ok := <-w.taskChan:
			if !ok {
				return
			}

			w.processOrder(ctx, orderNumber)
		}
	}
}

func (w *Worker) processOrder(ctx context.Context, number string) {
	resp, err := w.client.Get(fmt.Sprintf("%s/api/orders/%s", w.address, number))
	if err != nil {
		w.logger.Error("error when executing an order request", slog.String("number", number), slog.Any("error", err))
		return
	}
	defer resp.Body.Close()

	var acrrual AccrualResponse
	switch resp.StatusCode {
	case http.StatusOK:
		if err = json.NewDecoder(resp.Body).Decode(&acrrual); err != nil {
			w.logger.Error("Error when deserializing order in JSON", slog.String("number", number), slog.Any("error", err))
			return
		}

		var balance models.JSONBalance
		if acrrual.Accrual != nil {
			balance = *acrrual.Accrual
		}

		if err = w.repository.UpdateOrder(ctx, acrrual.Order, acrrual.Status, balance); err != nil {
			w.logger.Error("Error when updating the order status", slog.String("number", number), slog.Any("error", err))
			return
		}
	case http.StatusNoContent:
		return
	case http.StatusTooManyRequests:
		timeout, err := strconv.Atoi(resp.Header.Get("Retry-After"))
		if err != nil {
			w.logger.Error("Error when converting timeout to a number", slog.Any("error", err))
			return
		}

		if timeout == 0 {
			timeout = 60
		}

		select {
		case w.stopTimeout <- time.Duration(timeout) * time.Second:
			select {
			case w.taskChan <- number:
			default:
				w.repository.UpdateOrdersStatus(ctx, []string{number}, models.StatusREGISTERED)
			}
		case <-ctx.Done():
			return
		default:
			return
		}
	case http.StatusInternalServerError:
		return
	}
}

func (w *Worker) Wait() {
	w.wg.Wait()
}
