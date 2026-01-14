package framework

import (
	"github.com/BevisDev/godev/logx"
	"github.com/BevisDev/godev/redis"
	"github.com/BevisDev/godev/rest"
)

type Config struct {
	LoggerConfig logx.Config
	RestClient   *rest.Client
	RedisCache   *redis.Cache
}

type Bootstrap struct {
	Logger logx.Logger
}

func New() {

}
