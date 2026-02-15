package psql

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"

	"github.com/art-es/queue-service/internal/infra/log"
)

func Connect(source string, logger log.Logger) (Conn, error) {
	logger = logger.With("module", "internal/adapter/psql")

	for range 30 {
		db, err := sql.Open("postgres", source)
		if err != nil {
			logger.Log(log.LevelError).
				With("message", "connect error").
				With("error", err.Error()).
				Write()

			waitBetweenConnects()
			continue
		}

		if err := db.Ping(); err != nil {
			logger.Log(log.LevelError).
				With("message", "ping error").
				With("error", err.Error()).
				Write()

			waitBetweenConnects()
			continue
		}

		return newConnAdapter(db), nil
	}

	return nil, errors.New("reached max attempts to connect")
}

func waitBetweenConnects() {
	time.Sleep(time.Second)
}
