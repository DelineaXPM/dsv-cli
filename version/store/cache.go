package store

// CacheStrategy is the cache strategy
type CacheStrategy string

// Supported cache strategies
const (
	Never                      = CacheStrategy("server")
	ServerThenCache            = CacheStrategy("server.cache")
	CacheThenServer            = CacheStrategy("cache.server")
	CacheThenServerThenExpired = CacheStrategy("cache.server.expired")
)

// CacheAge is age in minutes to cache for
type CacheAge int
