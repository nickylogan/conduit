package conduit

type Source interface {
	Generate() (out <-chan interface{})
}

