package notifier

import "context"

// Notifier delivers a text notification to some channel.
// The Slack implementation is planned; for now see the log notifier.
type Notifier interface {
	Send(ctx context.Context, text string) error
}
