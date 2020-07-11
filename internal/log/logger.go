package log

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})

	With(args ...interface{}) Logger
	WithErr(err error) Logger

	Named(name string) Logger
}

type key string

var contextKey = key("com.fildr")

type sugaredLogWrapper struct {
	*zap.SugaredLogger
}

var _ Logger = (*sugaredLogWrapper)(nil)

func (s *sugaredLogWrapper) WithErr(err error) Logger {
	return &sugaredLogWrapper{s.SugaredLogger.With("err", err.Error())}
}

func (s *sugaredLogWrapper) With(args ...interface{}) Logger {
	return &sugaredLogWrapper{s.SugaredLogger.With(args...)}
}

func (s *sugaredLogWrapper) Named(name string) Logger {
	return &sugaredLogWrapper{s.SugaredLogger.Named(name)}
}

func Wrap(z *zap.SugaredLogger) Logger {
	return &sugaredLogWrapper{z}
}

func NopLogger() Logger {
	return Wrap(zap.NewNop().Sugar())
}

func WithLoggerContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

func From(ctx context.Context) Logger {
	if ctx == nil {
		return NopLogger()
	}
	v := ctx.Value(contextKey)
	l, ok := v.(Logger)
	if !ok || l == nil {
		return NopLogger()
	}
	return l
}

type InitOption func(config zap.Config) zap.Config

func Init(logLevel int, options ...InitOption) (*zap.Logger, error) {
	z, err := newZapLogger(logLevel, options...)
	if err != nil {
		return nil, fmt.Errorf("create zap logger: %w", err)
	}

	return z, nil

}

func newZapLogger(verboseLevel int, options ...InitOption) (*zap.Logger, error) {
	level := zapcore.InfoLevel - zapcore.Level(verboseLevel)
	if level < zapcore.DebugLevel || level > zapcore.FatalLevel {
		level = zapcore.DebugLevel
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	for _, option := range options {
		cfg = option(cfg)
	}

	return cfg.Build()
}
