package events

import (
	"context"
	"time"
)

type Event interface {
	OccuredOn() time.Time
	JSON() ([]byte, error)
}

type EventListener interface {
	Name() string
	Handle(ctx context.Context, event Event) error
}
