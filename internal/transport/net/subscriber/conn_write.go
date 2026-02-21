package subscriber

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/art-es/queue-service/internal/infra/log"
)

func (s *Server) writeMessages(ctx context.Context, conn net.Conn, writeChan <-chan []any, done func()) {
	defer done()

	var msgs []any
	for {
		select {
		case <-ctx.Done():
			return
		case msgs = <-writeChan:
		}

		for _, msg := range msgs {
			if err := write(conn, msg); err != nil {
				if !errors.Is(err, io.EOF) {
					s.logger.Log(log.LevelError).
						With("message", "write message error").
						With("error", err.Error()).
						Write()
				}

				return
			}
		}
	}
}
