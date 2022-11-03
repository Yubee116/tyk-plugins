package main

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/TykTechnologies/tyk/log"
)

var logger = log.Get()

// CompressResponse compresses response body using GZIP algorithm
func CompressResponse(rw http.ResponseWriter, res *http.Response, req *http.Request) {

	logger.Info("Response Compression Plugin Applied")
	logger.Info("Compression Algorithm:  Gzip")

	// get response from upstream
	respBody := res.Body
	body, _ := ioutil.ReadAll(respBody)
	defer respBody.Close()

	// apply gzip compression
	logger.Info("Compressing Response...")
	var compressedBody bytes.Buffer
	w := gzip.NewWriter(&compressedBody)
	w.Write([]byte(body))
	w.Close()

	// update & add response headers
	res.ContentLength = int64(compressedBody.Len())
	res.Header.Set("Content-Length", strconv.Itoa(compressedBody.Len()))
	res.Header.Set("Content-Encoding", "gzip")

	// replace upstream response body with compressed
	res.Body = ioutil.NopCloser(&compressedBody)
	logger.Info("Returned Compressed Response...")

}

func main() {}
