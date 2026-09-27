package main

import (
	"context"
	"fmt"
)

type Task struct {
	ID      string
	Prompt  string
	Allowed []string
}

type Result struct {
	TaskID string
	Text   string
}

type Worker interface {
	Run(context.Context, Task) (Result, error)
}

type FuncWorker func(context.Context, Task) (Result, error)

func (f FuncWorker) Run(ctx context.Context, task Task) (Result, error) {
	return f(ctx, task)
}

func delegate(ctx context.Context, worker Worker, task Task) (Result, error) {
	if task.ID == "" || task.Prompt == "" {
		return Result{}, fmt.Errorf("task id and prompt are required")
	}
	return worker.Run(ctx, task)
}

func main() {
	worker := FuncWorker(func(_ context.Context, task Task) (Result, error) {
		return Result{TaskID: task.ID, Text: "subagent completed scoped task: " + task.Prompt}, nil
	})
	result, err := delegate(context.Background(), worker, Task{
		ID: "research-1", Prompt: "summarize one paper", Allowed: []string{"read_file"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Text)
}
