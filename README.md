# conduit

[![GoDoc](https://godoc.org/github.com/nickylogan/conduit?status.svg)](https://godoc.org/github.com/nickylogan/conduit)

Package conduit is a helper for creating pipelines with worker pools with a rate limiter.

## Example

```go
// The pipeline is configured to create three workers, 
// where there can be at most ten jobs in queue. As the 
// rate limit is set to 5, only 5 jobs at most 
// can run in one second.
cfg := conduit.Config{
    MaxJobs:      10,
    MaxWorkers:   3,
    RateLimit:    5,
    OutputBuffer: 5,
}

// Create a source to generate 10 numbers onto a channel
generatorFunc := conduit.GeneratorFunc(func(out chan<- interface{}) {
    for i := 1; i <= 10; i++ {
        out <- i
    }
})
numbers := conduit.NewSource(cfg, generatorFunc).Generate()

// Create a process pipe that squares the incoming input
squareFunc := conduit.ProcessorFunc(func(in interface{}) (out interface{}) {
    x := in.(int)
    time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)
    return x * x
})
squares := conduit.NewPipe(cfg, squareFunc).Process(numbers)

// As a sink, print each incoming input
printer := conduit.ReceiverFunc(func(in interface{}) {
    fmt.Println(in)
})
done := conduit.NewSink(cfg, printer).Receive(squares)
<-done

// Output:
// 1
// 4
// 9
// 16
// 25
// 36
// 49
// 64
// 81
// 100
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details