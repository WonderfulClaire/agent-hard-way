package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestFanOutRespectsConcurrencyLimit(t *testing.T) {
	var active int32
	var peak int32
	worker := func(ctx context.Context, task Task) (string, error) {
		n := atomic.AddInt32(&active, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if n <= p || atomic.CompareAndSwapInt32(&peak, p, n) {
				break
			}
		}
		defer atomic.AddInt32(&active, -1)
		select {
		case <-time.After(20 * time.Millisecond):
			return task.ID, nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	tasks := []Task{{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}}
	results := fanOut(context.Background(), tasks, 2, worker)
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	if peak > 2 {
		t.Fatalf("peak concurrency=%d > 2", peak)
	}
}
