package main

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/doniyorbek7376/asia_pacific_bus_terminal/config"
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/passenger"
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/terminal"
)

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	passengers := make(chan passenger.PassengerI)
	terminal := terminal.NewTerminal(ctx, &wg, passengers)
	wg.Add(1)
	left := terminal.Run()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		defer close(passengers)
		leftCount := 0
		for leftCount < config.PassengersCount {
			<-left
			leftCount += 1
		}
	}()
	for i := 1; i <= config.PassengersCount; i++ {
		passengers <- passenger.NewPassenger(i)
		time.Sleep(time.Second * time.Duration(1+rand.Intn(config.PassengerArrivePeriod)))
	}

	wg.Wait()
}
