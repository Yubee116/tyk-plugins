package main

import (
	"context"
	"net/http"
	"time"

	"github.com/TykTechnologies/tyk/config"
	"github.com/TykTechnologies/tyk/ctx"
	"github.com/TykTechnologies/tyk/log"
	"github.com/TykTechnologies/tyk/storage"
)

// Global redis variables
var conf config.Config
var rc *storage.RedisController
var store storage.RedisCluster

const pluginKeyPrefix = "extract-og-rPath-plugin-"

var logger = log.Get()

func ExtractOriginalRequestPath(rw http.ResponseWriter, r *http.Request) {
	logger.Info("Plugin: Extract Original Request Path")

	// Get api definition. used later to get name of API called
	apidef := ctx.GetDefinition(r)

	// Get original request path from request
	original_request_path := r.RequestURI
	logger.Info("Original RequestURI:  ", original_request_path)

	// Store the original request path in redis with TTL of 5sec
	// API name included in Redis key to avoid collisions
	err := store.SetKey(apidef.Name, original_request_path, 5)
	if err != nil {
		logger.Info("There's an error storing key to redis", err)
		return
	}
}

func establishRedisConnection() {
	// Retrieve global configs
	conf = config.Global()

	// Create a Redis Controller, which will handle the Redis connection for the storage
	rc = storage.NewRedisController(context.Background())

	// Create a storage object, which will handle Redis operations using pluginKeyPrefix
	store = storage.RedisCluster{KeyPrefix: pluginKeyPrefix, RedisController: rc}

	// Perform Redis connection
	go rc.ConnectToRedis(context.Background(), nil, &conf)
	for i := 0; i < 5; i++ { // max 10 attempts - should only take 3
		time.Sleep(10 * time.Millisecond)
		if rc.Connected() {
			logger.Info("Redis Controller connected")
			break
		}
		logger.Warn("Redis Controller not connected, will retry")
	}

	// Error handling Redis connection
	if !rc.Connected() {
		logger.Error("Could not connect to storage")
		panic("Plugin Couldn't establish a connection to redis")
	}
}

func init() {
	logger.Info("---- Establishing redis connection in Request plugin ----")
	establishRedisConnection()
}

func main() {}
