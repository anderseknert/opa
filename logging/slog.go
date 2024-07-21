package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

type SlogLogger struct {
	logger *slog.Logger
	level  *slog.LevelVar
}

type SlogLoggerOptions struct {
	Sink            *io.Writer
	Level           string
	Format          string
	TimestampFormat string
}

func NewSlogLogger(options SlogLoggerOptions) *SlogLogger {
	var sink io.Writer
	if options.Sink != nil {
		sink = *options.Sink
	} else {
		sink = os.Stderr
	}

	var timestampFormat string
	if options.TimestampFormat != "" {
		timestampFormat = options.TimestampFormat
	} else {
		timestampFormat = os.Getenv("OPA_LOG_TIMESTAMP_FORMAT")
	}

	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.LevelKey:
			level := a.Value.Any().(slog.Level)
			a.Value = slog.StringValue(strings.ToLower(level.String()))
		case slog.TimeKey:
			t := a.Value.Any().(time.Time)
			a.Value = slog.StringValue(t.Format(timestampFormat))
		}

		return a
	}

	level := new(slog.LevelVar)

	// on error, the default level of info will be used, which seems reasonable, as
	// at this point the args have been validated by command line parsing anyway
	_ = level.UnmarshalText([]byte(options.Level))

	handlerOptions := &slog.HandlerOptions{ReplaceAttr: replaceAttr, Level: level}

	var handler slog.Handler
	switch options.Format {
	case "text":
		// TODO
		handler = slog.NewTextHandler(sink, handlerOptions)
	case "json-pretty":
		handler = NewJSONPrettyHandler(sink, handlerOptions)
	default:
		handler = slog.NewJSONHandler(sink, handlerOptions)
	}

	return &SlogLogger{logger: slog.New(handler), level: level}
}

func slogLevelToLevel(level slog.Level) Level {
	switch level {
	case slog.LevelDebug:
		return Debug
	case slog.LevelInfo:
		return Info
	case slog.LevelWarn:
		return Warn
	case slog.LevelError:
		return Error
	default:
		return Info
	}
}

func levelToSlogLevel(level Level) slog.Level {
	switch level {
	case Debug:
		return slog.LevelDebug
	case Info:
		return slog.LevelInfo
	case Warn:
		return slog.LevelWarn
	case Error:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *SlogLogger) Debug(fmt string, a ...any) {
	l.logger.Debug(fmt, a...)
}

func (l *SlogLogger) Info(fmt string, a ...any) {
	l.logger.Info(fmt, a...)
}

func (l *SlogLogger) Error(fmt string, a ...any) {
	l.logger.Error(fmt, a...)
}

func (l *SlogLogger) Warn(fmt string, a ...any) {
	l.logger.Warn(fmt, a...)
}

func (l *SlogLogger) WithFields(fields map[string]any) Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	cp := *l
	cp.logger = cp.logger.With(args...)

	return &cp
}

func (l *SlogLogger) GetLevel() Level {
	return slogLevelToLevel(l.level.Level())
}

func (l *SlogLogger) SetLevel(level Level) {
	l.level.Set(levelToSlogLevel(level))
}

type JSONPrettyHandler struct {
	w io.Writer
}

func (J JSONPrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	//TODO implement me
	panic("implement me")
}

func (J JSONPrettyHandler) Handle(_ context.Context, record slog.Record) error {
	fmt.Println(record)

	return nil
}

func (J JSONPrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// TODO implement me
	panic("implement me")
}

func (J JSONPrettyHandler) WithGroup(name string) slog.Handler {
	//TODO implement me
	panic("implement me")
}

func NewJSONPrettyHandler(w io.Writer, options *slog.HandlerOptions) slog.Handler {
	return &JSONPrettyHandler{}
}
