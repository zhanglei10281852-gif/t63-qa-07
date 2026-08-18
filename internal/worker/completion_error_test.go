package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"sanitation-operations/internal/clock"
	"sanitation-operations/internal/repository"
)

var completionPersistenceError = errors.New("completion persistence failed")

type completionFailureStore struct{ repository.Store }

func (completionFailureStore) ClaimDue(context.Context, time.Time, int) ([]repository.OutboxJob, error) {
	return []repository.OutboxJob{{ID: "job-1", Type: "trip.completed", Payload: []byte(`{}`)}}, nil
}
func (completionFailureStore) MarkJobDone(context.Context, string) error {
	return completionPersistenceError
}

func TestRunnerReportsCompletionPersistenceFailure(t *testing.T) {
	now := time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	runner := Runner{
		Store: completionFailureStore{}, Clock: clock.Fixed{Current: now}, Interval: time.Hour,
		Handler: HandlerFunc(func(context.Context, repository.OutboxJob) error { return nil }),
	}
	err := runner.Run(ctx)
	if !errors.Is(err, completionPersistenceError) {
		t.Fatalf("runner error=%v", err)
	}
}
