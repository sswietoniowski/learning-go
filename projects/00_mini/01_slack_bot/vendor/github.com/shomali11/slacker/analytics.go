package slacker

import (
	"time"

	"github.com/shomali11/proper"
)

// NewCommandEvent creates a new command event
func NewCommandEvent(command string, parameters *proper.Properties, event *MessageEvent) *CommandEvent {
	return &CommandEvent{
		Timestamp:  time.Now(),
		Command:    command,
		Parameters: parameters,
		Event:      event,
	}
}

// CommandEvent is an event to capture executed commands
type CommandEvent struct {
	Timestamp  time.Time
	Command    string
	Parameters *proper.Properties
	Event      *MessageEvent
}
