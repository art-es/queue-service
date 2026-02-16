package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/art-es/queue-service/internal/app/services/queue"
	"github.com/art-es/queue-service/internal/app/services/task"
	"github.com/art-es/queue-service/internal/cache/inmemory"
	"github.com/art-es/queue-service/internal/infra/clock"
	"github.com/art-es/queue-service/internal/infra/initial"
	"github.com/art-es/queue-service/internal/infra/log"
	"github.com/art-es/queue-service/internal/infra/log/logimpl"
	"github.com/art-es/queue-service/internal/repository/psql"
	psqltask "github.com/art-es/queue-service/internal/repository/psql/task"
	httpadapter "github.com/art-es/queue-service/internal/transport/http/adapter"
	httpendpoints "github.com/art-es/queue-service/internal/transport/http/endpoints"
)

var (
	appCtx, appCtxCancel = context.WithCancel(context.Background())

	logger     log.Logger
	baseLogger log.Logger
	httpServer *http.Server

	// Connections (should be closed)
	serviceListener net.Listener
	psqlConn        psql.Conn
)

func main() {
	baseLogger = logimpl.NewLogger(initial.GetLogOptions()...)
	logger = logger.With("module", "cmd/service")

	if err := setup(); err != nil {
		logger.Log(log.LevelError).
			With("message", "service setup error").
			With("error", err.Error()).
			Write()

		teardown()
	}

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		<-signalChan
		teardown()
	}()

	logger.Log(log.LevelInfo).
		With("message", "service started").
		With("addr", serviceListener.Addr().String()).
		Write()

	_ = httpServer.Serve(serviceListener)
}

func setup() error {
	var serviceAddr string
	var psqlSource string

	err := initial.ParseEnv(
		initial.Env{Name: "SERVICE_ADDR", Target: &serviceAddr, Required: true},
		initial.Env{Name: "PSQL_SOURCE", Target: &psqlSource, Required: true},
	)
	if err != nil {
		return err
	}

	serviceListener, err = net.Listen("tcp", serviceAddr)
	if err != nil {
		return fmt.Errorf("listen service addr: %w", err)
	}

	psqlConn, err = psql.Connect(psqlSource, baseLogger)
	if err != nil {
		return fmt.Errorf("psql connect: %w", err)
	}

	psqlExecGetter := psql.NewExecGetter(psqlConn)
	taskRepository := psqltask.NewRepository(psqlExecGetter)
	clockObj := clock.NewClock()
	idempotencyKeyCache := inmemory.NewIdempotencyKeyCache()

	queueService := queue.NewService(clockObj, idempotencyKeyCache, taskRepository, baseLogger)
	taskService := task.NewService(clockObj, idempotencyKeyCache, taskRepository, baseLogger)

	httpRouter := httpadapter.NewMuxRouter()
	httpendpoints.RegisterV1QueuesPop(httpRouter, queueService, baseLogger)
	httpendpoints.RegisterV1QueuesPush(httpRouter, queueService, baseLogger)
	httpendpoints.RegisterV1TasksAck(httpRouter, taskService, baseLogger)
	httpendpoints.RegisterV1TasksNack(httpRouter, taskService, baseLogger)
	httpServer = &http.Server{
		Handler: httpRouter.Mux,
		BaseContext: func(l net.Listener) context.Context {
			return appCtx
		},
	}

	return nil
}

func teardown() {
	appCtxCancel()

	if serviceListener != nil {
		if err := serviceListener.Close(); err != nil {
			logger.Log(log.LevelError).
				With("message", "service listener close error").
				With("error", err.Error()).
				Write()
		}
	}

	if psqlConn != nil {
		if err := psqlConn.Close(); err != nil {
			logger.Log(log.LevelError).
				With("message", "psql conn close error").
				With("error", err.Error()).
				Write()
		}
	}
}
