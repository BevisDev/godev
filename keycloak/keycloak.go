package keycloak

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
)

type KeyCloak struct {
	*Config
	client *gocloak.GoCloak
}

// New creates a new KeyCloak client connected to the specified host and port.
//
// The returned client can be used to authenticate users, manage realms, roles,
// and perform other Keycloak administrative tasks.
func New(cf *Config) Exec {
	return &KeyCloak{
		client: gocloak.NewClient(fmt.Sprintf("%s:%d", cf.Host, cf.Port)),
		Config: cf,
	}
}

func (k *KeyCloak) GetClient() *gocloak.GoCloak {
	return k.client
}

func (k *KeyCloak) Login(ctx context.Context, clientId, clientSecret, realm string) (*gocloak.JWT, error) {
	return k.client.LoginClient(ctx, clientId, clientSecret, realm)
}

func (k *KeyCloak) VerifyToken(ctx context.Context, token, clientId, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error) {
	return k.client.RetrospectToken(ctx, token, clientId, clientSecret, realm)
}

func (k *KeyCloak) GetUserInfo(ctx context.Context, token, realm string) (*gocloak.UserInfo, error) {
	return k.client.GetUserInfo(ctx, token, realm)
}

func (k *KeyCloak) RevokeToken(ctx context.Context, realm, clientId, clientSecret, token string) error {
	return k.client.RevokeToken(ctx, realm, clientId, clientSecret, token)
}
