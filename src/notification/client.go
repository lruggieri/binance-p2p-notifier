package notification

type Client interface {
	SendMessage(msg string) error
}
