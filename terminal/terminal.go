package terminal

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/doniyorbek7376/asia_pacific_bus_terminal/config"
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/entrance"
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/helper"
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/passenger"
)

var (
	ErrMaxCapacityReached = errors.New("building capacity reached")
	ErrNoPassengers       = errors.New("no passengers left")
)

type terminal struct {
	ctx             context.Context
	passengersCount int32
	closed          bool
	memoryAccess    sync.Mutex
	passengers      <-chan passenger.PassengerI

	entranceRequest    chan passenger.PassengerI
	ticketQueueRequest chan passenger.PassengerI
	waitingHallRequest chan passenger.PassengerI
	busRequest         chan passenger.PassengerI

	entranceResponse    []<-chan passenger.PassengerI
	ticketQueueResponse []<-chan passenger.PassengerI
	waitingHallResponse <-chan passenger.PassengerI
	busResponse         <-chan passenger.PassengerI

	entrances []entrance.EntranceI
}

type TerminalI interface {
	Run()
}

var _ TerminalI = &terminal{}

func NewTerminal(ctx context.Context, passangers <-chan passenger.PassengerI) TerminalI {
	return &terminal{ctx: ctx, passengers: passangers}
}

func (t *terminal) Run() {
	t.entranceRequest = make(chan passenger.PassengerI)
	defer close(t.entranceRequest)
	t.entrances = make([]entrance.EntranceI, config.EntrancesCount)
	t.entranceResponse = make([]<-chan passenger.PassengerI, config.EntrancesCount)

	for i := 0; i < config.EntrancesCount; i++ {
		t.entrances[i] = entrance.NewEntrance(t.ctx, fmt.Sprint(i+1), t.entranceRequest)
		t.entranceResponse[i] = t.entrances[i].Run()
	}

	pipeline := t.enterBuilding(t.passengers)
	// t.enterBus(
	// 	t.waitForBus(
	// 		t.temperatureScan(
	// 			t.ticketProcess(
	// 				t.goToWaitingHall(
	// 					t.waitForWaitingHall(
	// 						t.giveTicket(
	// 							t.queueTicket(
	// 								t.enterBuilding(t.passengers),
	// 							),
	// 						),
	// 					),
	// 				),
	// 			),
	// 		),
	// 	),
	// )

	for {
		select {
		case passenger, ok := <-pipeline:
			if !ok {
				return
			}
			fmt.Printf("Terminal: %v left the building\n", passenger)
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *terminal) enterBuilding(passengers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	go func() {
		for passenger := range passengers {
			t.memoryAccess.Lock()
			if t.passengersCount < config.BuildingCapacity {
				t.entranceRequest <- passenger
				t.passengersCount += 1
			} else {
				t.closed = true
				passenger.GoBack()
			}
			t.memoryAccess.Unlock()
		}
	}()

	return helper.Multiplex(t.ctx, t.entranceResponse...)
}
