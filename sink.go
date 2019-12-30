package conduit

type Sink interface {
	Receive(in chan<- interface{}) (done <-chan bool)
}
