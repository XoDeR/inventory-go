// Package log is a notifier that only writes notifications to the log.
// It is a stand-in until the Slack notifier is added.
package log

import (
	"context"

	"inventory-go/notification/internal/client/notifier"
	"inventory-go/platform/pkg/logger"
)

var _ notifier.Notifier = (*Notifier)(nil)

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) Send(ctx context.Context, text string) error {
	logger.Info(ctx, "Notification", "text", text)
	return nil
}
