package keycloak

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
)

type KeyCloak struct {
	Client *gocloak.GoCloak
}

// NewKeyCloak creates a new KeyCloak client connected to the specified host and port.
//
// The returned client can be used to authenticate users, manage realms, roles,
// and perform other Keycloak administrative tasks.
//
// Example:
//
//	kc := NewKeyCloak("http://localhost", 8080)
//	token, err := kc.Client.Login(...)
//	if err != nil {
//	    log.Fatalf("login failed: %v", err)
//	}
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
