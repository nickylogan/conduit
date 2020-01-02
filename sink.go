package conduit

import (
	"sync"

	"go.uber.org/ratelimit"
)

// Sink represents a data sink
type Sink interface {
	// Receive receives data from an input channel.
	// Once completed, done will send a finished signal.
	Receive(in chan interface{}) (done chan struct{})
}

// Receiver is an interface for wrapping a receiver function
type Receiver interface {
	// Receive is an input feeder
	Receive(in interface{})
}

// ReceiverFunc is a type adapter allowing regular functions to be Receiver. If
// f is a function with the appropriate signature, ReceiverFunc(f) is a Receiver
// the calls f.
type ReceiverFunc func(in interface{})

// Receive calls f(in).
func (f ReceiverFunc) Receive(in interface{}) {
	f(in)
}

// sink is a default implementation of the Sink interface
type sink struct {
	rl  ratelimit.Limiter // rate-limiter
	cfg Config
	r   Receiver
}

// NewSink creates a new sink with the given config. A Receiver argument
// is given to handle incoming data.
func NewSink(cfg Config, r Receiver) Sink {
	return &sink{
		rl:  ratelimit.New(cfg.RateLimit),
		cfg: cfg,
		r:   r,
	}
}

// Receive implements the Sink interface
//
// Process spawns a number of workers based on cfg.MaxWorkers. Data from the input channel
// will be added to the job queue, with queue size of cfg.MaxJobs. Jobs are processed at most
// cfg.RateLimit per second. Once all jobs are finished, the done channel with send a signal.
func (s *sink) Receive(in chan interface{}) (done chan struct{}) {
	jobs := make(chan Job, s.cfg.MaxJobs)
	done = make(chan struct{})

	go func(jobs chan Job, in <-chan interface{}, done chan struct{}) {
		var wg sync.WaitGroup
		wg.Add(s.cfg.MaxWorkers)

		// spawn a number of workers
		for i := 1; i <= s.cfg.MaxWorkers; i++ {
			go s.work(jobs, &wg)
		}

		// queue jobs
		s.delegateJobs(jobs, in)
		wg.Wait()

		done <- struct{}{}
	}(jobs, in, done)

	return done
}

func (s *sink) work(jobs <-chan Job, wg *sync.WaitGroup) {
	for j := range jobs {
		s.rl.Take()
		func(j Job) {
			s.r.Receive(j.Payload)
		}(j)
	}

	// Mark that the worker is done
	wg.Done()
}

func (s *sink) delegateJobs(jobs chan<- Job, in <-chan interface{}) {
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
