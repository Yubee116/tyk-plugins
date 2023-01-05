package main

import (
	"context"
	"time"

	"github.com/TykTechnologies/tyk-pump/analytics"
	"github.com/TykTechnologies/tyk/config"
	"github.com/TykTechnologies/tyk/log"
	"github.com/TykTechnologies/tyk/storage"
)

var logger = log.Get()

func ModifyAnalyticsRequestPath(record *analytics.AnalyticsRecord) {
	logger.Info("Processing Analytics Golang plugin")

	// Create key prefix from api name
	keyprefix := record.APIName + "-"

	// Get the global config - it's needed in various places
	conf := config.Global()

	// Create a Redis Controller, which will handle the Redis connection for the storage
	rc := storage.NewRedisController(context.Background())

	// Create a storage object, which will handle Redis operations using key keyprefix
	store := storage.RedisCluster{KeyPrefix: keyprefix, RedisController: rc}

	go rc.ConnectToRedis(context.Background(), nil, &conf)
	for i := 0; i < 10; i++ { // max 10 attempts - should only take 3
		if rc.Connected() {
			logger.Info("Redis Controller connected")
			break
		}
		logger.Warn("Redis Controller not connected, will retry")

		time.Sleep(10 * time.Millisecond)
	}

	if !rc.Connected() {
		logger.Error("Could not connect to storage")
		return
	}

	// Get original request path from Redis
	original_request_path, err := store.GetKey("original_request_path")
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
func main() {}
