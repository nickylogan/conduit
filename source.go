package conduit

type Source interface {
	Generate() (out chan interface{})
}

type Generator func(out chan<- interface{})

type source struct {
	cfg      Config
	generate Generator
}

func NewSource(cfg Config, g Generator) Source {
	return &source{
		cfg:      cfg,
		generate: g,
	}
}

func (s *source) Generate() (out chan interface{}) {
	out = make(chan interface{}, s.cfg.OutputBuffer)

	go func() {
		s.generate(out)
		close(out)
	}()

	return out
}
