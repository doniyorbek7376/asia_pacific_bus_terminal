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

type ticketBooth struct {
	ctx            context.Context
	passengers     <-chan passenger.PassengerI
	isWorking      bool
	wg             *sync.WaitGroup
	cond           *sync.Cond
	count          int32
	id             int
	processedCount int32
}

var _ TicketI = &ticketBooth{}

func NewTicketBooth(ctx context.Context, passengers <-chan passenger.PassengerI, wg *sync.WaitGroup, id int) TicketI {
	return &ticketBooth{
		ctx:        ctx,
		passengers: passengers,
		wg:         wg,
		id:         id,
		isWorking:  true,
		count:      0,
		cond:       sync.NewCond(&sync.Mutex{}),
	}
}

func (tb *ticketBooth) String() string {
	return fmt.Sprintf("Ticket-booth-%v", tb.id)
}

func (tb *ticketBooth) Run() <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI)

	pipeline := tb.process(
		tb.queue(tb.passengers),
	)
	tb.wg.Add(1)
	go func() {
		defer func() {
			fmt.Printf("%v:passengers processed: %v\n%v: closed\n", tb, tb.processedCount, tb)
		}()
		defer close(ch)
		defer tb.wg.Done()

		for {
			select {
			case <-tb.ctx.Done():
				return
			case p, ok := <-pipeline:
				if !ok {
					return
				}
				select {
				case <-tb.ctx.Done():
					return
				case ch <- p:
				}
			}
		}
	}()

	tb.wg.Add(1)
	go func() {
		defer tb.wg.Done()
		for {
			select {
			case <-tb.ctx.Done():
				return
			case <-time.After(config.ToiletBreakTime * 10):
				if rand.Intn(2) == 0 {
					tb.StopWorking()
					go func() {
						<-time.After(config.ToiletBreakTime)
						tb.ContinueWorking()
					}()
				}
			}
		}
	}()

	return ch
}

func (tb *ticketBooth) StopWorking() {
	tb.cond.L.Lock()
	tb.isWorking = false
	tb.cond.L.Unlock()
	fmt.Printf("%v: staff went to toilet\n", tb)
}

func (tb *ticketBooth) ContinueWorking() {
	tb.cond.L.Lock()
	tb.isWorking = true
	tb.cond.L.Unlock()
	fmt.Printf("%v: staff returned\n", tb)
	tb.cond.Signal()
}

func (tb *ticketBooth) queue(passengers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI, config.TicketQueueCapacity-1)
	tb.wg.Add(1)
	go func() {
		defer tb.wg.Done()
		defer close(ch)
		for p := range passengers {
			atomic.AddInt32(&tb.count, 1)
			fmt.Printf("%v: %v entered queue\n%v: people in %v: %v\n", tb, p, tb, tb, tb.count)
			select {
			case <-tb.ctx.Done():
				return
			case ch <- p:
			}
		}
	}()
	return ch
}

func (tb *ticketBooth) process(passengers <-chan passenger.PassengerI) <-chan passenger.PassengerI {
	ch := make(chan passenger.PassengerI)

	tb.wg.Add(1)
	go func() {
		defer tb.wg.Done()
		defer close(ch)

		for p := range passengers {
			tb.cond.L.Lock()
			if !tb.isWorking {
				tb.cond.Wait()
			}
			tb.count -= 1
			tb.cond.L.Unlock()
			fmt.Printf("%v: %v is getting served\n", tb, p)
			time.Sleep(config.TicketBoothProcessTime)

			tb.cond.L.Lock()
			if !tb.isWorking {
				tb.cond.Wait()
			}
			tb.processedCount++
			tb.cond.L.Unlock()
			fmt.Printf("%v: %v got a ticket\n", tb, p)
			select {
			case <-tb.ctx.Done():
				return
			case ch <- p:
			}
		}
	}()

	return ch
}
