package config

import "time"

const (
	TicketMachineCount = 1
	EntrancesCount     = 2
	TicketBoothCount   = 2
	WaitingAreasCount  = 3
	PassengersCount    = 150

	TicketQueueCapacity = 5
	WaitingAreaCapacity = 10
	BusCapacity         = 6
	BuildingCapacity    = 50

	TicketMachineProcessTime   = 4 * time.Second
	TicketBoothProcessTime     = 8 * time.Second
	TicketScannerProcessTime   = 5 * time.Second
	TicketInspectorProcessTime = 2 * time.Second
	BusProcessTime             = 1 * time.Second
	PassengerArrivePeriod      = 4
	BusArrivalPeriod           = 10 * time.Second

	ToiletBreakTime         = 5 * time.Second
	TicketMachineRepairTime = 5 * time.Second
)
