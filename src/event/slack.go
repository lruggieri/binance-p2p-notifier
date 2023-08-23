package event

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"p2p-check/src/logger"
)

type Slack struct {
	eventChannel chan Event
	c            *slack.Client

	callbacks map[Event]func(args string) (string, error)
}

func (s *Slack) execSlashCommand(command, args string) (message string, err error) {
	/*
		Markdown examples
		https://app.slack.com/block-kit-builder/TBLAK791U#%7B%22blocks%22:%5B%7B%22type%22:%22section%22,%22text%22:%7B%22type%22:%22mrkdwn%22,%22text%22:%22This%20is%20unquoted%20text%5Cn%3EThis%20is%20quoted%20text%5Cn%3EThis%20is%20still%20quoted%20text%5CnThis%20is%20unquoted%20text%20again%22%7D%7D%5D%7D
	*/
	switch command {
	case "/hello":
		return "Hi there :)", nil
	case "/blacklist":
		callback, ok := s.callbacks[Blacklist]
		if !ok {
			return "", errors.New("callback for blacklist not implemented")
		}

		return callback(args)
	case "/pause":
		callback, ok := s.callbacks[Pause]
		if !ok {
			return "", errors.New("callback for pause not implemented")
		}

		return callback(args)
	case "/restart":
		callback, ok := s.callbacks[Restart]
		if !ok {
			return "", errors.New("callback for restart not implemented")
		}

		return callback(args)
	default:
		logger.Default.WithField("command", command).Info("slash command not handled")
		return "", nil
	}
}

func (s *Slack) Listen(ctx context.Context) (chan Event, error) {
	sm := socketmode.New(
		s.c,
		// socketmode.OptionDebug(true),
	)

	// event receive example: https://github.com/slack-go/slack/blob/master/examples/socketmode/socketmode.go
	go func() {
		for evt := range sm.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				logger.Default.Info("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				logger.Default.Info("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				logger.Default.Info("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					logger.Default.Error("slash event is invalid")

					continue
				}

				logger.Default.WithField("command", cmd).Info("Slash command received")

				if message, err := s.execSlashCommand(cmd.Command, cmd.Text); err != nil {
					payload := map[string]interface{}{
						"blocks": []slack.Block{
							slack.NewSectionBlock(
								&slack.TextBlockObject{
									Type: slack.MarkdownType,
									Text: err.Error(),
								},
								nil,
								nil,
							),
						}}
					sm.Ack(*evt.Request, payload)
				} else {
					payload := map[string]interface{}{
						"blocks": []slack.Block{
							slack.NewSectionBlock(
								&slack.TextBlockObject{
									Type: slack.MarkdownType,
									Text: message,
								},
								nil,
								nil,
							),
						}}

					sm.Ack(*evt.Request, payload)
				}

			case socketmode.EventTypeHello:
				sm.Debugf("Hello event received")
			default:
				logger.Default.WithField("event-type", evt.Type).Error("Unexpected event type received")
			}
		}
	}()

	go func() {
		if err := sm.Run(); err != nil {
			logger.Default.WithError(err).Error("Slack socketmode error")
		}
	}()

	return s.eventChannel, nil
}

func (s *Slack) WithCallback(e Event, c func(args string) (reply string, err error)) Client {
	s.callbacks[e] = c

	return s
}

func NewSlack(appToken string) Client {
	api := slack.New(
		appToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	return &Slack{
		c:            api,
		eventChannel: make(chan Event),
		callbacks:    make(map[Event]func(args string) (string, error)),
	}
}
