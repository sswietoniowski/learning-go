// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import "testing"

func validVehicle(t *testing.T) *Vehicle {
	t.Helper()
	v, err := RegisterVehicle("AB-123", "Toyota Corolla")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return v
}

func TestRegisterVehicle(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		v := validVehicle(t)

		if v.PlateNumber() != "AB-123" {
			t.Errorf("got PlateNumber() = %q, want %q", v.PlateNumber(), "AB-123")
		}
		if v.Model() != "Toyota Corolla" {
			t.Errorf("got Model() = %q, want %q", v.Model(), "Toyota Corolla")
		}
		if v.Status() != VehicleAvailable {
			t.Errorf("got Status() = %q, want %q", v.Status(), VehicleAvailable)
		}
	})

	t.Run("empty_plate_number", func(t *testing.T) {
		_, err := RegisterVehicle("", "Toyota Corolla")
		if err == nil {
			t.Error("expected RegisterVehicle to reject an empty plate number")
		}
	})

	t.Run("empty_model", func(t *testing.T) {
		_, err := RegisterVehicle("AB-123", "")
		if err == nil {
			t.Error("expected RegisterVehicle to reject an empty model")
		}
	})
}

func TestMarkUnderMaintenance(t *testing.T) {
	t.Run("from_available", func(t *testing.T) {
		v := validVehicle(t)

		if err := v.MarkUnderMaintenance(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Status() != VehicleUnderMaintenance {
			t.Errorf("got Status() = %q, want %q", v.Status(), VehicleUnderMaintenance)
		}
	})

	t.Run("from_retired", func(t *testing.T) {
		v := validVehicle(t)
		if err := v.RetireVehicle(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err := v.MarkUnderMaintenance()
		if err == nil {
			t.Error("expected MarkUnderMaintenance to fail when vehicle is retired")
		}
	})
}

func TestMarkAvailable(t *testing.T) {
	t.Run("from_under_maintenance", func(t *testing.T) {
		v := validVehicle(t)
		if err := v.MarkUnderMaintenance(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := v.MarkAvailable(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Status() != VehicleAvailable {
			t.Errorf("got Status() = %q, want %q", v.Status(), VehicleAvailable)
		}
	})

	t.Run("from_available", func(t *testing.T) {
		v := validVehicle(t)

		err := v.MarkAvailable()
		if err == nil {
			t.Error("expected MarkAvailable to fail when vehicle is already available")
		}
	})
}

func TestRetireVehicle(t *testing.T) {
	t.Run("from_available", func(t *testing.T) {
		v := validVehicle(t)

		if err := v.RetireVehicle(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Status() != VehicleRetired {
			t.Errorf("got Status() = %q, want %q", v.Status(), VehicleRetired)
		}
	})

	t.Run("from_under_maintenance", func(t *testing.T) {
		v := validVehicle(t)
		if err := v.MarkUnderMaintenance(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := v.RetireVehicle(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Status() != VehicleRetired {
			t.Errorf("got Status() = %q, want %q", v.Status(), VehicleRetired)
		}
	})

	t.Run("already_retired", func(t *testing.T) {
		v := validVehicle(t)
		if err := v.RetireVehicle(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err := v.RetireVehicle()
		if err == nil {
			t.Error("expected RetireVehicle to fail when vehicle is already retired")
		}
	})
}

func validRental(t *testing.T) *Rental {
	t.Helper()
	r, err := StartRental("AB-123", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return r
}

func TestStartRental(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		r := validRental(t)

		if r.VehicleID() != "AB-123" {
			t.Errorf("got VehicleID() = %q, want %q", r.VehicleID(), "AB-123")
		}
		if r.CustomerName() != "Alice" {
			t.Errorf("got CustomerName() = %q, want %q", r.CustomerName(), "Alice")
		}
		if r.Status() != RentalStarted {
			t.Errorf("got Status() = %q, want %q", r.Status(), RentalStarted)
		}
	})

	t.Run("empty_vehicle_id", func(t *testing.T) {
		_, err := StartRental("", "Alice")
		if err == nil {
			t.Error("expected StartRental to reject an empty vehicle id")
		}
	})

	t.Run("empty_customer_name", func(t *testing.T) {
		_, err := StartRental("AB-123", "")
		if err == nil {
			t.Error("expected StartRental to reject an empty customer name")
		}
	})
}

func TestReportDamage(t *testing.T) {
	t.Run("from_started", func(t *testing.T) {
		r := validRental(t)

		if err := r.ReportDamage("scratched bumper"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Status() != RentalDamageReported {
			t.Errorf("got Status() = %q, want %q", r.Status(), RentalDamageReported)
		}
		if r.Damage() != "scratched bumper" {
			t.Errorf("got Damage() = %q, want %q", r.Damage(), "scratched bumper")
		}
	})

	t.Run("empty_description", func(t *testing.T) {
		r := validRental(t)

		err := r.ReportDamage("")
		if err == nil {
			t.Error("expected ReportDamage to reject an empty description")
		}
	})

	t.Run("from_ended", func(t *testing.T) {
		r := validRental(t)
		if err := r.EndRental(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err := r.ReportDamage("scratched bumper")
		if err == nil {
			t.Error("expected ReportDamage to fail when rental is ended")
		}
	})
}

func TestEndRental(t *testing.T) {
	t.Run("from_started", func(t *testing.T) {
		r := validRental(t)

		if err := r.EndRental(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Status() != RentalEnded {
			t.Errorf("got Status() = %q, want %q", r.Status(), RentalEnded)
		}
	})

	t.Run("from_damage_reported", func(t *testing.T) {
		r := validRental(t)
		if err := r.ReportDamage("scratched bumper"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := r.EndRental(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Status() != RentalEnded {
			t.Errorf("got Status() = %q, want %q", r.Status(), RentalEnded)
		}
	})

	t.Run("already_ended", func(t *testing.T) {
		r := validRental(t)
		if err := r.EndRental(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err := r.EndRental()
		if err == nil {
			t.Error("expected EndRental to fail when rental is already ended")
		}
	})
}
