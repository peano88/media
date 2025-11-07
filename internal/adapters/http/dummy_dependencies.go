package http

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/peano88/medias/internal/domain"
)

type DummyTagCreator struct{}

func (dtc DummyTagCreator) Execute(_ context.Context, tag domain.Tag) (domain.Tag, error) {
	return domain.Tag{
		ID:          uuid.New(),
		Name:        tag.Name,
		Description: tag.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

type DummyMetricsForwarder struct{}

func (dmf DummyMetricsForwarder) AddRequestHit(pattern string, code int, duration time.Duration) error {
	fmt.Printf("DummyMetricsForwarder: pattern=%s, code=%d, duration=%s\n", pattern, code, duration)
	return nil
}
