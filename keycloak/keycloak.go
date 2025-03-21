package keycloak

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
)

type KeyCloak struct {
	client *gocloak.GoCloak
}

func NewKeyCloak(host string, port int) *KeyCloak {
	return &KeyCloak{
		client: gocloak.NewClient(fmt.Sprintf("%s:%d", host, port)),
	}
}

func (k *KeyCloak) LoginClient(ctx context.Context, clientId, clientSecret, realm string) (*gocloak.JWT, error) {
	return k.client.LoginClient(ctx, clientId, clientSecret, realm)
}

func (k *KeyCloak) RetrospectToken(ctx context.Context, token, clientId, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error) {
	return k.client.RetrospectToken(ctx, token, clientId, clientSecret, realm)
}

func (k *KeyCloak) GetUserInfo(ctx context.Context, token, realm string) (*gocloak.UserInfo, error) {
	return k.client.GetUserInfo(ctx, token, realm)
}
