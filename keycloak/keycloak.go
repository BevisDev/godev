package keycloak

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
)

type KeyCloak struct {
	client       *gocloak.GoCloak
	realm        string
	clientId     string
	clientSecret string
}

// NewKeyCloak creates a new KeyCloak client connected to the specified host and port.
//
// The returned client can be used to authenticate users, manage realms, roles,
// and perform other Keycloak administrative tasks.
//
// Example:
//
//	kc := NewKeyCloak("http://localhost", 8080)
//	token, err := kc.LoginClient(...)
//	if err != nil {
//	    log.Fatalf("login failed: %v", err)
//	}
func NewKeyCloak(r *Request) KeyCloakExec {
	return &KeyCloak{
		client:       gocloak.NewClient(fmt.Sprintf("%s:%d", r.Host, r.Port)),
		realm:        r.Realm,
		clientId:     r.ClientId,
		clientSecret: r.ClientSecret,
	}
}

func (k *KeyCloak) GetClient() *gocloak.GoCloak {
	return k.client
}

func (k *KeyCloak) Login(ctx context.Context) (*gocloak.JWT, error) {
	return k.client.LoginClient(ctx, k.clientId, k.clientSecret, k.realm)
}

func (k *KeyCloak) VerifyToken(ctx context.Context, token string) (*gocloak.IntroSpectTokenResult, error) {
	return k.client.RetrospectToken(ctx, token, k.clientId, k.clientSecret, k.realm)
}

func (k *KeyCloak) GetUserInfo(ctx context.Context, token string) (*gocloak.UserInfo, error) {
	return k.client.GetUserInfo(ctx, token, k.realm)
}

func (k *KeyCloak) RevokeToken(ctx context.Context, token string) error {
	return k.client.RevokeToken(ctx, k.realm, k.clientId, k.clientSecret, token)
}
