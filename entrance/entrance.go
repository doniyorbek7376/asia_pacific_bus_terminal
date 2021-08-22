package entrance

import (
	"context"
	"fmt"
	"sync"

	"github.com/doniyorbek7376/asia_pacific_bus_terminal/passenger"
)

type entrance struct {
	ctx        context.Context
	id         int
	passengers <-chan passenger.PassengerI
	wg         *sync.WaitGroup
}

type EntranceI interface {
	Run() <-chan passenger.PassengerI
}

var _ EntranceI = &entrance{}

func NewEntrance(ctx context.Context, id int, passengers <-chan passenger.PassengerI, wg *sync.WaitGroup) EntranceI {
	return &entrance{
		ctx:        ctx,
		id:         id,
		passengers: passengers,
		wg:         wg,
	}
}

func (e *entrance) String() string {
	return fmt.Sprintf("Entrance-%v", e.id)
}

func (e *entrance) Run() <-chan passenger.PassengerI {
	return e.handlePassengers(e.passengers)
}

func (e *entrance) handlePassengers(passengers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI)
	go func() {
		defer e.wg.Done()
		defer fmt.Printf("%v: closed\n", e)
		defer close(ch)
		for p := range passengers {
			fmt.Printf("%v: %v came to the entrance\n", e, p)
			p.EnterBuilding()
			go func(p passenger.PassengerI) {
				select {
				case <-e.ctx.Done():
					return
				case ch <- p:
				}
			}(p)
		}
	}()

	return ch
}
