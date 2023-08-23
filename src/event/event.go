package event

type Event string

const (
	Pause     Event = "pause"
	Restart   Event = "restart"
	Blacklist Event = "blacklist"
)
