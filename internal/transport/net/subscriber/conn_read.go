package subscriber

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/infra/log"
)

func (s *Server) readMessages(ctx context.Context, conn net.Conn, writeChan chan<- []any, done func()) {
	defer done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := s.readMessage(ctx, conn, writeChan); err != nil {
			if !errors.Is(err, io.EOF) {
				s.logger.Log(log.LevelError).
					With("message", "read message error").
					With("error", err.Error()).
					Write()
			}

			return
		}
	}
}

func (s *Server) readMessage(ctx context.Context, conn net.Conn, writeChan chan<- []any) error {
	var msgType uint8
	if err := binary.Read(conn, binary.BigEndian, &msgType); err != nil {
		return err
	}

	switch msgType {
	case inTypeQueueSubscribe:
		return s.readSubscribeMessage(ctx, conn, writeChan)
	case inTypeTaskAck:
		return s.readTaskAck(ctx, conn, writeChan)
	case inTypeTaskNack:
		return s.readTaskNack(ctx, conn, writeChan)
	case inTypeConnClose:
		return conn.Close()
	default:
		return fmt.Errorf("unknown message type: %d", msgType)
	}
}

func (s *Server) readSubscribeMessage(ctx context.Context, conn net.Conn, writeChan chan<- []any) error {
	var queueNameBytes [sizeQueueName]byte
	if err := read(conn, &queueNameBytes); err != nil {
		return err
	}

	queueName := convertQueueNameToString(queueNameBytes)
	taskChan, err := s.queueService.Subscribe(ctx, queueName)
	if err != nil {
		s.logger.Log(log.LevelError).
			With("message", "queue subscribe error").
			With("queue_name", queueName).
			With("error", err.Error()).
			Write()

		writeChan <- []any{outTypeQueueSubscribeFail}
		return nil
	}

	s.logger.Log(log.LevelDebug).
		With("message", "subscribed to queue chan").
		With("queue_name", queueName).
		Write()

	writeChan <- []any{outTypeQueueSubscribePass}
	s.listenTaskChan(ctx, taskChan, writeChan)
	return nil
}

func (s *Server) readTaskAck(ctx context.Context, conn net.Conn, writeChan chan<- []any) error {
	var taskIDBytes [sizeTaskID]byte
	if err := read(conn, &taskIDBytes); err != nil {
		return err
	}

	taskID := convertTaskIDToString(taskIDBytes)
	if err := s.taskService.Ack(ctx, taskID); err != nil {
		s.logger.Log(log.LevelError).
			With("message", "task ack error").
			With("task_id", taskID).
			With("error", err.Error()).
			Write()

		writeChan <- []any{outTypeTaskAckFail}
		return nil
	}

	writeChan <- []any{outTypeTaskAckPass}
	return nil
}

func (s *Server) readTaskNack(ctx context.Context, conn net.Conn, writeChan chan<- []any) error {
	var taskIDBytes [sizeTaskID]byte
	if err := read(conn, &taskIDBytes); err != nil {
		return err
	}

	taskID := convertTaskIDToString(taskIDBytes)
	if err := s.taskService.Nack(ctx, taskID); err != nil {
		s.logger.Log(log.LevelError).
			With("message", "task nack error").
			With("task_id", taskID).
			With("error", err.Error()).
			Write()

		writeChan <- []any{outTypeTaskNackFail}
		return nil
	}

	writeChan <- []any{outTypeTaskNackPass}
	return nil
}

func (s *Server) listenTaskChan(ctx context.Context, taskChan <-chan *domain.Task, writeChan chan<- []any) {
	var task *domain.Task
	for {
		select {
		case <-ctx.Done():
			return
		case task = <-taskChan:
		}

		writeChan <- []any{
			outTypeTaskProcess,
			processTaskOut{
				ID:        convertTaskIDToBytes(task.ID),
				Payload:   convertTaskPayloadToBytes(task.Payload),
				CreatedAt: convertTaskCreatedAtToBytes(task.CreatedAt),
			},
		}
	}
}
