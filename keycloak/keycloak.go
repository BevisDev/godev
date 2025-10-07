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

func (k *KeyCloak) Login(ctx context.Context, clientId, clientSecret string) (*gocloak.JWT, error) {
	return k.client.LoginClient(ctx, clientId, clientSecret, k.Realm)
}

func (k *KeyCloak) VerifyToken(ctx context.Context, token, clientId, clientSecret string) (*gocloak.IntroSpectTokenResult, error) {
	return k.client.RetrospectToken(ctx, token, clientId, clientSecret, k.Realm)
}

func (k *KeyCloak) GetUserInfo(ctx context.Context, token string) (*gocloak.UserInfo, error) {
	return k.client.GetUserInfo(ctx, token, k.Realm)
}

func (k *KeyCloak) RevokeToken(ctx context.Context, clientId, clientSecret, token string) error {
	return k.client.RevokeToken(ctx, k.Realm, clientId, clientSecret, token)
}
