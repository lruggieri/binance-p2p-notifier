package notification

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type SlackMessage struct {
	Text string `json:"text"`
}

type Slack struct {
	webHookURL string
}

func (s *Slack) SendMessage(msg string) error {
	slackMsg := SlackMessage{Text: msg}
	msgBytes, err := json.Marshal(slackMsg)
	if err != nil {
		return err
	}

	resp, err := http.Post(s.webHookURL, "application/json", bytes.NewBuffer(msgBytes))
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

func NewSlack(webHookURL string) Client {
	return &Slack{
		webHookURL: webHookURL,
	}
}
