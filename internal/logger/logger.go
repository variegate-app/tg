package logger

import (
	"context"
	"log"
	"os"
	"sync"

	"log/slog"
)

var (
	Info  slog.Level = slog.LevelInfo
	Debug slog.Level = slog.LevelDebug
	Warn  slog.Level = slog.LevelWarn
	Error slog.Level = slog.LevelError
)

type fieldsKey struct{}
type Fields map[string]interface{}

func (zf Fields) Append(fields ...Field) Fields {
	zfCopy := make(Fields, len(zf)+len(fields))
	for k, v := range zf {
		zfCopy[k] = v
	}
	for _, f := range fields {
		zfCopy[f.Key] = f.Value
	}
	return zfCopy
}

type Field struct {
	Key   string
	Value interface{}
}

type Logger struct {
	mu     sync.Mutex
	logger *slog.Logger
	fields Fields
	level  slog.Level
}

// NewLogger создает новый логгер.
//
// level - уровень логирования.
//
// Возвращает логгер и ошибку, если возникла ошибка при создании логгера.
func NewLogger(l slog.Level) *Logger {
	return &Logger{
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
		fields: make(Fields),
		level:  l,
	}
}

func (z *Logger) WithContextFields(ctx context.Context, fields ...Field) context.Context {
	ctxFields, _ := ctx.Value(fieldsKey{}).(Fields)
	if ctxFields == nil {
		ctxFields = make(Fields)
	}
	merged := ctxFields.Append(fields...)
	return context.WithValue(ctx, fieldsKey{}, merged)
}

func (z *Logger) maskField(k string, v any) Field {
	if k == "password" {
		return Field{Key: k, Value: "******"}
	}
	return Field{Key: k, Value: v}
}

func (z *Logger) Sync() {
	// No-op for standard log package
}

func (z *Logger) withCtxFields(ctx context.Context, fields ...Field) []Field {
	ctxFields, _ := ctx.Value(fieldsKey{}).(Fields)
	if ctxFields == nil {
		ctxFields = make(Fields)
	}
	fs := ctxFields.Append(fields...)
	maskedFields := make([]Field, 0, len(fs))
	for k, f := range fs {
		maskedFields = append(maskedFields, z.maskField(k, f))
	}
	return maskedFields
}

func (z *Logger) log(l slog.Level, msg string, fields ...Field) {
	z.mu.Lock()
	defer z.mu.Unlock()

	if z.level == Debug || z.level == l {
		anyFields := make([]any, len(fields))
		for i, f := range fields {
			anyFields[i] = f
		}
		z.logger.Log(context.Background(), slog.Level(l), msg, anyFields...)
	}
}

func (z *Logger) InfoCtx(ctx context.Context, msg string, fields ...Field) {
	z.log(Info, msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Logger) DebugCtx(ctx context.Context, msg string, fields ...Field) {
	z.log(Debug, msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Logger) WarnCtx(ctx context.Context, msg string, fields ...Field) {
	z.log(Warn, msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Logger) ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	z.log(Error, msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Logger) SetLevel(l slog.Level) {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.level = l
}

func (z *Logger) Std() *log.Logger {
	return log.New(os.Stderr, "", log.LstdFlags)
}

// With добавляет поля в логгер и возвращает новый экземпляр логгера
func (z *Logger) With(fields ...Field) *Logger {
	z.mu.Lock()
	defer z.mu.Unlock()

	newFields := z.fields.Append(fields...)
	return &Logger{
		logger: z.logger,
		fields: newFields,
		level:  z.level,
	}
}
