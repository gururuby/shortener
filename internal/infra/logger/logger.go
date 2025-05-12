package logger

import (
	"go.uber.org/zap"
	"log"
	"sync"
)

var Log *zap.Logger

func Setup(appENV, logLevel string) {
	var initLogger sync.Once

	initLogger.Do(func() {
		var cfg zap.Config
		var err error

		switch appENV {
		case "production":
			cfg = zap.NewProductionConfig()
		default:
			cfg = zap.NewDevelopmentConfig()
		}

		cfg.Level = buildLogLevel(logLevel)

		if Log, err = cfg.Build(); err != nil {
			log.Fatalf("cannot init logger: %s", err)
		}
	})
}

func buildLogLevel(logLevel string) zap.AtomicLevel {
	var lvl zap.AtomicLevel

	switch logLevel {
	case "debug":
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		lvl = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		lvl = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return lvl
}
