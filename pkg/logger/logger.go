package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey string

const (
	loggerRequestIDKey ctxKey = "x-request-id"
	loggerTraceIDKey   ctxKey = "x-trace-id"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Sync()
}

type L struct {
	z *zap.Logger
}

func NewLogger(env string) Logger {
	var loggerCfg zap.Config

	switch env {
	case "prod":
		loggerCfg = zap.NewProductionConfig()
	case "test":
		loggerCfg = zap.NewProductionConfig()
	default: // dev
		loggerCfg = zap.NewDevelopmentConfig()
	}

	if env == "dev" {
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := loggerCfg.Build()
	if err != nil {
		panic(err)
	}

	return &L{z: logger}
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, loggerRequestIDKey, requestID)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, loggerTraceIDKey, traceID)
}

func (l *L) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Info(msg, l.appendCtxFields(ctx, fields...)...)
}

func (l *L) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Error(msg, l.appendCtxFields(ctx, fields...)...)
}

func (l *L) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Debug(msg, l.appendCtxFields(ctx, fields...)...)
}

func (l *L) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Warn(msg, l.appendCtxFields(ctx, fields...)...)
}

func (l *L) Sync() {
	_ = l.z.Sync()
}

func (l *L) appendCtxFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	requestID, _ := ctx.Value(loggerRequestIDKey).(string)
	traceID, _ := ctx.Value(loggerTraceIDKey).(string)

	fields = append(
		fields,
		zap.String(string(loggerRequestIDKey), requestID),
		zap.String(string(loggerTraceIDKey), traceID),
	)

	return fields
}
