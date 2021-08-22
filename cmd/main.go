package main

import (
	"context"
	"sync"
	"time"

	"github.com/doniyorbek7376/asia_pacific_bus_terminal/config"
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/passanger"
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/terminal"
)

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	terminal := terminal.NewTerminal(ctx)

	for i := 1; i <= config.PassangersCount; i++ {
		time.Sleep(config.PassangerArrivePeriod)
		passanger := passanger.NewPassanger(i)
		go terminal.HandleNewPassanger(passanger)
	}
	cancel()

	wg.Wait()
}
