package log

const (
	KeyCreated = "created" // log's created time
	KeyLevel   = "level"   // log level
)

const (
	LevelDisabled Level = iota
	LevelError
	LevelWarning
	LevelInfo
	LevelDebug
)

type Level int8

func (l Level) String() string {
	switch l {
	case LevelDisabled:
		return "disabled"
	case LevelError:
		return "error"
	case LevelWarning:
		return "warning"
	case LevelInfo:
		return "info"
	case LevelDebug:
		return "debug"
	default:
		return "unspecified"
	}
}

type Logger interface {
	With(key, value string) Logger
	Log(level Level) Log
}

type Log interface {
	With(key, value string) Log
	Write()
}
