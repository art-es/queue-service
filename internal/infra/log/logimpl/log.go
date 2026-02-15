package logimpl

import (
	"encoding/json"
	"io"

	core "github.com/art-es/queue-service/internal/infra/log"
)

var _ core.Logger = (*logger)(nil)

type loggerOptionFunc func(logger *logger)

type field struct {
	key   string
	value string
}

type fields []field

type logger struct {
	output             io.Writer
	level              core.Level
	createdFieldGetter func() string
	fields             fields
}

type disabledLogger struct{}

type log struct {
	logger logger
	level  core.Level
}

type hiddenLog struct{}

func (f loggerOptionFunc) apply(logger *logger) { f(logger) }

func (ff fields) set(key, value string) fields {
	if value == "" {
		return ff
	}

	exists := false
	out := make(fields, 0, len(ff)+1)

	for _, f := range ff {
		if f.key == key {
			exists = true
			f.value = value
		}
		out = append(out, f)
	}

	if !exists {
		out = append(out, field{key: key, value: value})
	}

	return out
}

func (l log) With(key, value string) core.Log {
	l.logger.fields = l.logger.fields.set(key, value)

	return l
}

func (l log) Write() {
	lmap := make(map[string]string)
	for _, f := range l.logger.fields {
		lmap[f.key] = f.value
	}

	lmap[core.KeyLevel] = l.level.String()
	lmap[core.KeyCreated] = l.logger.createdFieldGetter()

	_ = json.NewEncoder(l.logger.output).Encode(lmap)
}

func (l logger) With(key, value string) core.Logger {
	l.fields = l.fields.set(key, value)

	return l
}

func (l logger) Log(level core.Level) core.Log {
	if level > l.level || level == core.LevelDisabled {
		return &hiddenLog{}
	}

	return &log{logger: l, level: level}
}

func (*hiddenLog) With(key, value string) core.Log { return nil }

func (*hiddenLog) Write() {}

func (l disabledLogger) With(key, value string) core.Logger { return l }

func (l disabledLogger) Log(level core.Level) core.Log { return &hiddenLog{} }
