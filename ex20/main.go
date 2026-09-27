package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Task struct {
	ID     string
	Prompt string
}

type Result struct {
	TaskID string
	Text   string
	Err    error
}

type Worker func(context.Context, Task) (string, error)

func fanOut(ctx context.Context, tasks []Task, maxParallel int, worker Worker) []Result {
	if maxParallel < 1 {
		maxParallel = 1
	}
	sem := make(chan struct{}, maxParallel)
	out := make(chan Result, len(tasks))
	var wg sync.WaitGroup

	for _, task := range tasks {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				out <- Result{TaskID: task.ID, Err: ctx.Err()}
				return
			}
			text, err := worker(ctx, task)
			out <- Result{TaskID: task.ID, Text: text, Err: err}
		}()
	}

	wg.Wait()
	close(out)
	results := make([]Result, 0, len(tasks))
	for result := range out {
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].TaskID < results[j].TaskID })
	return results
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tasks := []Task{{ID: "a", Prompt: "inspect A"}, {ID: "b", Prompt: "inspect B"}, {ID: "c", Prompt: "inspect C"}}
	results := fanOut(ctx, tasks, 2, func(_ context.Context, task Task) (string, error) {
		return "done: " + task.Prompt, nil
	})
	for _, result := range results {
		fmt.Printf("%s %s
", result.TaskID, result.Text)
	}
}
