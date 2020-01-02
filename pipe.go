package conduit

import (
	"sync"

	"go.uber.org/ratelimit"
)

// Pipe represents a process pipe.
type Pipe interface {
	// Process receives data from an input channel, processes it,
	// and outputs the results onto a channel. The output channel
	// is returned for use.
	Process(in chan interface{}) (out chan interface{})
}

// Processor is an interface for wrapping a process function
type Processor interface {
	// Process receives a given input, processes it, and returns an output
	Process(in interface{}) (out interface{})
}

// ProcessorFunc is a type adapter allowing regular functions to be Processor. If
// f is a function with the appropriate signature, ProcessorFunc(f) is a
// Processor that calls f.
type ProcessorFunc func(in interface{}) (out interface{})

// Process calls f(in) and returns the result.
func (f ProcessorFunc) Process(in interface{}) (out interface{}) {
	return f(in)
}

// pipe is a default implementation of the Pipe interface.
type pipe struct {
	rl  ratelimit.Limiter
	cfg Config
	p   Processor
}

// NewPipe creates a new pipe with the given config. A Processor argument
// is given to process the incoming data.
func NewPipe(cfg Config, p Processor) Pipe {
	return &pipe{
		rl:  ratelimit.New(cfg.RateLimit),
		cfg: cfg,
		p:   p,
	}
}

// Process implements the Pipe interface.
//
// Process spawns a number of workers based on cfg.MaxWorkers. Data from the input channel
// will be added to the job queue, with queue size of cfg.MaxJobs. Jobs are processed at most
// cfg.RateLimit per second.
func (p *pipe) Process(in chan interface{}) (out chan interface{}) {
	out = make(chan interface{})
	jobs := make(chan Job, p.cfg.MaxJobs)

	go func() {
		var wg sync.WaitGroup

		// spawn a number of workers
		wg.Add(p.cfg.MaxWorkers)
		for i := 1; i <= p.cfg.MaxWorkers; i++ {
			go func() {
				defer wg.Done()
				p.work(jobs, out)
			}()
		}

		// queue jobs
		p.delegateJobs(in, jobs)
		wg.Wait()

		close(out)
	}()

	return
}

// delegateJobs receives data from the input channel and adds it
// to the queue.
func (p *pipe) delegateJobs(in <-chan interface{}, jobs chan<- Job) {
	jID := 1
	for d := range in {
		j := Job{
			ID:      jID,
			Payload: d,
		}
		jobs <- j
		jID++
	}
	close(jobs)
}

// work represents a worker, processing jobs at most cfg.RateLimit per second.
func (p *pipe) work(jobs <-chan Job, out chan<- interface{}) {
	for j := range jobs {
		p.rl.Take()
		func(j Job) {
			out <- p.p.Process(j.Payload)
		}(j)
	}
}
