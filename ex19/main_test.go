package main

import (
	"context"
	"testing"
)

func TestDelegateKeepsTaskIdentity(t *testing.T) {
	worker := FuncWorker(func(_ context.Context, task Task) (Result, error) {
		return Result{TaskID: task.ID, Text: task.Prompt}, nil
	})
	got, err := delegate(context.Background(), worker, Task{ID: "t1", Prompt: "inspect logs"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TaskID != "t1" || got.Text != "inspect logs" {
		t.Fatalf("unexpected result: %#v", got)
	}
}
