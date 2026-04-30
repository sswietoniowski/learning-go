package common

import (
	"database/sql/driver"

	"github.com/google/uuid"
)

type UUID [16]byte

func NewUUIDv7() UUID {
	u, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return UUID(u)
}

func MustUUIDFromString(s string) UUID {
	var u UUID
	err := u.UnmarshalText([]byte(s))
	if err != nil {
		panic(err)
	}

	return u
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func (u UUID) MarshalText() ([]byte, error) {
	return uuid.UUID(u).MarshalText()
}

func (u *UUID) UnmarshalText(data []byte) error {
	var guuid uuid.UUID
	if err := guuid.UnmarshalText(data); err != nil {
		return err
	}

	*u = UUID(guuid)
	return nil
}

func (u UUID) Value() (driver.Value, error) {
	return uuid.UUID(u).Value()
}

func (u *UUID) Scan(src any) error {
	var guuid uuid.UUID
	if err := guuid.Scan(src); err != nil {
		return err
	}

	*u = UUID(guuid)
	return nil
}

func (u UUID) IsZero() bool {
	return uuid.UUID(u) == uuid.Nil
}

func (u UUID) Equals(other UUID) bool {
	return u == other
}
