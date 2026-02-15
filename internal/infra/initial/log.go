package initial

import (
	"os"
	"strings"

	"github.com/art-es/queue-service/internal/infra/log"
	"github.com/art-es/queue-service/internal/infra/log/logimpl"
)

func GetLogOptions() []logimpl.LoggerOption {
	var logLevel log.Level
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "disabled":
		logLevel = log.LevelDisabled
	case "debug":
		logLevel = log.LevelDebug
	case "warning":
		logLevel = log.LevelWarning
	case "error":
		logLevel = log.LevelError
	default:
		logLevel = log.LevelInfo
	}

	return []logimpl.LoggerOption{
		logimpl.WithLevel(logLevel),
	}
}
