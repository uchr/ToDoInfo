package log

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

type Level zerolog.Level

const (
	ErrorLevel Level = Level(zerolog.ErrorLevel)
	WarnLevel  Level = Level(zerolog.WarnLevel)
	InfoLevel  Level = Level(zerolog.InfoLevel)
	DebugLevel Level = Level(zerolog.DebugLevel)
)

func Init(level Level) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.SetGlobalLevel(zerolog.Level(level))
}

func Error(err error) {
	log.Error().Err(errors.Wrap(err, "")).Send()
}

func Info(msg string) {
	log.Info().Msg(msg)
}

func Debug(msg string) {
	log.Debug().Msg(msg)
}
