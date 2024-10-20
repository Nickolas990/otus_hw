package hw06pipelineexecution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	//nolint:depguard
	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

var isFullTesting = true

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	if !isFullTesting {
		return
	}
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})
}

func TestExecutePipeline(t *testing.T) {
	t.Run("simple pipeline", testSimplePipeline)
	t.Run("pipeline with done signal", testPipelineWithDoneSignal)
	t.Run("pipeline with varying stage durations", testPipelineWithVaryingStageDurations)
	t.Run("pipeline with early done signal", testPipelineWithEarlyDoneSignal)
	t.Run("pipeline with large data set", testPipelineWithLargeDataSet)
}

func testSimplePipeline(t *testing.T) {
	in := make(chan interface{})
	done := make(chan interface{})

	stage1 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				out <- v.(int) * 2
			}
		}()
		return out
	}

	stage2 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				out <- v.(int) + 1
			}
		}()
		return out
	}

	out := ExecutePipeline(in, done, stage1, stage2)

	go func() {
		defer close(in)
		in <- 1
		in <- 2
		in <- 3
	}()

	results := make([]int, 0, 5)
	for v := range out {
		results = append(results, v.(int))
	}

	require.Equal(t, []int{3, 5, 7}, results)
}

func testPipelineWithDoneSignal(t *testing.T) {
	in := make(chan interface{})
	done := make(chan interface{})

	stage1 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				out <- v.(int) * 2
			}
		}()
		return out
	}

	stage2 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				out <- v.(int) + 1
			}
		}()
		return out
	}

	out := ExecutePipeline(in, done, stage1, stage2)

	go func() {
		defer close(in)
		in <- 1
		in <- 2
		in <- 3
	}()

	go func() {
		time.Sleep(1 * time.Second) // Give some time to process
		close(done)
	}()

	results := make([]int, 0, 5)
	for v := range out {
		results = append(results, v.(int))
	}

	// Since the pipeline should stop early, we may not get all results
	require.LessOrEqual(t, len(results), 3)
}

func testPipelineWithVaryingStageDurations(t *testing.T) {
	in := make(chan interface{})
	done := make(chan interface{})

	stage1 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				time.Sleep(50 * time.Millisecond)
				out <- v.(int) * 2
			}
		}()
		return out
	}

	stage2 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				time.Sleep(100 * time.Millisecond)
				out <- v.(int) + 1
			}
		}()
		return out
	}

	out := ExecutePipeline(in, done, stage1, stage2)

	go func() {
		defer close(in)
		for i := 1; i <= 5; i++ {
			in <- i
		}
	}()

	results := make([]int, 0, 5)
	for v := range out {
		results = append(results, v.(int))
	}

	require.Equal(t, []int{3, 5, 7, 9, 11}, results)
}

func testPipelineWithEarlyDoneSignal(t *testing.T) {
	in := make(chan interface{})
	done := make(chan interface{})

	stage1 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				time.Sleep(100 * time.Millisecond)
				out <- v.(int) * 2
			}
		}()
		return out
	}

	stage2 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				time.Sleep(100 * time.Millisecond)
				out <- v.(int) + 1
			}
		}()
		return out
	}

	out := ExecutePipeline(in, done, stage1, stage2)

	go func() {
		defer close(in)
		for i := 1; i <= 5; i++ {
			in <- i
		}
	}()

	go func() {
		time.Sleep(150 * time.Millisecond) // Close done before all stages complete
		close(done)
	}()

	results := make([]int, 0, 5)
	for v := range out {
		results = append(results, v.(int))
	}

	require.LessOrEqual(t, len(results), 2)
}

func testPipelineWithLargeDataSet(t *testing.T) {
	in := make(chan interface{})
	done := make(chan interface{})

	stage1 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				out <- v.(int) * 2
			}
		}()
		return out
	}

	stage2 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				out <- v.(int) + 1
			}
		}()
		return out
	}

	out := ExecutePipeline(in, done, stage1, stage2)

	go func() {
		defer close(in)
		for i := 1; i <= 1000; i++ {
			in <- i
		}
	}()

	results := make([]int, 0, 5)
	for v := range out {
		results = append(results, v.(int))
	}

	require.Equal(t, 1000, len(results))
	for i, v := range results {
		require.Equal(t, (i+1)*2+1, v)
	}
}
