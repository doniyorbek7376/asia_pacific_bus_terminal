package terminal

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/doniyorbek7376/asia_pacific_bus_terminal/config"
	"github.com/doniyorbek7376/asia_pacific_bus_terminal/passanger"
)

var (
	ErrMaxCapacityReached = errors.New("building capacity reached")
	ErrNoPassangers       = errors.New("no passangers left")
)

type terminal struct {
	ctx             context.Context
	passangersCount int32
	closed          bool
}

type TerminalI interface {
	HandleNewPassanger(p passanger.PassangerI)
}

var _ TerminalI = &terminal{}

func NewTerminal(ctx context.Context) TerminalI {
	return &terminal{ctx: ctx}
}

func (t *terminal) HandleNewPassanger(p passanger.PassangerI) {
	if t.passangersCount >= config.BuildingCapacity || t.closed {
		// return ErrMaxCapacityReached
		fmt.Println()
		return
	}
	atomic.AddInt32(&t.passangersCount, 1)
	// return nil
}
