package keycloak

import (
	"context"
	"github.com/Nerzal/gocloak/v13"
)

type Exec interface {
	// GetClient returns the GoCloak client instance.
	// This instance is configured with the base Keycloak URL and provides
	// methods to interact with the Keycloak Admin and User APIs
	GetClient() *gocloak.GoCloak

	// Login is responsible for authenticating a user (typically using credentials) in exchange for a token pair.
	// Returns *gocloak.JWT: A pointer to the struct containing the authenticated tokens, including the Access Token and Refresh Token.
	// Returns error: If the login process fails (e.g., invalid credentials, network error).
	Login(ctx context.Context, clientId, clientSecret, realm string) (*gocloak.JWT, error)

	// VerifyToken is used to check the validity and status of a given Access Token with the Keycloak server.
	// Returns *gocloak.IntroSpectTokenResult: A pointer to the introspection result,
	// which usually includes an Active boolean field indicating whether the token is currently valid.
	VerifyToken(ctx context.Context, token, clientId, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error)

	// GetUserInfo retrieves detailed information about the authenticated user associated with the given Access Token.
	// Returns *gocloak.UserInfo: A pointer to the struct containing user details like name, email, roles, etc.
	// Returns error: If the user info cannot be retrieved (e.g., token is invalid or expired).
	GetUserInfo(ctx context.Context, token, realm string) (*gocloak.UserInfo, error)

	// RevokeToken is used to immediately invalidate a given Refresh Token or Access Token,
	RevokeToken(ctx context.Context, realm, clientId, clientSecret, token string) error
}
