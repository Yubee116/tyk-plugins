package main

import (
	"net/http"
	"time"

	"github.com/TykTechnologies/tyk/config"
	"github.com/TykTechnologies/tyk/ctx"
	"github.com/TykTechnologies/tyk/log"
	"github.com/TykTechnologies/tyk/storage"
)

var logger = log.Get()

func ExtractOriginalRequestPath(rw http.ResponseWriter, r *http.Request) {
	logger.Info("Processing Request Golang plugin")

	// get api definition to get name to prefix keys stored in Redis
	apidef := ctx.GetDefinition(r)
	keyprefix := apidef.Name + "-"

	// Get the global config - it's needed in various places
	conf := config.Global()

	// Create a Redis Controller, which will handle the Redis connection for the storage
	rc := storage.NewRedisController(r.Context())

	// Create a storage object, which will handle Redis operations using key prefix
	store := storage.RedisCluster{KeyPrefix: keyprefix, RedisController: rc}

	go rc.ConnectToRedis(r.Context(), nil, &conf)
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
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get original request path from request
	original_request_path := r.RequestURI
	logger.Info("Original RequestURI:  ", original_request_path)

	// Store the original request path in redis with TTL of 120sec
	err := store.SetKey("original_request_path", original_request_path, 120)
	if err != nil {
		logger.Info("There's an error storing key to redis", err)
		return
	}
}
func main() {}
