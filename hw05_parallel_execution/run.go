package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}

	taskChan := make(chan Task)
	doneChan := make(chan struct{})
	var wg sync.WaitGroup
	var errs int
	var mu sync.Mutex
	var once sync.Once

	handleError := func() {
		mu.Lock()
		defer mu.Unlock()
		errs++
		if errs >= m {
			once.Do(func() {
				close(doneChan)
			})
		}
	}

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-doneChan:
				return
			case task, ok := <-taskChan:
				if !ok {
					return
				}
				if err := task(); err != nil {
					handleError()
				}
			}
		}
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		defer close(taskChan)
		for _, task := range tasks {
			select {
			case <-doneChan:
				return
			case taskChan <- task:
			}
		}
	}()

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if errs >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
