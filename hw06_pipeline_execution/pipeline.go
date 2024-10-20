package hw06pipelineexecution

import "log"

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
	go func() {
		defer close(out)
		stageOut := stage(in)
		for {
			select {
			case v, ok := <-stageOut:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done:
					go drain(stageOut)
					return
				}
			case <-done:
				go drain(stageOut)
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
