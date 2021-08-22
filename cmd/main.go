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
	terminal := terminal.NewTerminal(ctx, passengers)

	wg.Add(1)
	go func() {
		defer wg.Done()
		terminal.Run()
	}()

	for i := 1; i <= config.PassengersCount; i++ {
		passengers <- passenger.NewPassenger(i)
		time.Sleep(time.Second * time.Duration(1+rand.Intn(config.PassengerArrivePeriod)))
	}

	close(passengers)
	cancel()

	wg.Wait()
}
