package conduit

import (
	"sync"

	"go.uber.org/ratelimit"
)

type Sink interface {
	Receive(in chan interface{}) (done chan bool)
}

type Receiver func(in interface{})

type sink struct {
	rl      ratelimit.Limiter
	cfg     Config
	receive Receiver
}

func NewSink(cfg Config, r Receiver) Sink {
	return &sink{
		rl:      ratelimit.New(cfg.RateLimit),
		cfg:     cfg,
		receive: r,
	}
}

func (s *sink) Receive(in chan interface{}) (done chan bool) {
	jobs := make(chan Job, s.cfg.MaxJobs)
	done = make(chan bool)

	go func(jobs chan Job, in <-chan interface{}, done chan bool) {
		var wg sync.WaitGroup
		wg.Add(s.cfg.MaxWorkers)

		// spawn a number of workers
		for i := 1; i <= s.cfg.MaxWorkers; i++ {
			go s.work(jobs, &wg)
		}

		// queue jobs
		s.delegateJobs(jobs, in)
		wg.Wait()

		done <- true
	}(jobs, in, done)

	return done
}

func (s *sink) work(jobs <-chan Job, wg *sync.WaitGroup) {
	for j := range jobs {
		func(j Job) {
			s.receive(j.Payload)
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
