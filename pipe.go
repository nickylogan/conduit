package conduit

import (
	"log"
	"sync"

	"go.uber.org/ratelimit"
)

type Pipe interface {
	Process(in chan interface{}) (out chan interface{})
}

type Processor func(in interface{}) (out interface{}, err error)

type pipe struct {
	rl      ratelimit.Limiter
	cfg     Config
	process Processor
}

func NewPipe(cfg Config, p Processor) Pipe {
	return &pipe{
		rl:      ratelimit.New(cfg.RateLimit),
		cfg:     cfg,
		process: p,
	}
}

func (p *pipe) Process(in chan interface{}) (out chan interface{}) {
	out = make(chan interface{}, p.cfg.OutputBuffer)
	jobs := make(chan Job, p.cfg.MaxJobs)

	go func() {
		var wg sync.WaitGroup

		// spawn a number of workers
		wg.Add(p.cfg.MaxWorkers)
		for i := 1; i <= p.cfg.MaxWorkers; i++ {
			go p.work(jobs, out, &wg)
		}

		// queue jobs
		p.delegateJobs(jobs, in, &wg)
		wg.Wait()

		close(out)
	}()

	return
}

func (p *pipe) work(jobs <-chan Job, out chan<- interface{}, wg *sync.WaitGroup) {
	for j := range jobs {
		func(j Job) {
			res, err := p.process(j.Payload)
			if err != nil {
				log.Println("failed to process input:", err)
				return
			}
			out <- res
		}(j)
	}

	// Mark that the worker is done
	wg.Done()
}

func (p *pipe) delegateJobs(jobs chan<- Job, in <-chan interface{}, wg *sync.WaitGroup) {
	jID := 1
	for d := range in {
		p.rl.Take()

		j := Job{
			ID:      jID,
			Payload: d,
		}
		jobs <- j
		jID++
	}
	close(jobs)
}
