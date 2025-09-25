package keycloak

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v13"
)

type KeyCloak struct {
	Client       *gocloak.GoCloak
	Realm        string
	ClientId     string
	ClientSecret string
}

type Request struct {
	Host         string
	Port         int
	Realm        string
	ClientId     string
	ClientSecret string
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
func NewKeyCloak(r *Request) *KeyCloak {
	return &KeyCloak{
		Client:       gocloak.NewClient(fmt.Sprintf("%s:%d", r.Host, r.Port)),
		Realm:        r.Realm,
		ClientId:     r.ClientId,
		ClientSecret: r.ClientSecret,
	}
}

func (k *KeyCloak) Login(ctx context.Context) (*gocloak.JWT, error) {
	return k.Client.LoginClient(ctx, k.ClientId, k.ClientSecret, k.Realm)
}

func (k *KeyCloak) VerifyToken(ctx context.Context, token string) (*gocloak.IntroSpectTokenResult, error) {
	return k.Client.RetrospectToken(ctx, token, k.ClientId, k.ClientSecret, k.Realm)
}

func (k *KeyCloak) GetUserInfo(ctx context.Context, token string) (*gocloak.UserInfo, error) {
	return k.Client.GetUserInfo(ctx, token, k.Realm)
}

func (k *KeyCloak) RevokeToken(ctx context.Context, token string) error {
	return k.Client.RevokeToken(ctx, k.Realm, k.ClientId, k.ClientSecret, token)
}
