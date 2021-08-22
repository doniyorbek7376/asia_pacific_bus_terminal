package passenger

import (
	"errors"
	"fmt"
)

var (
	ErrTicketAlreadyScanned   = errors.New("ticket already scanned")
	ErrTicketAlreadyProcessed = errors.New("ticket already processed")
	ErrAlreadyHasTicket       = errors.New("passenger has a ticket")
	ErrNoTicket               = errors.New("passenger doesn't have a ticket")
)

type passenger struct {
	id              int
	hasTicket       bool
	ticketProcessed bool
	ticketScanned   bool
}

type PassengerI interface {
	EnterBuilding()
	LeaveBuilding()
	GoBack()
}

var _ PassengerI = &passenger{}

func NewPassenger(id int) PassengerI {
	return &passenger{id: id}
}

func (p *passenger) String() string {
	return fmt.Sprintf("Passenger-%v", p.id)
}

func (p *passenger) EnterBuilding() {
	fmt.Printf("%v entered the building\n", p)
}

func (p *passenger) LeaveBuilding() {
	fmt.Printf("%v left the building\n", p)
}

func (p *passenger) GoBack() {
	fmt.Printf("%v went back")
}
