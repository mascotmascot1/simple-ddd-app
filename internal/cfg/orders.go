package cfg

import (
	"log"
	"os"
	"strconv"
)

const (
	// -------------ENVS----------------------------
	envOrdersHTTPMaxUploadSize = "ORDERS_HTTP_MAXUPLOADSIZE" // UNNECESSARY

	// ------------default-values------------------
	defaultOrdersHTTPMaxUploadSize int64 = 1024 // KB
)

type Orders struct {
	MaxUploadSize int64
}

func newOrders(logger *log.Logger) Orders {
	uploadSize := defaultOrdersHTTPMaxUploadSize << 10

	eMaxUploadSize := os.Getenv(envOrdersHTTPMaxUploadSize)
	if eMaxUploadSize != "" {
		mus, err := strconv.ParseInt(eMaxUploadSize, 10, 64)
		if err != nil {
			logger.Printf("[WARNING] standard %q (%d KB) for http server was applied: %v\n",
				envOrdersHTTPMaxUploadSize, defaultOrdersHTTPMaxUploadSize, err)
		} else {
			uploadSize = mus << 10
		}
	}

	return Orders{
		MaxUploadSize: uploadSize,
	}
}
