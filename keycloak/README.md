# Keycloak Client

This package provides a simple Go wrapper around the [gocloak](https://github.com/Nerzal/gocloak) library to interact
with a Keycloak server.  
It allows you to easily manage authentication, token verification, and user information for confidential clients.

## Features

- Client login using `clientId` and `clientSecret`.
- Token verification and introspection.
- Fetch user information associated with access tokens.
- Revoke access or refresh tokens.
- Lightweight and easy-to-use API.

## Example Usage

```go
cfg := &keycloak.Config{
    Host:         "http://localhost",
    Port:         8080,
    Realm:        "myrealm",
    ClientId:     "myclient",
    ClientSecret: "mysecret",
}

kc := keycloak.New(cfg)
ctx := context.Background()

token, err := kc.Login(ctx)
if err != nil {
    log.Fatalf("login failed: %v", err)
}

userInfo, err := kc.GetUserInfo(ctx, token.AccessToken)
if err != nil {
    log.Fatalf("get user info failed: %v", err)
}
