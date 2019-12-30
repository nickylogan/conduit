package conduit

type Pipe interface {
	Process(in chan<- interface{}, ins ...chan<- interface{}) (outs []<-chan interface{})
}
