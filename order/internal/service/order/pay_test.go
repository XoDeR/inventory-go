package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"inventory-go/order/internal/model"
)

func (s *ServiceSuite) TestPaySuccess() {
	var (
		ctx           = context.Background()
		orderUUID     = uuid.New()
		userUUID      = uuid.New()
		transactionID = uuid.New()
		method        = model.PaymentMethodCard

		storedOrder = model.Order{
			Uuid:       orderUUID,
			UserUuid:   userUUID,
			TotalPrice: 100,
			Status:     model.OrderStatusPendingPayment,
		}

		payInfo = model.PayOrderInfo{OrderUuid: orderUUID, UserUuid: userUUID, PaymentMethod: method}

		paidOrder = model.Order{
			Uuid:            orderUUID,
			UserUuid:        userUUID,
			TotalPrice:      100,
			Status:          model.OrderStatusPaid,
			TransactionUuid: &transactionID,
			PaymentMethod:   &method,
		}
	)

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(storedOrder, nil)
	s.paymentClient.EXPECT().PayOrder(ctx, payInfo).Return(transactionID.String(), nil)
	s.orderRepository.EXPECT().Update(ctx, paidOrder).Return(nil)

	result, err := s.service.Pay(ctx, orderUUID, method)

	s.Require().NoError(err)
	s.Require().Equal(transactionID, result)
}

func (s *ServiceSuite) TestPayOrderNotFound() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
	)

	// The payment must not be attempted for a missing order: no PayOrder expectation is set.
	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(model.Order{}, model.ErrOrderNotFound)

	result, err := s.service.Pay(ctx, orderUUID, model.PaymentMethodCard)

	s.Require().ErrorIs(err, model.ErrOrderNotFound)
	s.Require().Equal(uuid.Nil, result)
}

func (s *ServiceSuite) TestPayPaymentError() {
	var (
		ctx        = context.Background()
		orderUUID  = uuid.New()
		userUUID   = uuid.New()
		method     = model.PaymentMethodCreditCard
		paymentErr = errors.New("payment is unavailable")

		storedOrder = model.Order{Uuid: orderUUID, UserUuid: userUUID, Status: model.OrderStatusPendingPayment}
		payInfo     = model.PayOrderInfo{OrderUuid: orderUUID, UserUuid: userUUID, PaymentMethod: method}
	)

	// The order must stay unchanged when the payment fails: no Update expectation is set.
	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(storedOrder, nil)
	s.paymentClient.EXPECT().PayOrder(ctx, payInfo).Return("", paymentErr)

	result, err := s.service.Pay(ctx, orderUUID, method)

	s.Require().ErrorIs(err, paymentErr)
	s.Require().Equal(uuid.Nil, result)
}

func (s *ServiceSuite) TestPayInvalidTransactionUuid() {
	var (
		ctx       = context.Background()
		orderUUID = uuid.New()
		userUUID  = uuid.New()
		method    = model.PaymentMethodCard

		storedOrder = model.Order{Uuid: orderUUID, UserUuid: userUUID, Status: model.OrderStatusPendingPayment}
		payInfo     = model.PayOrderInfo{OrderUuid: orderUUID, UserUuid: userUUID, PaymentMethod: method}
	)

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(storedOrder, nil)
	s.paymentClient.EXPECT().PayOrder(ctx, payInfo).Return("not-a-uuid", nil)

	result, err := s.service.Pay(ctx, orderUUID, method)

	s.Require().Error(err)
	s.Require().Equal(uuid.Nil, result)
}

func (s *ServiceSuite) TestPayUpdateError() {
	var (
		ctx           = context.Background()
		orderUUID     = uuid.New()
		userUUID      = uuid.New()
		transactionID = uuid.New()
		method        = model.PaymentMethodStoreCredit
		repoErr       = errors.New("database is unavailable")

		storedOrder = model.Order{Uuid: orderUUID, UserUuid: userUUID, Status: model.OrderStatusPendingPayment}
		payInfo     = model.PayOrderInfo{OrderUuid: orderUUID, UserUuid: userUUID, PaymentMethod: method}

		paidOrder = model.Order{
			Uuid:            orderUUID,
			UserUuid:        userUUID,
			Status:          model.OrderStatusPaid,
			TransactionUuid: &transactionID,
			PaymentMethod:   &method,
		}
	)

	s.orderRepository.EXPECT().Get(ctx, orderUUID.String()).Return(storedOrder, nil)
	s.paymentClient.EXPECT().PayOrder(ctx, payInfo).Return(transactionID.String(), nil)
	s.orderRepository.EXPECT().Update(ctx, paidOrder).Return(repoErr)

	result, err := s.service.Pay(ctx, orderUUID, method)

	s.Require().ErrorIs(err, repoErr)
	s.Require().Equal(uuid.Nil, result)
}
