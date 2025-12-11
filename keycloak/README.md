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
import (
    "context"
    "fmt"
    "log"
    
    "github.com/BevisDev/godev/keycloak"
)

func main() {
    var (
        clientId = "myclient"
        clientSecret = "mysecret"
    )
    cfg := &keycloak.Config{
        Host:  "http://localhost",
        Port:  8080,
        Realm: "myrealm",
    }

    kc := keycloak.NewClient(cfg)
    ctx := context.Background()

    token, err := kc.Login(ctx, clientId, clientSecret)
    if err != nil {
        log.Fatalf("login failed: %v", err)
    }

    userInfo, err := kc.GetUserInfo(ctx, token.AccessToken)
    if err != nil {
        log.Fatalf("get user info failed: %v", err)
    }

    fmt.Print(userInfo)
}

```