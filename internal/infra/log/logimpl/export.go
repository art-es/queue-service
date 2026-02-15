package logimpl

import (
	"io"
	"os"
	"time"

	core "github.com/art-es/queue-service/internal/infra/log"
)

type LoggerOption interface {
	apply(logger *logger)
}

func WithOutput(output io.Writer) LoggerOption {
	return loggerOptionFunc(func(logger *logger) {
		logger.output = output
	})
}

func WithLevel(level core.Level) LoggerOption {
	return loggerOptionFunc(func(logger *logger) {
		logger.level = level
	})
}

func WithCreatedFieldGetter(getter func() string) LoggerOption {
	return loggerOptionFunc(func(logger *logger) {
		logger.createdFieldGetter = getter
	})
}

func NewLogger(opts ...LoggerOption) core.Logger {
	out := logger{
		output: os.Stdout,
		level:  core.LevelInfo,
		createdFieldGetter: func() string {
			return time.Now().Format(time.DateTime)
		},
	}

	for _, opt := range opts {
		opt.apply(&out)
	}

	if out.level == core.LevelDisabled {
		return &disabledLogger{}
	}

	return out
}

func NewTestLogger() (core.Logger, core.Buffer) {
	buf := core.NewBuffer()
	out := NewLogger(
		WithOutput(buf),
		WithLevel(core.LevelDebug),
		WithCreatedFieldGetter(func() string {
			return time.DateTime
		}),
	)

	return out, buf
}
