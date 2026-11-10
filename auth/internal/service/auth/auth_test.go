package auth

import (
	"context"
	"errors"

	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"inventory-go/auth/internal/model"
)

const (
	testUserUUID    = "3b1f6c1e-6a5e-4b0e-9a54-2f6c1d9f7a10"
	testSessionUUID = "8c2a7d3e-1f4b-4c6a-8e55-0a9d3b7c2e11"
	testLogin       = "alice"
	testPassword    = "correct horse"
)

func (s *ServiceSuite) testUser() model.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	s.Require().NoError(err)

	return model.User{
		UUID:         testUserUUID,
		Login:        testLogin,
		Email:        "alice@example.com",
		PasswordHash: string(hash),
	}
}

func (s *ServiceSuite) TestRegisterHashesPassword() {
	ctx := context.Background()
	info := model.RegisterInfo{
		Login:               testLogin,
		Password:            testPassword,
		Email:               "alice@example.com",
		NotificationMethods: []model.NotificationMethod{{ProviderName: "email", Target: "alice@example.com"}},
	}

	var saved model.User
	s.userRepository.EXPECT().Create(ctx, mock.Anything).
		Run(func(_ context.Context, user model.User) { saved = user }).
		Return(nil)

	userUUID, err := s.service.Register(ctx, info)

	s.Require().NoError(err)
	s.Require().Equal(saved.UUID, userUUID)
	s.Require().NotEmpty(userUUID)
	s.Require().NotEqual(testPassword, saved.PasswordHash)
	s.Require().NoError(bcrypt.CompareHashAndPassword([]byte(saved.PasswordHash), []byte(testPassword)))
	s.Require().Equal(info.NotificationMethods, saved.NotificationMethods)
}

func (s *ServiceSuite) TestRegisterAlreadyExists() {
	ctx := context.Background()

	s.userRepository.EXPECT().Create(ctx, mock.Anything).Return(model.ErrUserAlreadyExists)

	_, err := s.service.Register(ctx, model.RegisterInfo{Login: testLogin, Password: testPassword})

	s.Require().ErrorIs(err, model.ErrUserAlreadyExists)
}

func (s *ServiceSuite) TestLoginCreatesSession() {
	ctx := context.Background()

	s.userRepository.EXPECT().GetByLogin(ctx, testLogin).Return(s.testUser(), nil)
	s.sessionRepository.EXPECT().Create(ctx, mock.Anything, testUserUUID, testSessionTTL).Return(nil)

	sessionUUID, err := s.service.Login(ctx, testLogin, testPassword)

	s.Require().NoError(err)
	s.Require().NotEmpty(sessionUUID)
}

func (s *ServiceSuite) TestLoginWrongPassword() {
	ctx := context.Background()

	s.userRepository.EXPECT().GetByLogin(ctx, testLogin).Return(s.testUser(), nil)

	_, err := s.service.Login(ctx, testLogin, "wrong")

	s.Require().ErrorIs(err, model.ErrInvalidCredentials)
}

func (s *ServiceSuite) TestLoginUnknownUserLooksLikeWrongPassword() {
	ctx := context.Background()

	s.userRepository.EXPECT().GetByLogin(ctx, testLogin).Return(model.User{}, model.ErrUserNotFound)

	_, err := s.service.Login(ctx, testLogin, testPassword)

	s.Require().ErrorIs(err, model.ErrInvalidCredentials)
}

func (s *ServiceSuite) TestLoginSessionStoreFails() {
	ctx := context.Background()
	storeErr := errors.New("redis down")

	s.userRepository.EXPECT().GetByLogin(ctx, testLogin).Return(s.testUser(), nil)
	s.sessionRepository.EXPECT().Create(ctx, mock.Anything, testUserUUID, testSessionTTL).Return(storeErr)

	_, err := s.service.Login(ctx, testLogin, testPassword)

	s.Require().ErrorIs(err, storeErr)
}

func (s *ServiceSuite) TestWhoami() {
	ctx := context.Background()
	user := s.testUser()

	s.sessionRepository.EXPECT().Get(ctx, testSessionUUID).Return(testUserUUID, nil)
	s.userRepository.EXPECT().GetByUUID(ctx, testUserUUID).Return(user, nil)

	got, err := s.service.Whoami(ctx, testSessionUUID)

	s.Require().NoError(err)
	s.Require().Equal(user, got)
}

func (s *ServiceSuite) TestWhoamiUnknownSession() {
	ctx := context.Background()

	s.sessionRepository.EXPECT().Get(ctx, testSessionUUID).Return("", model.ErrSessionNotFound)

	_, err := s.service.Whoami(ctx, testSessionUUID)

	s.Require().ErrorIs(err, model.ErrSessionNotFound)
}

func (s *ServiceSuite) TestGetUserNotFound() {
	ctx := context.Background()

	s.userRepository.EXPECT().GetByUUID(ctx, testUserUUID).Return(model.User{}, model.ErrUserNotFound)

	_, err := s.service.GetUser(ctx, testUserUUID)

	s.Require().ErrorIs(err, model.ErrUserNotFound)
}
