package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Guram-Gurych/LoadTest/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"net/url"
)

type service interface {
	Create(ctx context.Context, test domain.Test) error
	Get(ctx context.Context, id uuid.UUID) (domain.Test, error)
	Cancel(ctx context.Context, id uuid.UUID) error
}

type testHandler struct {
	serv service
}

func NewTestHandler(serv service) testHandler {
	return testHandler{serv: serv}
}

func (hand *testHandler) Create(w http.ResponseWriter, r *http.Request) {
	var requestTest createTestRequest
	if err := json.NewDecoder(r.Body).Decode(&requestTest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestTest.RPS < 0 || requestTest.DurationSeconds < 0 {
		http.Error(w, "rps и/или duration_seconds не могут быть меньше 0", http.StatusBadRequest)
		return
	}

	if _, err := url.Parse(requestTest.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	test := domain.Test{
		ID:              id,
		Status:          domain.StatusRunning,
		RPS:             requestTest.RPS,
		DurationSeconds: requestTest.DurationSeconds,
		URL:             requestTest.URL,
	}

	if err := hand.serv.Create(r.Context(), test); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(test)
}

func (hand *testHandler) Get(w http.ResponseWriter, r *http.Request) {
	IDStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(IDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	test, err := hand.serv.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "apllication/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(test)
}

type createTestRequest struct {
	URL             string `json:"url"`
	RPS             int    `json:"rps"`
	DurationSeconds int    `json:"duration_seconds"`
}

func (hand *testHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	IDStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(IDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := hand.serv.Cancel(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
