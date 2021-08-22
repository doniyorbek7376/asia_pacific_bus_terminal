package passanger

import (
	"errors"
	"fmt"
)

var (
	ErrTicketAlreadyScanned   = errors.New("ticket already scanned")
	ErrTicketAlreadyProcessed = errors.New("ticket already processed")
	ErrAlreadyHasTicket       = errors.New("passanger has a ticket")
	ErrNoTicket               = errors.New("passanger doesn't have a ticket")
)

type passanger struct {
	id              int
	hasTicket       bool
	ticketProcessed bool
	ticketScanned   bool
}

type PassangerI interface {
	GetID() int
	IsTicketProcessed() bool
	IsTicketScanned() bool
	HasTicket() bool
	GiveTicket() error
	ScanTicket() error
	ProcessTicket() error
}

var _ PassangerI = &passanger{}

func NewPassanger(id int) PassangerI {
	return &passanger{id: id}
}

func (p *passanger) String() string {
	return fmt.Sprintf("Passanger-%v", p.id)
}

func (p *passanger) GetID() int {
	return p.id
}

func (p *passanger) IsTicketProcessed() bool {
	return p.ticketProcessed
}

func (p *passanger) IsTicketScanned() bool {
	return p.ticketScanned
}

func (p *passanger) HasTicket() bool {
	return p.hasTicket
}

func (p *passanger) ScanTicket() error {
	if p.ticketScanned {
		return ErrTicketAlreadyScanned
	}

	if !p.hasTicket {
		return ErrNoTicket
	}

	p.ticketScanned = true
	return nil
}

func (p *passanger) ProcessTicket() error {
	if p.ticketProcessed {
		return ErrTicketAlreadyProcessed
	}

	if !p.hasTicket {
		return ErrNoTicket
	}

	p.ticketProcessed = true
	return nil
}

func (p *passanger) GiveTicket() error {
	if p.hasTicket {
		return ErrAlreadyHasTicket
	}

	p.hasTicket = true
	return nil
}
