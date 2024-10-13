package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("eventually complete", func(t *testing.T) {
		var wg sync.WaitGroup

		tasks := []Task{
			func() error {
				<-time.After(time.Second)
				return nil
			},
			func() error {
				<-time.After(time.Second * 2)
				return nil
			},
		}

		wg.Add(len(tasks))
		done := make(chan struct{})
		for _, task := range tasks {
			go func(task Task) {
				defer wg.Done()
				err := task()
				if err != nil {
					return
				}
			}(task)
		}
		go func() {
			wg.Wait()
			close(done)
		}()
		go func() {
			err := Run(tasks, 2, 1)
			if err != nil {
				return
			}
		}()
		require.Eventually(t, func() bool {
			select {
			case <-done:
				return true
			default:
				return false
			}
		}, time.Second*5, time.Millisecond*100)
	})

	t.Run("all tasks completed", func(t *testing.T) {
		tasks := make([]Task, 10)
		for i := range tasks {
			tasks[i] = func() error {
				return nil
			}
		}
		err := Run(tasks, 5, 1)
		require.NoError(t, err)
	})

	t.Run("all tasks failed", func(t *testing.T) {
		tasks := make([]Task, 10)
		for i := range tasks {
			tasks[i] = func() error {
				return errors.New("test error")
			}
		}
		err := Run(tasks, 5, 1)
		require.Error(t, err)
	})

	t.Run("first m tasks failed", func(t *testing.T) {
		tasks := make([]Task, 10)
		for i := range tasks {
			if i < 3 {
				tasks[i] = func() error {
					return errors.New("test error")
				}
			} else {
				tasks[i] = func() error {
					return nil
				}
			}
		}
		err := Run(tasks, 5, 3)
		require.Error(t, err)
	})

	t.Run("errors limit is zero or negative", func(t *testing.T) {
		tasks := []Task{
			func() error { return nil },
			func() error { return errors.New("error") },
		}

		err := Run(tasks, 2, 0)
		require.ErrorIs(t, err, ErrErrorsLimitExceeded)

		err = Run(tasks, 2, -1)
		require.ErrorIs(t, err, ErrErrorsLimitExceeded)
	})
}
