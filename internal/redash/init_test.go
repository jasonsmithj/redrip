package redash

import (
	"github.com/jasonsmithj/redrip/internal/logger"
)

func init() {
	// Initialize a null logger for all tests
	logger.InitNullLogger()
}
