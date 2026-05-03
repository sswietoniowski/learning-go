package main

import (
	"errors"
	"fmt"
)

type VehicleStatus string

const (
	VehicleAvailable        VehicleStatus = "available"
	VehicleUnderMaintenance VehicleStatus = "under_maintenance"
	VehicleRetired          VehicleStatus = "retired"
)

type Vehicle struct {
	plateNumber string
	model       string
	status      VehicleStatus
}

func RegisterVehicle(plateNumber, model string) (*Vehicle, error) {
	if plateNumber == "" {
		return nil, errors.New("plate number is required")
	}
	if model == "" {
		return nil, errors.New("model is required")
	}

	return &Vehicle{
		plateNumber: plateNumber,
		model:       model,
		status:      VehicleAvailable,
	}, nil
}

func (v *Vehicle) PlateNumber() string   { return v.plateNumber }
func (v *Vehicle) Model() string         { return v.model }
func (v *Vehicle) Status() VehicleStatus { return v.status }

func (v *Vehicle) MarkUnderMaintenance() error {
	if v.status != VehicleAvailable {
		return fmt.Errorf("cannot mark under maintenance: vehicle is %s", v.status)
	}
	v.status = VehicleUnderMaintenance
	return nil
}

func (v *Vehicle) MarkAvailable() error {
	if v.status != VehicleUnderMaintenance {
		return fmt.Errorf("cannot mark available: vehicle is %s", v.status)
	}
	v.status = VehicleAvailable
	return nil
}

func (v *Vehicle) RetireVehicle() error {
	if v.status == VehicleRetired {
		return errors.New("vehicle is already retired")
	}
	v.status = VehicleRetired
	return nil
}

type RentalStatus string

const (
	RentalStarted        RentalStatus = "started"
	RentalDamageReported RentalStatus = "damage_reported"
	RentalEnded          RentalStatus = "ended"
)

type Rental struct {
	vehicleID    string
	customerName string
	damage       string
	status       RentalStatus
}

func StartRental(vehicleID, customerName string) (*Rental, error) {
	if vehicleID == "" {
		return nil, errors.New("vehicle id is required")
	}
	if customerName == "" {
		return nil, errors.New("customer name is required")
	}

	return &Rental{
		vehicleID:    vehicleID,
		customerName: customerName,
		status:       RentalStarted,
	}, nil
}

func (r *Rental) VehicleID() string    { return r.vehicleID }
func (r *Rental) CustomerName() string { return r.customerName }
func (r *Rental) Damage() string       { return r.damage }
func (r *Rental) Status() RentalStatus { return r.status }

func (r *Rental) ReportDamage(description string) error {
	if r.status != RentalStarted {
		return fmt.Errorf("cannot report damage: rental is %s", r.status)
	}
	if description == "" {
		return errors.New("damage description is required")
	}
	r.damage = description
	r.status = RentalDamageReported
	return nil
}

func (r *Rental) EndRental() error {
	if r.status == RentalEnded {
		return errors.New("rental is already ended")
	}
	r.status = RentalEnded
	return nil
}
