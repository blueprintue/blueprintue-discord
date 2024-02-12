package logger

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/blueprintue/discord-bot/configuration"
	"github.com/ilya1st/rotatewriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const formatTimestampLogger = "2006-01-02T15:04:05.000000000Z07:00"

// Configure configures logger.
func Configure(confLog configuration.Log) error {
	var err error

	logFile := path.Clean(confLog.Filename)
	if err := os.MkdirAll(path.Dir(logFile), os.ModePerm); err != nil {
		log.Error().Err(err).Msg("Cannot create log folder")

		return fmt.Errorf("%w", err)
	}

	rwriter, err := rotatewriter.NewRotateWriter(logFile, confLog.NumberFilesRotation)
	if err != nil {
		log.Error().Err(err).Msg("Cannot create log file writer")

		return fmt.Errorf("%w", err)
	}

	sighupChan := make(chan os.Signal, 1)

	signal.Notify(sighupChan, syscall.SIGHUP)

	go func() {
		for {
			_, ok := <-sighupChan
			if !ok {
				return
			}

			if err := rwriter.Rotate(nil); err != nil {
				log.Error().Err(err).Msg("Cannot rotate log")
			}
		}
	}()

	logLevel, err := zerolog.ParseLevel(confLog.Level)
	if err != nil {
		log.Error().Err(err).Msg("Unknown log level")

		return fmt.Errorf("%w", err)
	}

	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimeFieldFormat = formatTimestampLogger

	log.Logger = zerolog.New(
		zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC1123,
			}, rwriter),
	).With().Timestamp().Caller().Logger()

	return nil
}
