package ticket

import (
	"context"

	"github.com/doniyorbek7376/asia_pacific_bus_terminal/passenger"
)

type ticketBooth struct {
	ctx        context.Context
	passengers <-chan passenger.PassengerI
	isWorking  bool
	id         int
}
