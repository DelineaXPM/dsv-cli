package constants

// store constants
const (
	StoreRoot             = "thy"
	TokenRoot             = "token"
	EncryptionKey         = "encryptionkey"
	CliConfigRoot         = "config"
	SecretRoot            = "secret"
	SecretDescriptionRoot = "description"
)

// cache constants
const (
	Cache         = "cache"
	CacheStrategy = "cache.strategy"
	CacheAge      = "cache.age"
	Age           = "age"
	Strategy      = "strategy"
)

// Supported cache strategies.
const (
	CacheStrategyNever                      = "server"
	CacheStrategyServerThenCache            = "server.cache"
	CacheStrategyCacheThenServer            = "cache.server"
	CacheStrategyCacheThenServerThenExpired = "cache.server.expired"
)
