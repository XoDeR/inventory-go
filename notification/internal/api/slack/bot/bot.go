// Package bot handles the Slack slash commands over Socket Mode: it opens an outbound
// websocket, so the service needs no public URL.
package bot

import (
	"context"
	"errors"
	"sync"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"inventory-go/platform/pkg/logger"
)

const (
	startCommand = "/start"
	startReply   = "Hi! I'll notify this channel about paid and assembled orders."
)

type Bot struct {
	api *slack.Client

	once sync.Once
	done chan struct{}
}

func NewBot(api *slack.Client) *Bot {
	return &Bot{api: api, done: make(chan struct{})}
}

// Run serves the slash commands until ctx is cancelled or the connection fails for good.
func (b *Bot) Run(ctx context.Context) error {
	defer b.once.Do(func() { close(b.done) })

	sm := socketmode.New(b.api)

	runErr := make(chan error, 1)
	go func() { runErr <- sm.RunContext(ctx) }()

	for {
		select {
		case err := <-runErr:
			if err != nil && !errors.Is(err, context.Canceled) {
				return err
			}

			return nil
		case evt := <-sm.Events:
			b.handle(ctx, sm, evt)
		}
	}
}

// Close waits until Run has returned. Cancel the context passed to Run first.
func (b *Bot) Close(ctx context.Context) error {
	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *Bot) handle(ctx context.Context, sm *socketmode.Client, evt socketmode.Event) {
	switch evt.Type {
	case socketmode.EventTypeConnected:
		logger.Info(ctx, "Slack Socket Mode connected")
	case socketmode.EventTypeConnectionError:
		logger.Warn(ctx, "Slack Socket Mode connection error", "data", evt.Data)
	case socketmode.EventTypeSlashCommand:
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok || evt.Request == nil {
			return
		}

		// The ack payload is the reply, and it has to be sent within 3 seconds.
		if cmd.Command != startCommand {
			sm.Ack(*evt.Request)
			return
		}

		sm.Ack(*evt.Request, map[string]any{
			"response_type": "ephemeral",
			"text":          startReply,
		})
	}
}
