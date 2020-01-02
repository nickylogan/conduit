package conduit

// Source represents a data source
type Source interface {
	// Generate generates a data to an output channel.
	Generate() (out chan interface{})
}

// Generator is an interface for wrapping a generator function
type Generator interface {
	// Generate generates data onto the given channel
	Generate(out chan<- interface{})
}

// GeneratorFunc is a type adapter allowing regular functions to be Receiver. If
// f is a function with the appropriate signature, GeneratorFunc(f) is a Generator
// the calls f.
type GeneratorFunc func(out chan<- interface{})

// Generate calls f(out)
func (f GeneratorFunc) Generate(out chan<- interface{}) {
	f(out)
}

// source is a default implementation of the Source interface
type source struct {
	cfg Config
	g   Generator
}

// NewSource creates a new sink with the given config. A Generator argument
// is used for generating data.
func NewSource(cfg Config, g Generator) Source {
	return &source{
		cfg: cfg,
		g:   g,
	}
}

// Generate implements the Source interface
//
// Generate simply calls the internal Generator to feed data to the
// returned output channel.
func (s *source) Generate() (out chan interface{}) {
	out = make(chan interface{}, s.cfg.OutputBuffer)

	go func() {
		s.g.Generate(out)
		close(out)
	}()

	return out
}
