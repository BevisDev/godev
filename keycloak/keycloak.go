package keycloak

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
)

type KeyCloak struct {
	Client *gocloak.GoCloak
}

func NewKeyCloak(host string, port int) *KeyCloak {
	return &KeyCloak{
		Client: gocloak.NewClient(fmt.Sprintf("%s:%d", host, port)),
	}
}

func (k *KeyCloak) LoginClient(ctx context.Context, clientId, clientSecret, realm string) (*gocloak.JWT, error) {
	return k.Client.LoginClient(ctx, clientId, clientSecret, realm)
}

func (k *KeyCloak) RetrospectToken(ctx context.Context, token, clientId, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error) {
	return k.Client.RetrospectToken(ctx, token, clientId, clientSecret, realm)
}

func (k *KeyCloak) GetUserInfo(ctx context.Context, token, realm string) (*gocloak.UserInfo, error) {
	return k.Client.GetUserInfo(ctx, token, realm)
}
