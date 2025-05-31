package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// GetLogger returns zap.Logger instance, but using singleton pattern creates only one reusable instace
// development config by default
func GetLogger() *zap.Logger {
	once.Do(func() {
		var err error
		logger, err = zap.NewDevelopment()
		if err != nil {
			panic("failed logger setup : " + err.Error())
		}

	})
	return logger
}
