package keycloak

import (
	"context"
	"errors"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockKeyCloak struct {
	mock.Mock
}

func (m *MockKeyCloak) LoginClient(ctx context.Context, clientID, secret, realm string) error {
	args := m.Called(ctx, clientID, secret, realm)
	return args.Error(0)
}

func (m *MockKeyCloak) RetrospectToken(ctx context.Context, token, clientID, secret, realm string) (*gocloak.IntroSpectTokenResult, error) {
	args := m.Called(ctx, token, clientID, secret, realm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gocloak.IntroSpectTokenResult), args.Error(1)
}

func (m *MockKeyCloak) GetUserInfo(ctx context.Context, token, realm string) (*gocloak.UserInfo, error) {
	args := m.Called(ctx, token, realm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gocloak.UserInfo), args.Error(1)
}

func TestLoginClient(t *testing.T) {
	authMock := new(MockKeyCloak)
	authMock.On("LoginClient", mock.Anything, "abc", "xyz", "test").Return(nil)

	err := authMock.LoginClient(context.TODO(), "abc", "xyz", "test")

	assert.NoError(t, err)
	authMock.AssertExpectations(t)
}

func TestLoginClient_Error(t *testing.T) {
	authMock := new(MockKeyCloak)

	authMock.On("LoginClient", context.TODO(), "wrong", "secret", "test").
		Return(errors.New("invalid client"))

	err := authMock.LoginClient(context.TODO(), "wrong", "secret", "test")

	assert.Error(t, err)
	assert.EqualError(t, err, "invalid client")
	authMock.AssertExpectations(t)
}

func TestRetrospectToken_Success(t *testing.T) {
	mockKC := new(MockKeyCloak)
	expected := &gocloak.IntroSpectTokenResult{
		Active: gocloak.BoolP(true),
	}

	ctx := context.TODO()
	mockKC.On("RetrospectToken", ctx, "token", "client", "secret", "realm").
		Return(expected, nil)

	result, err := mockKC.RetrospectToken(ctx, "token", "client", "secret", "realm")

	assert.NoError(t, err)
	assert.True(t, gocloak.PBool(result.Active))
	mockKC.AssertExpectations(t)
}

func TestRetrospectToken_Error(t *testing.T) {
	mockKC := new(MockKeyCloak)
	ctx := context.TODO()
	mockKC.On("RetrospectToken", ctx, "bad", "client", "secret", "realm").
		Return(nil, errors.New("invalid token"))

	result, err := mockKC.RetrospectToken(ctx, "bad", "client", "secret", "realm")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "invalid token")
	mockKC.AssertExpectations(t)
}

func TestGetUserInfo_Success(t *testing.T) {
	mockKC := new(MockKeyCloak)
	expected := &gocloak.UserInfo{
		Sub: gocloak.StringP("user123"),
	}

	ctx := context.TODO()
	mockKC.On("GetUserInfo", ctx, "token", "realm").
		Return(expected, nil)

	info, err := mockKC.GetUserInfo(ctx, "token", "realm")

	assert.NoError(t, err)

	assert.Equal(t, "user123", gocloak.PString(info.Sub))
	mockKC.AssertExpectations(t)
}

func TestGetUserInfo_Error(t *testing.T) {
	mockKC := new(MockKeyCloak)
	ctx := context.TODO()
	mockKC.On("GetUserInfo", ctx, "token", "realm").
		Return(nil, errors.New("unauthorized"))

	info, err := mockKC.GetUserInfo(ctx, "token", "realm")

	assert.Error(t, err)
	assert.Nil(t, info)
	assert.EqualError(t, err, "unauthorized")
	mockKC.AssertExpectations(t)
}
