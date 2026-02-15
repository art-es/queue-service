package logimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	core "github.com/art-es/queue-service/internal/infra/log"
)

func TestLogger(t *testing.T) {
	const mockedCreated = time.DateTime

	withMockedCreatedFieldGetter := func() LoggerOption {
		return WithCreatedFieldGetter(func() string {
			return mockedCreated
		})
	}

	t.Run("disabled logger", func(t *testing.T) {
		buffer := core.NewBuffer()
		logger := NewLogger(WithOutput(buffer), WithLevel(core.LevelDisabled))
		logger.Log(core.LevelError).Write()
		assert.Empty(t, buffer.Logs())
	})

	t.Run("disabled log", func(t *testing.T) {
		buffer := core.NewBuffer()
		logger := NewLogger(WithOutput(buffer), WithLevel(core.LevelDisabled))
		logger.Log(core.LevelDisabled).Write()
		assert.Empty(t, buffer.Logs())
	})

	t.Run("hidden level", func(t *testing.T) {
		buffer := core.NewBuffer()
		logger := NewLogger(WithOutput(buffer), WithLevel(core.LevelInfo))
		logger.Log(core.LevelDebug).Write()
		assert.Empty(t, buffer.Logs())
	})

	t.Run("ok level", func(t *testing.T) {
		buffer := core.NewBuffer()
		logger := NewLogger(WithOutput(buffer), WithLevel(core.LevelInfo), withMockedCreatedFieldGetter())
		logger.Log(core.LevelInfo).Write()
		logs := buffer.Logs()
		assert.Len(t, logs, 1)
		assert.JSONEq(t, `{"created":"2006-01-02 15:04:05","level":"info"}`, logs[0])
	})

	t.Run("logger field", func(t *testing.T) {
		buffer := core.NewBuffer()
		logger := NewLogger(WithOutput(buffer), WithLevel(core.LevelDebug), withMockedCreatedFieldGetter())
		logger = logger.With("foo", "bar")
		logger.Log(core.LevelInfo).Write()
		logs := buffer.Logs()
		assert.Len(t, logs, 1)
		assert.JSONEq(t, `{"created":"2006-01-02 15:04:05","level":"info","foo":"bar"}`, logs[0])
	})

	t.Run("log field", func(t *testing.T) {
		buffer := core.NewBuffer()
		logger := NewLogger(WithOutput(buffer), WithLevel(core.LevelDebug), withMockedCreatedFieldGetter())
		logobj := logger.Log(core.LevelInfo)
		logobj = logobj.With("foo", "bar")
		logobj.Write()
		logs := buffer.Logs()
		assert.Len(t, logs, 1)
		assert.JSONEq(t, `{"created":"2006-01-02 15:04:05","level":"info","foo":"bar"}`, logs[0])
	})

	t.Run("log field overloads logger field", func(t *testing.T) {
		buffer := core.NewBuffer()
		logger := NewLogger(WithOutput(buffer), WithLevel(core.LevelDebug), withMockedCreatedFieldGetter())
		logger = logger.With("foo", "bar")
		logobj := logger.Log(core.LevelInfo)
		logobj = logobj.With("foo", "baz")
		logobj.Write()
		logs := buffer.Logs()
		assert.Len(t, logs, 1)
		assert.JSONEq(t, `{"created":"2006-01-02 15:04:05","level":"info","foo":"baz"}`, logs[0])
	})
}
