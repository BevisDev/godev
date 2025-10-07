package keycloak

type Config struct {
	Host         string
	Port         int
	Realm        string
	ClientId     string
	ClientSecret string
}
