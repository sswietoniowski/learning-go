package common

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

type EchoSlogAdapter struct {
	logger *slog.Logger
	level  log.Lvl
	prefix string
	output io.Writer
}

func NewEchoSlogAdapter(logger *slog.Logger) *EchoSlogAdapter {
	return &EchoSlogAdapter{
		logger: logger,
		level:  log.INFO,
		output: os.Stdout,
	}
}

func (e *EchoSlogAdapter) Output() io.Writer {
	return e.output
}

func (e *EchoSlogAdapter) SetOutput(w io.Writer) {
	e.output = w
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: e.echoLevelToSlog(e.level),
	})
	e.logger = slog.New(handler)
}

func (e *EchoSlogAdapter) Prefix() string {
	return e.prefix
}

func (e *EchoSlogAdapter) SetPrefix(p string) {
	e.prefix = p
}

func (e *EchoSlogAdapter) Level() log.Lvl {
	return e.level
}

func (e *EchoSlogAdapter) SetLevel(v log.Lvl) {
	e.level = v
}

func (e *EchoSlogAdapter) SetHeader(h string) {
	// slog doesn't support custom headers, ignore
}

func (e *EchoSlogAdapter) echoLevelToSlog(level log.Lvl) slog.Level {
	switch level {
	case log.DEBUG:
		return slog.LevelDebug
	case log.INFO:
		return slog.LevelInfo
	case log.WARN:
		return slog.LevelWarn
	case log.ERROR:
		return slog.LevelError
	case log.OFF:
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

func (e *EchoSlogAdapter) shouldLog(level log.Lvl) bool {
	return level >= e.level
}

func (e *EchoSlogAdapter) logWithPrefix(msg string) string {
	if e.prefix != "" {
		return e.prefix + " " + msg
	}
	return msg
}

func (e *EchoSlogAdapter) Print(i ...interface{}) {
	if !e.shouldLog(log.INFO) {
		return
	}
	msg := fmt.Sprint(i...)
	e.logger.Info(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Printf(format string, args ...interface{}) {
	if !e.shouldLog(log.INFO) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	e.logger.Info(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Printj(j log.JSON) {
	if !e.shouldLog(log.INFO) {
		return
	}
	attrs := e.jsonToAttrs(j)
	e.logger.LogAttrs(context.Background(), slog.LevelInfo, e.logWithPrefix(""), attrs...)
}

func (e *EchoSlogAdapter) Debug(i ...interface{}) {
	if !e.shouldLog(log.DEBUG) {
		return
	}
	msg := fmt.Sprint(i...)
	e.logger.Debug(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Debugf(format string, args ...interface{}) {
	if !e.shouldLog(log.DEBUG) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	e.logger.Debug(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Debugj(j log.JSON) {
	if !e.shouldLog(log.DEBUG) {
		return
	}
	attrs := e.jsonToAttrs(j)
	e.logger.LogAttrs(context.Background(), slog.LevelDebug, e.logWithPrefix(""), attrs...)
}

func (e *EchoSlogAdapter) Info(i ...interface{}) {
	if !e.shouldLog(log.INFO) {
		return
	}
	msg := fmt.Sprint(i...)
	e.logger.Info(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Infof(format string, args ...interface{}) {
	if !e.shouldLog(log.INFO) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	e.logger.Info(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Infoj(j log.JSON) {
	if !e.shouldLog(log.INFO) {
		return
	}
	attrs := e.jsonToAttrs(j)
	e.logger.LogAttrs(context.Background(), slog.LevelInfo, e.logWithPrefix(""), attrs...)
}

func (e *EchoSlogAdapter) Warn(i ...interface{}) {
	if !e.shouldLog(log.WARN) {
		return
	}
	msg := fmt.Sprint(i...)
	e.logger.Warn(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Warnf(format string, args ...interface{}) {
	if !e.shouldLog(log.WARN) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	e.logger.Warn(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Warnj(j log.JSON) {
	if !e.shouldLog(log.WARN) {
		return
	}
	attrs := e.jsonToAttrs(j)
	e.logger.LogAttrs(context.Background(), slog.LevelWarn, e.logWithPrefix(""), attrs...)
}

func (e *EchoSlogAdapter) Error(i ...interface{}) {
	if !e.shouldLog(log.ERROR) {
		return
	}
	msg := fmt.Sprint(i...)
	e.logger.Error(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Errorf(format string, args ...interface{}) {
	if !e.shouldLog(log.ERROR) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	e.logger.Error(e.logWithPrefix(msg))
}

func (e *EchoSlogAdapter) Errorj(j log.JSON) {
	if !e.shouldLog(log.ERROR) {
		return
	}
	attrs := e.jsonToAttrs(j)
	e.logger.LogAttrs(context.Background(), slog.LevelError, e.logWithPrefix(""), attrs...)
}

func (e *EchoSlogAdapter) Fatal(i ...interface{}) {
	msg := fmt.Sprint(i...)
	e.logger.Error(e.logWithPrefix(msg))
	os.Exit(1)
}

func (e *EchoSlogAdapter) Fatalj(j log.JSON) {
	attrs := e.jsonToAttrs(j)
	e.logger.LogAttrs(context.Background(), slog.LevelError, e.logWithPrefix("fatal"), attrs...)
	os.Exit(1)
}

func (e *EchoSlogAdapter) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	e.logger.Error(e.logWithPrefix(msg))
	os.Exit(1)
}

func (e *EchoSlogAdapter) Panic(i ...interface{}) {
	msg := fmt.Sprint(i...)
	e.logger.Error(e.logWithPrefix(msg))
	panic(msg)
}

func (e *EchoSlogAdapter) Panicj(j log.JSON) {
	attrs := e.jsonToAttrs(j)
	e.logger.LogAttrs(context.Background(), slog.LevelError, e.logWithPrefix("panic"), attrs...)
	panic(j)
}

func (e *EchoSlogAdapter) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	e.logger.Error(e.logWithPrefix(msg))
	panic(msg)
}

func (e *EchoSlogAdapter) jsonToAttrs(j log.JSON) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(j))
	for k, v := range j {
		attrs = append(attrs, slog.Any(k, v))
	}
	return attrs
}
