package ticket

import "github.com/doniyorbek7376/asia_pacific_bus_terminal/passenger"

type TicketI interface {
	Run() <-chan passenger.PassengerI
	StopWorking()
	ContinueWorking()
}
