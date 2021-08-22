package helper

import (
	"context"
	"sync"

	"github.com/doniyorbek7376/asia_pacific_bus_terminal/passenger"
)

func Multiplex(ctx context.Context, channels ...<-chan passenger.PassengerI) <-chan passenger.PassengerI {
	multiplexedStream := make(chan passenger.PassengerI)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan passenger.PassengerI) {
			defer wg.Done()
			for p := range c {
				select {
				case <-ctx.Done():
					return
				case multiplexedStream <- p:
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()
	return multiplexedStream
}
