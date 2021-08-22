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
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/ticket"
)

var (
	ErrMaxCapacityReached = errors.New("building capacity reached")
	ErrNoPassengers       = errors.New("no passengers left")
)

type terminal struct {
	ctx             context.Context
	passengersCount int32
	closed          bool
	cond            *sync.Cond
	wg              *sync.WaitGroup
	passengers      <-chan passenger.PassengerI

	entranceRequest    chan passenger.PassengerI
	ticketQueueRequest chan passenger.PassengerI
	waitingHallRequest chan passenger.PassengerI
	busRequest         chan passenger.PassengerI

	entranceResponse    []<-chan passenger.PassengerI
	ticketQueueResponse []<-chan passenger.PassengerI
	waitingHallResponse <-chan passenger.PassengerI
	busResponse         <-chan passenger.PassengerI

	entrances        []entrance.EntranceI
	ticketProcessors []ticket.TicketI
}

type TerminalI interface {
	Run() <-chan passenger.PassengerI
}

var _ TerminalI = &terminal{}

func NewTerminal(ctx context.Context, wg *sync.WaitGroup, passengers <-chan passenger.PassengerI) TerminalI {
	return &terminal{
		ctx:        ctx,
		passengers: passengers,
		wg:         wg,
		cond:       sync.NewCond(&sync.Mutex{}),
	}
}

func (t *terminal) Run() <-chan passenger.PassengerI {
	// Init entrances
	t.entranceRequest = make(chan passenger.PassengerI)

	t.entrances = make([]entrance.EntranceI, config.EntrancesCount)
	t.entranceResponse = make([]<-chan passenger.PassengerI, config.EntrancesCount)
	t.wg.Add(config.EntrancesCount)
	for i := 0; i < config.EntrancesCount; i++ {
		t.entrances[i] = entrance.NewEntrance(t.ctx, i, t.entranceRequest, t.wg)
		t.entranceResponse[i] = t.entrances[i].Run()
	}

	// Init ticketProcessors
	t.ticketQueueRequest = make(chan passenger.PassengerI)

	t.ticketProcessors = make([]ticket.TicketI, config.TicketBoothCount+config.TicketMachineCount)
	t.ticketQueueResponse = make([]<-chan passenger.PassengerI, config.TicketBoothCount+config.TicketMachineCount)
	for i := 0; i < config.TicketBoothCount; i++ {
		t.ticketProcessors[i] = ticket.NewTicketBooth(t.ctx, t.ticketQueueRequest, t.wg, i)
		t.ticketQueueResponse[i] = t.ticketProcessors[i].Run()
	}

	for i := 0; i < config.TicketMachineCount; i++ {
		t.ticketProcessors[i+config.TicketBoothCount] = ticket.NewTicketMachine(t.ctx, t.ticketQueueRequest, t.wg, i)
		t.ticketQueueResponse[i+config.TicketBoothCount] = t.ticketProcessors[i+config.TicketBoothCount].Run()
	}

	pipeline := t.leaveTerminal(t.processTicket(
		t.enterBuilding(t.passengers),
	),
	)
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

	go func() {
		defer fmt.Println("Terminal: closed")
		defer t.wg.Done()
		defer close(t.entranceRequest)
		defer close(t.ticketQueueRequest)
		<-t.ctx.Done()
	}()

	return pipeline
}

func (t *terminal) enterBuilding(passengers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	go func() {
		for p := range passengers {
			t.cond.L.Lock()
			if t.passengersCount >= config.BuildingCapacity {
				t.closed = true
				fmt.Println("Terminal reached max capacity")
			}
			if t.closed {
				t.cond.Wait()
			}
			t.passengersCount++
			fmt.Printf("Terminal: %v came to the building\nTerminal: passengers count: %v\n", p, t.passengersCount)
			t.cond.L.Unlock()

			select {
			case <-t.ctx.Done():
				return
			default:
			}

			go func(p passenger.PassengerI) {
				select {
				case <-t.ctx.Done():
					return
				case t.entranceRequest <- p:
				}
			}(p)
		}
	}()

	return helper.Multiplex(t.ctx, t.entranceResponse...)
}

func (t *terminal) processTicket(passengers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		defer close(ch)
		for passenger := range passengers {
			select {
			case <-t.ctx.Done():
				return
			case t.ticketQueueRequest <- passenger:
			}
		}
	}()

	return helper.Multiplex(t.ctx, t.ticketQueueResponse...)
}

func (t *terminal) leaveTerminal(passangers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI)
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		defer close(ch)
		for p := range passangers {
			t.cond.L.Lock()
			t.passengersCount--
			fmt.Printf("Terminal: %v left the building\nTerminal: passengers count: %v\n", p, t.passengersCount)
			if t.closed && t.passengersCount <= config.BuildingCapacity*7/10 {
				fmt.Printf("Terminal: entrances are open\n")
				t.closed = false
				t.cond.L.Unlock()
				t.cond.Signal()
			} else {
				t.cond.L.Unlock()
			}

			select {
			case <-t.ctx.Done():
				return
			case ch <- p:
			}
		}
	}()
	return ch
}
