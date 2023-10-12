package main

import (
	"context"
	"time"

	"github.com/TykTechnologies/tyk-pump/analytics"
	"github.com/TykTechnologies/tyk/config"
	"github.com/TykTechnologies/tyk/log"
	"github.com/TykTechnologies/tyk/storage"
)

// Global redis variables
var conf config.Config
var rc *storage.RedisController
var store storage.RedisCluster

// this is set based on the other plugin that sets request path
const pluginKeyPrefix = "extract-og-rPath-plugin-"

var logger = log.Get()

func ModifyAnalyticsRequestPath(record *analytics.AnalyticsRecord) {
	logger.Info("Plugin: Modify Analytics Request Path")

	// Get original request path from Redis using API name
	original_request_path, err := store.GetKey(record.APIName)
	if err != nil {
		logger.Info("There's an error getting key from redis", err)
		return
	}

	// Add original request path to tags
	tags := record.Tags
	tags = append(tags, original_request_path)
	record.Tags = tags

	// Add original request path to request metadata
	record.RawPath += " rewritten from " + original_request_path

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
	for i := 0; i < 10; i++ { // max 10 attempts - should only take 3
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
	logger.Info("---- Establishing redis connection in Analytics plugin ----")
	establishRedisConnection()
}

func main() {}
