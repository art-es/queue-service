package subscriber

import (
	"context"
	"net"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/infra/log"
)

type queueService interface {
	Subscribe(ctx context.Context, queueName string) (<-chan *domain.Task, error)
}

type taskService interface {
	Ack(ctx context.Context, taskID string) error
	Nack(ctx context.Context, taskID string) error
}

type Server struct {
	queueService queueService
	taskService  taskService
	logger       log.Logger
	baseCtx      context.Context
}

func NewServer(
	queueService queueService,
	taskService taskService,
	logger log.Logger,
	baseCtx context.Context,
) *Server {
	logger = logger.With("module", "internal/transport/net/subscriber")

	return &Server{
		queueService: queueService,
		taskService:  taskService,
		logger:       logger,
		baseCtx:      baseCtx,
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

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	ctx, cancel := context.WithCancel(s.baseCtx)
	done := func() {
		cancel()
		conn.Close()
	}

	s.logger.Log(log.LevelDebug).
		With("message", "new conn").
		With("ip", conn.RemoteAddr().String()).
		Write()

	writeChan := make(chan []any, 1)
	go s.writeMessages(ctx, conn, writeChan, done)
	s.readMessages(ctx, conn, writeChan, done)
}
