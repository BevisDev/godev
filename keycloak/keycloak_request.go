package keycloak

type Request struct {
	Host         string
	Port         int
	Realm        string
	ClientId     string
	ClientSecret string
}
