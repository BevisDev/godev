package redis

// Config holds configuration options for connecting to a Redis instance.
//
// It includes host address, port, authentication credentials, selected DB index,
// connection pool size, and a default timeout (in seconds) for Redis operations.
type Config struct {
	Host       string // Redis server hostname or IP
	Port       int    // Redis server port
	Password   string // Password for authentication (if required)
	DB         int    // Redis database index (0 by default)
	PoolSize   int    // Maximum number of connections in the pool
	TimeoutSec int    // timeout for Redis operations in seconds
}
