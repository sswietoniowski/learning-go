package common

import (
	"database/sql/driver"
	"fmt"
)

type Enum[T Enumerable] struct {
	value string
}

func MustEnum[W ~struct{ Enum[T] }, T Enumerable](value string) W {
	return W{MustEnumFromString[T](value)}
}

func MustEnumFromString[T Enumerable](value string) Enum[T] {
	var e Enum[T]
	err := e.UnmarshalText([]byte(value))
	if err != nil {
		panic(fmt.Errorf("error unmarshalling enum value: %s", value))
	}

	return e
}

func (e *Enum[T]) Scan(src any) error {
	text, ok := src.(string)
	if !ok {
		return fmt.Errorf("invalid type for enum: %T, expected string", src)
	}

	return e.UnmarshalText([]byte(text))
}

func (e Enum[T]) Value() (driver.Value, error) {
	return e.value, nil
}

func (e Enum[T]) String() string {
	return e.value
}

func (e Enum[T]) IsZero() bool {
	return e.value == ""
}

func (e *Enum[T]) UnmarshalText(text []byte) error {
	var enum T
	valid := false
	expectedValues := enum.Values()

	if len(text) == 0 {
		e.value = ""
		return nil
	}

	for _, v := range expectedValues {
		if v == string(text) {
			valid = true
			e.value = v
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid enum value for %T: '%s', expected values %q", enum, string(text), expectedValues)
	}
	return nil
}

func (e Enum[T]) MarshalText() (text []byte, err error) {
	return []byte(e.value), nil
}

type Enumerable interface {
	Values() []string
}
