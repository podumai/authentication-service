package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

type Field struct {
	Key   string
	Value any
}

type Logger interface {
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Panic(msg string, fields ...Field)
}

type ZerologLogger struct {
	logger *zerolog.Logger
}

func (l *ZerologLogger) Info(msg string, fields ...Field) {
	event := l.logger.Info()
	for _, field := range fields {
		event.Any(field.Key, field.Value)
	}
	event.Msg(msg)
}

func (l *ZerologLogger) Warn(msg string, fields ...Field) {
	event := l.logger.Warn()
	for _, field := range fields {
		event.Any(field.Key, field.Value)
	}
	event.Msg(msg)
}

func (l *ZerologLogger) Debug(msg string, fields ...Field) {
	event := l.logger.Debug()
	for _, field := range fields {
		event.Any(field.Key, field.Value)
	}
	event.Msg(msg)
}

func (l *ZerologLogger) Error(msg string, fields ...Field) {
	event := l.logger.Error()
	for _, field := range fields {
		event.Any(field.Key, field.Value)
	}
	event.Msg(msg)
}

func (l *ZerologLogger) Fatal(msg string, fields ...Field) {
	event := l.logger.Fatal()
	for _, field := range fields {
		event.Any(field.Key, field.Value)
	}
	event.Msg(msg)
}

func (l *ZerologLogger) Panic(msg string, fields ...Field) {
	event := l.logger.Panic()
	for _, field := range fields {
		event.Any(field.Key, field.Value)
	}
	event.Msg(msg)
}

func NewLogger(level string, output io.Writer) *ZerologLogger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(lvl)

	if output == nil {
		output = os.Stderr
	}

	l := zerolog.New(output).With().Timestamp().Logger()
	return &ZerologLogger{logger: &l}
}
