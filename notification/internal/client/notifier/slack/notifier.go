// Package slack is a notifier that posts notifications to a Slack channel.
package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"

	"inventory-go/notification/internal/client/notifier"
)

var _ notifier.Notifier = (*Notifier)(nil)

// messagePoster is the part of *slack.Client the notifier uses.
type messagePoster interface {
	PostMessageContext(ctx context.Context, channelID string, options ...slack.MsgOption) (string, string, error)
}

type Notifier struct {
	api       messagePoster
	channelID string
}

func NewNotifier(api *slack.Client, channelID string) *Notifier {
	return &Notifier{api: api, channelID: channelID}
}

// Send posts the text (Slack mrkdwn) to the configured channel.
func (n *Notifier) Send(ctx context.Context, text string) error {
	if _, _, err := n.api.PostMessageContext(ctx, n.channelID, slack.MsgOptionText(text, false)); err != nil {
		return fmt.Errorf("failed to post the message to Slack channel %s: %w", n.channelID, err)
	}

	return nil
}
