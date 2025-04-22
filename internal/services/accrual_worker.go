package services

import (
	"context"
	"github.com/aseptimu/internal/repository"
	"github.com/aseptimu/internal/utils"
	"runtime"
	"sync"
	"time"
)

type accrualResult struct {
	Number string
	Err    error
}

type AccrualWorker struct {
	repo         *repository.OrderRepository
	orderService *OrderService
	period       time.Duration
	workers      int
}

func NewAccrualWorker(repo *repository.OrderRepository, orderService *OrderService, period time.Duration) *AccrualWorker {
	return &AccrualWorker{repo: repo, orderService: orderService, period: period, workers: runtime.NumCPU()}
}

func (w *AccrualWorker) Start(ctx context.Context) {
	jobs := make(chan string)
	results := make(chan accrualResult)

	var wg sync.WaitGroup
	wg.Add(w.workers)
	for i := 0; i < w.workers; i++ {
		go func(id int) {
			defer wg.Done()
			for number := range jobs {
				err := w.orderService.fetchAndStoreAccrual(ctx, number)
				results <- accrualResult{number, err}
			}
		}(i)
	}

	go func() {
		for res := range results {
			if res.Err != nil {
				utils.LogInfo(ctx, "[accrual-worker]: ", res.Number, res.Err)
			}
		}
	}()

	ticker := time.NewTicker(w.period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			close(results)
			return
		case <-ticker.C:
			orders, err := w.repo.GetOrdersByStatus(ctx, []string{"NEW", "PROCESSING"})
			if err != nil {
				utils.LogWithError(ctx, "[accrual-worker] failed to list pending orders: ", err)
				continue
			}
			for _, o := range orders {
				select {
				case jobs <- o.Number:
				case <-ctx.Done():
					close(jobs)
					wg.Wait()
					close(results)
					return
				}
			}
		}
	}

}
