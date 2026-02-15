package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuffer(t *testing.T) {
	buf := NewBuffer()
	buf.WriteString("first log\nsecond log\nthird log")

	logs := buf.Logs()
	require.Len(t, logs, 3)
	assert.Equal(t, []string{"first log", "second log", "third log"}, logs)
}
