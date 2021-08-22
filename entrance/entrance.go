package entrance

import (
	"context"
	"fmt"

	"github.com/doniyorbek7376/asia_pacific_bus_terminal/passenger"
)

type entrance struct {
	ctx        context.Context
	name       string
	passengers <-chan passenger.PassengerI
}

type EntranceI interface {
	Run() <-chan passenger.PassengerI
}

var _ EntranceI = &entrance{}

func NewEntrance(ctx context.Context, name string, passengers <-chan passenger.PassengerI) EntranceI {
	return &entrance{
		ctx:        ctx,
		name:       name,
		passengers: passengers,
	}
}

func (e *entrance) String() string {
	return fmt.Sprintf("Entrance-%v", e.name)
}

func (e *entrance) Run() <-chan passenger.PassengerI {
	return e.handlePassengers(e.passengers)
}

func (e *entrance) handlePassengers(passengers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI)

	go func() {
		defer close(ch)
		for p := range passengers {
			fmt.Printf("%v: %v came to the entrance\n", e, p)
			p.EnterBuilding()
			select {
			case <-e.ctx.Done():
				return
			case ch <- p:
			}
		}
	}()

	return ch
}
