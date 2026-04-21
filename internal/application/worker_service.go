package application

import (
	"context"
	"github.com/Guram-Gurych/LoadTest/internal/domain"
	"github.com/google/uuid"
	"log"
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

type metricsBroker interface {
	SendMetrics(ctx context.Context, metric domain.Metric) error
}

type workerService struct {
	mapTest map[uuid.UUID]context.CancelFunc
	mut     sync.RWMutex
	client  *http.Client
	ch      chan requestResult
	metrics metricsBroker
}

func NewWorkerService(metrics metricsBroker) workerService {
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

	return workerService{mapTest: mapTest, client: client, ch: ch, metrics: metrics}
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

type stats struct {
	total    int
	success  int
	errors   int
	totalDur time.Duration
}

func (w *workerService) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	current := make(map[uuid.UUID]*stats)

	for {
		select {
		case <-ctx.Done():
			return
		case res := <-w.ch:
			s, ok := current[res.TestID]
			if !ok {
				s = &stats{}
				current[res.TestID] = s
			}

			s.total++
			if res.Err != nil || res.StatusCode >= 400 {
				s.errors++
			} else {
				s.success++
			}

			s.totalDur += res.Duration

		case <-ticker.C:
			for testID, s := range current {
				if s.total == 0 {
					continue
				}

				metric := domain.Metric{
					TestID:        testID,
					BucketTime:    time.Now().Truncate(time.Second),
					RequestsCount: s.total,
					SuccessCount:  s.success,
					ErrorCount:    s.errors,
					AVGLatencyMs:  float64(s.totalDur.Milliseconds()) / float64(s.total),
				}

				err := w.metrics.SendMetrics(ctx, metric)
				if err != nil {
					log.Printf("failed to send metrics: %v", err)
				}
			}

			current = make(map[uuid.UUID]*stats)
		}
	}
}
