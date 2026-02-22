package consumer

import (
	"io"
	"net"

	"github.com/art-es/queue-service/internal/infra/log"
)

type consumeService interface {
	Consume(conn io.ReadWriteCloser)
}

type Server struct {
	consumeService consumeService
	logger         log.Logger
}

func NewServer(consumeService consumeService, logger log.Logger) *Server {
	return &Server{
		consumeService: consumeService,
		logger:         logger,
	}
}

func (s *Server) Serve(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			s.logger.Log(log.LevelError).
				With("message", "accept conn error").
				With("error", err.Error()).
				Write()
		}

		go s.consumeService.Consume(conn)
	}
}
