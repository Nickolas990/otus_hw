package hw06pipelineexecution

import (
	"log"
	"sync"
)

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	for _, stage := range stages {
		in = stageWrapper(in, done, stage)
	}
	return in
}

func stageWrapper(in In, done In, stage Stage) Out {
	out := make(Bi)
	proxy := make(Bi)
	var mu sync.Mutex

	go func() {
		defer close(proxy)
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case proxy <- v:
				case <-done:
					return
				}
			case <-done:
				return
			}
		}
	}()

	go func() {
		defer close(out)
		stageOut := stage(proxy)
		defer drain(stageOut)
		for {
			select {
			case v, ok := <-stageOut:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				default:
					mu.Lock()
					out <- v
					mu.Unlock()
				}
			case <-done:
				return
			}
		}
	}()
	return out
}

func drain(ch <-chan interface{}) {
	// Draining channel to avoid goroutine leak
	for range ch {
		// Perform a minimal action to avoid empty block
		log.Println("Draining channel")
	}
}
