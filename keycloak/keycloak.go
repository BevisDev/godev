package keycloak

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
)

type KC struct {
	*Config
	client *gocloak.GoCloak
}

// NewClient creates a new KeyCloak client connected to the specified host and port.
//
// The returned client can be used to authenticate users, manage realms, roles,
// and perform other Keycloak administrative tasks.
func NewClient(cf *Config) KeyCloak {
	return &KC{
		client: gocloak.NewClient(fmt.Sprintf("%s:%d", cf.Host, cf.Port)),
		Config: cf,
	}
}

func (k *KC) GetClient() *gocloak.GoCloak {
	return k.client
}

func (k *KC) Login(ctx context.Context, clientId, clientSecret string) (*gocloak.JWT, error) {
	return k.client.LoginClient(ctx, clientId, clientSecret, k.Realm)
}

func (k *KC) VerifyToken(ctx context.Context, token, clientId, clientSecret string) (*gocloak.IntroSpectTokenResult, error) {
	return k.client.RetrospectToken(ctx, token, clientId, clientSecret, k.Realm)
}

func (k *KC) GetUserInfo(ctx context.Context, token string) (*gocloak.UserInfo, error) {
	return k.client.GetUserInfo(ctx, token, k.Realm)
}

func (k *KC) RevokeToken(ctx context.Context, clientId, clientSecret, token string) error {
	return k.client.RevokeToken(ctx, k.Realm, clientId, clientSecret, token)
}
