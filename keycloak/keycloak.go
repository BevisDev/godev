package keycloak

import (
	"context"
	"fmt"
	"log"

	"github.com/Nerzal/gocloak/v13"
)

type KC struct {
	cf     *Config
	client *gocloak.GoCloak
}

// New creates a new keycloak client connected to the specified host and port.
//
// The returned client can be used to authenticate users, manage realms, roles,
// and perform other Keycloak administrative tasks.
func New(cf *Config) *KC {
	client := &KC{
		client: gocloak.NewClient(fmt.Sprintf("%s:%d", cf.Host, cf.Port)),
		cf:     cf,
	}

	log.Println("[keycloak] started successfully")
	return client
}

// GetClient returns the GoCloak client instance.
// This instance is configured with the base Keycloak URL and provides
// methods to interact with the Keycloak Admin and User APIs
func (k *KC) GetClient() *gocloak.GoCloak {
	return k.client
}

// Login is responsible for authenticating a user (typically using credentials) in exchange for a token pair.
// Returns *gocloak.JWT: A pointer to the struct containing the authenticated tokens, including the Access Token and Refresh Token.
// Returns error: If the login process fails (e.g., invalid credentials, network error).
func (k *KC) Login(ctx context.Context, clientId, clientSecret string) (*gocloak.JWT, error) {
	return k.client.LoginClient(ctx, clientId, clientSecret, k.cf.Realm)
}

// VerifyToken is used to check the validity and status of a given Access Token with the Keycloak server.
// Returns *gocloak.IntroSpectTokenResult: A pointer to the introspection result,
// which usually includes an Active boolean field indicating whether the token is currently valid.
func (k *KC) VerifyToken(ctx context.Context, token, clientId, clientSecret string) (*gocloak.IntroSpectTokenResult, error) {
	return k.client.RetrospectToken(ctx, token, clientId, clientSecret, k.cf.Realm)
}

// GetUserInfo retrieves detailed information about the authenticated user associated with the given Access Token.
// Returns *gocloak.UserInfo: A pointer to the struct containing user details like name, email, roles, etc.
// Returns error: If the user info cannot be retrieved (e.g., token is invalid or expired).
func (k *KC) GetUserInfo(ctx context.Context, token string) (*gocloak.UserInfo, error) {
	return k.client.GetUserInfo(ctx, token, k.cf.Realm)
}

// RevokeToken is used to immediately invalidate a given Refresh Token or Access Token,
func (k *KC) RevokeToken(ctx context.Context, clientId, clientSecret, token string) error {
	return k.client.RevokeToken(ctx, k.cf.Realm, clientId, clientSecret, token)
}
