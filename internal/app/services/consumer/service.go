package consumer

import (
	"context"
	"errors"
	"io"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/services/consumer/dto"
	"github.com/art-es/queue-service/internal/infra/log"
)

type messageIO interface {
	Read(r io.Reader) (*dto.Message, error)
	Write(w io.Writer, m *dto.Message) error
}

type queueService interface {
	Subscribe(ctx context.Context, queueName string) (<-chan *domain.Task, error)
}

type taskService interface {
	Ack(ctx context.Context, taskID string) error
	Nack(ctx context.Context, taskID string) error
}

type Service struct {
	baseCtx context.Context
	io      messageIO
	handle  func(ctx context.Context, in *dto.Message, outChan chan<- *dto.Message)
	logger  log.Logger
}

func NewService(
	baseCtx context.Context,
	io messageIO,
	queueService queueService,
	taskService taskService,
	logger log.Logger,
) *Service {
	logger = logger.With("module", "internal/app/services/consumer")
	handler := newMessageHandler(queueService, taskService, logger)

	return &Service{
		baseCtx: baseCtx,
		io:      io,
		handle:  handler.handle,
		logger:  logger,
	}
}

func (s *Service) Consume(conn io.ReadWriteCloser) {
	ctx, ctxCancel := context.WithCancel(s.baseCtx)
	done := func() {
		ctxCancel()
		conn.Close()
	}
	defer done()

	ch := make(chan *dto.Message, 1)

	go func() {
		defer done()
		s.write(ctx, conn, ch)
	}()

	s.read(ctx, conn, ch)
}

func (s *Service) read(ctx context.Context, r io.Reader, ch chan<- *dto.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := s.io.Read(r)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				s.logger.Log(log.LevelError).
					With("message", "read message error").
					With("error", err.Error()).
					Write()
			}
			return
		}

		if msg != nil {
			s.handle(ctx, msg, ch)
		}
	}
}

func (h *Service) write(ctx context.Context, w io.Writer, ch <-chan *dto.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			h.io.Write(w, msg)
		}
	}
}
