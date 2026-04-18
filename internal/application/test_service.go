package application

import (
	"context"
	"fmt"
	"github.com/Guram-Gurych/LoadTest/internal/domain"
	"github.com/google/uuid"
)

type storage interface {
	Create(ctx context.Context, test domain.Test) error
	Get(ctx context.Context, id uuid.UUID) (domain.Test, error)
	Cancel(ctx context.Context, id uuid.UUID) error
}

type broker interface {
	SendStart(ctx context.Context, test domain.Test) error
	SendStop(ctx context.Context, id uuid.UUID) error
}

type testService struct {
	rep  storage
	brok broker
}

func NewTestService(rep storage, brok broker) *testService {
	return &testService{rep: rep, brok: brok}
}

func (ts *testService) Create(ctx context.Context, test domain.Test) error {
	if err := ts.rep.Create(ctx, test); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	if err := ts.brok.SendStart(ctx, test); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return nil
}

func (ts *testService) Get(ctx context.Context, id uuid.UUID) (domain.Test, error) {
	return ts.rep.Get(ctx, id)
}

func (ts *testService) Cancel(ctx context.Context, id uuid.UUID) error {
	if err := ts.rep.Cancel(ctx, id); err != nil {
		return err
	}

	return ts.brok.SendStop(ctx, id)
}
