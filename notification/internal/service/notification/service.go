package notification

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"inventory-go/notification/internal/client/notifier"
	"inventory-go/notification/internal/model"
	def "inventory-go/notification/internal/service"
)

var _ def.NotificationService = (*service)(nil)

// The templates are written in Slack mrkdwn (*bold*, `code`), parsed once at startup.
//
//go:embed templates/*.tmpl
var templatesFS embed.FS

var (
	templates = template.Must(template.ParseFS(templatesFS, "templates/*.tmpl"))

	orderPaidTemplate         = templates.Lookup("paid_notification.tmpl")
	computerAssembledTemplate = templates.Lookup("assembled_notification.tmpl")
)

type service struct {
	notifier notifier.Notifier
}

func NewService(notifier notifier.Notifier) *service {
	return &service{notifier: notifier}
}

func (s *service) NotifyOrderPaid(ctx context.Context, event model.OrderPaidEvent) error {
	return s.send(ctx, orderPaidTemplate, event)
}

func (s *service) NotifyComputerAssembled(ctx context.Context, event model.ComputerAssembledEvent) error {
	return s.send(ctx, computerAssembledTemplate, event)
}

func (s *service) send(ctx context.Context, tmpl *template.Template, data any) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render the %s message: %w", tmpl.Name(), err)
	}

	return s.notifier.Send(ctx, strings.TrimSpace(buf.String()))
}
