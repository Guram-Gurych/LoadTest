package application

import (
	"context"
	"github.com/Guram-Gurych/LoadTest/internal/domain"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"
)

type requestResult struct {
	TestID     uuid.UUID
	Duration   time.Duration
	StatusCode int
	Err        error
}

type workerService struct {
	mapTest map[uuid.UUID]context.CancelFunc
	mut     sync.RWMutex
	client  *http.Client
	ch      chan requestResult
}

func NewWorkerService() workerService {
	mapTest := make(map[uuid.UUID]context.CancelFunc)
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 10 * time.Second,
	}
	ch := make(chan requestResult)

	return workerService{mapTest: mapTest, client: client, ch: ch}
}

func (w *workerService) StartTest(parentCtx context.Context, test domain.Test) {
	ctx, cancel := context.WithCancel(parentCtx)

	w.mut.Lock()
	w.mapTest[test.ID] = cancel
	w.mut.Unlock()

	go w.runLoad(ctx, test)
}

func (w *workerService) runLoad(ctx context.Context, test domain.Test) {
	ticker := time.NewTicker(time.Second / time.Duration(test.RPS))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go w.fire(ctx, test.ID, test.URL)
		}
	}
}

func (w *workerService) fire(ctx context.Context, testID uuid.UUID, url string) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	start := time.Now()

	resp, err := w.client.Do(req)

	duration := time.Since(start)

	result := requestResult{
		TestID:   testID,
		Duration: duration,
		Err:      err,
	}

	if err == nil {
		result.StatusCode = resp.StatusCode
		resp.Body.Close()
	}

	select {
	case w.ch <- result:
	default:
	}
}

func (w *workerService) StopTest(ctx context.Context, id uuid.UUID) error {
	w.mut.Lock()
	defer w.mut.Unlock()

	if cancel, ok := w.mapTest[id]; ok {
		cancel()
		delete(w.mapTest, id)
		return nil
	}

	return domain.ErrNotFound
}
