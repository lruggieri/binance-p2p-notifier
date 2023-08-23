package event

import "context"

type Client interface {
	// Listen : returns a channel where Events will be sent.
	// The channel closure is responsibility of Client
	Listen(ctx context.Context) (chan Event, error)

	// WithCallback : sets a function to be called once Event is received
	// args is a string including all arguments send for this Event, separated by a space
	WithCallback(Event, func(args string) (reply string, err error)) Client
}
