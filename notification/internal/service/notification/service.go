package notification

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"inventory-go/notification/internal/client/notifier"
	"inventory-go/notification/internal/model"
	def "inventory-go/notification/internal/service"
)

var _ def.NotificationService = (*service)(nil)

var (
	orderPaidTemplate = template.Must(template.New("order_paid").Parse(
		`Order {{.OrderUUID}} was paid.
User: {{.UserUUID}}
Payment method: {{.PaymentMethod}}
Transaction: {{.TransactionUUID}}`))

	computerAssembledTemplate = template.Must(template.New("computer_assembled").Parse(
		`The computer for order {{.OrderUUID}} is assembled in {{.BuildTimeSec}} s.
User: {{.UserUUID}}`))
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

	return s.notifier.Send(ctx, buf.String())
}
