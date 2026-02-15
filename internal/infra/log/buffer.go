package log

import (
	"bufio"
	"bytes"
)

type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() Buffer {
	return Buffer{
		Buffer: &bytes.Buffer{},
	}
}

func (b Buffer) Logs() []string {
	var logs []string
	for s := bufio.NewScanner(b); s.Scan(); {
		logs = append(logs, s.Text())
	}
	return logs
}
