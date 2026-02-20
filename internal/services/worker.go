package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Guram-Gurych/gophermart.git/internal/models"
	"github.com/Guram-Gurych/gophermart.git/internal/repository"
	"log"
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
	repository  repository.Repository
	client      *http.Client
	address     string
	taskChan    chan string
	stopTimeout chan time.Duration
	wg          sync.WaitGroup
}

func NewWorker(rep repository.Repository, client *http.Client, addr string) *Worker {
	task := make(chan string)
	stop := make(chan time.Duration)

	return &Worker{
		repository:  rep,
		client:      client,
		address:     addr,
		taskChan:    task,
		stopTimeout: stop,
	}
}

func (w *Worker) StartScheduler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case timeout := <-w.stopTimeout:
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
				log.Printf("Error when receiving orders from the database: %v", err)
				continue
			}

			if len(tasks) == 0 {
				pauseTimer := time.NewTimer(time.Second * 5)
				select {
				case <-pauseTimer.C:
				case <-ctx.Done():
					pauseTimer.Stop()
					return
				}
			}

			if err = w.repository.UpdateOrdersStatus(ctx, tasks, models.StatusProcessing); err != nil {
				log.Printf("Error when changing orders statuses in the database: %v", err)
				continue
			}

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
		log.Printf("error when executing an order request %s: %v", number, err)
		return
	}
	defer resp.Body.Close()

	var acrrual AccrualResponse
	switch resp.StatusCode {
	case http.StatusOK:
		if err = json.NewDecoder(resp.Body).Decode(&acrrual); err != nil {
			log.Printf("Error when deserializing order %s in JSON: %v", number, err)
			return
		}
		
		var balance models.JSONBalance
		if acrrual.Accrual != nil {
			balance = *acrrual.Accrual
		}

		if err = w.repository.UpdateOrder(ctx, acrrual.Order, acrrual.Status, balance); err != nil {
			log.Printf("Error when updating the order status %s: %v", number, err)
			return
		}
	case http.StatusNoContent:
		return
	case http.StatusTooManyRequests:
		timeout, err := strconv.Atoi(resp.Header.Get("Retry-After"))
		if err != nil {
			log.Printf("error when converting timeout to a number: %v", err)
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
