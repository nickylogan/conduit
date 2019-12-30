package conduit

// Config configures the batch job worker
type Config struct {
	MaxJobs      int
	MaxWorkers   int
	RateLimit    int
	OutputBuffer int
}

// Job is a representation of a batch job
type Job struct {
	ID      int
	Payload interface{}
}
