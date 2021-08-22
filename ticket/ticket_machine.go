package ticket

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/doniyorbek7376/asia_pacific_bus_terminal/config"
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/passenger"
)

type ticketMachine struct {
	ctx            context.Context
	passengers     <-chan passenger.PassengerI
	isWorking      bool
	wg             *sync.WaitGroup
	cond           *sync.Cond
	count          int32
	processedCount int32
	id             int
}

var _ TicketI = &ticketMachine{}

func NewTicketMachine(ctx context.Context, passengers <-chan passenger.PassengerI, wg *sync.WaitGroup, id int) TicketI {
	return &ticketMachine{
		ctx:        ctx,
		passengers: passengers,
		wg:         wg,
		id:         id,
		isWorking:  true,
		count:      0,
		cond:       sync.NewCond(&sync.Mutex{}),
	}
}

func (tm *ticketMachine) String() string {
	return fmt.Sprintf("Ticket-machine-%v", tm.id)
}

func (tm *ticketMachine) Run() <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI)

	pipeline := tm.process(
		tm.queue(tm.passengers),
	)
	tm.wg.Add(1)
	go func() {
		defer func() {
			fmt.Printf("%v:passengers processed: %v\n%v: closed\n", tm, tm.processedCount, tm)
		}()
		defer close(ch)
		defer tm.wg.Done()

		for {
			select {
			case <-tm.ctx.Done():
				return
			case p, ok := <-pipeline:
				if !ok {
					return
				}
				select {
				case <-tm.ctx.Done():
					return
				case ch <- p:
				}
			}
		}
	}()

	tm.wg.Add(1)
	go func() {
		defer tm.wg.Done()
		for {
			select {
			case <-tm.ctx.Done():
				return
			case <-time.After(config.TicketMachineRepairTime):
				if rand.Intn(10) == 0 {
					tm.StopWorking()
					go func() {
						select {
						case <-time.After(config.TicketMachineRepairTime):
							tm.ContinueWorking()
						case <-tm.ctx.Done():
							return
						}
					}()
				}
			}
		}
	}()

	return ch
}

func (tm *ticketMachine) StopWorking() {
	tm.cond.L.Lock()
	tm.isWorking = false
	tm.cond.L.Unlock()
	fmt.Printf("%v: broke down\n", tm)
}

func (tm *ticketMachine) ContinueWorking() {
	tm.cond.L.Lock()
	tm.isWorking = true
	tm.cond.L.Unlock()
	fmt.Printf("%v: repaired\n", tm)
	tm.cond.Signal()
}

func (tm *ticketMachine) queue(passengers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI, config.TicketQueueCapacity-1)
	tm.wg.Add(1)
	go func() {
		defer tm.wg.Done()
		defer close(ch)
		for p := range passengers {
			atomic.AddInt32(&tm.count, 1)
			fmt.Printf("%v: %v entered queue\n%v: people in %v: %v\n", tm, p, tm, tm, tm.count)
			select {
			case <-tm.ctx.Done():
				return
			case ch <- p:
			}
		}
	}()
	return ch
}

func (tm *ticketMachine) process(passengers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI)

	tm.wg.Add(1)
	go func() {
		defer tm.wg.Done()
		defer close(ch)

		for p := range passengers {
			tm.cond.L.Lock()
			if !tm.isWorking {
				tm.cond.Wait()
			}
			tm.count -= 1
			tm.cond.L.Unlock()
			fmt.Printf("%v: %v is getting served\n", tm, p)
			time.Sleep(config.TicketMachineProcessTime)

			tm.cond.L.Lock()
			if !tm.isWorking {
				tm.cond.Wait()
			}
			tm.processedCount++
			tm.cond.L.Unlock()
			fmt.Printf("%v: %v got a ticket\n", tm, p)
			select {
			case <-tm.ctx.Done():
				return
			case ch <- p:
			}
		}
	}()

	return ch
}
